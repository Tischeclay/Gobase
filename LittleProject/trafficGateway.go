// traffic_gateway.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ==================== 配置结构 ====================

type GatewayConfig struct {
	ListenAddr    string        `json:"listen_addr"`
	ReadTimeout   time.Duration `json:"read_timeout"`
	WriteTimeout  time.Duration `json:"write_timeout"`
	IdleTimeout   time.Duration `json:"idle_timeout"`
	MaxBodySize   int64         `json:"max_body_size"`
	EnableMetrics bool          `json:"enable_metrics"`
}

type RouteConfig struct {
	ID             string               `json:"id"`
	Path           string               `json:"path"`
	Methods        []string             `json:"methods"`
	Upstreams      []UpstreamConfig     `json:"upstreams"`
	LoadBalance    string               `json:"load_balance"` // round_robin, random, least_conn
	RetryCount     int                  `json:"retry_count"`
	Timeout        time.Duration        `json:"timeout"`
	RateLimit      int                  `json:"rate_limit"` // 每秒请求数
	Burst          int                  `json:"burst"`      // 突发请求数
	StripPrefix    bool                 `json:"strip_prefix"`
	RewritePath    string               `json:"rewrite_path"`
	Auth           AuthConfig           `json:"auth"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`
	Headers        map[string]string    `json:"headers"`
}

type UpstreamConfig struct {
	URL         string        `json:"url"`
	Weight      int           `json:"weight"`
	Healthy     bool          `json:"healthy"`
	MaxFails    int           `json:"max_fails"`
	FailTimeout time.Duration `json:"fail_timeout"`
}

type AuthConfig struct {
	Enabled   bool              `json:"enabled"`
	Type      string            `json:"type"` // api_key, jwt, basic
	APIKeys   map[string]string `json:"api_keys"`
	JWTSECRET string            `json:"jwt_secret"`
}

type CircuitBreakerConfig struct {
	Enabled          bool          `json:"enabled"`
	FailureThreshold int           `json:"failure_threshold"`
	Timeout          time.Duration `json:"timeout"`
	HalfOpenMax      int           `json:"half_open_max"`
}

// ==================== 核心组件 ====================

// 上游服务器
type Upstream struct {
	Config      UpstreamConfig
	URL         *url.URL
	FailCount   int32
	LastFail    time.Time
	Connections int32
	mu          sync.RWMutex
}

// 路由
type Route struct {
	Config         RouteConfig
	Upstreams      []*Upstream
	Proxy          *httputil.ReverseProxy
	RateLimiter    *RateLimiter
	CircuitBreaker *CircuitBreaker
	Stats          *RouteStats
}

// 限流器
type RateLimiter struct {
	tokens   chan struct{}
	burst    chan struct{}
	ticker   *time.Ticker
	mu       sync.Mutex
	stopChan chan struct{}
}

// 熔断器
type CircuitBreaker struct {
	Config       CircuitBreakerConfig
	State        int32 // 0=closed, 1=open, 2=half_open
	FailureCount int32
	LastFailure  time.Time
	mu           sync.RWMutex
}

// 路由统计
type RouteStats struct {
	TotalRequests  int64
	SuccessCount   int64
	ErrorCount     int64
	TimeoutCount   int64
	RateLimitCount int64
	TotalTime      time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	LastRequest    time.Time
	StatusCodes    map[int]int64
	UpstreamStats  map[string]*UpstreamStats
	mu             sync.Mutex
}

type UpstreamStats struct {
	Requests  int64
	Success   int64
	Errors    int64
	TotalTime time.Duration
}

// 流量网关主结构
type TrafficGateway struct {
	config     GatewayConfig
	routes     map[string]*Route
	httpServer *http.Server
	metrics    *MetricsCollector
	mu         sync.RWMutex
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// 指标收集器
type MetricsCollector struct {
	TotalRequests     int64
	ActiveConnections int64
	BytesIn           int64
	BytesOut          int64
	ErrorCount        int64
	mu                sync.Mutex
}

// ==================== 限流器实现 ====================

func NewRateLimiter(rate int, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, burst),
		burst:    make(chan struct{}, burst),
		ticker:   time.NewTicker(time.Second / time.Duration(rate)),
		stopChan: make(chan struct{}),
	}

	// 初始化令牌桶
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
		rl.burst <- struct{}{}
	}

	go rl.run()

	return rl
}

