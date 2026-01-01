package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// URLInfo 存储URL信息
type URLInfo struct {
	URL   string
	Depth int
}

// Crawler 爬虫结构体
type Crawler struct {
	startURL    string
	maxDepth    int
	visited     sync.Map
	results     chan string
	workerCount int
	delay       time.Duration
	domain      string
	outputDir   string
}

// NewCrawler 创建新的爬虫实例
func NewCrawler(startURL string, maxDepth, workers int, delay time.Duration) (*Crawler, error) {
	parsed, err := url.Parse(startURL)
	if err != nil {
		return nil, err
	}

	// 创建输出目录
	outputDir := fmt.Sprintf("crawled_%s_%d",
		strings.ReplaceAll(parsed.Host, ".", "_"),
		time.Now().Unix())

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	return &Crawler{
		startURL:    startURL,
		maxDepth:    maxDepth,
		results:     make(chan string, 1000),
		workerCount: workers,
		delay:       delay,
		domain:      parsed.Host,
		outputDir:   outputDir,
	}, nil
}

// Start 启动爬虫
func (c *Crawler) Start() {
	fmt.Printf("🚀 开始爬取: %s\n", c.startURL)
	fmt.Printf("📁 输出目录: %s\n", c.outputDir)

	var wg sync.WaitGroup

	// 启动结果处理器
	wg.Add(1)
	go c.resultProcessor(&wg)

	// 启动worker
	urlChan := make(chan URLInfo, 1000)
	for i := 0; i < c.workerCount; i++ {
		wg.Add(1)
		go c.worker(i+1, urlChan, &wg)
	}

	// 发送初始URL
	urlChan <- URLInfo{URL: c.startURL, Depth: 0}

	// 等待所有worker完成
	close(urlChan)
	wg.Wait()
	close(c.results)

	fmt.Println("\n✅ 爬取完成!")
}

// worker 爬虫工作线程
func (c *Crawler) worker(id int, urlChan chan URLInfo, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range urlChan {
		// 检查深度限制
		if task.Depth > c.maxDepth {
			continue
		}

		// 检查是否已访问
		if _, visited := c.visited.LoadOrStore(task.URL, true); visited {
			continue
		}

		// 延迟控制
		time.Sleep(c.delay)

		// 获取页面
		fmt.Printf("[Worker %d] 📥 获取: %s (深度: %d)\n",
			id, task.URL, task.Depth)

		content, links, err := c.fetchPage(task.URL)
		if err != nil {
			fmt.Printf("[Worker %d] ❌ 错误: %s - %v\n",
				id, task.URL, err)
			continue
		}

		// 保存页面内容
		c.results <- fmt.Sprintf("PAGE|%s|%s", task.URL, content)

		// 提取并处理链接
		for _, link := range links {
			// 转换为绝对URL
			absoluteURL := c.resolveURL(task.URL, link)
			if absoluteURL == "" {
				continue
			}

			// 检查域名限制（可选）
			if !c.isSameDomain(absoluteURL) {
				continue
			}

			// 发送到队列
			select {
			case urlChan <- URLInfo{URL: absoluteURL, Depth: task.Depth + 1}:
			default:
				// 队列满，丢弃
			}
		}
	}
}

// fetchPage 获取页面内容和链接
func (c *Crawler) fetchPage(urlStr string) (string, []string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", nil, err
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.BodyClose()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 读取内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	// 提取链接
	links := c.extractLinks(urlStr, body)

	return string(body), links, nil
}

// extractLinks 从HTML中提取链接
func (c *Crawler) extractLinks(baseURL string, content []byte) []string {
	var links []string

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return links
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var attrName string

			switch n.Data {
			case "a", "link":
				attrName = "href"
			case "img", "script":
				attrName = "src"
			case "iframe":
				attrName = "src"
			}

			if attrName != "" {
				for _, attr := range n.Attr {
					if attr.Key == attrName && attr.Val != "" {
						links = append(links, attr.Val)
					}
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}

	f(doc)
	return links
}

// resolveURL 解析相对URL为绝对URL
func (c *Crawler) resolveURL(base, relative string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}

	relURL, err := url.Parse(relative)
	if err != nil {
		return ""
	}

	absURL := baseURL.ResolveReference(relURL)
	return absURL.String()
}

// isSameDomain 检查是否同一域名
func (c *Crawler) isSameDomain(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsed.Host == c.domain || parsed.Host == ""
}

// resultProcessor 处理爬取结果
func (c *Crawler) resultProcessor(wg *sync.WaitGroup) {
	defer wg.Done()

	// 创建索引文件
	indexFile, err := os.Create(filepath.Join(c.outputDir, "index.txt"))
	if err != nil {
		fmt.Printf("❌ 创建索引文件失败: %v\n", err)
		return
	}
	defer indexFile.Close()

	for result := range c.results {
		parts := strings.SplitN(result, "|", 3)
		if len(parts) != 3 {
			continue
		}

		urlStr := parts[1]
		content := parts[2]

		// 保存到文件
		filename := c.generateFilename(urlStr)
		filepath := filepath.Join(c.outputDir, filename)

		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			fmt.Printf("❌ 保存文件失败: %s - %v\n", urlStr, err)
			continue
		}

		// 记录到索引
		indexFile.WriteString(fmt.Sprintf("%s -> %s\n", urlStr, filename))

		fmt.Printf("💾 已保存: %s\n", filename)
	}
}

// generateFilename 生成文件名
func (c *Crawler) generateFilename(urlStr string) string {
	// 移除协议和非法字符
	filename := strings.ReplaceAll(urlStr, "://", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "&", "_")
	filename = strings.ReplaceAll(filename, "=", "_")

	// 限制长度
	if len(filename) > 100 {
		filename = filename[:100]
	}

	return filename + ".html"
}

func main() {
	startURL := "https://example.com"
	maxDepth := 2
	workers := 5
	delay := 1 * time.Second

	crawler, err := NewCrawler(startURL, maxDepth, workers, delay)
	if err != nil {
		fmt.Printf("❌ 初始化失败: %v\n", err)
		return
	}

	crawler.Start()
}
