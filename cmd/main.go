package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
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
	mocks = make(map[string]map[string]MockResponse) // path -> method -> response
	mu    sync.RWMutex
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

func main() {
	http.HandleFunc("/__mock/ui", webUIHandler)
	http.HandleFunc("/__mock/list", listMocksHandler)
	http.HandleFunc("/__mock/add", addMockHandler)
	http.HandleFunc("/__mock/delete", deleteMockHandler)
	http.HandleFunc("/", mockHandler)

	log.Println("Dynamic mock server running on :8082")
	log.Println("Web UI available at: http://localhost:8082/__mock/ui")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
