// Package metrics 提供 Prometheus 格式的指标导出器
package metrics

import (
	"fmt"
	"io"
	"strings"
)

// PrometheusExporter Prometheus 格式指标导出器
//
// 将指标数据转换为 Prometheus 兼容的文本格式。
// 支持 Counter、Gauge、Histogram 等指标类型的导出。
type PrometheusExporter struct {
	writer io.Writer // 输出流
}

// NewPrometheusExporter 创建新的 Prometheus 导出器
func NewPrometheusExporter(writer io.Writer) Exporter {
	return &PrometheusExporter{
		writer: writer,
	}
}

// Export 将指标导出为 Prometheus 格式
func (e *PrometheusExporter) Export(metrics []Metric) error {
	for _, metric := range metrics {
		if err := e.writeMetric(metric); err != nil {
			return err
		}
	}
	return nil
}

// writeMetric 将单个指标写入输出流
func (e *PrometheusExporter) writeMetric(metric Metric) error {
	// 处理标签
	labels := ""
	if len(metric.Tags) > 0 {
		// 创建标签键的切片并排序，以确保输出顺序一致
		keys := make([]string, 0, len(metric.Tags))
		for k := range metric.Tags {
			keys = append(keys, k)
		}

		// 对键进行排序以确保一致性
		for i := 0; i < len(keys)-1; i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		pairs := make([]string, 0, len(metric.Tags))
		for _, k := range keys {
			pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, metric.Tags[k]))
		}
		labels = "{" + strings.Join(pairs, ",") + "}"
	}

	// 根据指标类型输出不同格式
	line := fmt.Sprintf("%s%s %g\n", metric.Name, labels, metric.Value)
	_, err := e.writer.Write([]byte(line))
	return err
}
