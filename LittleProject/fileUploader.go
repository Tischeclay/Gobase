// file_uploader.go
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 文件信息结构
type FileInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	Type          string    `json:"type"`
	UploadTime    time.Time `json:"upload_time"`
	Path          string    `json:"-"`
	ShareCode     string    `json:"share_code,omitempty"`
	DownloadCount int       `json:"download_count"`
	MD5           string    `json:"md5,omitempty"`
	IsPublic      bool      `json:"is_public"`
}

// 分块上传信息
type ChunkInfo struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	TotalChunks int    `json:"total_chunks"`
	Uploaded    map[int]bool
	FilePath    string
	mu          sync.Mutex
}

// 上传任务信息
type UploadTask struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	TotalSize   int64     `json:"total_size"`
	UploadedSize int64    `json:"uploaded_size"`
	Progress    float64   `json:"progress"`
	Status      string    `json:"status"` // uploading, completed, failed
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

// 上传服务结构
type UploadService struct {
	uploadDir     string
	tempDir       string
	maxSize       int64
	files         map[string]FileInfo
	shares        map[string]string // share_code -> file_id
	chunks        map[string]*ChunkInfo
	tasks         map[string]*UploadTask
	mu            sync.RWMutex
	chunkSize     int64
	allowedTypes  map[string]bool
	maxFileCount  int
	totalUploads  int64
	totalDownloads int64
}

// 创建上传服务
func NewUploadService(uploadDir string, maxSize int64) (*UploadService, error) {
	// 创建上传目录
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	// 创建临时目录
	tempDir := filepath.Join(uploadDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, err
	}

	// 允许的文件类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png": true,
		"image/gif": true,
		"image/webp": true,
		"application/pdf": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain": true,
		"text/csv": true,
		"application/zip": true,
		"application/x-gzip": true,
		"application/x-tar": true,
		"video/mp4": true,
		"audio/mpeg": true,
	}

	return &UploadService{
		uploadDir:     uploadDir,
		tempDir:       tempDir,
		maxSize:       maxSize,
		files:         make(map[string]FileInfo),
		shares:        make(map[string]string),
		chunks:        make(map[string]*ChunkInfo),
		tasks:         make(map[string]*UploadTask),
		chunkSize:     1024 * 1024, // 1MB chunks
		allowedTypes:  allowedTypes,
		maxFileCount:  1000,
		totalUploads:  0,
		totalDownloads: 0,
	}, nil
}

// 生成ID
func (s *UploadService) generateID() string {
	data := fmt.Sprintf("%d", time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// 计算文件MD5
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// 检查文件类型
func (s *UploadService) isAllowedFileType(filename string, contentType string) bool {
	// 如果指定了content-type，检查是否允许
	if contentType != "" {
		if s.allowedTypes[contentType] {
			return true
		}
	}

	// 根据扩展名检查
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".txt": true, ".csv": true, ".zip": true, ".gz": true, ".tar": true,
		".mp4": true, ".mp3": true,
	}

	return allowedExts[ext]
}

// 获取文件大小
func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// 保存文件
func (s *UploadService) saveFile(file multipart.File, header *multipart.FileHeader) (FileInfo, error) {
	id := s.generateID()
	filename := filepath.Base(header.Filename)

	// 检查文件类型
	if !s.isAllowedFileType(filename, header.Header.Get("Content-Type")) {
		return FileInfo{}, fmt.Errorf("不支持的文件类型")
	}

	// 生成安全的文件名
	safeFilename := fmt.Sprintf("%s_%s", id, sanitizeFilename(filename))
	savePath := filepath.Join(s.uploadDir, safeFilename)

	// 创建文件
	dst, err := os.Create(savePath)
	if err != nil {
		return FileInfo{}, err
	}
	defer dst.Close()

	// 复制文件内容
	size, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(savePath)
		return FileInfo{}, err
	}

	// 计算MD5
	md5sum, err := calculateMD5(savePath)
	if err != nil {
		md5sum = ""
	}

	fileInfo := FileInfo{
		ID:            id,
		Name:          filename,
		Size:          size,
		Type:          header.Header.Get("Content-Type"),
		UploadTime:    time.Now(),
		Path:          savePath,
		DownloadCount: 0,
		MD5:           md5sum,
		IsPublic:      false,
	}

	s.mu.Lock()
	s.files[id] = fileInfo
	s.totalUploads++
	s.mu.Unlock()

	return fileInfo, nil
}

