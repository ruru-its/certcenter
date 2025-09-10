package main

import (
	"encoding/json"
	"fmt"
	"log"
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

// 讀取註冊資訊自註冊檔案
func handleRegisterReadFile(w http.ResponseWriter, r *http.Request) {
	reg, err := loadRegisterInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("register info not found error: %v", err), 500)
		return
	}
	// 組合回傳 JSON，額外加上 tips
	resp := map[string]interface{}{
		"username":   reg.Username,
		"password":   reg.Password,
		"subdomain":  reg.Subdomain,
		"fulldomain": reg.FullDomain,
		"tips":       fmt.Sprintf("Azure DNS zones Recordsets add _acme-challenge type CNAME Alias %s", reg.FullDomain),
	}

	json.NewEncoder(w).Encode(resp)
}

// 讀取註冊資訊自環境變數
func handleRegisterReadEnv(w http.ResponseWriter, r *http.Request) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	subdomain := os.Getenv("SUBDOMAIN")
	fulldomain := os.Getenv("FQDN")

	if username == "" || password == "" || subdomain == "" || fulldomain == "" {
		http.Error(w, "required environment variables not found", 500)
		return
	}
	// 組合回傳 JSON，額外加上 tips
	resp := map[string]interface{}{
		"username":   username,
		"password":   password,
		"subdomain":  subdomain,
		"fulldomain": fulldomain,
		"tips":       fmt.Sprintf("Azure DNS zones Recordsets: add a CNAME record '_acme-challenge.<your-domain>' → %s", fulldomain),
	}
	json.NewEncoder(w).Encode(resp)
}

func issueCert(domain string) error {
	domainDir := filepath.Join(baseDir, domain)
	os.MkdirAll(domainDir, 0700)

	// 如果已經有 fullchain.cer，就不要重複 issue
	certFile := filepath.Join(domainDir, "fullchain.cer")
	if _, err := os.Stat(certFile); err == nil {
		return fmt.Errorf("certificate already exists for %s", domain)
	}

	reg, err := loadRegisterInfo()
	if err != nil {
		return fmt.Errorf("no register info: %v", err)
	}

	cmd := exec.Command(
		"/root/.acme.sh/acme.sh",
		"--issue",
		"--dns",
		"dns_acmedns",
		"-d", domain,
		"--key-file", filepath.Join(domainDir, "its-certcenter.key"),
		"--fullchain-file", filepath.Join(domainDir, "fullchain.cer"),
		"--cert-file", filepath.Join(domainDir, "its-certcenter.cer"),
		"--ca-file", filepath.Join(domainDir, "ca.cer"),
	)

	cmd.Env = append(os.Environ(),
		"ACMEDNS_BASE_URL=https://auth.acme-dns.io",
		"ACMEDNS_USERNAME="+reg.Username,
		"ACMEDNS_PASSWORD="+reg.Password,
		"ACMEDNS_SUBDOMAIN="+reg.Subdomain,
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

	files := []string{"fullchain.cer", "its-certcenter.key", "ca.cer"}
	domainDir := filepath.Join(baseDir, domain)

	zipName := filepath.Join(os.TempDir(), "live.zip")
	args := []string{"-j", zipName}

	validFiles := 0
	for _, f := range files {
		path := filepath.Join(domainDir, f)
		if _, err := os.Stat(path); err == nil {
			log.Printf("[handleGetCert] found file for domain=%s: %s", domain, f)
			args = append(args, path)
			validFiles++
		} else {
			log.Printf("[handleGetCert] missing file for domain=%s: %s (err=%v)",
				domain, f, err)
		}
	}

	if validFiles == 0 {
		http.Error(w, fmt.Sprintf("no cert files found for domain %s", domain), 404)
		return
	}

	zipCmd := exec.Command("zip", args...)
	if err := zipCmd.Run(); err != nil {
		http.Error(w, "zip failed", 500)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=live.zip")
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
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

	// 解析憑證到期日
	t, _ := time.Parse("Jan 2 15:04:05 2006 MST", parts[1])
	now := time.Now().UTC()
	daysRemaining := int(t.Sub(now).Hours() / 24)

	status := "OK"
	if daysRemaining <= 0 {
		status = "ERROR"
		log.Printf("[handleHealth] ERROR: domain=%s expired at %s", domain, t.UTC().Format(time.RFC3339))
	} else if daysRemaining <= 30 {
		status = "WARN"
		log.Printf("[handleHealth] WARN: domain=%s will expire in %d days (%s)", domain, daysRemaining, t.UTC().Format(time.RFC3339))
	}

	resp := map[string]interface{}{
		"domain":        domain,
		"expireAt":      t.UTC().Format(time.RFC3339),
		"status":        status,
		"daysRemaining": daysRemaining,
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
	//http.HandleFunc("/register", handleRegisterReadFile) //來自註冊檔案
	http.HandleFunc("/register", handleRegisterReadEnv) //來自環境變數
	http.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleIssue(w, r)
		} else {
			handleGetCert(w, r)
		}
	})
	http.HandleFunc("/expire", handleExpire)
	http.HandleFunc("/renew", handleRenew)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("[its-certcenter] Server started at :9250")
	http.ListenAndServe(":9250", nil)
}
