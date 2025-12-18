package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	mu       sync.RWMutex
	config   *Config
	viper    *viper.Viper
	watchers []ConfigWatcher
	watcher  *fsnotify.Watcher
}

// ConfigWatcher 配置变更监听器
type ConfigWatcher interface {
	OnConfigChange(oldConfig, newConfig *Config) error
}

// ConfigWatcherFunc 配置变更监听器函数
type ConfigWatcherFunc func(oldConfig, newConfig *Config) error

// OnConfigChange 实现 ConfigWatcher 接口
func (f ConfigWatcherFunc) OnConfigChange(oldConfig, newConfig *Config) error {
	return f(oldConfig, newConfig)
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		viper:    viper.New(),
		watchers: make([]ConfigWatcher, 0),
	}
}

// LoadConfig 加载配置
func (cl *ConfigLoader) LoadConfig(opts *LoadOptions) (*Config, error) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if opts == nil {
		opts = DefaultLoadOptions()
	}

	// 配置 viper
	cl.setupViper(opts)

	// 读取配置文件
	if err := cl.readConfigFile(); err != nil {
		return nil, err
	}

	// 解析配置
	config, err := cl.parseConfig()
	if err != nil {
		return nil, err
	}

	// 验证配置
	if !opts.SkipValidate {
		if err := validateConfig(config); err != nil {
			return nil, err
		}
	}

	cl.config = config
	return config, nil
}

// setupViper 设置 viper
func (cl *ConfigLoader) setupViper(opts *LoadOptions) {
	// 设置配置文件信息
	cl.viper.SetConfigName(opts.ConfigName)
	cl.viper.SetConfigType(opts.ConfigType)

	// 添加配置文件搜索路径
	if opts.ConfigPath != "" {
		cl.viper.AddConfigPath(opts.ConfigPath)
	} else {
		cl.addDefaultConfigPaths()
	}

	// 设置环境变量
	cl.viper.SetEnvPrefix(opts.EnvPrefix)
	cl.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	cl.viper.AutomaticEnv()

	// 设置默认值
	setDefaults(cl.viper)
}

// addDefaultConfigPaths 添加默认配置路径
func (cl *ConfigLoader) addDefaultConfigPaths() {
	paths := []string{
		".",
		"./configs",
		"./config",
		"/etc/backend-go",
	}

	// 添加用户主目录
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".backend-go"))
	}

	// 添加可执行文件目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		paths = append(paths, execDir)
		paths = append(paths, filepath.Join(execDir, "configs"))
	}

	for _, path := range paths {
		cl.viper.AddConfigPath(path)
	}
}

// readConfigFile 读取配置文件
func (cl *ConfigLoader) readConfigFile() error {
	if err := cl.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认值和环境变量
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}
	return nil
}

// parseConfig 解析配置
func (cl *ConfigLoader) parseConfig() (*Config, error) {
	var config Config
	if err := cl.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 后处理配置
	if err := postProcessConfig(&config); err != nil {
		return nil, fmt.Errorf("failed to post-process config: %w", err)
	}

	return &config, nil
}

// GetConfig 获取当前配置
func (cl *ConfigLoader) GetConfig() *Config {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.config
}

// ReloadConfig 重新加载配置
func (cl *ConfigLoader) ReloadConfig() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	oldConfig := cl.config

	// 重新读取配置文件
	if err := cl.readConfigFile(); err != nil {
		return err
	}

	// 解析新配置
	newConfig, err := cl.parseConfig()
	if err != nil {
		return err
	}

	// 验证新配置
	if err := validateConfig(newConfig); err != nil {
		return err
	}

	// 通知监听器
	for _, watcher := range cl.watchers {
		if err := watcher.OnConfigChange(oldConfig, newConfig); err != nil {
			return fmt.Errorf("config watcher error: %w", err)
		}
	}

	cl.config = newConfig
	return nil
}

// AddWatcher 添加配置变更监听器
func (cl *ConfigLoader) AddWatcher(watcher ConfigWatcher) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.watchers = append(cl.watchers, watcher)
}

// RemoveWatcher 移除配置变更监听器
func (cl *ConfigLoader) RemoveWatcher(watcher ConfigWatcher) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for i, w := range cl.watchers {
		if w == watcher {
			cl.watchers = append(cl.watchers[:i], cl.watchers[i+1:]...)
			break
		}
	}
}

