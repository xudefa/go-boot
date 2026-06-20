// Package constants 定义 go-boot 框架中所有 Bean ID 常量。
//
// 按功能模块分组，用于 IoC 容器中的 Bean 注册和查找。
package constants

const (
	// Tracing 相关
	TracerProviderBeanID = "tracerProvider"
	TracerBeanID         = "tracer"

	// Metrics 相关
	MeterRegistryBeanID = "meterRegistry"

	// Actuator 相关
	ActuatorBeanID = "actuator"

	// Schedule 相关
	ScheduleSchedulerBeanID = "scheduleScheduler"

	// JWT 相关
	JWTUtilBeanID             = "jwtUtil"
	JWTAuthenticationFilterID = "jwtAuthenticationFilter"

	// Casbin 相关
	CasbinEnforcerBeanID = "casbinEnforcer"

	// Hertz 相关
	HertzServerBeanID = "hertzServer"

	// Gin 相关
	GinServerBeanID = "ginServer"

	// FastHTTP 相关
	FastHTTPClientBeanID = "fastHttpClient"

	// Kitex 相关
	KitexServerBeanID = "kitexServer"
	KitexClientBeanID = "kitexClient"

	// GORM 相关
	GORMDBBeanID                  = "gormDB"
	DatabaseHealthIndicatorBeanID = "databaseHealthIndicator"

	// Redis 相关
	RedisCacheBeanID           = "redisCache"
	RedisHealthIndicatorBeanID = "redisHealthIndicator"

	// XORM 相关
	XORMDBBeanID                      = "xormDB"
	XORMDatabaseHealthIndicatorBeanID = "xormDatabaseHealthIndicator"

	// Email 相关
	EmailClientBeanID = "emailClient"

	// Viper 相关
	ViperConfigBeanID = "viperConfig"

	// ZeroLog 相关
	ZeroLogLoggerBeanID = "zerologLogger"

	// Zap 相关
	ZapLoggerBeanID = "zapLogger"

	// Nacos 相关
	NacosRegistryBeanID = "nacosRegistry"

	// gRPC 相关
	GRPCServerBeanID = "grpcServer"
	GRPCClientBeanID = "grpcClient"

	// Config Center 相关
	ConfigCenterBeanID       = "configCenter"
	EtcdConfigCenterBeanID   = "etcdConfigCenter"
	NacosConfigCenterBeanID  = "nacosConfigCenter"
	ConsulConfigCenterBeanID = "consulConfigCenter"

	// Etcd 相关
	EtcdRegistryBeanID = "etcdRegistry"

	// Consul 相关
	ConsulRegistryBeanID = "consulRegistry"

	// Security 相关
	UserDetailsServiceBeanID    = "userDetailsService"
	PasswordEncoderBeanID       = "passwordEncoder"
	AuthenticationManagerBeanID = "authenticationManager"
	SecurityFilterChainBeanID   = "securityFilterChain"

	// WebSocket 相关
	WebSocketServerBeanID = "webSocketServer"

	// Prometheus 相关
	PrometheusExporterBeanID = "prometheus.exporter"

	// Validation 相关
	TagValidatorBeanID = "tagValidator"
)
