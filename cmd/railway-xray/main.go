package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type Config struct {
	Log       LogConfig        `json:"log"`
	Inbounds  []Inbound        `json:"inbounds"`
	Outbounds []map[string]any `json:"outbounds"`
}

type LogConfig struct {
	Loglevel string `json:"loglevel"`
}

type Inbound struct {
	Listen   string        `json:"listen"`
	Port     int           `json:"port"`
	Protocol string        `json:"protocol"`
	Settings VLESSSettings `json:"settings"`
	Stream   StreamSetting `json:"streamSettings"`
	Sniffing Sniffing      `json:"sniffing"`
}

type VLESSSettings struct {
	Clients    []Client `json:"clients"`
	Decryption string   `json:"decryption"`
}

type Client struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

type StreamSetting struct {
	Network    string   `json:"network"`
	Security   string   `json:"security"`
	WSSettings WSConfig `json:"wsSettings"`
}

type WSConfig struct {
	Path string `json:"path"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

func env(name, fallback string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return fallback
	}
	return v
}

func envInt(name string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(name))
	var n int
	if v != "" {
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16])), nil
}

func buildConfig(port int, uuid, path string) Config {
	return Config{
		Log: LogConfig{Loglevel: env("XRAY_LOG_LEVEL", "warning")},
		Inbounds: []Inbound{{
			Listen:   "0.0.0.0",
			Port:     port,
			Protocol: "vless",
			Settings: VLESSSettings{
				Clients:    []Client{{ID: uuid, Email: "railway"}},
				Decryption: "none",
			},
			Stream: StreamSetting{
				Network:  "ws",
				Security: "none",
				WSSettings: WSConfig{
					Path: path,
				},
			},
			Sniffing: Sniffing{
				Enabled:      true,
				DestOverride: []string{"http", "tls", "quic"},
			},
		}},
		Outbounds: []map[string]any{
			{"protocol": "freedom", "tag": "direct"},
			{"protocol": "blackhole", "tag": "blocked"},
		},
	}
}

func makeLink(domain string, port int, uuid, path string) string {
	host := domain
	if strings.Contains(host, "://") {
		if u, err := url.Parse(host); err == nil {
			host = u.Hostname()
		}
	}
	q := url.Values{}
	q.Set("encryption", "none")
	q.Set("security", "tls")
	q.Set("sni", host)
	q.Set("type", "ws")
	q.Set("host", host)
	q.Set("path", path)
	q.Set("fp", "chrome")
	return "vless://" + uuid + "@" + host + ":" + fmt.Sprint(443) + "?" + q.Encode() + "#WhiteDNS-Railway"
}

func writeConfig(cfg Config) (string, error) {
	dir := "/tmp/whitedns-railway"
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func main() {
	port := envInt("PORT", 8080)
	uuid := strings.TrimSpace(os.Getenv("UUID"))
	if uuid == "" {
		var err error
		uuid, err = newUUID()
		if err != nil {
			panic(err)
		}
		fmt.Println("UUID was not set; generated a temporary UUID for this deployment.")
		fmt.Println("Set UUID as a Railway secret to keep it stable across redeploys.")
	}

	path := env("WS_PATH", "/whitedns")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.ContainsAny(path, " ?#") {
		panic("WS_PATH must be a simple URL path without spaces, ?, or #")
	}

	domain := env("DOMAIN", env("RAILWAY_PUBLIC_DOMAIN", ""))
	if domain == "" {
		fmt.Println("DOMAIN is not set yet. After Railway creates a public domain, redeploy or set DOMAIN to that hostname.")
	}

	cfg := buildConfig(port, uuid, path)
	configPath, err := writeConfig(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("WhiteDNS Railway edition")
	fmt.Printf("Listening: 0.0.0.0:%d\n", port)
	fmt.Printf("WebSocket path: %s\n", path)
	fmt.Printf("Config: %s\n", configPath)
	if domain != "" {
		fmt.Println("Client import:")
		fmt.Println(makeLink(domain, port, uuid, path))
	}
	fmt.Println("Note: TLS is terminated by Railway's public HTTPS edge; Xray intentionally listens for plain WS behind that edge.")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cmd := exec.CommandContext(ctx, "/usr/local/bin/xray", "run", "-config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return
		}
		panic(errors.Join(errors.New("xray exited"), err))
	}
}
