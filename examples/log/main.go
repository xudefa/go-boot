package main

import (
	"context"
	"github.com/xudefa/go-boot/log"
	"github.com/xudefa/go-boot/log/slog"
	"github.com/xudefa/go-boot/log/zap"
	"github.com/xudefa/go-boot/log/zero"
	"time"
)

func main() {
	zapConfig := &zap.ZapConfig{
		Level:  "debug",
		Format: "json",
	}
	l := zap.NewZapAdapter(zapConfig)

	l.Debug(context.Background(), "debug msg", log.KeyValue{
		Key:   "key",
		Value: "l",
	})

	zeroConfig := &zero.ZerologConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02 15:04:05",
	}
	zl := zero.NewZerologAdapter(zeroConfig)
	zl.Debug(context.Background(), "debug msg", log.KeyValue{
		Key:   "key",
		Value: "zl",
	})

	slogConfig := &slog.Config{
		Level: "debug",
	}
	sl := slog.NewSlogLogger(slogConfig)
	sl.Debug(context.Background(), "debug msg", log.KeyValue{
		Key:   "key",
		Value: "slog",
	})
	time.Sleep(1 * time.Second)
}
