// Package environment 提供配置文件加载器，支持基础配置和 Profile 配置的分层加载。
//
// ConfigLoader 负责在预设的搜索路径中查找并加载配置文件，
// 支持 application-{profile}.json 形式的 Profile 配置覆盖。
package environment

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigType 配置文件类型枚举
type ConfigType string

const (
	// ConfigTypeJSON JSON 配置文件类型
	ConfigTypeJSON ConfigType = "json"
)

// ConfigLoader 配置文件加载器
//
// 在预设搜索路径中查找配置文件，支持基础配置和 Profile 配置的分层加载。
// Profile 配置会覆盖基础配置中的同名属性。
//
// 加载流程：
//  1. 如果指定了 configLocation，直接加载该文件
//  2. 否则搜索 configName.{configType} 作为基础配置
//  3. 依次搜索 configName-{profile}.{configType} 作为 Profile 配置
type ConfigLoader struct {
	configName     string     // 配置文件名（不含扩展名）
	configType     ConfigType // 配置文件类型
	configLocation string     // 自定义配置文件路径（优先级最高）
	profiles       []string   // 激活的 Profile 列表
}

// NewConfigLoader 创建配置文件加载器
//
// 参数：
//   - configName: 配置文件名（不含扩展名），如 "application"
//   - configType: 配置文件类型，如 ConfigTypeJSON
//   - configLocation: 自定义配置文件路径，为空时自动搜索
//   - profiles: 激活的 Profile 列表
func NewConfigLoader(configName string, configType ConfigType, configLocation string, profiles []string) *ConfigLoader {
	return &ConfigLoader{
		configName:     configName,
		configType:     configType,
		configLocation: configLocation,
		profiles:       profiles,
	}
}

// Load 加载配置文件并返回配置源列表
//
// 返回的配置源按加载顺序排列：基础配置在前，Profile 配置在后。
// Profile 配置的优先级高于基础配置。
func (l *ConfigLoader) Load() ([]PropertySource, error) {
	var sources []PropertySource

	if l.configLocation != "" {
		source, err := l.loadConfigFile(l.configLocation, "custom-config")
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
		return sources, nil
	}

	baseConfigPath, err := l.findConfigFile(l.configName)
	if err != nil {
		return nil, err
	}

	if baseConfigPath != "" {
		baseSource, err := l.loadConfigFile(baseConfigPath, "base-config")
		if err != nil {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
		sources = append(sources, baseSource)
	}

	for _, profile := range l.profiles {
		profileConfigName := fmt.Sprintf("%s-%s", l.configName, profile)
		profileConfigPath, err := l.findConfigFile(profileConfigName)
		if err != nil {
			return nil, err
		}

		if profileConfigPath != "" {
			profileSource, err := l.loadConfigFile(profileConfigPath, fmt.Sprintf("profile-config-%s", profile))
			if err != nil {
				return nil, fmt.Errorf("failed to load profile config %s: %w", profile, err)
			}
			sources = append(sources, profileSource)
		}
	}

	return sources, nil
}

// findConfigFile 在预设搜索路径中查找配置文件
//
// 搜索路径优先级：
//  1. /etc/config
//  2. 当前目录
//  3. ./config
//  4. 可执行文件所在目录及其 ./config 子目录
func (l *ConfigLoader) findConfigFile(configName string) (string, error) {
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
		configFile := filepath.Join(path, fmt.Sprintf("%s.%s", configName, l.configType))
		if _, err := os.Stat(configFile); err == nil {
			return configFile, nil
		}
	}

	return "", nil
}

// loadConfigFile 加载单个配置文件为 PropertySource
func (l *ConfigLoader) loadConfigFile(filePath, sourceName string) (PropertySource, error) {
	return NewJSONPropertySource(sourceName, filePath)
}

// GetConfigFileExtension 根据配置类型返回文件扩展名
func GetConfigFileExtension(configType ConfigType) string {
	return string(configType)
}

// ParseConfigType 根据文件路径解析配置类型（当前仅支持 JSON）
func ParseConfigType(filePath string) ConfigType {
	return ConfigTypeJSON
}
