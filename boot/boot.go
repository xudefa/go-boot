package boot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/config"
	contextpkg "github.com/xudefa/go-boot/context"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
	"github.com/xudefa/go-boot/life"
)

// ConfigCenterFactory 配置中心工厂函数类型
type ConfigCenterFactory func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error)

var (
	configCenterFactories = make(map[string]ConfigCenterFactory)
	factoryMutex          sync.RWMutex
)

// RegisterConfigCenterFactory 注册配置中心工厂函数
func RegisterConfigCenterFactory(centerType string, factory ConfigCenterFactory) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	configCenterFactories[centerType] = factory
}

// BootConfig 启动配置
type BootConfig struct {
	ConfigLocation string   // 配置文件路径
	ConfigType     string   // 配置文件类型 (json)
	Profiles       []string // 激活的 Profile
	AppName        string   // 应用名称
	Version        string   // 版本号

	AutoExecute bool // 是否自动执行自动配置（默认 true）
	Starters    bool // 是否自动管理启动器生命周期（默认 true）

	CustomPropertySources []environment.PropertySource // 用户自定义配置源

	// 配置中心配置
	ConfigCenterEnabled bool          // 是否启用配置中心（默认 false）
	ConfigCenterType    string        // 配置中心类型 (nacos/etcd/consul)
	ConfigCenterAddr    []string      // 配置中心地址
	ConfigCenterDataID  string        // 配置中心数据ID
	ConfigCenterGroup   string        // 配置中心分组
	ConfigCenterPrefix  string        // 配置中心前缀
	ConfigCenterTimeout time.Duration // 配置中心超时时间
}

// defaultBootConfig 返回默认启动配置
func defaultBootConfig() *BootConfig {
	return &BootConfig{
		AppName:     "go-boot-app",
		Version:     "1.0.0",
		ConfigType:  "json",
		AutoExecute: true,
		Starters:    true,
	}
}

// BootOption 启动选项函数
type BootOption func(*BootConfig)

// WithConfigLocation 设置配置文件路径
func WithConfigLocation(location string) BootOption {
	return func(cfg *BootConfig) {
		cfg.ConfigLocation = location
	}
}

// WithProfiles 设置激活的 Profile
func WithProfiles(profiles ...string) BootOption {
	return func(cfg *BootConfig) {
		cfg.Profiles = append(cfg.Profiles, profiles...)
	}
}

// WithAppName 设置应用名称
func WithAppName(name string) BootOption {
	return func(cfg *BootConfig) {
		cfg.AppName = name
	}
}

// WithVersion 设置版本号
func WithVersion(version string) BootOption {
	return func(cfg *BootConfig) {
		cfg.Version = version
	}
}

// WithoutAutoConfig 禁用自动配置执行
func WithoutAutoConfig() BootOption {
	return func(cfg *BootConfig) {
		cfg.AutoExecute = false
	}
}

// WithoutStarters 禁用启动器自动管理
func WithoutStarters() BootOption {
	return func(cfg *BootConfig) {
		cfg.Starters = false
	}
}

// WithConfigType 设置配置文件类型（如 json），为空时使用默认值。
func WithConfigType(configType string) BootOption {
	return func(cfg *BootConfig) {
		if configType != "" {
			cfg.ConfigType = configType
		}
	}
}

// WithPropertySource 添加自定义配置源，优先级最高。
func WithPropertySource(source environment.PropertySource) BootOption {
	return func(cfg *BootConfig) {
		cfg.CustomPropertySources = append(cfg.CustomPropertySources, source)
	}
}

// WithConfigCenter 启用配置中心
func WithConfigCenter(centerType string, addr []string, opts ...ConfigCenterOption) BootOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterEnabled = true
		cfg.ConfigCenterType = centerType
		cfg.ConfigCenterAddr = addr
		cfg.ConfigCenterTimeout = 5 * time.Second
		for _, opt := range opts {
			opt(cfg)
		}
	}
}

// ConfigCenterOption 配置中心选项函数
type ConfigCenterOption func(*BootConfig)

// WithConfigCenterDataID 设置配置中心数据ID
func WithConfigCenterDataID(dataID string) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterDataID = dataID
	}
}

// WithConfigCenterGroup 设置配置中心分组
func WithConfigCenterGroup(group string) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterGroup = group
	}
}

