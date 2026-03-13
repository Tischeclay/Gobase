// apigateway.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Route struct {
	Path        string            `yaml:"path"`
	Target      string            `yaml:"target"`
	Methods     []string          `yaml:"methods"`
	StripPrefix bool              `yaml:"strip_prefix"`
	Headers     map[string]string `yaml:"headers"`
	RateLimit   int               `yaml:"rate_limit"` // 每秒请求数
}

type Gateway struct {
	routes     []Route
	proxy      map[string]*httputil.ReverseProxy
	rateLimits map[string]*RateLimiter
	mu         sync.RWMutex
	stats      *Stats
	authToken  string
	enableAuth bool
}

type RateLimiter struct {
	tokens chan struct{}
	ticker *time.Ticker
}

type Stats struct {
	TotalRequests  int64
	SuccessCount   int64
	ErrorCount     int64
	BytesIn        int64
	BytesOut       int64
	LastRequest    time.Time
	RequestsPerSec float64
	mu             sync.Mutex
}

func NewRateLimiter(rate int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, rate),
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
	}

	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func NewGateway() *Gateway {
	return &Gateway{
		routes:     make([]Route, 0),
		proxy:      make(map[string]*httputil.ReverseProxy),
		rateLimits: make(map[string]*RateLimiter),
		stats:      &Stats{},
	}
}

func (g *Gateway) LoadRoutes(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// 简单格式：每行 "path target method1,method2 stripPrefix"
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			log.Printf("第%d行格式错误: %s", i+1, line)
			continue
		}

		route := Route{
			Path:        parts[0],
			Target:      parts[1],
			Methods:     strings.Split(parts[2], ","),
			StripPrefix: len(parts) > 3 && parts[3] == "true",
		}

		if len(parts) > 4 {
			fmt.Sscanf(parts[4], "%d", &route.RateLimit)
		}

		g.AddRoute(route)
	}

	return nil
}

func (g *Gateway) AddRoute(route Route) {
	g.mu.Lock()
	defer g.mu.Unlock()

	targetURL, err := url.Parse(route.Target)
	if err != nil {
		log.Printf("无效的目标URL %s: %v", route.Target, err)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义Director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 修改Host头
		req.Host = targetURL.Host

		// 添加自定义头
		for k, v := range route.Headers {
			req.Header.Set(k, v)
		}

		// 添加网关标识
		req.Header.Set("X-Gateway", "Simple-API-Gateway/1.0")
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("代理错误: %v", err)
		g.stats.mu.Lock()
		g.stats.ErrorCount++
		g.stats.mu.Unlock()
		http.Error(w, "代理服务错误", http.StatusBadGateway)
	}

	g.routes = append(g.routes, route)
	g.proxy[route.Path] = proxy

	if route.RateLimit > 0 {
		g.rateLimits[route.Path] = NewRateLimiter(route.RateLimit)
	}

	log.Printf("添加路由: %s -> %s [%v]", route.Path, route.Target, route.Methods)
}

func (g *Gateway) findRoute(path string) (*Route, *httputil.ReverseProxy, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 精确匹配
	if proxy, ok := g.proxy[path]; ok {
		for _, route := range g.routes {
			if route.Path == path {
				return &route, proxy, true
			}
		}
	}

	// 前缀匹配
	for _, route := range g.routes {
		if strings.HasPrefix(path, route.Path) {
			if proxy, ok := g.proxy[route.Path]; ok {
				return &route, proxy, true
			}
		}
	}

	return nil, nil, false
}

func (g *Gateway) authenticate(r *http.Request) bool {
	if !g.enableAuth || g.authToken == "" {
		return true
	}

	auth := r.Header.Get("Authorization")
	return auth == "Bearer "+g.authToken || auth == "Token "+g.authToken
}

