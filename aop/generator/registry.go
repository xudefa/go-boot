package generator

import (
	"encoding/json"
	"os"
	"sync"
)

// Registry 代理注册表
//
// 记录 Bean 标识与生成文件路径的映射关系，支持持久化到 JSON 文件。
type Registry struct {
	mu      sync.RWMutex      // 读写锁
	proxies map[string]string // Bean 标识 -> 代理文件路径映射
}

// NewRegistry 创建代理注册表
func NewRegistry() *Registry {
	return &Registry{
		proxies: make(map[string]string),
	}
}

// Register 注册代理映射
func (r *Registry) Register(beanID, filePath string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proxies[beanID] = filePath
}

// Get 根据 Bean 标识获取代理文件路径
func (r *Registry) Get(beanID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	path, exists := r.proxies[beanID]
	return path, exists
}

// List 获取所有代理映射的副本
func (r *Registry) List() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range r.proxies {
		result[k] = v
	}
	return result
}

// Save 将注册表持久化到 JSON 文件
func (r *Registry) Save(filePath string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.MarshalIndent(r.proxies, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Load 从 JSON 文件加载注册表
func (r *Registry) Load(filePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &r.proxies)
}

// Clear 清空注册表
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proxies = make(map[string]string)
}
