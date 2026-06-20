package actuator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/environment"
)

// ActuatorHttpStarter Actuator HTTP 启动器
//
// 负责将 Actuator 端点自动挂载到现有的 HTTP 服务器上。
// 支持优先级：Gin → Hertz → net/http 默认 mux
// 通过配置项控制各端点的暴露
type ActuatorHttpStarter struct {
	actuator *Actuator
}

// Name 返回启动器名称
func (s *ActuatorHttpStarter) Name() string {
	return "actuator-http"
}

// Dependencies 返回依赖的其他启动器名称
func (s *ActuatorHttpStarter) Dependencies() []string {
	return []string{}
}

// GetCondition 返回启动条件
func (s *ActuatorHttpStarter) GetCondition() condition.Condition {
	return condition.OnProperty("actuator.enabled", "true")
}

// Configure 配置阶段：从容器中获取 Actuator 实例
func (s *ActuatorHttpStarter) Configure(ctx boot.ApplicationContext) error {
	bean, err := ctx.Container().Get("actuator")
	if err != nil {
		return nil
	}
	s.actuator = bean.(*Actuator)
	return nil
}

// Start 启动阶段：将 Actuator 端点挂载到 HTTP 服务器
func (s *ActuatorHttpStarter) Start(ctx boot.ApplicationContext) error {
	if s.actuator == nil {
		return nil
	}

	env := ctx.Environment()

	// 从配置中读取各端点的暴露开关，默认全部暴露
	exposeHealth := env.GetBool("actuator.expose.health", true)
	exposeMetrics := env.GetBool("actuator.expose.metrics", true)
	exposeEnv := env.GetBool("actuator.expose.env", true)
	exposeBeans := env.GetBool("actuator.expose.beans", true)
	exposeInfo := env.GetBool("actuator.expose.info", true)
	exposePrometheus := env.GetBool("actuator.expose.prometheus", true)

	if !exposeHealth && !exposeMetrics && !exposeEnv && !exposeBeans && !exposeInfo && !exposePrometheus {
		return nil
	}

	// 优先尝试挂载到 Gin 服务器
	if err := s.registerToGin(ctx, exposeHealth, exposeMetrics, exposeEnv, exposeBeans, exposeInfo, exposePrometheus); err == nil {
		return nil
	}

	// 其次尝试挂载到 Hertz 服务器
	if err := s.registerToHertz(ctx, exposeHealth, exposeMetrics, exposeEnv, exposeBeans, exposeInfo, exposePrometheus); err == nil {
		return nil
	}

	// 最后使用默认的 net/http mux
	s.registerToDefaultMux(env, exposeHealth, exposeMetrics, exposeEnv, exposeBeans, exposeInfo, exposePrometheus)
	return nil
}

// Stop 停止阶段：无需特殊处理
func (s *ActuatorHttpStarter) Stop(ctx boot.ApplicationContext) error {
	return nil
}

// registerToGin 尝试将 Actuator 端点注册到 Gin 服务器
func (s *ActuatorHttpStarter) registerToGin(ctx boot.ApplicationContext, exposeHealth, exposeMetrics, exposeEnv, exposeBeans, exposeInfo, exposePrometheus bool) error {
	beans := ctx.Container().ListBeans()
	for _, bean := range beans {
		// 查找 Gin 服务器 Bean
		if strings.Contains(bean.Type, "GinServer") || strings.Contains(bean.Type, "gin.GinServer") {
			beanInstance, err := ctx.Container().Get(bean.ID)
			if err != nil {
				continue
			}

			// 获取 Gin 引擎并注册路由
			if ginServer, ok := beanInstance.(interface{ Engine() any }); ok {
				engine := ginServer.Engine()
				if engine == nil {
					continue
				}

				if ginEngine, ok := engine.(interface {
					GET(path string, handlers ...any)
					NoRoute(handlers ...any)
				}); ok {
					if exposeHealth {
						ginEngine.GET("/actuator/health", s.wrapGinHandler(s.actuator.HealthHandler))
					}
					if exposeMetrics {
						ginEngine.GET("/actuator/metrics", s.wrapGinHandler(s.actuator.MetricsHandler))
					}
					if exposeEnv {
						ginEngine.GET("/actuator/env", s.wrapGinHandler(s.actuator.EnvHandler))
					}
					if exposeBeans {
						ginEngine.GET("/actuator/beans", s.wrapGinHandler(s.actuator.BeansHandler))
					}
					if exposeInfo {
						ginEngine.GET("/actuator/info", s.wrapGinHandler(s.actuator.InfoHandler))
					}
					if exposePrometheus {
						ginEngine.GET("/metrics", s.wrapGinHandler(s.actuator.PrometheusHandler))
					}
					return nil
				}
			}
		}
	}
	return fmt.Errorf("no gin server found")
}

