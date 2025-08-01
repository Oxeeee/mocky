package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type MockResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type MockRoute struct {
	Method   string       `json:"method"`
	Path     string       `json:"path"`
	Response MockResponse `json:"response"`
}

var (
	mocks        = make(map[string]map[string]MockResponse) // path -> method -> response
	mu           sync.RWMutex
	enableTunnel = flag.Bool("tunnel", false, "Enable VK tunnel for external access")
	tunnelShort  = flag.Bool("t", false, "Enable VK tunnel for external access (short form)")
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	if methodMap, ok := mocks[r.URL.Path]; ok {
		if resp, ok := methodMap[r.Method]; ok {
			for k, v := range resp.Headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(resp.StatusCode)
			w.Write([]byte(resp.Body))
			return
		}
	}

	http.NotFound(w, r)
}

func webUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	htmlPath := filepath.Join("templates", "index.html")

	tmpl, err := template.ParseFiles(htmlPath)
	if err != nil {
		http.Error(w, "Template loading error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func listMocksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mocks)
}

func addMockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var route MockRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := mocks[route.Path]; !ok {
		mocks[route.Path] = make(map[string]MockResponse)
	}
	mocks[route.Path][route.Method] = route.Response

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Mock added"))
}

func deleteMockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE allowed", http.StatusMethodNotAllowed)
		return
	}

	var route MockRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if methodMap, ok := mocks[route.Path]; ok {
		delete(methodMap, route.Method)
		if len(methodMap) == 0 {
			delete(mocks, route.Path)
		}
		w.Write([]byte("Mock deleted"))
		return
	}

	http.NotFound(w, r)
}

func checkVKTunnelInstalled() bool {
	log.Println("Checking if VK tunnel is installed...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "vk-tunnel", "--version")
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("VK tunnel version check timed out (probably means it's installed but hangs)")
		return true 
	}

	if err != nil {
		return false
	}

	return true
}

func installVKTunnel() error {
	log.Println("Installing VK tunnel via npm...")
	cmd := exec.Command("npm", "install", "@vkontakte/vk-tunnel", "-g")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	log.Println("VK tunnel installed successfully")
	return nil
}

func startVKTunnel() {
	log.Println("Starting VK tunnel...")

	if !checkVKTunnelInstalled() {
		log.Println("VK tunnel not found, installing...")
		if err := installVKTunnel(); err != nil {
			log.Printf("Failed to install VK tunnel: %v", err)
			return
		}
	}

	cmd := exec.Command("vk-tunnel",
		"--insecure=1",
		"--http-protocol=http",
		"--ws-protocol=ws",
		"--host=localhost",
		"--port=8082",
		"--timeout=5000")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Starting VK tunnel process...")
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start VK tunnel: %v", err)
		return
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("VK tunnel process finished with error: %v", err)
		} else {
			log.Println("VK tunnel process finished successfully")
		}
	}()
}

func main() {
	flag.Parse()

	shouldStartTunnel := *enableTunnel || *tunnelShort

	http.HandleFunc("/__mock/ui", webUIHandler)
	http.HandleFunc("/__mock/list", listMocksHandler)
	http.HandleFunc("/__mock/add", addMockHandler)
	http.HandleFunc("/__mock/delete", deleteMockHandler)
	http.HandleFunc("/", mockHandler)

	log.Println("Dynamic mock server running on :8082")
	log.Println("Web UI available at: http://localhost:8082/__mock/ui")

	if shouldStartTunnel {
		log.Println("VK tunnel mode enabled - external access will be available shortly")
		startVKTunnel()
	}

	log.Println("Starting HTTP server...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}
