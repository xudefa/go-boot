// Package environment 提供 Profile 相关的辅助函数。
//
// 支持通过命令行参数（--profile=xxx）或环境变量（GO_BOOT_PROFILE）
// 设置激活的 Profile。
package environment

import (
	"os"
	"strings"
)

// GetProfileActive 从命令行参数或环境变量中获取激活的 Profile
//
// 优先级：命令行参数 --profile=xxx > 环境变量 GO_BOOT_PROFILE。
// 命令行参数格式：--profile=dev
func GetProfileActive(args []string) string {
	for _, arg := range args {
		if after, ok := strings.CutPrefix(arg, "--profile="); ok {
			return after
		}
	}
	return os.Getenv("GO_BOOT_PROFILE")
}

// ParseProfiles 解析逗号分隔的 Profile 字符串为切片
//
// 例如 "dev,test" -> ["dev", "test"]，会自动去除空格和空项。
func ParseProfiles(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
