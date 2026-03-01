// file_uploader.go
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	Type          string    `json:"type"`
	UploadTime    time.Time `json:"upload_time"`
	Path          string    `json:"-"`
	ShareCode     string    `json:"share_code,omitempty"`
	DownloadCount int       `json:"download_count"`
}

type UploadService struct {
	uploadDir string
	maxSize   int64
	files     map[string]FileInfo
	shares    map[string]string // share_code -> file_id
	mu        sync.RWMutex
	chunkSize int64
}

func NewUploadService(uploadDir string, maxSize int64) (*UploadService, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	return &UploadService{
		uploadDir: uploadDir,
		maxSize:   maxSize,
		files:     make(map[string]FileInfo),
		shares:    make(map[string]string),
		chunkSize: 1024 * 1024, // 1MB chunks
	}, nil
}

func (s *UploadService) generateID() string {
	data := fmt.Sprintf("%d", time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func (s *UploadService) saveFile(file multipart.File, header *multipart.FileHeader) (FileInfo, error) {
	id := s.generateID()
	filename := filepath.Base(header.Filename)
	savePath := filepath.Join(s.uploadDir, id+"_"+filename)

	dst, err := os.Create(savePath)
	if err != nil {
		return FileInfo{}, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(savePath)
		return FileInfo{}, err
	}

	fileInfo := FileInfo{
		ID:            id,
		Name:          filename,
		Size:          size,
		Type:          header.Header.Get("Content-Type"),
		UploadTime:    time.Now(),
		Path:          savePath,
		DownloadCount: 0,
	}

	s.mu.Lock()
	s.files[id] = fileInfo
	s.mu.Unlock()

	return fileInfo, nil
}

func (s *UploadService) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查文件大小
	r.Body = http.MaxBytesReader(w, r.Body, s.maxSize)
	if err := r.ParseMultipartForm(s.maxSize); err != nil {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileInfo, err := s.saveFile(file, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfo)
}

func (s *UploadService) HandleChunkUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chunkIndex := r.FormValue("chunk")
	totalChunks := r.FormValue("total")
	filename := r.FormValue("filename")

	file, _, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建临时目录存储块
	tempDir := filepath.Join(s.uploadDir, "temp", filename)
	os.MkdirAll(tempDir, 0755)

	chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%s", chunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 如果是最后一块，合并文件
	if chunkIndex == totalChunks {
		go s.mergeChunks(filename, tempDir, totalChunks)
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "chunk uploaded",
		"chunk":  chunkIndex,
	})
}

func (s *UploadService) mergeChunks(filename, tempDir, totalChunks string) {
	// 这里实现块合并逻辑
	// 可以异步处理，返回任务ID
}

func (s *UploadService) HandleDownload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/download/")

	s.mu.RLock()
	fileInfo, exists := s.files[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(fileInfo.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 更新下载计数
	s.mu.Lock()
	fileInfo.DownloadCount++
	s.files[id] = fileInfo
	s.mu.Unlock()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name))
	w.Header().Set("Content-Type", fileInfo.Type)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))

	http.ServeContent(w, r, fileInfo.Name, fileInfo.UploadTime, file)
}

func (s *UploadService) HandleList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	files := make([]FileInfo, 0, len(s.files))
	for _, file := range s.files {
		files = append(files, file)
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (s *UploadService) HandleInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	s.mu.RLock()
	fileInfo, exists := s.files[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfo)
}

func (s *UploadService) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/delete/")

	s.mu.Lock()
	fileInfo, exists := s.files[id]
	if exists {
		os.Remove(fileInfo.Path)
		delete(s.files, id)

		// 删除相关的分享
		for code, fileID := range s.shares {
			if fileID == id {
				delete(s.shares, code)
			}
		}
	}
	s.mu.Unlock()

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *UploadService) HandleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FileID string `json:"file_id"`
		Expiry int    `json:"expiry_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	_, exists := s.files[req.FileID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	shareCode := s.generateID()

	s.mu.Lock()
	s.shares[shareCode] = req.FileID
	s.mu.Unlock()

	shareURL := fmt.Sprintf("http://localhost:8080/share/%s", shareCode)

	json.NewEncoder(w).Encode(map[string]string{
		"share_code": shareCode,
		"share_url":  shareURL,
	})
}
