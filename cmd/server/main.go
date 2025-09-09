package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	baseDir    = "/its-certcenter"
	registerFN = "/its-certcenter/register.json"
)

type RegisterInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Subdomain  string `json:"subdomain"`
	FullDomain string `json:"fulldomain"`
}

func loadRegisterInfo() (*RegisterInfo, error) {
	data, err := os.ReadFile(registerFN)
	if err != nil {
		return nil, err
	}
	var reg RegisterInfo
	err = json.Unmarshal(data, &reg)
	return &reg, err
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	reg, err := loadRegisterInfo()
	if err != nil {
		http.Error(w, "register info not found", 500)
		return
	}
	json.NewEncoder(w).Encode(reg)
}

func issueCert(domain string) error {
	domainDir := filepath.Join(baseDir, domain)
	os.MkdirAll(domainDir, 0700)

	cmd := exec.Command(
		"/root/.acme.sh/acme.sh",
		"--issue",
		"--dns",
		"dns_acmedns",
		"-d", domain,
		"--key-file", filepath.Join(domainDir, "itsower.com.tw.key"),
		"--fullchain-file", filepath.Join(domainDir, "fullchain.cer"),
		"--ca-file", filepath.Join(domainDir, "ca.cer"),
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func handleIssue(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain required", 400)
		return
	}
	if err := issueCert(domain); err != nil {
		http.Error(w, fmt.Sprintf("issue failed: %v", err), 500)
		return
	}
	fmt.Fprintf(w, `{"status":"issued","domain":"%s"}`, domain)
}

func handleGetCert(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain required", 400)
		return
	}
	files := []string{"fullchain.cer", "itsower.com.tw.key", "ca.cer"}
	domainDir := filepath.Join(baseDir, domain)

	zipName := filepath.Join(os.TempDir(), fmt.Sprintf("%s-cert.zip", strings.ReplaceAll(domain, "*", "star")))
	zipCmd := exec.Command("zip", "-j", zipName,
		filepath.Join(domainDir, files[0]),
		filepath.Join(domainDir, files[1]),
		filepath.Join(domainDir, files[2]),
	)
	if err := zipCmd.Run(); err != nil {
		http.Error(w, "zip failed", 500)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=cert.zip")
	http.ServeFile(w, r, zipName)
}

func handleExpire(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain required", 400)
		return
	}
	certFile := filepath.Join(baseDir, domain, "fullchain.cer")
	cmd := exec.Command("openssl", "x509", "-enddate", "-noout", "-in", certFile)
	out, err := cmd.Output()
	if err != nil {
		http.Error(w, "cannot read cert", 500)
		return
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "=")
	if len(parts) != 2 {
		http.Error(w, "bad cert format", 500)
		return
	}
	t, _ := time.Parse("Jan 2 15:04:05 2006 MST", parts[1])
	resp := map[string]string{
		"domain":   domain,
		"expireAt": t.UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

func handleRenew(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	var cmd *exec.Cmd
	if domain == "" {
		cmd = exec.Command("/root/.acme.sh/acme.sh", "--renew-all", "--force")
	} else {
		cmd = exec.Command("/root/.acme.sh/acme.sh", "--renew", "-d", domain, "--force")
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		http.Error(w, "renew failed", 500)
		return
	}
	fmt.Fprintf(w, `{"status":"renewed","domain":"%s"}`, domain)
}

func main() {
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleIssue(w, r)
		} else {
			handleGetCert(w, r)
		}
	})
	http.HandleFunc("/expire", handleExpire)
	http.HandleFunc("/renew", handleRenew)

	fmt.Println("[its-certcenter] Server started at :9250")
	http.ListenAndServe(":9250", nil)
}
