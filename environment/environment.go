// Package environment 提供分层配置源管理和 Profile 机制。
//
// 参考 Spring 的 Environment 抽象，支持从命令行参数、环境变量、配置文件等
// 多种来源读取配置，按优先级覆盖。支持 ${...} 占位符解析、类型安全的属性
// 获取和结构体绑定。
//
// 配置源优先级（从高到低）：
//  1. 命令行参数（--key=value）
//  2. 环境变量（GO_BOOT_ 前缀）
//  3. 用户添加的配置源
package environment

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/xudefa/go-boot/refresh"
)

// Environment 环境配置管理器
//
// 参考 Spring 的 Environment 抽象，提供分层配置源管理和 Profile 机制。
// 配置源按优先级排序，高优先级覆盖低优先级。
// 支持命令行参数（最高优先级）> 环境变量 > 配置文件（最低优先级）。
type Environment struct {
	mu              sync.RWMutex                      // 保护所有字段的读写锁
	sources         []PropertySource                  // 配置源列表，按优先级升序排列
	activeProfiles  []string                          // 当前激活的 Profile 列表
	configListeners []func(refresh.ConfigChangeEvent) // 配置变更监听器
}

// NewEnvironment 创建环境配置管理器
//
// 默认配置源优先级（从高到低）：
//  1. 命令行参数（--key=value）
//  2. 环境变量（GO_BOOT_ 前缀）
//  3. 应用 JSON 配置文件（application.json）
func NewEnvironment() *Environment {
	args := NewArgsPropertySource("args", os.Args)
	envSource := NewEnvPropertySource("env", "GO_BOOT")

	sources := make([]PropertySource, 2, 5)
	sources[0] = args
	sources[1] = envSource

	// 尝试加载应用 JSON 配置文件
	if applicationConfigFile := FindApplicationConfigFile(); applicationConfigFile != "" {
		if jsonSource := NewJSONPropertySourceOrDefault("application-config", applicationConfigFile); jsonSource != nil {
			sources = append(sources, jsonSource)
		}
	}

	environment := &Environment{
		sources: sources,
	}

	// 排序配置源以确保正确的优先级顺序
	environment.sortSources()

	return environment
}

// getRawProperty 从所有配置源中获取属性原始值（不解析占位符）
func (e *Environment) getRawProperty(key string) (any, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for i := len(e.sources) - 1; i >= 0; i-- {
		if val, ok := e.sources[i].GetProperty(key); ok {
			return val, true
		}
	}
	return nil, false
}

// GetProperty 从所有配置源中获取属性值
//
// 遍历配置源（从高优先级到低优先级），返回第一个匹配的值。
// 如果值是字符串且包含 ${...} 占位符，自动递归解析。
// 高优先级的配置源会覆盖低优先级的同名属性。
func (e *Environment) GetProperty(key string) (any, bool) {
	val, ok := e.getRawProperty(key)
	if !ok {
		return nil, false
	}
	if s, ok := val.(string); ok {
		return e.resolvePlaceholders(s, make(map[string]bool)), true
	}
	return val, true
}

// GetString 获取字符串类型的属性值.
//
// 参数:
//   - key: 属性键名
//   - defaultVal: 属性不存在时返回的默认值
//
// 返回:
//   - string: 属性值,不存在时返回 defaultVal
func (e *Environment) GetString(key, defaultVal string) string {
	if val, ok := e.GetProperty(key); ok {
		converted, err := globalTypeConverter.ConvertTo(val, reflect.TypeOf(""))
		if err == nil {
			return converted.String()
		}
	}
	return defaultVal
}

// GetInt 获取整数类型的属性值.
//
// 支持 int、float64 和字符串形式的整数值.
//
// 参数:
//   - key: 属性键名
//   - defaultVal: 属性不存在时返回的默认值
//
// 返回:
//   - int: 属性值,不存在时返回 defaultVal
func (e *Environment) GetInt(key string, defaultVal int) int {
	if val, ok := e.GetProperty(key); ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return defaultVal
}

// GetBool 获取布尔类型的属性值.
//
// 支持 bool 和字符串 "true"/"false" 形式.
//
// 参数:
//   - key: 属性键名
//   - defaultVal: 属性不存在时返回的默认值
//
// 返回:
//   - bool: 属性值,不存在时返回 defaultVal
func (e *Environment) GetBool(key string, defaultVal bool) bool {
	if val, ok := e.GetProperty(key); ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		}
	}
	return defaultVal
}

// ContainsProperty 检查属性是否存在.
//
// 参数:
//   - key: 属性键名
//
// 返回:
//   - bool: 是否存在
func (e *Environment) ContainsProperty(key string) bool {
	_, ok := e.GetProperty(key)
	return ok
}

// GetRequiredProperty 获取必需属性，不存在时返回错误
func (e *Environment) GetRequiredProperty(key string) (any, error) {
	val, ok := e.GetProperty(key)
	if !ok {
		return nil, fmt.Errorf("required property not found: %s", key)
	}
	return val, nil
}

