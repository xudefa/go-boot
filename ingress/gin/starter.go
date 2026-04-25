// Package gin 提供 Gin 框架的 boot starter 支持。
//
// 该包允许使用配置文件自动配置 Gin HTTP 服务器，
// 并将其注册到 go-boot 容器中。
package gin

import (
	"github.com/xudefa/go-boot/config"
	"time"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/ingress"
	"github.com/xudefa/go-boot/log"
)

const HttpClientBeanId = "httpClient"

func Starter() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	viper := cfg.Viper()

	ginConfig := &GinConfig{}
	if err := viper.UnmarshalKey("server.gin", ginConfig); err != nil {
		ginConfig = defaultGinConfig()
	}

	logger := log.DefaultLogger()

	container := core.New()

	server := New(container, ginConfig, logger)
	err = container.Register(ServerBeanId, core.Bean(server))
	if err != nil {
		return err
	}

	return nil
}

func defaultGinConfig() *GinConfig {
	return &GinConfig{
		HostPorts:    ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Mode:         "debug",
	}
}

func Register(container core.Container, ginConfig *GinConfig, logger log.Logger) (ingress.Server, error) {
	server := New(container, ginConfig, logger)
	return server, nil
}