// WithConfigCenterPrefix 设置配置中心前缀
func WithConfigCenterPrefix(prefix string) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterPrefix = prefix
	}
}

// WithConfigCenterTimeout 设置配置中心超时时间
func WithConfigCenterTimeout(timeout time.Duration) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterTimeout = timeout
	}
}

// NewApplication 创建新的应用
//
// 这是 go-boot 框架的推荐入口，替代旧的 start.BuildApp()。
// 自动配置（AutoConfiguration）和启动器（Starter）会在 Start() 中按生命周期阶段自动执行。
//
// 示例：
//
//	ctx, err := boot.NewApplication(
//	    boot.WithAppName("my-app"),
//	    boot.WithVersion("1.0.0"),
//	    boot.WithProfiles("dev"),
//	)
//	if err != nil { log.Fatal(err) }
//	ctx.Start()
//	defer ctx.Stop()
//	ctx.WaitForSignal()
func NewApplication(opts ...BootOption) (*Boot, error) {
	cfg := defaultBootConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	container := core.New()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	for _, p := range cfg.Profiles {
		appCtx.Environment().AddActiveProfile(p)
	}

	configLoader := environment.NewConfigLoader(
		"application",
		environment.ConfigType(cfg.ConfigType),
		cfg.ConfigLocation,
		cfg.Profiles,
	)

	return &Boot{
		ctx:          appCtx,
		config:       cfg,
		configLoader: configLoader,
	}, nil
}

// Boot 应用启动器，管理应用的完整生命周期
//
// 参考 Spring Boot 的 SpringApplication，负责：
//   - 自动配置执行（AutoConfiguration）
//   - 启动器管理（Starter 的 Configure/Start/Stop）
//   - 生命周期阶段流转
//   - 事件发布
//   - 优雅关闭
type Boot struct {
	ctx          *contextpkg.DefaultApplicationContext
	config       *BootConfig
	configLoader *environment.ConfigLoader
	starters     []Starter
}

// Context 返回应用上下文
func (b *Boot) Context() *contextpkg.DefaultApplicationContext {
	return b.ctx
}

// Container 返回 IoC 容器
func (b *Boot) Container() core.Container {
	return b.ctx.Container()
}

// Environment 返回环境配置
func (b *Boot) Environment() *environment.Environment {
	return b.ctx.Environment()
}

// Start 启动应用，执行完整的生命周期
//
// 启动流程参考 Spring Boot：
//  1. PhaseInitializing → PhaseConfiguring：执行自动配置
//  2. PhaseConfiguring → PhaseContextRefreshed：配置启动器
//  3. PhaseContextRefreshed → PhaseReady：启动启动器
//  4. 打印横幅
//  5. PhaseReady → PhaseRunning：发布事件
func (b *Boot) Start() error {
	if b.ctx.IsRunning() {
		return nil
	}

	// === 阶段 1：配置阶段 ===
	if err := b.ctx.Lifecycle().SetPhase(life.PhaseConfiguring); err != nil {
		return b.reportError("configuring", err)
	}

	if b.configLoader != nil {
		configSources, err := b.configLoader.Load()
		if err != nil {
			return b.reportError("configuring", fmt.Errorf("failed to load config files: %w", err))
		}

		for _, source := range configSources {
			b.ctx.Environment().AddPropertySource(source)
		}
	}

	if b.config.ConfigCenterEnabled {
		if err := b.loadConfigCenterConfig(); err != nil {
			return b.reportError("configuring", fmt.Errorf("failed to load config center: %w", err))
		}
	}

	for _, source := range b.config.CustomPropertySources {
		b.ctx.Environment().AddPropertySourceFirst(source)
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventEnvironmentPrepared})

	if b.config.AutoExecute {
		entries := GlobalRegistry().GetMatching(newConditionCtx(b.ctx))
		for _, entry := range entries {
			if err := entry.Config.Configure(newAppCtx(b.ctx)); err != nil {
				return b.reportError("configuring", fmt.Errorf("auto-config %T failed: %w", entry.Config, err))
			}
		}
	}

	b.starters = GlobalStarterRegistry().GetOrdered()

	if b.config.Starters {
		for _, s := range b.starters {
			if !b.starterMatches(s) {
				continue
			}
			if err := s.Configure(newAppCtx(b.ctx)); err != nil {
				return b.reportError("configuring", fmt.Errorf("starter %s configure failed: %w", s.Name(), err))
			}
		}
	}

	// === 阶段 2：上下文刷新 ===
	if err := b.ctx.Lifecycle().SetPhase(life.PhaseContextRefreshed); err != nil {
		return b.reportError("context_refreshed", err)
	}
	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventContextRefreshed})

	// === 阶段 3：就绪阶段 ===
	if err := b.ctx.Lifecycle().SetPhase(life.PhaseReady); err != nil {
		return b.reportError("ready", err)
	}

	if b.config.Starters {
		for _, s := range b.starters {
			if !b.starterMatches(s) {
				continue
			}
			if err := s.Start(newAppCtx(b.ctx)); err != nil {
				return b.reportError("ready", fmt.Errorf("starter %s start failed: %w", s.Name(), err))
			}
		}
	}

	banner := DefaultBanner
	banner.Print(os.Stdout, b.config.AppName, b.config.Version, b.ctx.Environment().GetActiveProfiles())

	// === 阶段 4：运行阶段 ===
	if err := b.ctx.Lifecycle().SetPhase(life.PhaseRunning); err != nil {
		return b.reportError("running", err)
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationReady})

	return nil
}

