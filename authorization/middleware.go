package authorization

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

var apiToken string

type agentConfig struct {
	AllowedIPs []string `json:"allowed_ips"`
}

var (
	allowAllIPs   bool
	allowedIPNets []*net.IPNet
	allowedIPsSet map[string]struct{}
)

func InitAuthWithPath(projectPath string) {
	// Load APP_KEY from the panel .env
	envPath := filepath.Join(projectPath, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	apiToken = os.Getenv("APP_KEY")
	if apiToken == "" {
		log.Fatal("APP_KEY not set in .env")
	}

	// Resolve agent config path
	configPath := os.Getenv("AGENT_CONFIG")
	if configPath == "" {
		configPath = "/raweb/apps/agent/config.json"
	}
	loadAllowedIPs(configPath)
}

func loadAllowedIPs(configPath string) {
	allowAllIPs = false
	allowedIPNets = nil
	allowedIPsSet = make(map[string]struct{})

	data, err := os.ReadFile(configPath)
	if err != nil {
		// If config isn't available, default to allow-all (key-only)
		allowAllIPs = true
		log.Printf("authorization: could not read config at %s: %v; defaulting to allow-all IPs", configPath, err)
		return
	}
	var cfg agentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		allowAllIPs = true
		log.Printf("authorization: could not parse config at %s: %v; defaulting to allow-all IPs", configPath, err)
		return
	}
	// Default to allow-all if not set
	if len(cfg.AllowedIPs) == 0 {
		allowAllIPs = true
		return
	}
	// Process entries
	for _, entry := range cfg.AllowedIPs {
		e := strings.TrimSpace(entry)
		if e == "" {
			continue
		}
		if e == "0.0.0.0" {
			allowAllIPs = true
			// No need to process others; keep flag set
			continue
		}
		// CIDR?
		if _, cidr, err := net.ParseCIDR(e); err == nil && cidr != nil {
			allowedIPNets = append(allowedIPNets, cidr)
			continue
		}
		// Exact IP?
		if ip := net.ParseIP(e); ip != nil {
			allowedIPsSet[ip.String()] = struct{}{}
			continue
		}
		log.Printf("authorization: ignoring invalid allowed_ips entry: %q", e)
	}
}

func clientIP(r *http.Request) string {
	// Prefer X-Forwarded-For (first IP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			return ip
		}
	}
	// Fallback to X-Real-IP
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	// RemoteAddr host part
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func ipAllowed(ipStr string) bool {
	if allowAllIPs {
		return true
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	if _, ok := allowedIPsSet[ip.String()]; ok {
		return true
	}
	for _, n := range allowedIPNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer "+apiToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		cip := clientIP(r)
		if !ipAllowed(cip) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
