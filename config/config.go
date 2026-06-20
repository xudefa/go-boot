package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config 配置接口
type Config interface {
	Get(key string) any
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetAll() map[string]any
	Set(key string, value any)
	Load(path string) error
	Save(path string) error
}

// memoryConfig 内存配置实现
type memoryConfig struct {
	data map[string]any
	mu   sync.RWMutex
}

// NewConfig 创建配置
func NewConfig() Config {
	return &memoryConfig{
		data: make(map[string]any),
	}
}

// Get 获取配置值
func (c *memoryConfig) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// GetString 获取字符串值
func (c *memoryConfig) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt 获取整数值
func (c *memoryConfig) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool 获取布尔值
func (c *memoryConfig) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetAll 获取所有配置值
func (c *memoryConfig) GetAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]any, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Set 设置配置值
func (c *memoryConfig) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Load 从文件加载配置
func (c *memoryConfig) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return json.Unmarshal(data, &c.data)
}

// Save 保存配置到文件
func (c *memoryConfig) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
