package redis

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics Redis 性能指标
type Metrics struct {
	// 操作统计
	TotalOperations   int64 `json:"total_operations"`
	SuccessOperations int64 `json:"success_operations"`
	FailedOperations  int64 `json:"failed_operations"`

	// 缓存统计
	CacheHits   int64 `json:"cache_hits"`
	CacheMisses int64 `json:"cache_misses"`

	// 性能统计
	TotalLatency int64 `json:"total_latency_ns"`
	MinLatency   int64 `json:"min_latency_ns"`
	MaxLatency   int64 `json:"max_latency_ns"`

	// 操作类型统计
	operationStats   map[string]*OperationStats
	operationStatsMu sync.RWMutex

	// 时间窗口统计
	windowStats *WindowStats

	// 启动时间
	StartTime time.Time `json:"start_time"`
}

// OperationStats 操作统计
type OperationStats struct {
	Count        int64 `json:"count"`
	Errors       int64 `json:"errors"`
	TotalLatency int64 `json:"total_latency_ns"`
	MinLatency   int64 `json:"min_latency_ns"`
	MaxLatency   int64 `json:"max_latency_ns"`
}

// WindowStats 时间窗口统计
type WindowStats struct {
	mu           sync.RWMutex
	windows      []*TimeWindow
	windowSize   time.Duration
	maxWindows   int
	currentIndex int
}

// TimeWindow 时间窗口
type TimeWindow struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Operations  int64     `json:"operations"`
	Errors      int64     `json:"errors"`
	CacheHits   int64     `json:"cache_hits"`
	CacheMisses int64     `json:"cache_misses"`
	AvgLatency  float64   `json:"avg_latency_ms"`
}

// NewMetrics 创建新的指标实例
func NewMetrics() *Metrics {
	return &Metrics{
		operationStats: make(map[string]*OperationStats),
		windowStats:    NewWindowStats(time.Minute, 60), // 60个1分钟窗口
		StartTime:      time.Now(),
		MinLatency:     int64(^uint64(0) >> 1), // 最大int64值
	}
}

// NewWindowStats 创建时间窗口统计
func NewWindowStats(windowSize time.Duration, maxWindows int) *WindowStats {
	ws := &WindowStats{
		windowSize: windowSize,
		maxWindows: maxWindows,
		windows:    make([]*TimeWindow, maxWindows),
	}

	// 初始化第一个窗口
	now := time.Now()
	ws.windows[0] = &TimeWindow{
		StartTime: now,
		EndTime:   now.Add(windowSize),
	}

	return ws
}

// RecordOperation 记录操作
func (m *Metrics) RecordOperation(operation string, latency time.Duration, err error) {
	atomic.AddInt64(&m.TotalOperations, 1)

	latencyNs := latency.Nanoseconds()
	atomic.AddInt64(&m.TotalLatency, latencyNs)

	// 更新最小延迟
	for {
		current := atomic.LoadInt64(&m.MinLatency)
		if latencyNs >= current || atomic.CompareAndSwapInt64(&m.MinLatency, current, latencyNs) {
			break
		}
	}

	// 更新最大延迟
	for {
		current := atomic.LoadInt64(&m.MaxLatency)
		if latencyNs <= current || atomic.CompareAndSwapInt64(&m.MaxLatency, current, latencyNs) {
			break
		}
	}

	if err != nil {
		atomic.AddInt64(&m.FailedOperations, 1)
	} else {
		atomic.AddInt64(&m.SuccessOperations, 1)
	}

	// 记录操作类型统计
	m.recordOperationStats(operation, latency, err)

	// 记录时间窗口统计
	m.windowStats.RecordOperation(latency, err)
}

// RecordCacheHit 记录缓存命中
func (m *Metrics) RecordCacheHit(key string) {
	atomic.AddInt64(&m.CacheHits, 1)
	m.windowStats.RecordCacheHit()
}