// Stop 停止应用
//
// 流程：
//  1. 根据当前阶段执行相应清理
//  2. 逆序停止启动器
//  3. 发布停止事件
//
// 注意：允许从任何阶段调用 Stop（而不仅仅是 PhaseRunning），
// 以支持 Start() 部分失败时的资源清理。
func (b *Boot) Stop() error {
	phase := b.ctx.Lifecycle().GetPhase()

	// 已停止、停止中或在初始化之前，无需清理
	if phase >= life.PhaseStopping || phase == life.PhaseInitializing {
		return nil
	}

	// 尝试推进到 PhaseStopping（允许从 Configuring / ContextRefreshed / Ready / Running 进入）
	if err := b.ctx.Lifecycle().SetPhase(life.PhaseStopping); err != nil {
		_ = b.reportError("stopping", err)
		fmt.Fprintf(os.Stderr, "[go-boot] failed to set phase to stopping: %v\n", err)
	}

	// 逆序停止启动器（仅当启动器已启动后才执行停止）
	if b.config.Starters && phase >= life.PhaseReady {
		for i := len(b.starters) - 1; i >= 0; i-- {
			s := b.starters[i]
			if !b.starterMatches(s) {
				continue
			}
			if err := s.Stop(newAppCtx(b.ctx)); err != nil {
				fmt.Fprintf(os.Stderr, "starter %s stop error: %v\n", s.Name(), err)
			}
		}
	}

	if err := b.ctx.Lifecycle().SetPhase(life.PhaseStopped); err != nil {
		_ = b.reportError("stopping", err)
		fmt.Fprintf(os.Stderr, "[go-boot] failed to set phase to stopped: %v\n", err)
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
	return nil
}

// IsRunning 检查应用是否运行中
func (b *Boot) IsRunning() bool {
	return b.ctx.Lifecycle().GetPhase() == life.PhaseRunning
}

// WaitForSignal 等待终止信号，收到信号后自动执行优雅关闭
func (b *Boot) WaitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\n接收到终止信号，正在优雅关闭...")
	signal.Stop(sigCh)
	if stopErr := b.Stop(); stopErr != nil {
		fmt.Fprintf(os.Stderr, "[go-boot] failed to stop application: %v\n", stopErr)
	}
}

// starterMatches 检查启动器条件是否匹配
func (b *Boot) starterMatches(s Starter) bool {
	cond := s.GetCondition()
	if cond == nil {
		return true
	}
	return cond.Matches(newConditionCtx(b.ctx))
}

// reportError 通过 FailureAnalyzer 输出友好错误提示并返回结构化错误
func (b *Boot) reportError(phase string, err error) *BootError {
	bootErr := NewBootError(phase, err)

	// 使用 FailureAnalyzer 分析
	report := globalAnalyzerRegistry.Analyze(err)
	if report != nil {
		_ = bootErr.WithAnalysis(report.Description)
		_ = bootErr.WithSuggestions(report.PossibleSolutions...)
		fmt.Fprintf(os.Stderr, "\n%s\n", formatFailure(report))
	}

	return bootErr
}

