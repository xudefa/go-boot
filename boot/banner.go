package boot

import (
	"fmt"
	"io"
	"os"
)

// Banner 启动横幅接口
//
// 参考 Spring Boot 的 Banner，支持多种格式的启动横幅显示。
type Banner interface {
	Print(ctx ApplicationContext)
}

// TextBanner 文本横幅
//
// 使用模板和属性渲染启动横幅。
type TextBanner struct {
	Template   string         // 横幅文本模板
	Properties map[string]any // 模板属性键值对
}

// Print 输出文本横幅到标准输出
func (t *TextBanner) Print(ctx ApplicationContext) {
	text := t.Template
	if text == "" {
		text = defaultBannerText
	}
	fmt.Println(text)
	if len(t.Properties) > 0 {
		for k, v := range t.Properties {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	fmt.Println()
}

// ASCIIArtBanner ASCII 艺术横幅
type ASCIIArtBanner struct {
	Art   string // ASCII 艺术文本
	Color string // 显示颜色（预留）
}

// Print 输出 ASCII 艺术横幅到标准输出
func (a *ASCIIArtBanner) Print(ctx ApplicationContext) {
	art := a.Art
	if art == "" {
		art = defaultBannerText
	}
	fmt.Println(art)
	fmt.Println()
}

// CustomTemplateBanner 自定义模板横幅
//
// 支持从文件加载横幅模板。
type CustomTemplateBanner struct {
	TemplatePath string         // 模板文件路径
	Data         map[string]any // 模板数据
}

// Print 输出自定义模板横幅到标准输出
func (c *CustomTemplateBanner) Print(ctx ApplicationContext) {
	if c.TemplatePath == "" {
		fmt.Println(defaultBannerText)
		fmt.Println()
		return
	}
	fmt.Printf("Banner from: %s\n", c.TemplatePath)
	fmt.Println()
}

// LegacyBanner 旧版横幅实现
//
// 使用预定义的 ASCII 艺术行列表渲染启动横幅。
type LegacyBanner struct {
	lines []string // ASCII 艺术行列表
}

// DefaultBanner 默认横幅实例
var DefaultBanner = &LegacyBanner{
	lines: []string{
		"  ________                  ____             __",
		" /  _____/  ____    ____   / __ )  ____     / /_",
		"/   \\  ___ /  _ \\  /  _ \\ / __  | / __ \\   / __ \\",
		"\\    \\_\\  (  <_> )(  <_> ) /_/ / / /_/ /  / /_/ /",
		" \\______  /\\____/  \\____/_____/  \\____/   /___/",
		"        \\/",
	},
}

const defaultBannerText = `
   ____  __  ______    ___  _____
  / __ \/ / / / __ \  /   |/  ___/
 / /_/ / /_/ / /_/ / / /|  |\__ \
/ ____/ __  / _, _/ / ___ |/__/ /
/_/   /_/ /_/_/ |_| /_/  |_/____/`

// NewBanner 创建旧版横幅
//
// 参数：
//   - lines: ASCII 艺术行列表
func NewBanner(lines []string) *LegacyBanner {
	return &LegacyBanner{lines: lines}
}

// Print 输出横幅到指定 writer
//
// 显示 ASCII 艺术横幅，附带应用名称、版本号和激活的 Profile 信息。
func (b *LegacyBanner) Print(w io.Writer, appName, version string, profiles []string) {
	for _, line := range b.lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			fmt.Printf("[go-boot] failed to print banner line: %v\n", err)
			return
		}
	}
	profileStr := ""
	for i, p := range profiles {
		if i > 0 {
			profileStr += ", "
		}
		profileStr += p
	}
	if profileStr == "" {
		profileStr = "default"
	}
	if _, err := fmt.Fprintf(w, ":: %s :: v%s :: profiles(%s)\n\n", appName, version, profileStr); err != nil {
		fmt.Printf("[go-boot] failed to print banner info: %v\n", err)
	}
}

// PrintBanner 使用默认横幅输出到 stdout
func PrintBanner(appName, version string, profiles []string) {
	DefaultBanner.Print(os.Stdout, appName, version, profiles)
}