func (g *Gateway) updateStats(r *http.Request, status int, size int64) {
	g.stats.mu.Lock()
	defer g.stats.mu.Unlock()

	g.stats.TotalRequests++
	g.stats.LastRequest = time.Now()

	if status >= 200 && status < 300 {
		g.stats.SuccessCount++
	} else {
		g.stats.ErrorCount++
	}

	if r.ContentLength > 0 {
		g.stats.BytesIn += r.ContentLength
	}
	g.stats.BytesOut += size
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 统计处理器
	if r.URL.Path == "/_stats" {
		g.handleStats(w, r)
		return
	}

	// 健康检查
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// 路由匹配
	route, proxy, found := g.findRoute(r.URL.Path)
	if !found {
		http.Error(w, "Not Found", http.StatusNotFound)
		g.updateStats(r, http.StatusNotFound, 0)
		return
	}

	// 方法检查
	methodAllowed := false
	for _, m := range route.Methods {
		if m == "*" || strings.EqualFold(m, r.Method) {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		g.updateStats(r, http.StatusMethodNotAllowed, 0)
		return
	}

	// 认证检查
	if !g.authenticate(r) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="gateway"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		g.updateStats(r, http.StatusUnauthorized, 0)
		return
	}

	// 限流检查
	if route.RateLimit > 0 {
		limiter := g.rateLimits[route.Path]
		if limiter != nil && !limiter.Allow() {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", route.RateLimit))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			g.updateStats(r, http.StatusTooManyRequests, 0)
			return
		}
	}

	// 修改请求路径（如果需要）
	if route.StripPrefix {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Path)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
	}

	// 记录请求
	log.Printf("[%s] %s -> %s", r.Method, r.URL.Path, route.Target)

	// 捕获响应
	recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

	// 转发请求
	proxy.ServeHTTP(recorder, r)

	// 更新统计
	g.updateStats(r, recorder.statusCode, int64(recorder.size))

	// 记录响应时间
	duration := time.Since(start)
	log.Printf("完成: %s %s [%d] %v", r.Method, r.URL.Path, recorder.statusCode, duration)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func (g *Gateway) handleStats(w http.ResponseWriter, r *http.Request) {
	g.stats.mu.Lock()
	defer g.stats.mu.Unlock()

	// 计算QPS
	var qps float64
	if time.Since(g.stats.LastRequest) < time.Minute {
		qps = float64(g.stats.TotalRequests) / time.Since(g.stats.LastRequest).Seconds()
	}

	stats := map[string]interface{}{
		"total_requests": g.stats.TotalRequests,
		"success_count":  g.stats.SuccessCount,
		"error_count":    g.stats.ErrorCount,
		"bytes_in":       g.stats.BytesIn,
		"bytes_out":      g.stats.BytesOut,
		"last_request":   g.stats.LastRequest,
		"current_qps":    qps,
		"uptime":         time.Since(g.stats.LastRequest).String(),
		"routes_count":   len(g.routes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	// 命令行参数
	port := flag.Int("p", 8080, "监听端口")
	configFile := flag.String("c", "routes.txt", "路由配置文件")
	authToken := flag.String("auth", "", "认证令牌")
	flag.Parse()

	// 创建网关
	gateway := NewGateway()
	gateway.enableAuth = *authToken != " "
	gateway.authToken = *authToken

	// 加载路由配置
	if err := gateway.LoadRoutes(*configFile); err != nil {
		log.Printf("警告: 无法加载路由配置: %v", err)
		// 添加默认路由
		gateway.AddRoute(Route{
			Path:        "/api",
			Target:      "http://localhost:3000",
			Methods:     []string{"*"},
			StripPrefix: false,
		})
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("API网关启动在 %s", addr)
	log.Printf("认证: %v", gateway.enableAuth)
	log.Fatal(http.ListenAndServe(addr, gateway))
}

// routes.txt 示例:
// /api http://localhost:3000 * false 100
// /auth http://localhost:3001 POST,GET true
// /static http://localhost:3002 GET true