// registerToHertz 尝试将 Actuator 端点注册到 Hertz 服务器
func (s *ActuatorHttpStarter) registerToHertz(ctx boot.ApplicationContext, exposeHealth, exposeMetrics, exposeEnv, exposeBeans, exposeInfo, exposePrometheus bool) error {
	beans := ctx.Container().ListBeans()
	for _, bean := range beans {
		// 查找 Hertz 服务器 Bean
		if strings.Contains(bean.Type, "HertzServer") || strings.Contains(bean.Type, "hertz.HertzServer") {
			beanInstance, err := ctx.Container().Get(bean.ID)
			if err != nil {
				continue
			}

			// 获取 Hertz 引擎并注册路由
			if hertzServer, ok := beanInstance.(interface{ Engine() any }); ok {
				engine := hertzServer.Engine()
				if engine == nil {
					continue
				}

				if hertzEngine, ok := engine.(interface {
					GET(path string, handlers ...any)
				}); ok {
					if exposeHealth {
						hertzEngine.GET("/actuator/health", s.wrapHertzHandler(s.actuator.HealthHandler))
					}
					if exposeMetrics {
						hertzEngine.GET("/actuator/metrics", s.wrapHertzHandler(s.actuator.MetricsHandler))
					}
					if exposeEnv {
						hertzEngine.GET("/actuator/env", s.wrapHertzHandler(s.actuator.EnvHandler))
					}
					if exposeBeans {
						hertzEngine.GET("/actuator/beans", s.wrapHertzHandler(s.actuator.BeansHandler))
					}
					if exposeInfo {
						hertzEngine.GET("/actuator/info", s.wrapHertzHandler(s.actuator.InfoHandler))
					}
					if exposePrometheus {
						hertzEngine.GET("/metrics", s.wrapHertzHandler(s.actuator.PrometheusHandler))
					}
					return nil
				}
			}
		}
	}
	return fmt.Errorf("no hertz server found")
}

// registerToDefaultMux 使用默认的 net/http mux 注册端点
func (s *ActuatorHttpStarter) registerToDefaultMux(env *environment.Environment, exposeHealth, exposeMetrics, exposeEnv, exposeBeans, exposeInfo, exposePrometheus bool) {
	mux := http.NewServeMux()
	if exposeHealth {
		mux.HandleFunc("/actuator/health", func(w http.ResponseWriter, r *http.Request) {
			s.actuator.HealthHandler(w, r)
		})
	}
	if exposeMetrics {
		mux.HandleFunc("/actuator/metrics", func(w http.ResponseWriter, r *http.Request) {
			s.actuator.MetricsHandler(w, r)
		})
	}
	if exposeEnv {
		mux.HandleFunc("/actuator/env", func(w http.ResponseWriter, r *http.Request) {
			s.actuator.EnvHandler(w, r)
		})
	}
	if exposeBeans {
		mux.HandleFunc("/actuator/beans", func(w http.ResponseWriter, r *http.Request) {
			s.actuator.BeansHandler(w, r)
		})
	}
	if exposeInfo {
		mux.HandleFunc("/actuator/info", func(w http.ResponseWriter, r *http.Request) {
			s.actuator.InfoHandler(w, r)
		})
	}
	if exposePrometheus {
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			s.actuator.PrometheusHandler(w, r)
		})
	}

	// 从配置读取端口，默认 8081
	port := env.GetString("actuator.port", "8081")
	go func() {
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			fmt.Printf("Actuator server error: %v\n", err)
		}
	}()
}

// wrapGinHandler 将 HTTP 处理器包装为 Gin 兼容的处理器
func (s *ActuatorHttpStarter) wrapGinHandler(handler http.HandlerFunc) any {
	return func(c any) {
		s.adaptGinContext(c, handler)
	}
}

// adaptGinContext 适配 Gin 上下文到标准 HTTP 处理器
func (s *ActuatorHttpStarter) adaptGinContext(c any, handler http.HandlerFunc) {
	type ginContext interface {
		Writer() http.ResponseWriter
		Request() *http.Request
	}

	if gc, ok := c.(ginContext); ok {
		handler(gc.Writer(), gc.Request())
	}
}

// wrapHertzHandler 将 HTTP 处理器包装为 Hertz 兼容的处理器
func (s *ActuatorHttpStarter) wrapHertzHandler(handler http.HandlerFunc) any {
	return func(ctx any, c any) {
		s.adaptHertzContext(ctx, c, handler)
	}
}

// adaptHertzContext 适配 Hertz 上下文到标准 HTTP 处理器
func (s *ActuatorHttpStarter) adaptHertzContext(ctx any, c any, handler http.HandlerFunc) {
	type hertzContext interface {
		GetRequest() *http.Request
		GetResponseWriter() http.ResponseWriter
	}

	if hc, ok := c.(hertzContext); ok {
		handler(hc.GetResponseWriter(), hc.GetRequest())
	}
}

func init() {
	boot.RegisterStarter(&ActuatorHttpStarter{})
}