// 清理文件名
func sanitizeFilename(filename string) string {
	// 移除路径分隔符
	filename = filepath.Base(filename)
	// 替换危险字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}

// ==================== HTTP处理函数 ====================

// 上传文件
func (s *UploadService) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 检查文件数量
	s.mu.RLock()
	fileCount := len(s.files)
	s.mu.RUnlock()

	if fileCount >= s.maxFileCount {
		sendJSONError(w, "文件数量已达上限", http.StatusForbidden)
		return
	}

	// 限制请求大小
	r.Body = http.MaxBytesReader(w, r.Body, s.maxSize)

	// 解析multipart表单
	if err := r.ParseMultipartForm(s.maxSize); err != nil {
		sendJSONError(w, "文件太大或格式错误", http.StatusRequestEntityTooLarge)
		return
	}

	// 获取文件
	file, header, err := r.FormFile("file")
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > s.maxSize {
		sendJSONError(w, fmt.Sprintf("文件大小不能超过 %d MB", s.maxSize/1024/1024), http.StatusRequestEntityTooLarge)
		return
	}

	// 保存文件
	fileInfo, err := s.saveFile(file, header)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建上传任务记录
	task := &UploadTask{
		ID:          fileInfo.ID,
		FileName:    fileInfo.Name,
		TotalSize:   fileInfo.Size,
		UploadedSize: fileInfo.Size,
		Progress:    100,
		Status:      "completed",
		StartTime:   fileInfo.UploadTime,
		EndTime:     time.Now(),
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fileInfo)
}