// WatchConfig 监听配置文件变更
func (cl *ConfigLoader) WatchConfig() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.watcher != nil {
		return fmt.Errorf("config watcher already started")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	cl.watcher = watcher

	// 获取配置文件路径
	configFile := cl.viper.ConfigFileUsed()
	if configFile == "" {
		return fmt.Errorf("no config file in use")
	}

	// 监听配置文件
	if err := watcher.Add(configFile); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	// 启动监听协程
	go cl.watchLoop()

	return nil
}

// StopWatching 停止监听配置文件变更
func (cl *ConfigLoader) StopWatching() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.watcher == nil {
		return nil
	}

	err := cl.watcher.Close()
	cl.watcher = nil
	return err
}

// watchLoop 监听循环
func (cl *ConfigLoader) watchLoop() {
	for {
		select {
		case event, ok := <-cl.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入和创建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if err := cl.ReloadConfig(); err != nil {
					// 这里可以记录日志
					fmt.Printf("Failed to reload config: %v\n", err)
				}
			}

		case err, ok := <-cl.watcher.Errors:
			if !ok {
				return
			}
			// 这里可以记录日志
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}
}

// GetConfigValue 获取配置值
func (cl *ConfigLoader) GetConfigValue(key string) interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.viper.Get(key)
}

// SetConfigValue 设置配置值
func (cl *ConfigLoader) SetConfigValue(key string, value interface{}) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.viper.Set(key, value)
}

// IsSet 检查配置键是否已设置
func (cl *ConfigLoader) IsSet(key string) bool {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.viper.IsSet(key)
}

// GetConfigFile 获取当前使用的配置文件路径
func (cl *ConfigLoader) GetConfigFile() string {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.viper.ConfigFileUsed()
}

// WriteConfig 写入配置到文件
func (cl *ConfigLoader) WriteConfig() error {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.viper.WriteConfig()
}

// WriteConfigAs 写入配置到指定文件
func (cl *ConfigLoader) WriteConfigAs(filename string) error {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.viper.WriteConfigAs(filename)
}

// MergeConfig 合并配置
func (cl *ConfigLoader) MergeConfig(configMap map[string]interface{}) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for key, value := range configMap {
		cl.viper.Set(key, value)
	}

	// 重新解析配置
	newConfig, err := cl.parseConfig()
	if err != nil {
		return err
	}

	// 验证新配置
	if err := validateConfig(newConfig); err != nil {
		return err
	}

	oldConfig := cl.config
	cl.config = newConfig

	// 通知监听器
	for _, watcher := range cl.watchers {
		if err := watcher.OnConfigChange(oldConfig, newConfig); err != nil {
			return fmt.Errorf("config watcher error: %w", err)
		}
	}

	return nil
}

// GetAllSettings 获取所有配置设置
func (cl *ConfigLoader) GetAllSettings() map[string]interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.viper.AllSettings()
}

// ConfigSnapshot 配置快照
type ConfigSnapshot struct {
	Config    *Config                `json:"config"`
	File      string                 `json:"file"`
	Settings  map[string]interface{} `json:"settings"`
	Timestamp int64                  `json:"timestamp"`
}

// CreateSnapshot 创建配置快照
func (cl *ConfigLoader) CreateSnapshot() *ConfigSnapshot {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	return &ConfigSnapshot{
		Config:    cl.config,
		File:      cl.viper.ConfigFileUsed(),
		Settings:  cl.viper.AllSettings(),
		Timestamp: time.Now().Unix(),
	}
}

// RestoreFromSnapshot 从快照恢复配置
func (cl *ConfigLoader) RestoreFromSnapshot(snapshot *ConfigSnapshot) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	oldConfig := cl.config

	// 验证快照配置
	if err := validateConfig(snapshot.Config); err != nil {
		return err
	}

	// 恢复配置
	cl.config = snapshot.Config

	// 通知监听器
	for _, watcher := range cl.watchers {
		if err := watcher.OnConfigChange(oldConfig, snapshot.Config); err != nil {
			return fmt.Errorf("config watcher error: %w", err)
		}
	}

	return nil
}
