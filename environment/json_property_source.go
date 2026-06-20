// Package environment 提供 JSON 配置源实现和配置文件查找功能。
//
// 核心组件：
//   - JSONPropertySource: 基于 JSON 文件的配置源
//   - FindApplicationConfigFile: 应用配置文件查找
package environment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JSONPropertySource 基于 JSON 文件的配置源
//
// 从 JSON 文件加载配置，支持嵌套结构。
// 使用 PriorityLowest 优先级，作为默认配置源。
type JSONPropertySource struct {
	name     string
	data     map[string]any
	priority Priority
	filePath string
}

// NewJSONPropertySource 从 JSON 文件创建配置源
//
// filePath 是 JSON 文件的完整路径。
// 如果文件不存在或解析失败，返回空配置源。
func NewJSONPropertySource(name, filePath string) (*JSONPropertySource, error) {
	data, err := loadJSONFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON config from %s: %w", filePath, err)
	}

	return &JSONPropertySource{
		name:     name,
		data:     data,
		priority: PriorityLowest,
		filePath: filePath,
	}, nil
}

// NewJSONPropertySourceOrDefault 从 JSON 文件创建配置源，如果失败则返回空配置源
//
// filePath 是 JSON 文件的完整路径。
// 如果文件不存在或解析失败，返回空配置源但不报错。
func NewJSONPropertySourceOrDefault(name, filePath string) *JSONPropertySource {
	data, err := loadJSONFile(filePath)
	if err != nil {
		return &JSONPropertySource{
			name:     name,
			data:     make(map[string]any),
			priority: PriorityLowest,
			filePath: filePath,
		}
	}

	return &JSONPropertySource{
		name:     name,
		data:     data,
		priority: PriorityLowest,
		filePath: filePath,
	}
}

func (j *JSONPropertySource) Name() string {
	return j.name
}

func (j *JSONPropertySource) Priority() Priority {
	return j.priority
}

func (j *JSONPropertySource) Contains(key string) bool {
	_, ok := j.GetProperty(key)
	return ok
}

// GetProperty 从 JSON 数据中获取配置
//
// 支持点分隔的嵌套键名，如 "server.port"。
func (j *JSONPropertySource) GetProperty(key string) (any, bool) {
	if j.data == nil {
		return nil, false
	}

	parts := strings.Split(key, ".")
	var current any = j.data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// Keys 返回所有扁平化的键名
//
// 将嵌套结构转换为点分隔的键名，如 "server.port"。
func (j *JSONPropertySource) Keys() []string {
	if j.data == nil {
		return []string{}
	}
	return flattenKeys(j.data, "")
}

// loadJSONFile 从文件加载 JSON 数据
func loadJSONFile(filePath string) (map[string]any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("[go-boot] failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	var data map[string]any
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	if data == nil {
		return make(map[string]any), nil
	}

	return data, nil
}

// flattenKeys 将嵌套的 map 扁平化为点分隔的键名
func flattenKeys(data map[string]any, prefix string) []string {
	keys := make([]string, 0, len(data))
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			keys = append(keys, flattenKeys(v, fullKey)...)
		default:
			keys = append(keys, fullKey)
		}
	}
	return keys
}

// FindApplicationConfigFile 查找应用配置文件
//
// 在以下路径中查找 application.json：
//  1. 当前目录
//  2. ./config 目录
//  3. 可执行文件所在目录
//  4. 可执行文件所在目录的 ./config 子目录
func FindApplicationConfigFile() string {
	searchPaths := []string{
		"/etc/config",
		".",
		"./config",
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		searchPaths = append(searchPaths, exeDir, filepath.Join(exeDir, "config"), filepath.Join(exeDir, "src", "config"))
	}

	for _, path := range searchPaths {
		configFile := filepath.Join(path, "application.json")
		if _, err := os.Stat(configFile); err == nil {
			return configFile
		}
	}

	return ""
}

// FindDefaultConfigFile 查找默认配置文件（向后兼容）
//
// @deprecated 使用 FindApplicationConfigFile 代替
func FindDefaultConfigFile() string {
	return FindApplicationConfigFile()
}
