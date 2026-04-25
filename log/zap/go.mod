module github.com/xudefa/go-boot/log/zap

go 1.25.0

require (
	github.com/xudefa/go-boot/log v0.0.0-20260419045311-edf60d5b228f
	go.uber.org/zap v1.27.1
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/xudefa/go-boot/log => ../
