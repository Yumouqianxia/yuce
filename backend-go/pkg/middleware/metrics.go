package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求大小
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// HTTP响应大小
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status"},
	)

	// 当前活跃连接数
	httpActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// 业务指标
	userRegistrations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
		[]string{"source"},
	)

	userLogins = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
		[]string{"status"},
	)

	predictionsCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "predictions_created_total",
			Help: "Total number of predictions created",
		},
		[]string{"match_type"},
	)

	votesCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "votes_created_total",
			Help: "Total number of votes created",
		},
		[]string{"prediction_type"},
	)

	cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	databaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Number of database connections",
		},
		[]string{"state"}, // open, idle, in_use
	)

	databaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	redisConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_connections",
			Help: "Number of Redis connections",
		},
		[]string{"state"}, // active, idle
	)

	redisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// WebSocket metrics removed
)

// MetricsConfig 指标配置
type MetricsConfig struct {
	// SkipPaths 跳过记录指标的路径
	SkipPaths []string
	// NormalizePath 是否标准化路径（将参数替换为占位符）
	NormalizePath bool
	// MaxPathLabels 最大路径标签数量
	MaxPathLabels int
}

// DefaultMetricsConfig 默认指标配置
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		SkipPaths: []string{
			"/metrics",
			"/health",
			"/favicon.ico",
		},
		NormalizePath: true,
		MaxPathLabels: 100,
	}
}

// MetricsMiddleware Prometheus指标中间件
func MetricsMiddleware(config *MetricsConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	pathCounter := make(map[string]int)

	return func(c *gin.Context) {
		// 检查是否跳过此路径
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 标准化路径
		normalizedPath := path
		if config.NormalizePath {
			normalizedPath = normalizePath(path)
		}

		// 检查路径标签数量限制
		if config.MaxPathLabels > 0 {
			if count, exists := pathCounter[normalizedPath]; exists {
				if count > config.MaxPathLabels {
					normalizedPath = "other"
				}
			} else {
				pathCounter[normalizedPath] = 1
			}
		}

		// 增加活跃连接数
		httpActiveConnections.Inc()
		defer httpActiveConnections.Dec()

		// 记录请求大小
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(
				c.Request.Method,
				normalizedPath,
			).Observe(float64(c.Request.ContentLength))
		}

		// 记录开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算持续时间
		duration := time.Since(startTime)
		status := strconv.Itoa(c.Writer.Status())

		// 记录指标
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			normalizedPath,
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			normalizedPath,
			status,
		).Observe(duration.Seconds())

		// 记录响应大小
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(
				c.Request.Method,
				normalizedPath,
				status,
			).Observe(float64(responseSize))
		}
	}
}

// normalizePath 标准化路径，将参数替换为占位符
func normalizePath(path string) string {
	// 简单的路径标准化，可以根据需要扩展
	// 例如：/api/v1/users/123 -> /api/v1/users/:id
	
	// 这里可以实现更复杂的路径标准化逻辑
	// 目前返回原路径
	return path
}

// BusinessMetrics 业务指标记录器
type BusinessMetrics struct{}

// NewBusinessMetrics 创建业务指标记录器
func NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{}
}

// RecordUserRegistration 记录用户注册
func (m *BusinessMetrics) RecordUserRegistration(source string) {
	userRegistrations.WithLabelValues(source).Inc()
}

// RecordUserLogin 记录用户登录
func (m *BusinessMetrics) RecordUserLogin(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	userLogins.WithLabelValues(status).Inc()
}

// RecordPredictionCreated 记录预测创建
func (m *BusinessMetrics) RecordPredictionCreated(matchType string) {
	predictionsCreated.WithLabelValues(matchType).Inc()
}

// RecordVoteCreated 记录投票创建
func (m *BusinessMetrics) RecordVoteCreated(predictionType string) {
	votesCreated.WithLabelValues(predictionType).Inc()
}

// RecordCacheHit 记录缓存命中
func (m *BusinessMetrics) RecordCacheHit(cacheType string) {
	cacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss 记录缓存未命中
func (m *BusinessMetrics) RecordCacheMiss(cacheType string) {
	cacheMisses.WithLabelValues(cacheType).Inc()
}

// RecordDatabaseConnections 记录数据库连接数
func (m *BusinessMetrics) RecordDatabaseConnections(open, idle, inUse int) {
	databaseConnections.WithLabelValues("open").Set(float64(open))
	databaseConnections.WithLabelValues("idle").Set(float64(idle))
	databaseConnections.WithLabelValues("in_use").Set(float64(inUse))
}

// RecordDatabaseQuery 记录数据库查询
func (m *BusinessMetrics) RecordDatabaseQuery(operation, table string, duration time.Duration) {
	databaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordRedisConnections 记录Redis连接数
func (m *BusinessMetrics) RecordRedisConnections(active, idle int) {
	redisConnections.WithLabelValues("active").Set(float64(active))
	redisConnections.WithLabelValues("idle").Set(float64(idle))
}

// RecordRedisOperation 记录Redis操作
func (m *BusinessMetrics) RecordRedisOperation(operation string, duration time.Duration) {
	redisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// WebSocket metrics methods removed

// GetBusinessMetrics 获取全局业务指标记录器
var globalBusinessMetrics = NewBusinessMetrics()

func GetBusinessMetrics() *BusinessMetrics {
	return globalBusinessMetrics
}

// CustomMetrics 自定义指标
type CustomMetrics struct {
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
}

// NewCustomMetrics 创建自定义指标
func NewCustomMetrics() *CustomMetrics {
	return &CustomMetrics{
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

// RegisterCounter 注册计数器
func (m *CustomMetrics) RegisterCounter(name, help string, labels []string) {
	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.counters[name] = counter
}

// RegisterGauge 注册仪表盘
func (m *CustomMetrics) RegisterGauge(name, help string, labels []string) {
	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.gauges[name] = gauge
}

// RegisterHistogram 注册直方图
func (m *CustomMetrics) RegisterHistogram(name, help string, labels []string, buckets []float64) {
	if buckets == nil {
		buckets = prometheus.DefBuckets
	}
	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	m.histograms[name] = histogram
}

// IncCounter 增加计数器
func (m *CustomMetrics) IncCounter(name string, labelValues ...string) {
	if counter, exists := m.counters[name]; exists {
		counter.WithLabelValues(labelValues...).Inc()
	}
}

// SetGauge 设置仪表盘值
func (m *CustomMetrics) SetGauge(name string, value float64, labelValues ...string) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Set(value)
	}
}

// ObserveHistogram 观察直方图
func (m *CustomMetrics) ObserveHistogram(name string, value float64, labelValues ...string) {
	if histogram, exists := m.histograms[name]; exists {
		histogram.WithLabelValues(labelValues...).Observe(value)
	}
}