// RecordCacheMiss 记录缓存未命中
func (m *Metrics) RecordCacheMiss(key string) {
	atomic.AddInt64(&m.CacheMisses, 1)
	m.windowStats.RecordCacheMiss()
}

// recordOperationStats 记录操作类型统计
func (m *Metrics) recordOperationStats(operation string, latency time.Duration, err error) {
	m.operationStatsMu.Lock()
	defer m.operationStatsMu.Unlock()

	stats, exists := m.operationStats[operation]
	if !exists {
		stats = &OperationStats{
			MinLatency: int64(^uint64(0) >> 1), // 最大int64值
		}
		m.operationStats[operation] = stats
	}

	latencyNs := latency.Nanoseconds()

	atomic.AddInt64(&stats.Count, 1)
	atomic.AddInt64(&stats.TotalLatency, latencyNs)

	if err != nil {
		atomic.AddInt64(&stats.Errors, 1)
	}

	// 更新最小延迟
	for {
		current := atomic.LoadInt64(&stats.MinLatency)
		if latencyNs >= current || atomic.CompareAndSwapInt64(&stats.MinLatency, current, latencyNs) {
			break
		}
	}

	// 更新最大延迟
	for {
		current := atomic.LoadInt64(&stats.MaxLatency)
		if latencyNs <= current || atomic.CompareAndSwapInt64(&stats.MaxLatency, current, latencyNs) {
			break
		}
	}
}

// GetSummary 获取指标摘要
func (m *Metrics) GetSummary() map[string]interface{} {
	totalOps := atomic.LoadInt64(&m.TotalOperations)
	successOps := atomic.LoadInt64(&m.SuccessOperations)
	failedOps := atomic.LoadInt64(&m.FailedOperations)
	cacheHits := atomic.LoadInt64(&m.CacheHits)
	cacheMisses := atomic.LoadInt64(&m.CacheMisses)
	totalLatency := atomic.LoadInt64(&m.TotalLatency)
	minLatency := atomic.LoadInt64(&m.MinLatency)
	maxLatency := atomic.LoadInt64(&m.MaxLatency)

	var avgLatency float64
	if totalOps > 0 {
		avgLatency = float64(totalLatency) / float64(totalOps) / 1e6 // 转换为毫秒
	}

	var successRate float64
	if totalOps > 0 {
		successRate = float64(successOps) / float64(totalOps) * 100
	}

	var cacheHitRate float64
	totalCacheOps := cacheHits + cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalCacheOps) * 100
	}

	uptime := time.Since(m.StartTime)

	summary := map[string]interface{}{
		"uptime_seconds":     uptime.Seconds(),
		"total_operations":   totalOps,
		"success_operations": successOps,
		"failed_operations":  failedOps,
		"success_rate":       successRate,
		"cache_hits":         cacheHits,
		"cache_misses":       cacheMisses,
		"cache_hit_rate":     cacheHitRate,
		"avg_latency_ms":     avgLatency,
		"min_latency_ms":     float64(minLatency) / 1e6,
		"max_latency_ms":     float64(maxLatency) / 1e6,
		"operations_per_sec": float64(totalOps) / uptime.Seconds(),
	}

	// 添加操作类型统计
	m.operationStatsMu.RLock()
	operationStats := make(map[string]interface{})
	for op, stats := range m.operationStats {
		count := atomic.LoadInt64(&stats.Count)
		errors := atomic.LoadInt64(&stats.Errors)
		totalLat := atomic.LoadInt64(&stats.TotalLatency)
		minLat := atomic.LoadInt64(&stats.MinLatency)
		maxLat := atomic.LoadInt64(&stats.MaxLatency)

		var avgLat float64
		if count > 0 {
			avgLat = float64(totalLat) / float64(count) / 1e6
		}

		var errorRate float64
		if count > 0 {
			errorRate = float64(errors) / float64(count) * 100
		}

		operationStats[op] = map[string]interface{}{
			"count":          count,
			"errors":         errors,
			"error_rate":     errorRate,
			"avg_latency_ms": avgLat,
			"min_latency_ms": float64(minLat) / 1e6,
			"max_latency_ms": float64(maxLat) / 1e6,
		}
	}
	m.operationStatsMu.RUnlock()

	summary["operations"] = operationStats

	// 添加时间窗口统计
	summary["recent_windows"] = m.windowStats.GetRecentWindows(10)

	return summary
}