func (rl *RateLimiter) run() {
	for {
		select {
		case <-rl.ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		case <-rl.stopChan:
			rl.ticker.Stop()
			return
		}
	}
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) AllowBurst() bool {
	select {
	case <-rl.burst:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// ==================== 熔断器实现 ====================

const (
	StateClosed   int32 = 0
	StateOpen     int32 = 1
	StateHalfOpen int32 = 2
)

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		Config: config,
		State:  StateClosed,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.Allow() {
		return fmt.Errorf("circuit breaker open")
	}

	err := fn()
	cb.RecordResult(err == nil)

	return err
}

func (cb *CircuitBreaker) Allow() bool {
	state := atomic.LoadInt32(&cb.State)

	if state == StateOpen {
		cb.mu.RLock()
		lastFail := cb.LastFailure
		cb.mu.RUnlock()

		if time.Since(lastFail) > cb.Config.Timeout {
			atomic.CompareAndSwapInt32(&cb.State, StateOpen, StateHalfOpen)
			return true
		}
		return false
	}

	return true
}

func (cb *CircuitBreaker) RecordResult(success bool) {
	if !success {
		failures := atomic.AddInt32(&cb.FailureCount, 1)

		if int(failures) >= cb.Config.FailureThreshold {
			cb.mu.Lock()
			cb.LastFailure = time.Now()
			cb.mu.Unlock()
			atomic.StoreInt32(&cb.State, StateOpen)
			atomic.StoreInt32(&cb.FailureCount, 0)
		}
	} else {
		if atomic.LoadInt32(&cb.State) == StateHalfOpen {
			atomic.StoreInt32(&cb.State, StateClosed)
			atomic.StoreInt32(&cb.FailureCount, 0)
		} else {
			atomic.StoreInt32(&cb.FailureCount, 0)
		}
	}
}

// ==================== 负载均衡器 ====================

type LoadBalancer struct {
	upstreams []*Upstream
	strategy  string
	current   int32
}

func NewLoadBalancer(upstreams []*Upstream, strategy string) *LoadBalancer {
	return &LoadBalancer{
		upstreams: upstreams,
		strategy:  strategy,
	}
}

func (lb *LoadBalancer) Next() *Upstream {
	if len(lb.upstreams) == 0 {
		return nil
	}

	// 过滤健康的upstream
	healthy := make([]*Upstream, 0)
	for _, u := range lb.upstreams {
		if u.IsHealthy() {
			healthy = append(healthy, u)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	switch lb.strategy {
	case "random":
		return healthy[rand.Intn(len(healthy))]
	case "least_conn":
		var minConn int32 = 1 << 30
		var selected *Upstream
		for _, u := range healthy {
			if u.Connections < minConn {
				minConn = u.Connections
				selected = u
			}
		}
		return selected
	default: // round_robin
		idx := atomic.AddInt32(&lb.current, 1)
		return healthy[int(idx)%len(healthy)]
	}
}

func (u *Upstream) IsHealthy() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if !u.Config.Healthy {
		return false
	}

	if u.FailCount >= int32(u.Config.MaxFails) {
		if time.Since(u.LastFail) > u.Config.FailTimeout {
			u.mu.RUnlock()
			u.mu.Lock()
			u.FailCount = 0
			u.mu.Unlock()
			u.mu.RLock()
			return true
		}
		return false
	}

	return true
}

func (u *Upstream) MarkFailed() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.FailCount++
	u.LastFail = time.Now()
	u.Config.Healthy = u.FailCount < int32(u.Config.MaxFails)
}

func (u *Upstream) MarkSuccess() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.FailCount = 0
	u.Config.Healthy = true
}

// ==================== 路由处理器 ====================

