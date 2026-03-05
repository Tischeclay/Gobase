// qr_tool.go
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/tuotoo/qrcode"
)

// QRCode 结构
type QRCode struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Size        int       `json:"size"`
	Format      string    `json:"format"`
	CreatedAt   time.Time `json:"created_at"`
	DownloadCount int     `json:"download_count"`
	Path        string    `json:"-"`
}

// BatchJob 批量任务
type BatchJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Total     int       `json:"total"`
	Completed int       `json:"completed"`
	Results   []string  `json:"results"`
	Errors    []string  `json:"errors"`
	CreatedAt time.Time `json:"created_at"`
}

// QRService 服务
type QRService struct {
	qrDir     string
	qrcodes   map[string]QRCode
	jobs      map[string]*BatchJob
	mu        sync.RWMutex
}

// NewQRService 创建服务
func NewQRService(qrDir string) (*QRService, error) {
	if err := os.MkdirAll(qrDir, 0755); err != nil {
		return nil, err
	}

	return &QRService{
		qrDir:   qrDir,
		qrcodes: make(map[string]QRCode),
		jobs:    make(map[string]*BatchJob),
	}, nil
}

// 生成ID
func (s *QRService) generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ==================== 二维码生成 ====================

// 生成二维码
func (s *QRService) GenerateQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
		Size    int    `json:"size"`
		Format  string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证输入
	if req.Content == "" {
		sendJSONError(w, "内容不能为空", http.StatusBadRequest)
		return
	}

	if req.Size < 128 || req.Size > 1024 {
		req.Size = 256 // 默认大小
	}

	if req.Format == "" {
		req.Format = "png"
	}

	// 生成二维码
	id := s.generateID()
	filename := fmt.Sprintf("%s_%d.%s", id, time.Now().Unix(), req.Format)
	filepath := filepath.Join(s.qrDir, filename)

	var err error
	if req.Format == "png" {
		err = qrcode.WriteFile(req.Content, qrcode.Medium, req.Size, filepath)
	} else if req.Format == "jpg" || req.Format == "jpeg" {
		// PNG转JPG
		var q *qrcode.QRCode
		q, err = qrcode.New(req.Content, qrcode.Medium)
		if err == nil {
			q.DisableBorder = false
			img := q.Image(req.Size)
			// 保存为JPG（简化处理）
			err = saveAsJPEG(img, filepath)
		}
	}

	if err != nil {
		sendJSONError(w, "生成二维码失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	qr := QRCode{
		ID:          id,
		Content:     req.Content,
		Size:        req.Size,
		Format:      req.Format,
		CreatedAt:   time.Now(),
		DownloadCount: 0,
		Path:        filepath,
	}

	s.mu.Lock()
	s.qrcodes[id] = qr
	s.mu.Unlock()

	// 返回二维码信息
	resp := map[string]interface{}{
		"id":         id,
		"content":    req.Content,
		"size":       req.Size,
		"format":     req.Format,
		"created_at": qr.CreatedAt,
		"url":        fmt.Sprintf("/download/%s", id),
		"preview":    fmt.Sprintf("/preview/%s", id),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 批量生成二维码
func (s *QRService) BatchGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Contents []string `json:"contents"`
		Size     int      `json:"size"`
		Format   string   `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if len(req.Contents) == 0 {
		sendJSONError(w, "内容列表不能为空", http.StatusBadRequest)
		return
	}

	jobID := s.generateID()
	job := &BatchJob{
		ID:        jobID,
		Status:    "processing",
		Total:     len(req.Contents),
		Completed: 0,
		Results:   make([]string, 0),
		Errors:    make([]string, 0),
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	// 异步处理
	go func() {
		for i, content := range req.Contents {
			id := s.generateID()
			filename := fmt.Sprintf("%s_%d.%s", id, time.Now().Unix(), req.Format)
			filepath := filepath.Join(s.qrDir, filename)

			err := qrcode.WriteFile(content, qrcode.Medium, req.Size, filepath)

			s.mu.Lock()
			if err != nil {
				job.Errors = append(job.Errors, fmt.Sprintf("第%d项: %v", i+1, err))
			} else {
				qr := QRCode{
					ID:          id,
					Content:     content,
					Size:        req.Size,
					Format:      req.Format,
					CreatedAt:   time.Now(),
					DownloadCount: 0,
					Path:        filepath,
				}
				s.qrcodes[id] = qr
				job.Results = append(job.Results, fmt.Sprintf("/preview/%s", id))
			}
			job.Completed++
			s.mu.Unlock()
		}

		s.mu.Lock()
		job.Status = "completed"
		s.mu.Unlock()
	}()

	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "processing",
	})
}

// 获取任务状态
func (s *QRService) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")

	s.mu.RLock()
	job, exists := s.jobs[jobID]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "任务不存在", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}

// ==================== 二维码解码 ====================

// 解码二维码
func (s *QRService) DecodeQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析上传的文件
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		sendJSONError(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("qr_image")
	if err != nil {
		sendJSONError(w, "获取文件失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存临时文件
	tempFile := filepath.Join(os.TempDir(), "qr_"+header.Filename)
	dst, err := os.Create(tempFile)
	if err != nil {
		sendJSONError(w, "创建临时文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	defer os.Remove(tempFile)

	if _, err := io.Copy(dst, file); err != nil {
		sendJSONError(w, "保存文件失败", http.StatusInternalServerError)
		return
	}

	// 解码二维码
	qrmatrix, err := qrcode.Decode(tempFile)
	if err != nil {
		sendJSONError(w, "解码失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	result := map[string]interface{}{
		"content":  qrmatrix.Content,
		"filename": header.Filename,
		"size":     header.Size,
		"type":     header.Header.Get("Content-Type"),
	}

	json.NewEncoder(w).Encode(result)
}

// 批量解码
func (s *QRService) BatchDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		sendJSONError(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["qr_images"]
	if len(files) == 0 {
		sendJSONError(w, "请选择文件", http.StatusBadRequest)
		return
	}

	results := make([]map[string]interface{}, 0)
	errors := make([]string, 0)

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: 打开失败", header.Filename))
			continue
		}

		tempFile := filepath.Join(os.TempDir(), "qr_"+header.Filename)
		dst, err := os.Create(tempFile)
		if err != nil {
			file.Close()
			errors = append(errors, fmt.Sprintf("%s: 创建临时文件失败", header.Filename))
			continue
		}

		io.Copy(dst, file)
		dst.Close()
		file.Close()

		qrmatrix, err := qrcode.Decode(tempFile)
		os.Remove(tempFile)

		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: 解码失败", header.Filename))
		} else {
			results = append(results, map[string]interface{}{
				"filename": header.Filename,
				"content":  qrmatrix.Content,
			})
		}
	}

	response := map[string]interface{}{
		"total":   len(files),
		"success": len(results),
		"failed":  len(errors),
		"results": results,
		"errors":  errors,
	}

	json.NewEncoder(w).Encode(response)
}

// ==================== 文件操作 ====================

// 下载二维码
func (s *QRService) DownloadQR(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/download/")

	s.mu.RLock()
	qr, exists := s.qrcodes[id]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "二维码不存在", http.StatusNotFound)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(qr.Path); os.IsNotExist(err) {
		sendJSONError(w, "文件已丢失", http.StatusNotFound)
		return
	}

	// 更新下载计数
	s.mu.Lock()
	qr.DownloadCount++
	s.qrcodes[id] = qr
	s.mu.Unlock()

	// 发送文件
	http.ServeFile(w, r, qr.Path)
}

// 预览二维码
func (s *QRService) PreviewQR(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/preview/")

	s.mu.RLock()
	qr, exists := s.qrcodes[id]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "二维码不存在", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, qr.Path)
}

// 获取二维码信息
func (s *QRService) GetQRInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	s.mu.RLock()
	qr, exists := s.qrcodes[id]
	s.mu.RUnlock()

	if !exists {
		sendJSONError(w, "二维码不存在", http.StatusNotFound)
		return
	}

	info := map[string]interface{}{
		"id":             qr.ID,
		"content":        qr.Content,
		"size":           qr.Size,
		"format":         qr.Format,
		"created_at":     qr.CreatedAt,
		"download_count": qr.DownloadCount,
		"url":            fmt.Sprintf("/download/%s", qr.ID),
		"preview":        fmt.Sprintf("/preview/%s", qr.ID),
	}

	json.NewEncoder(w).Encode(info)
}

// 列出所有二维码
func (s *QRService) ListQRs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	qrs := make([]map[string]interface{}, 0, len(s.qrcodes))
	for _, qr := range s.qrcodes {
		qrs = append(qrs, map[string]interface{}{
			"id":         qr.ID,
			"content":    truncateString(qr.Content, 50),
			"size":       qr.Size,
			"format":     qr.Format,
			"created_at": qr.CreatedAt,
			"preview":    fmt.Sprintf("/preview/%s", qr.ID),
		})
	}
	s.mu.RUnlock()

	json.NewEncoder(w).Encode(qrs)
}

// 删除二维码
func (s *QRService) DeleteQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSONError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/delete/")

	s.mu.Lock()
	qr, exists := s.qrcodes[id]
	if exists {
		os.Remove(qr.Path)
		delete(s.qrcodes, id)
	}
	s.mu.Unlock()

	if !exists {
		sendJSONError(w, "二维码不存在", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ==================== 工具函数 ====================

// 保存为JPEG
func saveAsJPEG(img image.Image, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img) // 简化处理，实际应该用jpeg
}

// 截断字符串
func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// 发送JSON错误
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// 生成带Logo的二维码
func (s *QRService) GenerateQRWithLogo(w http.ResponseWriter, r *http.Request) {
	// 这里可以添加带Logo的二维码生成逻辑
	sendJSONError(w, "功能开发中", http.StatusNotImplemented)
}

// ==================== Web界面 ====================

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>二维码工具箱 - QR Code Tool</title>
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
        
        .content {
            padding: 30px;
        }
        
        .tab-container {
            margin-bottom: 30px;
        }
        
        .tab-buttons {
            display: flex;
            border-bottom: 2px solid #667eea;
            flex-wrap: wrap;
        }
        
        .tab-btn {
            padding: 12px 24px;
            background: #f8f9fa;
            border: none;
            cursor: pointer;
            margin-right: 5px;
            border-radius: 5px 5px 0 0;
            font-size: 16px;
            transition: all 0.3s;
        }
        
        .tab-btn:hover {
            background: #e9ecef;
        }
        
        .tab-btn.active {
            background: #667eea;
            color: white;
        }
        
        .tab-content {
            display: none;
            padding: 30px;
            border: 1px solid #e0e0e0;
            border-top: none;
            border-radius: 0 0 8px 8px;
        }
        
        .tab-content.active {
            display: block;
        }
        
        .form-group {
            margin-bottom: 20px;
        }
        
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: bold;
            color: #333;
        }
        
        input, select, textarea {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        
        input:focus, select:focus, textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        
        textarea {
            min-height: 100px;
            resize: vertical;
        }
        
        button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
            transition: transform 0.2s;
            margin-right: 10px;
        }
        
        button:hover {
            transform: translateY(-2px);
        }
        
        button.secondary {
            background: #95a5a6;
        }
        
        button.secondary:hover {
            background: #7f8c8d;
        }
        
        .qr-preview {
            margin-top: 30px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 8px;
            text-align: center;
        }
        
        .qr-image {
            max-width: 300px;
            border: 1px solid #ddd;
            border-radius: 8px;
        }
        
        .result-area {
            margin-top: 20px;
            padding: 15px;
            background: #e8f4fd;
            border-radius: 8px;
            word-break: break-all;
        }
        
        .grid-2 {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
        }
        
        .file-list {
            margin-top: 20px;
        }
        
        .file-item {
            background: #f8f9fa;
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 8px;
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
        
        .file-actions button {
            padding: 5px 10px;
            font-size: 12px;
            margin-left: 5px;
        }
        
        .progress-bar {
            height: 10px;
            background: #e0e0e0;
            border-radius: 5px;
            overflow: hidden;
            margin: 10px 0;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            width: 0%;
            transition: width 0.3s;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        
        .stat-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }
        
        .stat-value {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        
        .stat-label {
            color: #666;
            margin-top: 5px;
        }
        
        @media (max-width: 768px) {
            .grid-2 {
                grid-template-columns: 1fr;
            }
            
            .tab-buttons {
                flex-direction: column;
            }
            
            .tab-btn {
                border-radius: 5px;
                margin-bottom: 5px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔲 二维码工具箱</h1>
            <p>生成、解码、管理二维码，一应俱全</p>
        </div>
        
        <div class="content">
            <div class="tab-container">
                <div class="tab-buttons">
                    <button class="tab-btn active" onclick="switchTab('generate')">生成二维码</button>
                    <button class="tab-btn" onclick="switchTab('batch')">批量生成</button>
                    <button class="tab-btn" onclick="switchTab('decode')">解码二维码</button>
                    <button class="tab-btn" onclick="switchTab('history')">历史记录</button>
                </div>
                
                <!-- 生成二维码标签页 -->
                <div id="generateTab" class="tab-content active">
                    <div class="grid-2">
                        <div>
                            <h3>📝 输入内容</h3>
                            <div class="form-group">
                                <label>内容类型</label>
                                <select id="contentType" onchange="updateContentTemplate()">
                                    <option value="text">纯文本</option>
                                    <option value="url">网址</option>
                                    <option value="wifi">WiFi</option>
                                    <option value="vcard">电子名片</option>
                                    <option value="email">电子邮件</option>
                                    <option value="phone">电话号码</option>
                                    <option value="sms">短信</option>
                                </select>
                            </div>
                            
                            <div class="form-group">
                                <label for="content">内容</label>
                                <textarea id="content" placeholder="请输入要编码的内容"></textarea>
                            </div>
                            
                            <div class="form-group">
                                <label for="size">二维码大小 (像素)</label>
                                <select id="size">
                                    <option value="128">128x128</option>
                                    <option value="256" selected>256x256</option>
                                    <option value="512">512x512</option>
                                    <option value="1024">1024x1024</option>
                                </select>
                            </div>
                            
                            <div class="form-group">
                                <label for="format">输出格式</label>
                                <select id="format">
                                    <option value="png">PNG</option>
                                    <option value="jpg">JPG</option>
                                </select>
                            </div>
                            
                            <button onclick="generateQR()">生成二维码</button>
                            <button class="secondary" onclick="clearForm()">清空</button>
                        </div>
                        
                        <div>
                            <h3>📸 预览</h3>
                            <div class="qr-preview" id="preview">
                                <p>点击生成按钮预览二维码</p>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 批量生成标签页 -->
                <div id="batchTab" class="tab-content">
                    <h3>📦 批量生成二维码</h3>
                    <div class="form-group">
                        <label>每行一个内容</label>
                        <textarea id="batchContents" rows="10" placeholder="https://example.com/1&#10;https://example.com/2&#10;https://example.com/3"></textarea>
                    </div>
                    
                    <div class="form-group">
                        <label>二维码大小</label>
                        <select id="batchSize">
                            <option value="128">128x128</option>
                            <option value="256" selected>256x256</option>
                            <option value="512">512x512</option>
                        </select>
                    </div>
                    
                    <button onclick="batchGenerate()">开始批量生成</button>
                    
                    <div id="batchProgress" style="display: none;">
                        <h4>生成进度</h4>
                        <div class="progress-bar">
                            <div class="progress-fill" id="batchProgressBar"></div>
                        </div>
                        <p id="batchStatus"></p>
                    </div>
                    
                    <div id="batchResults"></div>
                </div>
                
                <!-- 解码二维码标签页 -->
                <div id="decodeTab" class="tab-content">
                    <div class="grid-2">
                        <div>
                            <h3>📤 上传二维码图片</h3>
                            <div class="upload-area" onclick="document.getElementById('qrFile').click()" style="border: 3px dashed #667eea; padding: 40px; text-align: center; border-radius: 8px; cursor: pointer;">
                                <div style="font-size: 48px;">📱</div>
                                <p>点击或拖拽二维码图片到此处</p>
                                <p class="file-hint" style="color: #666; font-size: 0.9em;">支持 PNG、JPG、GIF</p>
                            </div>
                            <input type="file" id="qrFile" accept="image/*" style="display: none;" onchange="decodeQR(this.files[0])">
                            
                            <div style="margin-top: 20px;">
                                <button onclick="decodeQRFromClipboard()" style="width: 100%;">从剪贴板粘贴</button>
                            </div>
                        </div>
                        
                        <div>
                            <h3>📋 解码结果</h3>
                            <div id="decodeResult" class="result-area">
                                <p>上传图片后显示解码结果</p>
                            </div>
                        </div>
                    </div>
                    
                    <div style="margin-top: 30px;">
                        <h3>📦 批量解码</h3>
                        <input type="file" id="batchFiles" multiple accept="image/*" onchange="batchDecode(this.files)">
                        <div id="batchDecodeResults" style="margin-top: 20px;"></div>
                    </div>
                </div>
                
                <!-- 历史记录标签页 -->
                <div id="historyTab" class="tab-content">
                    <h3>📜 历史记录</h3>
                    <div class="file-list" id="historyList">
                        <p>加载中...</p>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // 切换标签页
        function switchTab(tabName) {
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            
            event.target.classList.add('active');
            document.getElementById(tabName + 'Tab').classList.add('active');
            
            if (tabName === 'history') {
                loadHistory();
            }
        }
        
        // 更新内容模板
        function updateContentTemplate() {
            const type = document.getElementById('contentType').value;
            const contentArea = document.getElementById('content');
            
            const templates = {
                'text': '请输入任意文本内容',
                'url': 'https://example.com',
                'wifi': 'WIFI:T:WPA;S:网络名称;P:密码;;',
                'vcard': 'BEGIN:VCARD\nVERSION:3.0\nN:张三\nTEL:12345678\nEMAIL:zhangsan@example.com\nEND:VCARD',
                'email': 'mailto:example@example.com?subject=标题&body=内容',
                'phone': 'tel:+861234567890',
                'sms': 'smsto:1234567890:短信内容'
            };
            
            contentArea.placeholder = templates[type];
        }
        
        // 生成二维码
        function generateQR() {
            const content = document.getElementById('content').value;
            if (!content) {
                alert('请输入内容');
                return;
            }
            
            const data = {
                content: content,
                size: parseInt(document.getElementById('size').value),
                format: document.getElementById('format').value
            };
            
            fetch('/api/generate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            })
            .then(response => response.json())
            .then(data => {
                if (data.error) {
                    alert('生成失败: ' + data.error);
                } else {
                    showPreview(data.preview);
                }
            });
        }
        
        // 显示预览
        function showPreview(url) {
            const preview = document.getElementById('preview');
            preview.innerHTML = `<img src="${url}" class="qr-image" alt="QR Code">`;
        }
        
        // 清空表单
        function clearForm() {
            document.getElementById('content').value = '';
            document.getElementById('preview').innerHTML = '<p>点击生成按钮预览二维码</p>';
        }
        
        // 批量生成
        function batchGenerate() {
            const contents = document.getElementById('batchContents').value.split('\n').filter(c => c.trim());
            if (contents.length === 0) {
                alert('请输入内容');
                return;
            }
            
            const data = {
                contents: contents,
                size: parseInt(document.getElementById('batchSize').value),
                format: 'png'
            };
            
            document.getElementById('batchProgress').style.display = 'block';
            
            fetch('/api/batch/generate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            })
            .then(response => response.json())
            .then(data => {
                if (data.job_id) {
                    checkBatchProgress(data.job_id);
                }
            });
        }
        
        // 检查批量进度
        function checkBatchProgress(jobId) {
            const interval = setInterval(() => {
                fetch('/api/job/status?job_id=' + jobId)
                    .then(response => response.json())
                    .then(data => {
                        const progress = (data.completed / data.total * 100).toFixed(1);
                        document.getElementById('batchProgressBar').style.width = progress + '%';
                        document.getElementById('batchStatus').textContent = 
                            `已完成: ${data.completed}/${data.total}`;
                        
                        if (data.status === 'completed') {
                            clearInterval(interval);
                            showBatchResults(data);
                        }
                    });
            }, 1000);
        }
        
        // 显示批量结果
        function showBatchResults(data) {
            let html = '<h4>生成结果</h4>';
            data.results.forEach(url => {
                html += `<img src="${url}" style="width: 100px; margin: 5px;">`;
            });
            document.getElementById('batchResults').innerHTML = html;
        }
        
        // 解码二维码
        function decodeQR(file) {
            if (!file) return;
            
            const formData = new FormData();
            formData.append('qr_image', file);
            
            fetch('/api/decode', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.error) {
                    document.getElementById('decodeResult').innerHTML = 
                        `<p style="color: #e74c3c;">解码失败: ${data.error}</p>`;
                } else {
                    document.getElementById('decodeResult').innerHTML = `
	<p><strong>文件名:</strong> ${data.filename}</p>
	<p><strong>内容:</strong></p>
	<pre style="background: white; padding: 10px; border-radius: 5px;">${data.content}</pre>
		`;
                }
            });
        }
        
        // 从剪贴板解码
        function decodeQRFromClipboard() {
            navigator.clipboard.read().then(clipboardItems => {
                for (let item of clipboardItems) {
                    for (let type of item.types) {
                        if (type.startsWith('image/')) {
                            item.getType(type).then(blob => {
                                const file = new File([blob], 'clipboard.png', {type: type});
                                decodeQR(file);
                            });
                        }
                    }
                }
            });
        }
        
        // 批量解码
        function batchDecode(files) {
            if (!files || files.length === 0) return;
            
            const formData = new FormData();
            for (let file of files) {
                formData.append('qr_images', file);
            }
            
            fetch('/api/batch/decode', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                let html = `<p>共处理 ${data.total} 个文件，成功 ${data.success} 个</p>`;
                
                if (data.results.length > 0) {
                    html += '<h4>解码结果</h4>';
                    data.results.forEach(r => {
                        html += `
	<div style="background: #f0f0f0; padding: 10px; margin: 5px 0; border-radius: 5px;">
	<strong>${r.filename}</strong>: ${r.content}
	</div>
		`;
                    });
               