// loadConfigCenterConfig 从配置中心加载配置
func (b *Boot) loadConfigCenterConfig() error {
	if len(b.config.ConfigCenterAddr) == 0 {
		return fmt.Errorf("config center address is required")
	}

	factoryMutex.RLock()
	factory, ok := configCenterFactories[b.config.ConfigCenterType]
	factoryMutex.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported config center type: %s (no factory registered)", b.config.ConfigCenterType)
	}

	cfg := b.buildConfigCenterConfig()
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ConfigCenterTimeout)
	defer cancel()

	center, err := factory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create config center client: %w", err)
	}
	defer func() {
		if closeErr := center.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "[go-boot] failed to close config center: %v\n", closeErr)
		}
	}()

	data, err := center.Load()
	if err != nil {
		return fmt.Errorf("failed to load config from center: %w", err)
	}

	if len(data) > 0 {
		source := environment.NewMapPropertySource("config-center", environment.PriorityNormal, data)
		b.ctx.Environment().AddPropertySource(source)
	}

	return nil
}

// buildConfigCenterConfig 构建配置中心配置
func (b *Boot) buildConfigCenterConfig() *config.ConfigCenterConfig {
	cfg := &config.ConfigCenterConfig{
		Endpoints: b.config.ConfigCenterAddr,
		Timeout:   b.config.ConfigCenterTimeout,
	}

	switch b.config.ConfigCenterType {
	case "nacos":
		dataID := b.config.ConfigCenterDataID
		if dataID == "" {
			dataID = "app-config"
		}
		group := b.config.ConfigCenterGroup
		if group == "" {
			group = "DEFAULT_GROUP"
		}
		cfg.DataID = dataID
		cfg.Group = group

	case "etcd", "consul":
		prefix := b.config.ConfigCenterPrefix
		if prefix == "" {
			if b.config.ConfigCenterType == "etcd" {
				prefix = "/config"
			} else {
				prefix = "config"
			}
		}
		cfg.Prefix = prefix
	}

	return cfg
}

// appCtxAdapter 适配 DefaultApplicationContext 到 boot.ApplicationContext
//
// DefaultApplicationContext.EventBus() 返回 *event.EventBus，
// 而 boot.ApplicationContext.EventBus() 要求返回 interface{ Publish(...) }，
// 在 Go 中这被视为不同签名，需要显式适配。
type appCtxAdapter struct {
	ctx *contextpkg.DefaultApplicationContext
}

func newAppCtx(ctx *contextpkg.DefaultApplicationContext) *appCtxAdapter {
	return &appCtxAdapter{ctx: ctx}
}

func (a *appCtxAdapter) Container() core.Container {
	return a.ctx.Container()
}

func (a *appCtxAdapter) Environment() *environment.Environment {
	return a.ctx.Environment()
}

func (a *appCtxAdapter) Register(name string, opts ...core.BuilderOption) error {
	return a.ctx.Register(name, opts...)
}

func (a *appCtxAdapter) Get(name string) (any, error) {
	return a.ctx.Get(name)
}

func (a *appCtxAdapter) EventBus() interface {
	Publish(event event.ApplicationEvent)
} {
	return a.ctx.EventBus()
}

// conditionCtx 适配 DefaultApplicationContext 到 condition.ConditionContext
//
// DefaultApplicationContext 的方法签名（如 Environment() *environment.Environment）
// 与 condition.ConditionContext 要求的签名不完全一致，因此需要此适配器桥接。
type conditionCtx struct {
	ctx *contextpkg.DefaultApplicationContext
}

func newConditionCtx(ctx *contextpkg.DefaultApplicationContext) *conditionCtx {
	return &conditionCtx{ctx: ctx}
}

func (c *conditionCtx) Environment() interface{ GetProperty(key string) (any, bool) } {
	return c.ctx.Environment()
}

func (c *conditionCtx) Container() interface{ Has(id string) bool } {
	return c.ctx.Container()
}

func (c *conditionCtx) ClassLoader() interface{ HasClass(name string) bool } {
	return c.ctx.ClassLoader()
}

func (c *conditionCtx) GetBean(beanID string) (any, bool) {
	return c.ctx.GetBean(beanID)
}

func (c *conditionCtx) HasProperty(key string) bool {
	return c.ctx.HasProperty(key)
}

func (c *conditionCtx) GetProperty(key string) (any, bool) {
	return c.ctx.GetProperty(key)
}

var _ condition.ConditionContext = (*conditionCtx)(nil)
