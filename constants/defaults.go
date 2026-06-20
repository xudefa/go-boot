// Package constants 定义 go-boot 框架中所有默认值常量。
//
// 按功能模块分组，与 config_keys.go 中的配置键一一对应。
package constants

const (
	// Viper 默认值
	DefaultViperConfigName  = "config"
	DefaultViperConfigType  = "json"
	DefaultViperConfigPaths = "./,./config"
	DefaultViperEnvPrefix   = ""

	// 应用默认值
	DefaultAppName    = "go-boot-app"
	DefaultAppVersion = "1.0.0"

	// 服务器默认值
	DefaultServerHost = ":8080"
	DefaultServerMode = "debug"

	// 日志默认值
	DefaultLogLevel = "info"

	// 数据源默认值
	DefaultDatasourceHost     = "localhost"
	DefaultDatasourcePort     = 3306
	DefaultDatasourceUsername = "gate"
	DefaultDatasourcePassword = "123456"
	DefaultDatasourceDatabase = "gate"
	DefaultDatasourceCharset  = "utf8mb4"

	// 缓存默认值
	DefaultCacheTTL = 300

	// JWT 默认值
	DefaultJWTSecretKey              = "go-bootJwtSecret"
	DefaultJWTIssuer                 = "go-boot"
	DefaultJWTExpiresDuration        = 600
	DefaultJWTRefreshExpiresDuration = 3600
	DefaultJWTExcludePaths           = "/login,/register,/health,/actuator/health"

	// Casbin 默认值
	DefaultCasbinDBTable = "casbin_rule"

	// Hertz 默认值
	DefaultHertzReadTimeout  = 10
	DefaultHertzWriteTimeout = 10

	// Gin 默认值
	DefaultGinReadTimeout  = 10
	DefaultGinWriteTimeout = 10
	DefaultGinIdleTimeout  = 60

	// Tracing 默认值
	DefaultTracingServiceName          = "go-boot-app"
	DefaultTracingServiceVersion       = "1.0.0"
	DefaultTracingExporterType         = "otlpgrpc"
	DefaultTracingExporterEndpoint     = "localhost:4317"
	DefaultTracingExporterInsecure     = true
	DefaultTracingEnvironment          = "production"
	DefaultTracingExporterBatchTimeout = 5000
	DefaultTracingExporterBatchMaxSize = 512
	DefaultTracingSampling             = 1.0

	// FastHTTP 默认值
	DefaultFastHTTPTimeout         = 30
	DefaultFastHTTPMaxConnsPerHost = 256
	DefaultFastHTTPReadTimeout     = 10
	DefaultFastHTTPWriteTimeout    = 10

	// Kitex 默认值
	DefaultKitexClientTimeout = 5

	// GORM 默认值
	DefaultGORMHost            = "localhost"
	DefaultGORMPort            = 3306
	DefaultGORMUsername        = "gate"
	DefaultGORMPassword        = "123456"
	DefaultGORMDatabase        = "gate"
	DefaultGORMCharset         = "utf8mb4"
	DefaultGORMTimezone        = "Local"
	DefaultGORMMaxOpenConns    = 100
	DefaultGORMMaxIdleConns    = 10
	DefaultGORMConnMaxLifetime = 3600

	// Redis 默认值
	DefaultRedisAddress              = "localhost:6379"
	DefaultRedisDB                   = 0
	DefaultRedisPoolSize             = 10
	DefaultRedisMaxActiveConnections = 0
	DefaultRedisMinIdleConnections   = 5
	DefaultRedisUseCluster           = false
	DefaultRedisCachePrefix          = "webcache:"

	// Security 默认值
	DefaultSecurityLoginURL              = "/login"
	DefaultSecurityCorsAllowedOrigins    = "*"
	DefaultSecurityCorsAllowedMethods    = "GET,POST,PUT,DELETE,OPTIONS"
	DefaultSecurityCorsAllowedHeaders    = "Content-Type,Authorization,X-Requested-With"
	DefaultSecurityCorsExposedHeaders    = ""
	DefaultSecurityCorsMaxAge            = 3600
	DefaultSecurityCorsAllowCredentials  = false
	DefaultSecurityCorsEnabled           = false
	DefaultSecurityRateLimitRate         = 100
	DefaultSecurityRateLimitBurst        = 200
	DefaultSecurityRateLimitEnabled      = false
	DefaultSecurityRateLimitExcludePaths = "/health,/actuator/health"

	// Actuator 默认值
	DefaultActuatorPort             = "8081"
	DefaultActuatorExposeHealth     = true
	DefaultActuatorExposeMetrics    = true
	DefaultActuatorExposeEnv        = true
	DefaultActuatorExposeBeans      = true
	DefaultActuatorExposeInfo       = true
	DefaultActuatorExposePrometheus = true

	// Schedule 默认值
	DefaultSchedulePoolSize        = 10
	DefaultScheduleScanAnnotations = true

	// XORM 默认值
	DefaultXORMType         = "mysql"
	DefaultXORMHost         = "localhost"
	DefaultXORMPort         = 3306
	DefaultXORMUsername     = "gate"
	DefaultXORMPassword     = "123456"
	DefaultXORMDatabase     = "gate"
	DefaultXORMMaxOpenConns = 100
	DefaultXORMMaxIdleConns = 10
	DefaultXORMShowSQL      = false

	// Email 默认值
	DefaultEmailSMTP     = "smtp.163.com"
	DefaultEmailPort     = 25
	DefaultEmailUsername = ""
	DefaultEmailPassword = ""

	// Zap 默认值
	DefaultZapCallerSkip = 1
	DefaultZapAddCaller  = false

	// gRPC 默认值
	DefaultGRPCClientTimeout = 5

	// Prometheus 默认值
	DefaultPrometheusEnabled  = false
	DefaultPrometheusEndpoint = "/metrics"
	DefaultPrometheusPort     = 9090
)