// GetOperationStats 获取操作统计
func (m *Metrics) GetOperationStats() map[string]*OperationStats {
	m.operationStatsMu.RLock()
	defer m.operationStatsMu.RUnlock()

	result := make(map[string]*OperationStats)
	for op, stats := range m.operationStats {
		result[op] = &OperationStats{
			Count:        atomic.LoadInt64(&stats.Count),
			Errors:       atomic.LoadInt64(&stats.Errors),
			TotalLatency: atomic.LoadInt64(&stats.TotalLatency),
			MinLatency:   atomic.LoadInt64(&stats.MinLatency),
			MaxLatency:   atomic.LoadInt64(&stats.MaxLatency),
		}
	}

	return result
}

// Reset 重置指标
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.TotalOperations, 0)
	atomic.StoreInt64(&m.SuccessOperations, 0)
	atomic.StoreInt64(&m.FailedOperations, 0)
	atomic.StoreInt64(&m.CacheHits, 0)
	atomic.StoreInt64(&m.CacheMisses, 0)
	atomic.StoreInt64(&m.TotalLatency, 0)
	atomic.StoreInt64(&m.MinLatency, int64(^uint64(0)>>1))
	atomic.StoreInt64(&m.MaxLatency, 0)

	m.operationStatsMu.Lock()
	m.operationStats = make(map[string]*OperationStats)
	m.operationStatsMu.Unlock()

	m.windowStats = NewWindowStats(time.Minute, 60)
	m.StartTime = time.Now()
}

// WindowStats 方法实现

// RecordOperation 记录操作到时间窗口
func (ws *WindowStats) RecordOperation(latency time.Duration, err error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	window := ws.getCurrentWindow()
	atomic.AddInt64(&window.Operations, 1)

	if err != nil {
		atomic.AddInt64(&window.Errors, 1)
	}

	// 更新平均延迟（简化计算）
	currentAvg := window.AvgLatency
	ops := atomic.LoadInt64(&window.Operations)
	newLatencyMs := float64(latency.Nanoseconds()) / 1e6
	window.AvgLatency = (currentAvg*float64(ops-1) + newLatencyMs) / float64(ops)
}

// RecordCacheHit 记录缓存命中到时间窗口
func (ws *WindowStats) RecordCacheHit() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	window := ws.getCurrentWindow()
	atomic.AddInt64(&window.CacheHits, 1)
}

// RecordCacheMiss 记录缓存未命中到时间窗口
func (ws *WindowStats) RecordCacheMiss() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	window := ws.getCurrentWindow()
	atomic.AddInt64(&window.CacheMisses, 1)
}

// getCurrentWindow 获取当前时间窗口
func (ws *WindowStats) getCurrentWindow() *TimeWindow {
	now := time.Now()
	currentWindow := ws.windows[ws.currentIndex]

	// 检查是否需要创建新窗口
	if currentWindow == nil || now.After(currentWindow.EndTime) {
		ws.currentIndex = (ws.currentIndex + 1) % ws.maxWindows
		ws.windows[ws.currentIndex] = &TimeWindow{
			StartTime: now.Truncate(ws.windowSize),
			EndTime:   now.Truncate(ws.windowSize).Add(ws.windowSize),
		}
		currentWindow = ws.windows[ws.currentIndex]
	}

	return currentWindow
}