func NewRoute(config RouteConfig) (*Route, error) {
	// 解析上游地址
	upstreams := make([]*Upstream, 0)
	for _, uc := range config.Upstreams {
		u, err := url.Parse(uc.URL)
		if err != nil {
			return nil, err
		}

		upstreams = append(upstreams, &Upstream{
			Config: uc,
			URL:    u,
		})
	}

	// 创建反向代理
	lb := NewLoadBalancer(upstreams, config.LoadBalance)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			upstream := lb.Next()
			if upstream == nil {
				return
			}

			// 更新连接数
			atomic.AddInt32(&upstream.Connections, 1)

			// 修改请求
			req.URL.Scheme = upstream.URL.Scheme
			req.URL.Host = upstream.URL.Host

			// 路径重写
			if config.StripPrefix && config.Path != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, config.Path)
			}
			if config.RewritePath != "" {
				req.URL.Path = config.RewritePath
			}

			// 添加网关头
			req.Header.Set("X-Gateway", "Traffic-Gateway/1.0")
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)

			// 添加自定义头
			for k, v := range config.Headers {
				req.Header.Set(k, v)
			}
		},
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("代理错误: %v", err)
			http.Error(w, "Gateway Error", http.StatusBadGateway)
		},
		ModifyResponse: func(resp *http.Response) error {
			// 更新上游统计
			return nil
		},
	}

	// 创建限流器
	var rateLimiter *RateLimiter
	if config.RateLimit > 0 {
		burst := config.Burst
		if burst == 0 {
			burst = config.RateLimit
		}
		rateLimiter = NewRateLimiter(config.RateLimit, burst)
	}

	// 创建熔断器
	var cb *CircuitBreaker
	if config.CircuitBreaker.Enabled {
		cb = NewCircuitBreaker(config.CircuitBreaker)
	}

	return &Route{
		Config:         config,
		Upstreams:      upstreams,
		Proxy:          proxy,
		RateLimiter:    rateLimiter,
		CircuitBreaker: cb,
		Stats:          &RouteStats{StatusCodes: make(map[int]int64), UpstreamStats: make(map[string]*UpstreamStats)},
	}, nil
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// 更新统计
	atomic.AddInt64(&r.Stats.TotalRequests, 1)
	r.Stats.mu.Lock()
	r.Stats.LastRequest = time.Now()
	r.Stats.mu.Unlock()

	// 方法检查
	if len(r.Config.Methods) > 0 {
		methodAllowed := false
		for _, m := range r.Config.Methods {
			if strings.EqualFold(m, req.Method) {
				methodAllowed = true
				break
			}
		}
		if !methodAllowed {
			r.updateStats(http.StatusMethodNotAllowed, 0, time.Since(start))
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	// 限流检查
	if r.RateLimiter != nil && !r.RateLimiter.Allow() {
		r.updateStats(http.StatusTooManyRequests, 0, time.Since(start))
		atomic.AddInt64(&r.Stats.RateLimitCount, 1)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", r.Config.RateLimit))
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// 认证检查
	if r.Config.Auth.Enabled {
		if !r.authenticate(req) {
			r.updateStats(http.StatusUnauthorized, 0, time.Since(start))
			w.Header().Set("WWW-Authenticate", `Bearer realm="gateway"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// 熔断器检查
	if r.CircuitBreaker != nil {
		err := r.CircuitBreaker.Call(func() error {
			// 记录请求
			recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			r.Proxy.ServeHTTP(recorder, req)
			return nil
		})

		if err != nil {
			r.updateStats(http.StatusServiceUnavailable, 0, time.Since(start))
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		r.updateStats(recorder.statusCode, recorder.size, time.Since(start))
		return
	}

	// 正常代理
	recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	r.Proxy.ServeHTTP(recorder, req)

	r.updateStats(recorder.statusCode, recorder.size, time.Since(start))
}

func (r *Route) authenticate(req *http.Request) bool {
	auth := r.Config.Auth

	switch auth.Type {
	case "api_key":
		apiKey := req.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = req.URL.Query().Get("api_key")
		}
		expected, ok := auth.APIKeys[apiKey]
		return ok && expected == apiKey
	case "jwt":
		// JWT验证（简化实现）
		token := req.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			return false
		}
		// 实际应该验证JWT
		return true
	default:
		return true
	}
}

func (r *Route) updateStats(statusCode int, size int, duration time.Duration) {
	atomic.AddInt64(&r.Stats.SuccessCount, 1)
	if statusCode >= 400 {
		atomic.AddInt64(&r.Stats.ErrorCount, 1)
	}

	r.Stats.mu.Lock()
	defer r.Stats.mu.Unlock()

	r.Stats.StatusCodes[statusCode]++
	r.Stats.TotalTime += duration

	if r.Stats.MinTime == 0 || duration < r.Stats.MinTime {
		r.Stats.MinTime = duration
	}
	if duration > r.Stats.MaxTime {
		r.Stats.MaxTime = duration
	}
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
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// ==================== 网关主逻辑 ====================

func NewTrafficGateway(config GatewayConfig) *TrafficGateway {
	return &TrafficGateway{
		config:   config,
		routes:   make(map[string]*Route),
		metrics:  &MetricsCollector{},
		stopChan: make(chan struct{}),
	}
}

func (g *TrafficGateway) LoadRoutes(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var routes []RouteConfig
	if err := json.Unmarshal(data, &routes); err != nil {
		return err
	}

	for _, rc := range routes {
		route, err := NewRoute(rc)
		if err != nil {
			log.Printf("加载路由 %s 失败: %v", rc.ID, err)
			continue
		}

		g.mu.Lock()
		g.routes[rc.Path] = route
		g.mu.Unlock()

		log.Printf("加载路由: %s -> %v", rc.Path, rc.Upstreams)
	}

	return nil
}

func (g *TrafficGateway) findRoute(path string) (*Route, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 精确匹配
	if route, ok := g.routes[path]; ok {
		return route, true
	}

	// 前缀匹配
	for prefix, route := range g.routes {
		if strings.HasPrefix(path, prefix) {
			return route, true
		}
	}

	return nil, false
}

func (g *TrafficGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 统计
	atomic.AddInt64(&g.metrics.TotalRequests, 1)
	atomic.AddInt64(&g.metrics.ActiveConnections, 1)
	defer atomic.AddInt64(&g.metrics.ActiveConnections, -1)

	// 健康检查
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// 指标接口
	if r.URL.Path == "/metrics" && g.config.EnableMetrics {
		g.handleMetrics(w, r)
		return
	}

	// 路由匹配
	route, found := g.findRoute(r.URL.Path)
	if !found {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// 转发
	route.ServeHTTP(w, r)

	// 更新统计
	duration := time.Since(start)
	atomic.AddInt64(&g.metrics.BytesIn, r.ContentLength)
}

func (g *TrafficGateway) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"total_requests":     atomic.LoadInt64(&g.metrics.TotalRequests),
		"active_connections": atomic.LoadInt64(&g.metrics.ActiveConnections),
		"bytes_in":           atomic.LoadInt64(&g.metrics.BytesIn),
		"bytes_out":          atomic.LoadInt64(&g.metrics.BytesOut),
		"error_count":        atomic.LoadInt64(&g.metrics.ErrorCount),
	}

	// 路由统计
	routeStats := make(map[string]interface{})
	g.mu.RLock()
	for path, route := range g.routes {
		route.Stats.mu.Lock()
		stats := map[string]interface{}{
			"total_requests": route.Stats.TotalRequests,
			"success_count":  route.Stats.SuccessCount,
			"error_count":    route.Stats.ErrorCount,
			"rate_limit":     route.Stats.RateLimitCount,
			"avg_time_ms":    route.Stats.TotalTime.Milliseconds() / max(1, route.Stats.TotalRequests),
			"status_codes":   route.Stats.StatusCodes,
		}
		route.Stats.mu.Unlock()
		routeStats[path] = stats
	}
	g.mu.RUnlock()

	metrics["routes"] = routeStats

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (g *TrafficGateway) Start() error {
	g.httpServer = &http.Server{
		Addr:         g.config.ListenAddr,
		Handler:      g,
		ReadTimeout:  g.config.ReadTimeout,
		WriteTimeout: g.config.WriteTimeout,
		IdleTimeout:  g.config.IdleTimeout,
	}

	go func() {
		log.Printf("流量网关启动在 %s", g.config.ListenAddr)
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	return nil
}

func (g *TrafficGateway) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("正在关闭网关...")
	return g.httpServer.Shutdown(ctx)
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ==================== 健康检查 ====================

func (g *TrafficGateway) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.mu.RLock()
			for _, route := range g.routes {
				for _, upstream := range route.Upstreams {
					go g.checkUpstreamHealth(upstream)
				}
			}
			g.mu.RUnlock()
		case <-g.stopChan:
			return
		}
	}
}

func (g *TrafficGateway) checkUpstreamHealth(upstream *Upstream) {
	client := &http.Client{Timeout: 5 * time.Second}
	healthURL := upstream.URL.String() + "/health"

	resp, err := client.Get(healthURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		upstream.MarkFailed()
		return
	}
	defer resp.Body.Close()

	upstream.MarkSuccess()
}

// ==================== 主函数 ====================

func main() {
	var configFile string
	var listenAddr string
	var routesFile string

	flag.StringVar(&configFile, "config", "", "网关配置文件")
	flag.StringVar(&listenAddr, "listen", ":8080", "监听地址")
	flag.StringVar(&routesFile, "routes", "routes.json", "路由配置文件")
	flag.Parse()

	// 默认配置
	config := GatewayConfig{
		ListenAddr:    listenAddr,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		IdleTimeout:   120 * time.Second,
		MaxBodySize:   10 * 1024 * 1024,
		EnableMetrics: true,
	}

	// 启动健康检查
	go gateway.healthCheck()
	
}