// 分块上传初始化
func (s *UploadService) HandleChunkInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FileName    string `json:"file_name"`
		TotalSize   int64  `json:"total_size"`
		TotalChunks int    `json:"total_chunks"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 检查文件大小
	if req.TotalSize > s.maxSize {
		sendJSONError(w, fmt.Sprintf("文件太大，最大允许 %d MB", s.maxSize/1024/1024), http.StatusBadRequest)
		return
	}

	fileID := s.generateID()

	// 创建分块上传目录
	chunkDir := filepath.Join(s.tempDir, fileID)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chunkInfo := &ChunkInfo{
		FileID:      fileID,
		FileName:    req.FileName,
		TotalChunks: req.TotalChunks,
		Uploaded:    make(map[int]bool),
		FilePath:    chunkDir,
	}

	// 创建上传任务
	task := &UploadTask{
		ID:          fileID,
		FileName:    req.FileName,
		TotalSize:   req.TotalSize,
		UploadedSize: 0,
		Progress:    0,
		Status:      "uploading",
		StartTime:   time.Now(),
	}

	s.mu.Lock()
	s.chunks[fileID] = chunkInfo
	s.tasks[fileID] = task
	s.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{
		"file_id": fileID,
		"message": "分块上传初始化成功",
	})
}

// 上传分块
func (s *UploadService) HandleChunkUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	fileID := r.FormValue("file_id")
	chunkIndexStr := r.FormValue("chunk_index")
	totalChunksStr := r.FormValue("total_chunks")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		sendJSONError(w, "无效的分块索引", http.StatusBadRequest)
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		sendJSONError(w, "无效的总分块数", http.StatusBadRequest)
		return
	}

	// 获取分块信息
	s.mu.RLock()
	chunkInfo, exists := s.chunks[fileID]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "上传会话不存在", http.StatusNotFound)
		return
	}

	// 获取上传的文件块
	file, _, err := r.FormFile("chunk")
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存分块
	chunkPath := filepath.Join(chunkInfo.FilePath, fmt.Sprintf("chunk_%d", chunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 更新上传状态
	chunkInfo.mu.Lock()
	chunkInfo.Uploaded[chunkIndex] = true

	// 更新任务进度
	s.mu.Lock()
	if task, ok := s.tasks[fileID]; ok {
		task.UploadedSize += size
		task.Progress = float64(task.UploadedSize) / float64(task.TotalSize) * 100
	}
	s.mu.Unlock()

	chunkInfo.mu.Unlock()

	// 检查是否所有分块都已上传
	if len(chunkInfo.Uploaded) == totalChunks {
		go s.mergeChunks(fileID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"chunk":   chunkIndex,
		"message": fmt.Sprintf("分块 %d 上传成功", chunkIndex),
	})
}

// 合并分块
func (s *UploadService) mergeChunks(fileID string) {
	s.mu.RLock()
	chunkInfo, exists := s.chunks[fileID]
	s.mu.RUnlock()

	if !exists {
		log.Printf("合并失败: 找不到文件ID %s", fileID)
		return
	}

	// 生成最终文件名
	safeFilename := fmt.Sprintf("%s_%s", fileID, sanitizeFilename(chunkInfo.FileName))
	finalPath := filepath.Join(s.uploadDir, safeFilename)

	// 创建最终文件
	finalFile, err := os.Create(finalPath)
	if err != nil {
		log.Printf("创建最终文件失败: %v", err)
		return
	}
	defer finalFile.Close()

	// 按顺序合并分块
	var totalSize int64
	for i := 0; i < chunkInfo.TotalChunks; i++ {
		chunkPath := filepath.Join(chunkInfo.FilePath, fmt.Sprintf("chunk_%d", i))

		// 检查分块是否存在
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			log.Printf("分块 %d 不存在", i)
			continue
		}

		// 打开分块
		chunk, err := os.Open(chunkPath)
		if err != nil {
			log.Printf("打开分块 %d 失败: %v", i, err)
			continue
		}

		// 复制到最终文件
		size, err := io.Copy(finalFile, chunk)
		chunk.Close()

		if err != nil {
			log.Printf("合并分块 %d 失败: %v", i, err)
			continue
		}

		totalSize += size

		// 删除分块
		os.Remove(chunkPath)
	}

	// 删除临时目录
	os.RemoveAll(chunkInfo.FilePath)

	// 计算MD5
	md5sum, err := calculateMD5(finalPath)
	if err != nil {
		md5sum = ""
	}

	// 创建文件记录
	fileInfo := FileInfo{
		ID:            fileID,
		Name:          chunkInfo.FileName,
		Size:          totalSize,
		Type:          "application/octet-stream",
		UploadTime:    time.Now(),
		Path:          finalPath,
		DownloadCount: 0,
		MD5:           md5sum,
		IsPublic:      false,
	}

	s.mu.Lock()
	s.files[fileID] = fileInfo
	delete(s.chunks, fileID)

	if task, ok := s.tasks[fileID]; ok {
		task.Status = "completed"
		task.EndTime = time.Now()
		task.Progress = 100
	}

	s.totalUploads++
	s.mu.Unlock()

	log.Printf("文件合并完成: %s", chunkInfo.FileName)
}

// 获取上传进度
func (s *UploadService) GetUploadProgress(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("file_id")

	s.mu.RLock()
	task, exists := s.tasks[fileID]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "任务不存在", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(task)
}

// 下载文件
func (s *UploadService) HandleDownload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/download/")

	s.mu.RLock()
	fileInfo, exists := s.files[id]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(fileInfo.Path); os.IsNotExist(err) {
		sendJSONError(w, "文件已丢失", http.StatusNotFound)
		return
	}

	// 打开文件
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 更新下载计数
	s.mu.Lock()
	fileInfo.DownloadCount++
	s.files[id] = fileInfo
	s.totalDownloads++
	s.mu.Unlock()

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name))
	w.Header().Set("Content-Type", fileInfo.Type)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	w.Header().Set("Content-MD5", fileInfo.MD5)

	// 发送文件
	http.ServeContent(w, r, fileInfo.Name, fileInfo.UploadTime, file)
}

// 预览文件
func (s *UploadService) HandlePreview(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/preview/")

	s.mu.RLock()
	fileInfo, exists := s.files[id]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 只允许预览图片和PDF
	if !strings.HasPrefix(fileInfo.Type, "image/") && fileInfo.Type != "application/pdf" {
		sendJSONError(w, "该文件类型不支持预览", http.StatusBadRequest)
		return
	}

	// 打开文件
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 设置响应头
	w.Header().Set("Content-Type", fileInfo.Type)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))

	// 发送文件
	http.ServeContent(w, r, fileInfo.Name, fileInfo.UploadTime, file)
}

// 获取文件列表
func (s *UploadService) HandleList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	files := make([]FileInfo, 0, len(s.files))
	for _, file := range s.files {
		// 不返回文件路径
		file.Path = ""
		files = append(files, file)
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// 获取文件信息
func (s *UploadService) HandleInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	s.mu.RLock()
	fileInfo, exists := s.files[id]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 不返回文件路径
	fileInfo.Path = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfo)
}

// 删除文件
func (s *UploadService) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/delete/")

	s.mu.Lock()
	fileInfo, exists := s.files[id]
	if exists {
		// 删除物理文件
		os.Remove(fileInfo.Path)
		delete(s.files, id)

		// 删除相关的分享
		for code, fileID := range s.shares {
			if fileID == id {
				delete(s.shares, code)
			}
		}

		// 删除上传任务
		delete(s.tasks, id)
	}
	s.mu.Unlock()

	if !exists {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// 创建分享
func (s *UploadService) HandleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FileID   string `json:"file_id"`
		Expiry   int    `json:"expiry_hours"`
		IsPublic bool   `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	fileInfo, exists := s.files[req.FileID]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 生成分享码
	shareCode := s.generateID()

	// 更新文件信息
	fileInfo.ShareCode = shareCode
	fileInfo.IsPublic = req.IsPublic

	s.mu.Lock()
	s.files[req.FileID] = fileInfo
	s.shares[shareCode] = req.FileID
	s.mu.Unlock()

	shareURL := fmt.Sprintf("http://localhost:8080/share/%s", shareCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"share_code": shareCode,
		"share_url":  shareURL,
		"expiry":     req.Expiry,
		"is_public":  req.IsPublic,
	})
}