// GetFloat64 获取 float64 类型属性
func (e *Environment) GetFloat64(key string, defaultVal float64) float64 {
	if val, ok := e.GetProperty(key); ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultVal
}

// IsPropertyEmpty 检查属性是否为空（不存在或空字符串）
func (e *Environment) IsPropertyEmpty(key string) bool {
	val, ok := e.GetProperty(key)
	if !ok {
		return true
	}
	s, ok := val.(string)
	if !ok {
		return false
	}
	return s == ""
}

// RemovePropertySource 按名称移除配置源
func (e *Environment) RemovePropertySource(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, src := range e.sources {
		if src.Name() == name {
			e.sources = append(e.sources[:i], e.sources[i+1:]...)
			return
		}
	}
}

// RemoveProfile 移除激活的 Profile
func (e *Environment) RemoveProfile(profile string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, p := range e.activeProfiles {
		if p == profile {
			e.activeProfiles = append(e.activeProfiles[:i], e.activeProfiles[i+1:]...)
			return
		}
	}
}

// Bind 将整个环境配置绑定到目标结构体
func (e *Environment) Bind(target any) error {
	val := reflectValueOf(target)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	return e.bindStruct(val.Elem(), "")
}

// BindKey 将指定键的值绑定到目标
func (e *Environment) BindKey(key string, target any) error {
	val := reflectValueOf(target)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	v, ok := e.GetProperty(key)
	if !ok {
		return fmt.Errorf("property not found: %s", key)
	}
	val.Elem().Set(reflectValueOf(v))
	return nil
}

// BindPrefix 将指定前缀的配置绑定到目标结构体
func (e *Environment) BindPrefix(prefix string, target any) error {
	val := reflectValueOf(target)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	return e.bindStruct(val.Elem(), prefix+".")
}

// Validate 验证配置，返回所有错误
func (e *Environment) Validate() []error {
	var errs []error
	for _, src := range e.GetPropertySources() {
		if v, ok := src.(interface{ Validate() []error }); ok {
			errs = append(errs, v.Validate()...)
		}
	}
	return errs
}

// ResolvePlaceholders 解析 ${...} 占位符（公有 API）
//
// 支持语法：
//   - ${key} — 引用配置参数 key
//   - ${key:defaultValue} — 引用 key，不存在时使用 defaultValue
func (e *Environment) ResolvePlaceholders(val string) string {
	return e.resolvePlaceholders(val, make(map[string]bool))
}

// parsePlaceholder 解析 ${...} 占位符内容
//
// 从 val[startIdx] 处的 '$' 开始解析，找到匹配的 '}'，
// 提取占位符内的 key 和 defaultValue。
//
// 参数:
//   - val: 原始字符串
//   - startIdx: '$' 字符的位置
//
// 返回:
//   - key: 占位符中的配置键名
//   - defaultValue: 冒号后的默认值（无默认值时返回空字符串）
//   - hasDefault: 是否存在默认值
//   - endIdx: 匹配的 '}' 的位置
//   - ok: 是否成功解析
func (e *Environment) parsePlaceholder(val string, startIdx int) (key, defaultValue string, hasDefault bool, endIdx int, ok bool) {
	depth := 1
	j := startIdx + 2
	for j < len(val) && depth > 0 {
		if val[j] == '$' && j+1 < len(val) && val[j+1] == '{' {
			depth++
			j += 2
			continue
		}
		if val[j] == '}' {
			depth--
			if depth == 0 {
				break
			}
		}
		j++
	}
	if depth != 0 {
		return "", "", false, 0, false
	}

	inner := val[startIdx+2 : j]

	keyEnd := -1
	nd := 0
	for k := 0; k < len(inner); k++ {
		if inner[k] == '$' && k+1 < len(inner) && inner[k+1] == '{' {
			nd++
			k++
			continue
		}
		if inner[k] == '}' {
			if nd > 0 {
				nd--
			}
			continue
		}
		if inner[k] == ':' && nd == 0 {
			keyEnd = k
			break
		}
	}

	if keyEnd >= 0 {
		return inner[:keyEnd], inner[keyEnd+1:], true, j, true
	}
	return inner, "", false, j, true
}

