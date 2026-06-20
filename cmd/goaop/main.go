package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/xudefa/go-boot/aop/generator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "generate":
		runGenerate()
	case "clean":
		runClean()
	case "validate":
		runValidate()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("goaop - AOP code generator for Go")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  goaop <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate    Generate AOP proxy code")
	fmt.Println("  clean       Clean generated AOP proxy code")
	fmt.Println("  validate    Validate AOP annotations")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -dir <path>       Directory to scan (default: .)")
	fmt.Println("  -enable-aop       Enable AOP proxy generation (default: simple proxy)")
	fmt.Println("  -h, --help        Show help")
}

func runGenerate() {
	dir := "."
	enableAOP := false
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.StringVar(&dir, "dir", ".", "Directory to scan")
	fs.BoolVar(&enableAOP, "enable-aop", false, "Enable AOP proxy generation")
	_ = fs.Parse(os.Args[2:])

	slog.Info("goaop: generating AOP proxies", "dir", dir, "enableAOP", enableAOP)

	gen, err := generator.NewGenerator()
	if err != nil {
		slog.Error("goaop: failed to create generator", "error", err)
		os.Exit(1)
	}

	if err := gen.Generate(dir, enableAOP); err != nil {
		slog.Error("goaop: failed to generate proxies", "error", err)
		os.Exit(1)
	}

	registry := gen.GetRegistry()
	list := registry.List()
	slog.Info("goaop: generated proxies", "count", len(list))

	for beanID, path := range list {
		slog.Info("goaop: generated proxy", "bean", beanID, "path", path)
	}

	if enableAOP {
		fmt.Println("AOP proxies generated successfully!")
		fmt.Println("Build with: go build")
	} else {
		fmt.Println("Simple proxies generated successfully!")
		fmt.Println("Build with: go build")
		fmt.Println("Enable AOP with: goaop generate -dir . --enable-aop")
	}
}

func runClean() {
	dir := "."
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	fs.StringVar(&dir, "dir", ".", "Directory to clean")
	_ = fs.Parse(os.Args[2:])

	slog.Info("goaop: cleaning generated proxies", "dir", dir)

	gen, err := generator.NewGenerator()
	if err != nil {
		slog.Error("goaop: failed to create generator", "error", err)
		os.Exit(1)
	}

	if err := gen.Clean(dir); err != nil {
		slog.Error("goaop: failed to clean proxies", "error", err)
		os.Exit(1)
	}

	fmt.Println("Generated proxies cleaned successfully!")
}

func runValidate() {
	dir := "."
	enableAOP := false
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.StringVar(&dir, "dir", ".", "Directory to validate")
	fs.BoolVar(&enableAOP, "enable-aop", false, "Enable AOP proxy generation")
	_ = fs.Parse(os.Args[2:])

	slog.Info("goaop: validating annotations", "dir", dir, "enableAOP", enableAOP)

	gen, err := generator.NewGenerator()
	if err != nil {
		slog.Error("goaop: failed to create generator", "error", err)
		os.Exit(1)
	}

	if err := gen.Generate(dir, enableAOP); err != nil {
		slog.Error("goaop: validation failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("Annotations validated successfully!")
}
