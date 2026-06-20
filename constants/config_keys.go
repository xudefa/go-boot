// Package constants 定义 go-boot 框架中所有配置键常量。
//
// 使用集中式常量管理避免硬编码字符串，提高可维护性。
// 按功能模块分组：Viper、应用、服务器、数据源、缓存、特性开关、
// JWT、Casbin、Hertz、Gin、Tracing、FastHTTP、Kitex、GORM、Redis、
// Security、Metrics、Actuator、Schedule、XORM、Email、Zap、ZeroLog、
// Nacos、gRPC、Etcd、Consul、WebSocket、Prometheus。
package constants

const (
	// Viper 配置相关
	ViperEnabled     = "viper.enabled"
	ViperConfigName  = "viper.config-name"
	ViperConfigType  = "viper.config-type"
	ViperConfigPaths = "viper.config-paths"
	ViperEnvPrefix   = "viper.env-prefix"
	ViperEnv         = "viper.env"

	// 应用配置
	AppName    = "app.name"    // 应用名称
	AppVersion = "app.version" // 应用版本
	AppPort    = "app.port"    // 应用端口
	BuildTime  = "build.time"  // 构建时间

	// 服务器配置
	ServerHost = "server.host" // 服务器监听地址
	ServerMode = "server.mode" // 服务器运行模式（debug/release）
	ServerPort = "server.port" // 服务器监听端口

	// 日志配置
	LogLevel = "log.level" // 日志级别（debug/info/warn/error）

	// 数据源配置
	BootDatasourceHost     = "boot.datasource.host"     // 数据库主机地址
	BootDatasourcePort     = "boot.datasource.port"     // 数据库端口
	BootDatasourceUsername = "boot.datasource.username" // 数据库用户名
	BootDatasourcePassword = "boot.datasource.password" // 数据库密码
	BootDatasourceDatabase = "boot.datasource.database" // 数据库名称
	BootDatasourceCharset  = "boot.datasource.charset"  // 字符编码
	DatasourceURL          = "datasource.url"           // 数据源连接 URL

	// 缓存配置
	CacheTTL = "cache.ttl" // 缓存默认过期时间（秒）

	// 特性开关
	FeatureX          = "feature.x"
	FeatureY          = "feature.y"
	FeatureXEnabled   = "feature.x.enable"
	AppFeatureEnabled = "app.feature.enabled"

	// JWT 配置
	JWTEnabled                = "jwt.enabled"                  // 是否启用 JWT
	JWTSecretKey              = "jwt.secret-key"               // JWT 签名密钥
	JWTIssuer                 = "jwt.issuer"                   // JWT 签发者
	JWTExpiresDuration        = "jwt.expires-duration"         // JWT 过期时间（秒）
	JWTRefreshExpiresDuration = "jwt.refresh-expires-duration" // JWT 刷新令牌过期时间（秒）
	JWTExcludePaths           = "jwt.exclude-paths"            // JWT 排除路径（逗号分隔）

	// Casbin 配置
	CasbinEnabled   = "casbin.enabled"    // 是否启用 Casbin
	CasbinModel     = "casbin.model"      // Casbin 模型文件路径
	CasbinAdapter   = "casbin.adapter"    // Casbin 适配器类型
	CasbinDBAdapter = "casbin.db-adapter" // Casbin 数据库适配器类型
	CasbinDBTable   = "casbin.db-table"   // Casbin 数据库表名

	// Hertz 配置
	HertzEnabled      = "hertz.enabled"       // 是否启用 Hertz 服务器
	HertzHost         = "hertz.host"          // Hertz 监听地址
	HertzReadTimeout  = "hertz.read-timeout"  // Hertz 读超时（秒）
	HertzWriteTimeout = "hertz.write-timeout" // Hertz 写超时（秒）

	// Gin 配置
	GinEnabled      = "gin.enabled"       // 是否启用 Gin 服务器
	GinHost         = "gin.host"          // Gin 监听地址
	GinMode         = "gin.mode"          // Gin 运行模式（debug/release）
	GinReadTimeout  = "gin.read-timeout"  // Gin 读超时（秒）
	GinWriteTimeout = "gin.write-timeout" // Gin 写超时（秒）
	GinIdleTimeout  = "gin.idle-timeout"  // Gin 空闲超时（秒）

	// Tracing 配置
	TracingEnabled              = "tracing.enabled"                 // 是否启用链路追踪
	TracingProvider             = "tracing.provider"                // 追踪提供者
	TracingServiceName          = "tracing.service.name"            // 服务名称
	TracingServiceVersion       = "tracing.service.version"         // 服务版本
	TracingExporterType         = "tracing.exporter.type"           // 导出器类型
	TracingExporterEndpoint     = "tracing.exporter.endpoint"       // 导出器端点
	TracingExporterInsecure     = "tracing.exporter.insecure"       // 是否使用非安全连接
	TracingEnvironment          = "tracing.environment"             // 运行环境
	TracingResourceAttributes   = "tracing.resource.attributes"     // 资源属性
	TracingExporterBatchTimeout = "tracing.exporter.batch.timeout"  // 批量导出超时（毫秒）
	TracingExporterBatchMaxSize = "tracing.exporter.batch.max-size" // 批量导出最大条数
	TracingSampling             = "tracing.sampling"                // 采样率（0.0-1.0）

	// FastHTTP 配置
	FastHTTPEnabled         = "fasthttp.enabled"            // 是否启用 FastHTTP 客户端
	FastHTTPBaseURL         = "fasthttp.base-url"           // FastHTTP 基础 URL
	FastHTTPTimeout         = "fasthttp.timeout"            // FastHTTP 请求超时（秒）
	FastHTTPMaxConnsPerHost = "fasthttp.max-conns-per-host" // 每主机最大连接数
	FastHTTPReadTimeout     = "fasthttp.read-timeout"       // FastHTTP 读超时（秒）
	FastHTTPWriteTimeout    = "fasthttp.write-timeout"      // FastHTTP 写超时（秒）

	// Kitex 配置
	KitexServerEnabled = "kitex.server.enabled" // 是否启用 Kitex 服务器
	KitexServerAddress = "kitex.server.address" // Kitex 服务器地址
	KitexClientAddress = "kitex.client.address" // Kitex 客户端地址
	KitexClientTimeout = "kitex.client.timeout" // Kitex 客户端超时（秒）

	// GORM 配置
	GORMEnabled         = "gorm.enabled"           // 是否启用 GORM
	GORMHost            = "gorm.host"              // 数据库主机
	GORMPort            = "gorm.port"              // 数据库端口
	GORMUsername        = "gorm.username"          // 数据库用户名
	GORMPassword        = "gorm.password"          // 数据库密码
	GORMDatabase        = "gorm.database"          // 数据库名称
	GORMCharset         = "gorm.charset"           // 字符编码
	GORMTimezone        = "gorm.timezone"          // 时区
	GORMMaxOpenConns    = "gorm.max-open-conns"    // 最大打开连接数
	GORMMaxIdleConns    = "gorm.max-idle-conns"    // 最大空闲连接数
	GORMConnMaxLifetime = "gorm.conn-max-lifetime" // 连接最大生命时间（秒）

	// Redis 配置
	RedisEnabled              = "redis.enabled"                // 是否启用 Redis
	RedisAddress              = "redis.address"                // Redis 地址
	RedisUsername             = "redis.username"               // Redis 用户名
	RedisPassword             = "redis.password"               // Redis 密码
	RedisDB                   = "redis.db"                     // Redis 数据库编号
	RedisPoolSize             = "redis.pool-size"              // 连接池大小
	RedisMaxActiveConnections = "redis.max-active-connections" // 最大活跃连接数
	RedisMinIdleConnections   = "redis.min-idle-connections"   // 最小空闲连接数
	RedisUseCluster           = "redis.use-cluster"            // 是否使用集群模式
	RedisCachePrefix          = "redis.cache-prefix"           // 缓存键前缀

	// Security 配置
	SecurityEnabled               = "security.enabled"                  // 是否启用安全模块
	SecurityRules                 = "security.rules"                    // 安全规则
	SecurityLoginURL              = "security.login-url"                // 登录 URL
	SecurityCorsEnabled           = "security.cors.enabled"             // 是否启用 CORS
	SecurityCorsAllowedOrigins    = "security.cors.allowed-origins"     // 允许的来源
	SecurityCorsAllowedMethods    = "security.cors.allowed-methods"     // 允许的方法
	SecurityCorsAllowedHeaders    = "security.cors.allowed-headers"     // 允许的请求头
	SecurityCorsExposedHeaders    = "security.cors.exposed-headers"     // 暴露的响应头
	SecurityCorsMaxAge            = "security.cors.max-age"             // 预检缓存时间（秒）
	SecurityCorsAllowCredentials  = "security.cors.allow-credentials"   // 是否允许凭证
	SecurityRateLimitEnabled      = "security.rate-limit.enabled"       // 是否启用限流
	SecurityRateLimitExcludePaths = "security.rate-limit.exclude-paths" // 限流排除路径
	SecurityRateLimitRate         = "security.rate-limit.rate"          // 限流速率（请求/秒）
	SecurityRateLimitBurst        = "security.rate-limit.burst"         // 限流突发量

	// Metrics 配置
	MetricsEnabled = "metrics.enabled" // 是否启用指标收集

	// Actuator 配置
	ActuatorEnabled          = "actuator.enabled"           // 是否启用 Actuator
	ActuatorPort             = "actuator.port"              // Actuator HTTP 端口
	ActuatorExposeHealth     = "actuator.expose.health"     // 是否暴露健康检查端点
	ActuatorExposeMetrics    = "actuator.expose.metrics"    // 是否暴露指标端点
	ActuatorExposeEnv        = "actuator.expose.env"        // 是否暴露环境信息端点
	ActuatorExposeBeans      = "actuator.expose.beans"      // 是否暴露 Bean 列表端点
	ActuatorExposeInfo       = "actuator.expose.info"       // 是否暴露应用信息端点
	ActuatorExposePrometheus = "actuator.expose.prometheus" // 是否暴露 Prometheus 端点

	// Schedule 配置
	ScheduleEnabled         = "schedule.enabled"          // 是否启用定时任务
	SchedulePoolSize        = "schedule.pool-size"        // 定时任务线程池大小
	ScheduleScanAnnotations = "schedule.scan-annotations" // 是否扫描定时任务注解

	// XORM 配置
	XORMEnabled      = "xorm.enabled"        // 是否启用 XORM
	XORMType         = "xorm.type"           // 数据库类型
	XORMHost         = "xorm.host"           // 数据库主机
	XORMPort         = "xorm.port"           // 数据库端口
	XORMUsername     = "xorm.username"       // 数据库用户名
	XORMPassword     = "xorm.password"       // 数据库密码
	XORMDatabase     = "xorm.database"       // 数据库名称
	XORMCharset      = "xorm.charset"        // 字符编码
	XORMMaxOpenConns = "xorm.max-open-conns" // 最大打开连接数
	XORMMaxIdleConns = "xorm.max-idle-conns" // 最大空闲连接数
	XORMShowSQL      = "xorm.show-sql"       // 是否打印 SQL

	// Email 配置
	EmailEnabled  = "email.enabled"  // 是否启用邮件
	EmailSMTP     = "email.smtp"     // SMTP 服务器地址
	EmailPort     = "email.port"     // SMTP 端口
	EmailUsername = "email.username" // 邮箱用户名
	EmailPassword = "email.password" // 邮箱密码

	// Validation 配置
	ValidationEnabled = "validation.enabled" // 是否启用数据验证

	// Zap 配置
	ZapEnabled    = "zap.enabled"     // 是否启用 Zap 日志
	ZapCallerSkip = "zap.caller-skip" // Zap 调用者跳过层数
	ZapAddCaller  = "zap.add-caller"  // Zap 是否添加调用者信息

	// ZeroLog 配置
	ZeroLogEnabled = "zerolog.enabled" // 是否启用 ZeroLog

	// Nacos 配置
	NacosEnabled = "nacos.enabled" // 是否启用 Nacos

	// gRPC 配置
	GRPCServerEnabled = "grpc.server.enabled" // 是否启用 gRPC 服务器
	GRPCServerAddress = "grpc.server.address" // gRPC 服务器地址
	GRPCClientAddress = "grpc.client.address" // gRPC 客户端地址
	GRPCClientTimeout = "grpc.client.timeout" // gRPC 客户端超时（秒）

	// Etcd 配置
	EtcdEnabled = "etcd.enabled" // 是否启用 Etcd

	// Consul 相关
	ConsulEnabled = "consul.enabled" // 是否启用 Consul

	// WebSocket 配置
	WebSocketEnabled = "websocket.enabled" // 是否启用 WebSocket

	// Prometheus 配置
	PrometheusEnabled  = "prometheus.enabled"  // 是否启用 Prometheus 集成
	PrometheusEndpoint = "prometheus.endpoint" // 指标暴露端点
	PrometheusPort     = "prometheus.port"     // HTTP 服务器端口
)
