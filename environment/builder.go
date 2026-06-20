package environment

// EnvironmentBuilder 环境配置构建器，支持链式配置
type EnvironmentBuilder struct {
	profiles        []string
	propertySources []PropertySource
}

// NewEnvironmentBuilder 创建环境配置构建器
func NewEnvironmentBuilder() *EnvironmentBuilder {
	return &EnvironmentBuilder{
		profiles:        []string{},
		propertySources: []PropertySource{},
	}
}

// WithProfile 添加激活的 Profile
func (b *EnvironmentBuilder) WithProfile(profile string) *EnvironmentBuilder {
	b.profiles = append(b.profiles, profile)
	return b
}

// WithProfiles 批量添加激活的 Profiles
func (b *EnvironmentBuilder) WithProfiles(profiles ...string) *EnvironmentBuilder {
	b.profiles = append(b.profiles, profiles...)
	return b
}

// WithPropertySource 添加配置源
func (b *EnvironmentBuilder) WithPropertySource(source PropertySource) *EnvironmentBuilder {
	b.propertySources = append(b.propertySources, source)
	return b
}

// WithPropertySourceFirst 添加最高优先级的配置源
func (b *EnvironmentBuilder) WithPropertySourceFirst(source PropertySource) *EnvironmentBuilder {
	b.propertySources = append([]PropertySource{source}, b.propertySources...)
	return b
}

// WithJSONConfig 添加JSON配置文件
func (b *EnvironmentBuilder) WithJSONConfig(filePath string) *EnvironmentBuilder {
	if source := NewJSONPropertySourceOrDefault("json-config", filePath); source != nil {
		b.propertySources = append(b.propertySources, source)
	}
	return b
}

// WithEnvPrefix 添加带前缀的环境变量配置源
func (b *EnvironmentBuilder) WithEnvPrefix(prefix string) *EnvironmentBuilder {
	b.propertySources = append(b.propertySources, NewEnvPropertySource("env-"+prefix, prefix))
	return b
}

// WithArgs 添加命令行参数配置源
func (b *EnvironmentBuilder) WithArgs(args ...string) *EnvironmentBuilder {
	b.propertySources = append(b.propertySources, NewArgsPropertySource("args", args))
	return b
}

// Build 构建环境配置
func (b *EnvironmentBuilder) Build() *Environment {
	env := NewEnvironment()

	// 添加自定义配置源
	for _, source := range b.propertySources {
		env.AddPropertySource(source)
	}

	// 添加激活的 Profiles
	for _, profile := range b.profiles {
		env.AddActiveProfile(profile)
	}

	return env
}

// MustBuild 构建环境配置，失败则panic
func (b *EnvironmentBuilder) MustBuild() *Environment {
	env := b.Build()
	if env == nil {
		panic("failed to build environment")
	}
	return env
}

// EnvironmentHelper 环境辅助工具，简化环境配置操作
type EnvironmentHelper struct {
	env    *Environment
	prefix string
}

// NewEnvironmentHelper 创建环境辅助工具
func NewEnvironmentHelper(env *Environment) *EnvironmentHelper {
	return &EnvironmentHelper{env: env}
}

// WithPrefix 创建带前缀的新辅助工具
func (h *EnvironmentHelper) WithPrefix(prefix string) *EnvironmentHelper {
	return &EnvironmentHelper{
		env:    h.env,
		prefix: prefix,
	}
}

// key 生成带前缀的键
func (h *EnvironmentHelper) key(key string) string {
	if h.prefix == "" {
		return key
	}
	return h.prefix + "." + key
}

// GetString 获取字符串类型属性
func (h *EnvironmentHelper) GetString(key string, defaultVal string) string {
	return h.env.GetString(h.key(key), defaultVal)
}

// GetInt 获取整数类型属性
func (h *EnvironmentHelper) GetInt(key string, defaultVal int) int {
	return h.env.GetInt(h.key(key), defaultVal)
}

// GetBool 获取布尔类型属性
func (h *EnvironmentHelper) GetBool(key string, defaultVal bool) bool {
	return h.env.GetBool(h.key(key), defaultVal)
}

// GetFloat64 获取浮点数类型属性
func (h *EnvironmentHelper) GetFloat64(key string, defaultVal float64) float64 {
	return h.env.GetFloat64(h.key(key), defaultVal)
}

// ContainsProperty 检查属性是否存在
func (h *EnvironmentHelper) ContainsProperty(key string) bool {
	return h.env.ContainsProperty(h.key(key))
}

// GetRequiredProperty 获取必需属性
func (h *EnvironmentHelper) GetRequiredProperty(key string) (any, error) {
	return h.env.GetRequiredProperty(h.key(key))
}

// IsDev 检查是否为开发环境
func (h *EnvironmentHelper) IsDev() bool {
	return h.env.AcceptsProfile("dev")
}