// resolvePlaceholders 内部占位符解析，带循环检测
//
// 支持语法：
//   - ${key} — 引用配置参数 key
//   - ${key:defaultValue} — 引用 key，不存在时使用 defaultValue
//     defaultValue 中支持嵌套 ${...} 占位符
func (e *Environment) resolvePlaceholders(val string, resolving map[string]bool) string {
	var result strings.Builder
	i := 0
	for i < len(val) {
		if val[i] == '$' && i+1 < len(val) && val[i+1] == '{' {
			key, defaultVal, hasDefault, j, ok := e.parsePlaceholder(val, i)
			if !ok {
				result.WriteByte(val[i])
				i++
				continue
			}

			if resolving[key] {
				slog.Warn("[environment] circular placeholder reference", "key", key)
				result.WriteString(val[i : j+1])
				i = j + 1
				continue
			}

			var replacement string
			if rawVal, rawOk := e.getRawProperty(key); rawOk {
				resolving[key] = true
				if s, ok := rawVal.(string); ok {
					replacement = e.resolvePlaceholders(s, resolving)
				} else {
					replacement = e.resolvePlaceholders(fmt.Sprintf("%v", rawVal), resolving)
				}
				delete(resolving, key)
			} else if hasDefault {
				resolving[key] = true
				replacement = e.resolvePlaceholders(defaultVal, resolving)
				delete(resolving, key)
			} else {
				replacement = val[i : j+1]
			}

			result.WriteString(replacement)
			i = j + 1
			continue
		}
		result.WriteByte(val[i])
		i++
	}
	return result.String()
}

func (e *Environment) bindStruct(val reflect.Value, prefix string) error {
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			continue
		}
		key := field.Name
		if tag, ok := field.Tag.Lookup("mapstructure"); ok {
			key = tag
		} else if tag, ok := field.Tag.Lookup("env"); ok {
			key = tag
		}
		fullKey := prefix + key
		if fieldVal.Kind() == reflect.Struct {
			if err := e.bindStruct(fieldVal, fullKey+"."); err != nil {
				return err
			}
			continue
		}
		if pv, ok := e.GetProperty(fullKey); ok {
			setField(fieldVal, pv)
		}
	}
	return nil
}

var globalTypeConverter = NewTypeConverter()

func setField(fieldVal reflect.Value, val any) {
	converted, err := globalTypeConverter.ConvertTo(val, fieldVal.Type())
	if err != nil {
		// 转换失败，忽略
		return
	}
	fieldVal.Set(converted)
}

func reflectValueOf(v any) reflect.Value {
	rv, ok := v.(reflect.Value)
	if ok {
		return rv
	}
	return reflect.ValueOf(v)
}

// AddPropertySource 添加配置源到环境.
//
// 新添加的配置源会按优先级排序,高优先级覆盖低优先级.
//
// 参数:
//   - source: 配置源实例
func (e *Environment) AddPropertySource(source PropertySource) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sources = append(e.sources, source)
	e.sortSources()
}

// AddPropertySourceFirst 添加最高优先级的配置源.
//
// 新添加的配置源会插入到配置源列表头部,成为最高优先级的来源.
//
// 参数:
//   - source: 配置源实例
func (e *Environment) AddPropertySourceFirst(source PropertySource) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sources = append([]PropertySource{source}, e.sources...)
	e.sortSources()
}

// GetPropertySources 获取所有配置源列表.
//
// 返回按优先级排序的配置源副本,高优先级在后.
//
// 返回:
//   - []PropertySource: 配置源列表
func (e *Environment) GetPropertySources() []PropertySource {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]PropertySource, len(e.sources))
	copy(result, e.sources)
	return result
}

// GetActiveProfiles 获取当前激活的 Profile 列表.
//
// 返回:
//   - []string: Profile 名称列表
func (e *Environment) GetActiveProfiles() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]string, len(e.activeProfiles))
	copy(result, e.activeProfiles)
	return result
}

// AddActiveProfile 激活指定 Profile.
//
// 如果该 Profile 已经激活,则忽略.
//
// 参数:
//   - profile: Profile 名称
func (e *Environment) AddActiveProfile(profile string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if slices.Contains(e.activeProfiles, profile) {
		return
	}
	e.activeProfiles = append(e.activeProfiles, profile)
}

// AcceptsProfile 检查指定 Profile 是否被当前环境接受
//
// 支持否定前缀 "!"，如 "!dev" 表示非 dev 环境时匹配。
// 检查流程：
//  1. 检查是否有否定前缀
//  2. 在激活的 Profile 列表中查找
//  3. 含否定前缀时：Profile 不在激活列表中返回 true
//  4. 无否定前缀时：Profile 在激活列表中返回 true
func (e *Environment) AcceptsProfile(profile string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	negate := false
	if strings.HasPrefix(profile, "!") {
		negate = true
		profile = profile[1:]
	}
	if slices.Contains(e.activeProfiles, profile) {
		return !negate
	}
	return negate
}

func (e *Environment) sortSources() {
	sort.Slice(e.sources, func(i, j int) bool {
		return e.sources[i].Priority() < e.sources[j].Priority()
	})
}

// AddConfigChangeListener 添加配置变更监听器
func (e *Environment) AddConfigChangeListener(listener func(refresh.ConfigChangeEvent)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.configListeners = append(e.configListeners, listener)
}

// notifyConfigChange 通知所有配置变更监听器
func (e *Environment) notifyConfigChange(event refresh.ConfigChangeEvent) {
	e.mu.RLock()
	listeners := make([]func(refresh.ConfigChangeEvent), len(e.configListeners))
	copy(listeners, e.configListeners)
	e.mu.RUnlock()

	for _, listener := range listeners {
		go listener(event) // 异步通知
	}
}
