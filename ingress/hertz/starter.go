// Package hertz 提供 Hertz 框架的 boot starter 支持。
//
// 该包允许使用配置文件自动配置 Hertz HTTP 服务器，
// 并将其注册到 go-boot 容器中。
package hertz

import (
	"github.com/xudefa/go-boot/config"
	"time"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/log"
)

const HttpClientBeanId = "hertzHttpClient"

func Starter() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	viper := cfg.Viper()

	hertzConfig := &HertzConfig{}
	if err := viper.UnmarshalKey("server.hertz", hertzConfig); err != nil {
		hertzConfig = defaultHertzConfig()
	}
	logger := log.DefaultLogger()

	container := core.New()

	server := New(container, hertzConfig, logger)
	err = container.Register(ServerBeanId, core.Bean(server))
	if err != nil {
		return err
	}

	return nil
}

func defaultHertzConfig() *HertzConfig {
	return &HertzConfig{
		HostPorts:    "0.0.0.0:8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		KeepAlive:    true,
		BathPath:     "/",
	}
}

func Register(container core.Container, hertzConfig *HertzConfig, logger log.Logger) (*Hertz, error) {
	server := New(container, hertzConfig, logger)
	return server, nil
}