// IsProd 检查是否为生产环境
func (h *EnvironmentHelper) IsProd() bool {
	return h.env.AcceptsProfile("prod")
}

// IsTest 检查是否为测试环境
func (h *EnvironmentHelper) IsTest() bool {
	return h.env.AcceptsProfile("test")
}

// GetActiveProfile 获取当前激活的 Profile
func (h *EnvironmentHelper) GetActiveProfile() string {
	profiles := h.env.GetActiveProfiles()
	if len(profiles) > 0 {
		return profiles[0]
	}
	return "default"
}

// EnvironmentTemplate 环境模板，提供常用的环境配置模板
type EnvironmentTemplate struct {
	env *Environment
}

// NewEnvironmentTemplate 创建环境模板
func NewEnvironmentTemplate(env *Environment) *EnvironmentTemplate {
	return &EnvironmentTemplate{env: env}
}

// GetDatabaseURL 获取数据库连接URL
func (t *EnvironmentTemplate) GetDatabaseURL(defaultVal string) string {
	return t.env.GetString("database.url", defaultVal)
}

// GetDatabaseHost 获取数据库主机
func (t *EnvironmentTemplate) GetDatabaseHost(defaultVal string) string {
	return t.env.GetString("database.host", defaultVal)
}

// GetDatabasePort 获取数据库端口
func (t *EnvironmentTemplate) GetDatabasePort(defaultVal int) int {
	return t.env.GetInt("database.port", defaultVal)
}

// GetDatabaseName 获取数据库名称
func (t *EnvironmentTemplate) GetDatabaseName(defaultVal string) string {
	return t.env.GetString("database.name", defaultVal)
}

// GetServerHost 获取服务器主机
func (t *EnvironmentTemplate) GetServerHost(defaultVal string) string {
	return t.env.GetString("server.host", defaultVal)
}

// GetServerPort 获取服务器端口
func (t *EnvironmentTemplate) GetServerPort(defaultVal int) int {
	return t.env.GetInt("server.port", defaultVal)
}

// GetLogLevel 获取日志级别
func (t *EnvironmentTemplate) GetLogLevel(defaultVal string) string {
	return t.env.GetString("log.level", defaultVal)
}

// GetRedisHost 获取Redis主机
func (t *EnvironmentTemplate) GetRedisHost(defaultVal string) string {
	return t.env.GetString("redis.host", defaultVal)
}

// GetRedisPort 获取Redis端口
func (t *EnvironmentTemplate) GetRedisPort(defaultVal int) int {
	return t.env.GetInt("redis.port", defaultVal)
}

// GetRedisPassword 获取Redis密码
func (t *EnvironmentTemplate) GetRedisPassword(defaultVal string) string {
	return t.env.GetString("redis.password", defaultVal)
}

// IsDebugMode 检查是否启用调试模式
func (t *EnvironmentTemplate) IsDebugMode() bool {
	return t.env.GetBool("debug", false)
}

// IsVerbose 检查是否启用详细日志
func (t *EnvironmentTemplate) IsVerbose() bool {
	return t.env.GetBool("verbose", false)
}

// EnvironmentConfig 环境配置
type EnvironmentConfig struct {
	Profiles           []string         // 激活的 Profiles
	PropertySources    []PropertySource // 配置源列表
	DefaultProfile     string           // 默认 Profile
	AutoDetectProfiles bool             // 是否自动检测 Profiles
}

// EnvironmentOption 环境配置选项
type EnvironmentOption func(*EnvironmentConfig)

// WithProfiles 设置激活的 Profiles
func WithProfiles(profiles ...string) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.Profiles = append(c.Profiles, profiles...)
	}
}

// WithDefaultProfile 设置默认 Profile
func WithDefaultProfile(profile string) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.DefaultProfile = profile
	}
}

// WithAutoDetectProfiles 设置是否自动检测 Profiles
func WithAutoDetectProfiles(enabled bool) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.AutoDetectProfiles = enabled
	}
}

// WithPropertySources 设置配置源列表
func WithPropertySources(sources ...PropertySource) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.PropertySources = append(c.PropertySources, sources...)
	}
}

// DefaultEnvironmentConfig 返回默认环境配置
func DefaultEnvironmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		DefaultProfile:     "default",
		AutoDetectProfiles: true,
	}
}

// ApplyOptions 应用配置选项
func (c *EnvironmentConfig) ApplyOptions(opts []EnvironmentOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// CreateEnvironment 创建并配置环境
func CreateEnvironment(opts ...EnvironmentOption) *Environment {
	config := DefaultEnvironmentConfig()
	config.ApplyOptions(opts)

	env := NewEnvironment()

	// 添加配置源
	for _, source := range config.PropertySources {
		env.AddPropertySource(source)
	}

	// 添加 Profiles
	if config.DefaultProfile != "" {
		env.AddActiveProfile(config.DefaultProfile)
	}
	for _, profile := range config.Profiles {
		env.AddActiveProfile(profile)
	}

	return env
}