// 访问分享文件
func (s *UploadService) HandleSharedFile(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/share/")

	s.mu.RLock()
	fileID, exists := s.shares[code]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "分享链接无效或已过期", http.StatusNotFound)
		return
	}

	s.mu.RLock()
	fileInfo, exists := s.files[fileID]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 如果不公开，检查权限（简化版，实际应该有用户认证）
	if !fileInfo.IsPublic {
		// 这里应该检查用户权限
	}

	// 重定向到下载或预览
	if strings.HasPrefix(fileInfo.Type, "image/") || fileInfo.Type == "application/pdf" {
		http.Redirect(w, r, fmt.Sprintf("/preview/%s", fileID), http.StatusFound)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/download/%s", fileID), http.StatusFound)
	}
}

// 获取统计信息
func (s *UploadService) GetStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 计算总存储大小
	var totalSize int64
	for _, file := range s.files {
		totalSize += file.Size
	}

	// 计算活跃上传任务
	var activeUploads int
	for _, task := range s.tasks {
		if task.Status == "uploading" {
			activeUploads++
		}
	}

	stats := map[string]interface{}{
		"total_files":      len(s.files),
		"total_size":       totalSize,
		"total_size_mb":    float64(totalSize) / 1024 / 1024,
		"total_uploads":    s.totalUploads,
		"total_downloads":  s.totalDownloads,
		"active_uploads":   activeUploads,
		"active_shares":    len(s.shares),
		"max_file_count":   s.maxFileCount,
		"max_file_size_mb": s.maxSize / 1024 / 1024,
	}

	json.NewEncoder(w).Encode(stats)
}

