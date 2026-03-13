// main.go
package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type URLStore struct {
	mu     sync.RWMutex
	urls   map[string]string
	stats  map[string]int
	expiry map[string]time.Time
}

type URLData struct {
	OriginalURL string `json:"original_url"`
	CustomAlias string `json:"custom_alias,omitempty"`
	ExpiryDays  int    `json:"expiry_days,omitempty"`
}

func NewURLStore() *URLStore {
	return &URLStore{
		urls:   make(map[string]string),
		stats:  make(map[string]int),
		expiry: make(map[string]time.Time),
	}
}

func (s *URLStore) generateShortCode(originalURL string) string {
	hash := md5.Sum([]byte(originalURL + time.Now().String()))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}

func (s *URLStore) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data URLData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	shortCode := data.CustomAlias
	if shortCode == "" {
		shortCode = s.generateShortCode(data.OriginalURL)
	}

	if _, exists := s.urls[shortCode]; exists {
		http.Error(w, "Short code already exists", http.StatusConflict)
		return
	}

	s.urls[shortCode] = data.OriginalURL
	s.stats[shortCode] = 0

	if data.ExpiryDays > 0 {
		s.expiry[shortCode] = time.Now().AddDate(0, 0, data.ExpiryDays)
	}

	response := map[string]string{
		"short_url":    fmt.Sprintf("http://localhost:8080/%s", shortCode),
		"original_url": data.OriginalURL,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *URLStore) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]

	s.mu.RLock()
	originalURL, exists := s.urls[shortCode]
	expiry, hasExpiry := s.expiry[shortCode]
	s.mu.RUnlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	if hasExpiry && time.Now().After(expiry) {
		s.mu.Lock()
		delete(s.urls, shortCode)
		delete(s.stats, shortCode)
		delete(s.expiry, shortCode)
		s.mu.Unlock()
		http.Error(w, "URL has expired", http.StatusGone)
		return
	}

	s.mu.Lock()
	s.stats[shortCode]++
	s.mu.Unlock()

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (s *URLStore) GetStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Query().Get("code")

	s.mu.RLock()
	visits := s.stats[shortCode]
	originalURL := s.urls[shortCode]
	expiry := s.expiry[shortCode]
	s.mu.RUnlock()

	response := map[string]interface{}{
		"short_code":   shortCode,
		"original_url": originalURL,
		"visits":       visits,
		"expiry":       expiry,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *URLStore) DeleteURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Path[len("/delete/"):]

	s.mu.Lock()
	delete(s.urls, shortCode)
	delete(s.stats, shortCode)
	delete(s.expiry, shortCode)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *URLStore) ListURLs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type URLInfo struct {
		ShortCode   string    `json:"short_code"`
		OriginalURL string    `json:"original_url"`
		Visits      int       `json:"visits"`
		Expiry      time.Time `json:"expiry"`
	}

	var urls []URLInfo
	for code, original := range s.urls {
		urls = append(urls, URLInfo{
			ShortCode:   code,
			OriginalURL: original,
			Visits:      s.stats[code],
			Expiry:      s.expiry[code],
		})
	}

	json.NewEncoder(w).Encode(urls)
}

func main() {
	store := NewURLStore()

	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API路由
	http.HandleFunc("/api/shorten", store.CreateShortURL)
	http.HandleFunc("/api/stats", store.GetStats)
	http.HandleFunc("/api/list", store.ListURLs)
	http.HandleFunc("/delete/", store.DeleteURL)

	// 重定向路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		store.Redirect(w, r)
	})

	log.Println("Server starting on :8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
}