// GetRecentWindows 获取最近的时间窗口
func (ws *WindowStats) GetRecentWindows(count int) []*TimeWindow {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if count > ws.maxWindows {
		count = ws.maxWindows
	}

	result := make([]*TimeWindow, 0, count)

	for i := 0; i < count; i++ {
		index := (ws.currentIndex - i + ws.maxWindows) % ws.maxWindows
		if ws.windows[index] != nil {
			// 创建副本避免并发问题
			window := &TimeWindow{
				StartTime:   ws.windows[index].StartTime,
				EndTime:     ws.windows[index].EndTime,
				Operations:  atomic.LoadInt64(&ws.windows[index].Operations),
				Errors:      atomic.LoadInt64(&ws.windows[index].Errors),
				CacheHits:   atomic.LoadInt64(&ws.windows[index].CacheHits),
				CacheMisses: atomic.LoadInt64(&ws.windows[index].CacheMisses),
				AvgLatency:  ws.windows[index].AvgLatency,
			}
			result = append(result, window)
		}
	}

	return result
}

// GetCurrentQPS 获取当前 QPS
func (m *Metrics) GetCurrentQPS() float64 {
	windows := m.windowStats.GetRecentWindows(1)
	if len(windows) == 0 {
		return 0
	}

	window := windows[0]
	duration := window.EndTime.Sub(window.StartTime).Seconds()
	if duration > 0 {
		return float64(window.Operations) / duration
	}

	return 0
}

// GetCurrentErrorRate 获取当前错误率
func (m *Metrics) GetCurrentErrorRate() float64 {
	windows := m.windowStats.GetRecentWindows(1)
	if len(windows) == 0 {
		return 0
	}

	window := windows[0]
	if window.Operations > 0 {
		return float64(window.Errors) / float64(window.Operations) * 100
	}

	return 0
}

// GetCurrentCacheHitRate 获取当前缓存命中率
func (m *Metrics) GetCurrentCacheHitRate() float64 {
	windows := m.windowStats.GetRecentWindows(1)
	if len(windows) == 0 {
		return 0
	}

	window := windows[0]
	totalCacheOps := window.CacheHits + window.CacheMisses
	if totalCacheOps > 0 {
		return float64(window.CacheHits) / float64(totalCacheOps) * 100
	}

	return 0
}

// IsHealthy 检查指标是否健康
func (m *Metrics) IsHealthy() bool {
	// 检查错误率是否过高（超过5%）
	if m.GetCurrentErrorRate() > 5.0 {
		return false
	}

	// 检查缓存命中率是否过低（低于50%）
	if m.GetCurrentCacheHitRate() < 50.0 && atomic.LoadInt64(&m.CacheHits)+atomic.LoadInt64(&m.CacheMisses) > 100 {
		return false
	}

	return true
}

// GetHealthScore 获取健康评分 (0-100)
func (m *Metrics) GetHealthScore() float64 {
	score := 100.0

	// 错误率影响 (最多扣30分)
	errorRate := m.GetCurrentErrorRate()
	if errorRate > 0 {
		score -= errorRate * 6 // 每1%错误率扣6分
		if score < 70 {
			score = 70
		}
	}

	// 缓存命中率影响 (最多扣20分)
	cacheHitRate := m.GetCurrentCacheHitRate()
	totalCacheOps := atomic.LoadInt64(&m.CacheHits) + atomic.LoadInt64(&m.CacheMisses)
	if totalCacheOps > 100 { // 只有在有足够样本时才考虑缓存命中率
		if cacheHitRate < 90 {
			score -= (90 - cacheHitRate) * 0.5 // 每低于90%的1%扣0.5分
		}
	}

	// 延迟影响 (最多扣10分)
	avgLatency := float64(atomic.LoadInt64(&m.TotalLatency)) / float64(atomic.LoadInt64(&m.TotalOperations)) / 1e6
	if avgLatency > 10 { // 超过10ms开始扣分
		score -= (avgLatency - 10) * 0.5 // 每超过1ms扣0.5分
		if score < 90 {
			score = 90
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}