// 发送JSON错误响应
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// ==================== 前端界面 ====================

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文件上传服务 - File Uploader</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Microsoft YaHei', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 15px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        
        .header p {
            opacity: 0.9;
            font-size: 1.1em;
        }
        
        .stats-bar {
            display: flex;
            justify-content: space-around;
            background: #f8f9fa;
            padding: 15px;
            border-bottom: 1px solid #e0e0e0;
        }
        
        .stat-item {
            text-align: center;
        }
        
        .stat-value {
            font-size: 1.5em;
            font-weight: bold;
            color: #667eea;
        }
        
        .stat-label {
            font-size: 0.9em;
            color: #666;
        }
        
        .content {
            padding: 30px;
        }
        
        .upload-area {
            border: 3px dashed #667eea;
            border-radius: 10px;
            padding: 40px;
            text-align: center;
            background: #f8f9fa;
            cursor: pointer;
            transition: all 0.3s;
            margin-bottom: 30px;
        }
        
        .upload-area:hover {
            background: #e9ecef;
            border-color: #5a67d8;
        }
        
        .upload-area.dragover {
            background: #e2e8f0;
            border-color: #48bb78;
        }
        
        .upload-icon {
            font-size: 48px;
            color: #667eea;
            margin-bottom: 10px;
        }
        
        .upload-text {
            font-size: 1.2em;
            color: #333;
            margin-bottom: 5px;
        }
        
        .upload-hint {
            color: #666;
            font-size: 0.9em;
        }
        
        .file-input {
            display: none;
        }
        
        .progress-area {
            margin-bottom: 30px;
        }
        
        .progress-item {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 10px;
        }
        
        .progress-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
        }
        
        .progress-bar {
            height: 10px;
            background: #e0e0e0;
            border-radius: 5px;
            overflow: hidden;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            width: 0%;
            transition: width 0.3s;
        }
        
        .file-list {
            margin-top: 20px;
        }
        
        .file-item {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 10px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .file-info {
            flex: 1;
        }
        
        .file-name {
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .file-meta {
            font-size: 0.9em;
            color: #666;
        }
        
        .file-actions {
            display: flex;
            gap: 10px;
        }
        
        .btn {
            padding: 8px 15px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.3s;
            text-decoration: none;
            display: inline-block;
        }
        
        .btn-primary {
            background: #667eea;
            color: white;
        }
        
        .btn-primary:hover {
            background: #5a67d8;
        }
        
        .btn-success {
            background: #48bb78;
            color: white;
        }
        
        .btn-success:hover {
            background: #38a169;
        }
        
        .btn-danger {
            background: #f56565;
            color: white;
        }
        
        .btn-danger:hover {
            background: #e53e3e;
        }
        
        .btn-sm {
            padding: 5px 10px;
            font-size: 12px;
        }
        
        .tab-container {
            margin-top: 30px;
        }
        
        .tab-buttons {
            display: flex;
            border-bottom: 2px solid #667eea;
        }
        
        .tab-btn {
            padding: 10px 20px;
            background: #f8f9fa;
            border: none;
            cursor: pointer;
            margin-right: 5px;
            border-radius: 5px 5px 0 0;
        }
        
        .tab-btn.active {
            background: #667eea;
            color: white;
        }
        
        .tab-content {
            display: none;
            padding: 20px;
            border: 1px solid #e0e0e0;
            border-top: none;
        }
        
        .tab-content.active {
            display: block;
        }
        
        .share-link {
            background: #f0f0f0;
            padding: 10px;
            border-radius: 5px;
            font-family: monospace;
            word-break: break-all;
        }
        
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            z-index: 1000;
        }
        
        .modal-content {
            background: white;
            max-width: 500px;
            margin: 100px auto;
            padding: 30px;
            border-radius: 10px;
            position: relative;
        }
        
        .close {
            position: absolute;
            top: 10px;
            right: 15px;
            font-size: 24px;
            cursor: pointer;
        }
        
        .chunk-upload-info {
            margin-top: 10px;
            padding: 10px;
            background: #f0f0f0;
            border-radius: 5px;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 文件上传服务</h1>
            <p>安全、快速、可靠的文件存储与分享</p>
        </div>
        
        <div class="stats-bar" id="statsBar">
            <div class="stat-item">
                <div class="stat-value" id="totalFiles">0</div>
                <div class="stat-label">总文件数</div>
            </div>
            <div class="stat-item">
                <div class="stat-value" id="totalSize">0 MB</div>
                <div class="stat-label">总存储</div>
            </div>
            <div class="stat-item">
                <div class="stat-value" id="activeUploads">0</div>
                <div class="stat-label">上传中</div>
            </div>
            <div class="stat-item">
                <div class="stat-value" id="totalDownloads">0</div>
                <div class="stat-label">下载次数</div>
            </div>
        </div>
        
        <div class="content">
            <div class="tab-container">
                <div class="tab-buttons">
                    <button class="tab-btn active" onclick="switchTab('upload')">上传文件</button>
                    <button class="tab-btn" onclick="switchTab('files')">文件列表</button>
                    <button class="tab-btn" onclick="switchTab('chunk')">分块上传</button>
                    <button class="tab-btn" onclick="switchTab('stats')">详细统计</button>
                </div>
                
                <!-- 上传标签页 -->
                <div id="uploadTab" class="tab-content active">
                    <div class="upload-area" id="dropArea" onclick="document.getElementById('fileInput').click()">
                        <div class="upload-icon">📤</div>
                        <div class="upload-text">点击或拖拽文件到此处上传</div>
                        <div class="upload-hint">支持所有常见文件格式，单个文件最大 100MB</div>
                    </div>
                    <input type="file" id="fileInput" class="file-input" multiple onchange="handleFileSelect(this.files)">
                    
                    <div class="progress-area" id="progressArea"></div>
                </div>
                
                <!-- 文件列表标签页 -->
                <div id="filesTab" class="tab-content">
                    <div class="file-list" id="fileList"></div>
                </div>
                
                <!-- 分块上传标签页 -->
                <div id="chunkTab" class="tab-content">
                    <div class="upload-area" id="chunkDropArea" onclick="document.getElementById('chunkFileInput').click()">
                        <div class="upload-icon">🔀</div>
                        <div class="upload-text">选择大文件进行分块上传</div>
                        <div class="upload-hint">支持超大文件，自动分块上传</div>
                    </div>
                    <input type="file" id="chunkFileInput" class="file-input" onchange="handleChunkFileSelect(this.files[0])">
                    
                    <div class="chunk-upload-info" id="chunkInfo"></div>
                    <div class="progress-area" id="chunkProgressArea"></div>
                </div>
                
                <!-- 统计标签页 -->
                <div id="statsTab" class="tab-content">
                    <div id="detailedStats"></div>
                </div>
            </div>
        </div>
    </div>
    
    <!-- 分享模态框 -->
    <div id="shareModal" class="modal">
        <div class="modal-content">
            <span class="close" onclick="closeModal()">&times;</span>
            <h3>分享链接</h3>
            <div class="share-link" id="shareLink"></div>
            <button class="btn btn-success" onclick="copyShareLink()" style="margin-top: 15px; width: 100%;">复制链接</button>
        </div>
    </div>

    <script>
        let uploadTasks = {};
        let chunkUploadTask = null;
        
        // 初始化
        window.onload = function() {
            loadFileList();
            loadStats();
        };
        
        // 标签页切换
        function switchTab(tabName) {
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            
            event.target.classList.add('active');
            document.getElementById(tabName + 'Tab').classList.add('active');
        }
        
        // 拖拽上传
        const dropArea = document.getElementById('dropArea');
        
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            dropArea.addEventListener(eventName, preventDefaults, false);
        });
        
        function preventDefaults(e) {
            e.preventDefault();
            e.stopPropagation();
        }
        
        ['dragenter', 'dragover'].forEach(eventName => {
            dropArea.addEventListener(eventName, highlight, false);
        });
        
        ['dragleave', 'drop'].forEach(eventName => {
            dropArea.addEventListener(eventName, unhighlight, false);
        });
        
        function highlight() {
            dropArea.classList.add('dragover');
        }
        
        function unhighlight() {
            dropArea.classList.remove('dragover');
        }
        
        dropArea.addEventListener('drop', handleDrop, false);
        
        function handleDrop(e) {
            const dt = e.dataTransfer;
            const files = dt.files;
            handleFileSelect(files);
        }
        
        // 处理文件选择
        function handleFileSelect(files) {
            for (let file of files) {
                uploadFile(file);
            }
        }
        
        // 上传文件
        function uploadFile(file) {
            const taskId = Date.now() + '-' + file.name;
            
            uploadTasks[taskId] = {
                file: file,
                progress: 0,
                status: 'uploading'
            };
            
            updateProgressDisplay();
            
            const formData = new FormData();
            formData.append('file', file);
            
            const xhr = new XMLHttpRequest();
            
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const progress = (e.loaded / e.total) * 100;
                    uploadTasks[taskId].progress = progress;
                    updateProgressDisplay();
                }
            });
            
            xhr.addEventListener('load', () => {
                if (xhr.status === 201) {
                    uploadTasks[taskId].status = 'completed';
                    const response = JSON.parse(xhr.responseText);
                    showNotification('文件上传成功: ' + file.name, 'success');
                    loadFileList();
                    loadStats();
                } else {
                    uploadTasks[taskId].status = 'failed';
                    showNotification('上传失败: ' + file.name, 'error');
                }
                updateProgressDisplay();
            });
            
            xhr.addEventListener('error', () => {
                uploadTasks[taskId].status = 'failed';
                showNotification('上传出错: ' + file.name, 'error');
                updateProgressDisplay();
            });
            
            xhr.open('POST', '/api/upload', true);
            xhr.send(formData);
        }
        
        // 更新进度显示
        function updateProgressDisplay() {
            const progressArea = document.getElementById('progressArea');
            let html = '';
            
            for (let [id, task] of Object.entries(uploadTasks)) {
                if (task.status === 'uploading') {
                    html += `
	<div class="progress-item">
	<div class="progress-header">
	<span>${task.file.name}</span>
	<span>${task.progress.toFixed(1)}%</span>
	</div>
	<div class="progress-bar">
	<div class="progress-fill" style="width: ${task.progress}%"></div>
	</div>
	</div>
		`;
                }
            }
            
            progressArea.innerHTML = html;
        }
        
        // 处理分块文件选择
        function handleChunkFileSelect(file) {
            if (!file) return;
            
            const totalChunks = Math.ceil(file.size / (1024 * 1024)); // 1MB chunks
            
            document.getElementById('chunkInfo').innerHTML = `
	<strong>文件名:</strong> ${file.name}<br>
	<strong>大小:</strong> ${(file.size / 1024 / 1024).toFixed(2)} MB<br>
	<strong>分块数:</strong> ${totalChunks}<br>
	<strong>状态:</strong> 准备上传...
	`;
            
            // 初始化分块上传
            fetch('/api/chunk/init', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    file_name: file.name,
                    total_size: file.size,
                    total_chunks: totalChunks
                })
            })
            .then(response => response.json())
            .then(data => {
                if (data.file_id) {
                    chunkUploadTask = {
                        fileId: data.file_id,
                        fileName: file.name,
                        totalChunks: totalChunks,
                        uploadedChunks: 0
                    };
                    uploadChunks(file, data.file_id, totalChunks);
                }
            });
        }
        
        // 上传分块
        async function uploadChunks(file, fileId, totalChunks) {
            const chunkSize = 1024 * 1024; // 1MB
            
            for (let i = 0; i < totalChunks; i++) {
                const start = i * chunkSize;
                const end = Math.min(start + chunkSize, file.size);
                const chunk = file.slice(start, end);
                
                const formData = new FormData();
                formData.append('chunk', chunk);
                formData.append('file_id', fileId);
                formData.append('chunk_index', i);
                formData.append('total_chunks', totalChunks);
                
                try {
                    const response = await fetch('/api/chunk/upload', {
                        method: 'POST',
                        body: formData
                    });
                    
                    if (response.ok) {
                        chunkUploadTask.uploadedChunks = i + 1;
                        const progress = ((i + 1) / totalChunks * 100).toFixed(1);
                        
                        document.getElementById('chunkInfo').innerHTML = `
	<strong>文件名:</strong> ${file.name}<br>
	<strong>大小:</strong> ${(file.size / 1024 / 1024).toFixed(2)} MB<br>
	<strong>分块数:</strong> ${totalChunks}<br>
	<strong>已上传:</strong> ${i + 1}/${totalChunks} (${progress}%)<br>
	<div class="progress-bar" style="margin-top: 10px;">
	<div class="progress-fill" style="width: ${progress}%"></div>
	</div>
		`;
                    }
                } catch (error) {
                    console.error('Chunk upload failed:', error);
                    showNotification('分块上传失败', 'error');
                    break;
                }
            }
            
            showNotification('文件上传完成: ' + file.name, 'success');
            loadFileList();
            loadStats();
        }
        
        // 加载文件列表
        function loadFileList() {
            fetch('/api/files')
                .then(response => response.json())
                .then(files => {
                    const fileList = document.getElementById('fileList');
                    let html = '';
                    
                    files.forEach(file => {
                        const size = (file.size / 1024 / 1024).toFixed(2);
                        const date = new Date(file.upload_time).toLocaleString();
                        
                        html += `
	<div class="file-item">
	<div class="file-info">
	<div class="file-name">${file.name}</div>
	<div class="file-meta">
	${size} MB | ${file.type} | 上传于 ${date} | 下载 ${file.download_count} 次
${file.md5 ? ' | MD5: ' + file.md5.substr(0, 8) : ''}
</div>
</div>
<div class="file-actions">
<button class="btn btn-sm btn-primary" onclick="downloadFile('${file.id}')">下载</button>
<button class="btn btn-sm btn-success" onclick="shareFile('${file.id}')">分享</button>
<button class="btn btn-sm btn-danger" onclick="deleteFile('${file.id}')">删除</button>
</div>
</div>
`;
                    });

                    if (files.length === 0) {
                        html = '<p style="text-align: center; color: #666;">暂无文件</p>';
                    }

                    fileList.innerHTML = html;
                });
        }

        // 加载统计信息
        function loadStats() {
            fetch('/api/stats')
                .then(response => response.json())
                .then(stats => {
                    document.getElementById('totalFiles').textContent = stats.total_files;
                    document.getElementById('totalSize').textContent = stats.total_size_mb.toFixed(2) + ' MB';
                    document.getElementById('activeUploads').textContent = stats.active_uploads;
                    document.getElementById('totalDownloads').textContent = stats.total_downloads;

                    // 详细统计
                    document.getElementById('detailedStats').innerHTML = `
<div class="stats-grid" style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 15px;">
<div class="stat-card" style="background: #f8f9fa; padding: 15px; border-radius: 8px;">
<h4>基本统计</h4>
<p>总文件数: ${stats.total_files}</p>
<p>总存储: ${stats.total_size_mb.toFixed(2)} MB</p>
<p>总上传次数: ${stats.total_uploads}</p>
<p>总下载次数: ${stats.total_downloads}</p>
</div>
<div class="stat-card" style="background: #f8f9fa; padding: 15px; border-radius: 8px;">
<h4>实时状态</h4>
<p>活跃上传: ${stats.active_uploads}</p>
<p>活跃分享: ${stats.active_shares}</p>
<p>最大文件数: ${stats.max_file_count}</p>
<p>最大文件大小: ${stats.max_file_size_mb} MB</p>
</div>
</div>
`;
                });
        }

        // 下载文件
        function downloadFile(fileId) {
            window.location.href = '/download/' + fileId;
        }

        // 分享文件
        function shareFile(fileId) {
            fetch('/api/share', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    file_id: fileId,
                    expiry_hours: 24,
                    is_public: true
                })
            })
            .then(response => response.json())
            .then(data => {
                document.getElementById('shareLink').textContent = data.share_url;
                document.getElementById('shareModal').style.display = 'block';
            });
        }

        // 删除文件
        function deleteFile(fileId) {
            if (confirm('确定要删除这个文件吗？')) {
                fetch('/delete/' + fileId, {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.status === 204) {
                        showNotification('文件删除成功', 'success');
                        loadFileList();
                        loadStats();
                    }
                });
            }
        }

        // 复制分享链接
        function copyShareLink() {
            const link = document.getElementById('shareLink').textContent;
            navigator.clipboard.writeText(link).then(() => {
                showNotification('链接已复制到剪贴板', 'success');
                closeModal();
            });
        }

        // 关闭模态框
        function closeModal() {
            document.getElementById('shareModal').style.display = 'none';
        }

        // 显示通知
        function showNotification(message, type) {
            const notification = document.createElement('div');
            notification.style.cssText = `
position: fixed;
top: 20px;
right: 20px;
padding: 15px 20px;
background: ${type === 'success' ? '#48bb78' : '#f56565'};
color: white;
border-radius: 5px;
box-shadow: 0 3px 10px rgba(0,0,0,0.2);
z-index: 2000;
animation: slideIn 0.3s ease;
`;
            notification.textContent = message;

            document.body.appendChild(notification);

            setTimeout(() => {
                notification.style.animation = 'slideOut 0.3s ease';
                setTimeout(() => {
                    document.body.removeChild(notification);
                }, 300);
            }, 3000);
        }

        // 添加动画样式
        const style = document.createElement('style');
        style.textContent = `
@keyframes slideIn {
from { transform: translateX(100%); opacity: 0; }
to { transform: translateX(0); opacity: 1; }
}
@keyframes slideOut {
from { transform: translateX(0); opacity: 1; }
to { transform: translateX(100%); opacity: 0; }
}
`;
        document.head.appendChild(style);
    </script>
</body>
</html>`

w.Header().Set("Content-Type", "text/html; charset=utf-8")
fmt.Fprint(w, html)
}

func main() {
	// 创建上传服务
	uploadService, err := NewUploadService("./uploads", 100*1024*1024) // 100MB max
	if err != nil {
		log.Fatal("创建上传服务失败:", err)
	}

	// API路由
	http.HandleFunc("/api/upload", uploadService.HandleUpload)
	http.HandleFunc("/api/chunk/init", uploadService.HandleChunkInit)
	http.HandleFunc("/api/chunk/upload", uploadService.HandleChunkUpload)
	http.HandleFunc("/api/progress", uploadService.GetUploadProgress)
	http.HandleFunc("/api/files", uploadService.HandleList)
	http.HandleFunc("/api/info", uploadService.HandleInfo)
	http.HandleFunc("/api/share", uploadService.HandleShare)
	http.HandleFunc("/api/stats", uploadService.GetStats)

	// 文件操作路由
	http.HandleFunc("/download/", uploadService.HandleDownload)
	http.HandleFunc("/preview/", uploadService.HandlePreview)
	http.HandleFunc("/delete/", uploadService.HandleDelete)
	http.HandleFunc("/share/", uploadService.HandleSharedFile)

	// 首页
	http.HandleFunc("/", serveHTML)

	log.Println("文件上传服务启动在 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}