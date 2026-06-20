// Package schedule 提供定时任务调度框架，包含 Cron 表达式解析、任务注册和调度执行。
//
// 核心组件：
//   - Scheduler: 定时任务调度器，基于最小堆实现，支持动态注册/注销
//   - CronExpression: 6字段 Spring 风格 Cron 表达式解析器
//   - Task: 定时任务接口，包含名称、Cron 表达式和执行逻辑
//   - @Scheduled 注解扫描：通过 AST 解析源码中的注解自动注册任务
//
// 使用示例：
//
//	scheduler := schedule.NewScheduler(schedule.WithPoolSize(5))
//	scheduler.Register(schedule.NewTask("cleanup", "0 */5 * * * ?", func(ctx context.Context) error {
//	    return nil
//	}))
//	scheduler.Start(ctx)
package schedule

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/xudefa/go-boot/core"
)

// annotationInfo 保存从 Go 源码 AST 中解析出的 @Scheduled 注解信息
// 包含所属结构体名称、方法名称和 Cron 表达式
type annotationInfo struct {
	structName string
	methodName string
	cronExpr   string
}

// ScanScheduledTasks 扫描容器中所有 bean 的 @Scheduled 注解方法，返回 Task 列表
//
// 在 bean 方法上使用注解：
//
//	// @Scheduled(cron="0/5 * * * * ?")
//	func (s *MyService) Cleanup(ctx context.Context) error {
//	    return nil
//	}
//
// scanDir 参数指定要扫描的源码目录。
// 扫描过程：遍历目录 → 解析 AST → 收集注解 → 从容器查找 bean → 反射调用。
func ScanScheduledTasks(container core.Container, scanDir string) ([]Task, error) {
	annotations, err := parseAnnotationsFromDir(scanDir)
	if err != nil {
		return nil, err
	}
	if len(annotations) == 0 {
		return nil, nil
	}
	return buildTasksFromAnnotations(container, annotations)
}

// parseAnnotationsFromDir 扫描目录下所有 Go 源文件，解析 @Scheduled 注解
func parseAnnotationsFromDir(scanDir string) ([]annotationInfo, error) {
	fset := token.NewFileSet()
	files, err := walkGoFiles(scanDir)
	if err != nil {
		return nil, err
	}

	var annotations []annotationInfo
	for _, filePath := range files {
		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		structNames := collectStructNames(f)
		fileAnnotations := collectScheduledAnnotations(f, structNames)
		annotations = append(annotations, fileAnnotations...)
	}
	return annotations, nil
}

// walkGoFiles 递归遍历目录，收集所有非 _test.go 的 Go 源文件
func walkGoFiles(scanDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(scanDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, "_test.go") && strings.HasSuffix(name, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("schedule: failed to walk directory %s: %w", scanDir, err)
	}
	return files, nil
}

// buildTasksFromAnnotations 根据注解信息从容器中查找 bean 并构造 Task 列表
//
// 通过反射查找 bean 上的对应方法并包装为 Task，方法签名必须为 func(context.Context) error
func buildTasksFromAnnotations(container core.Container, annotations []annotationInfo) ([]Task, error) {
	var tasks []Task
	for _, ann := range annotations {
		beanID := toFirstCharLower(ann.structName)
		bean, err := container.Get(beanID)
		if err != nil {
			// 通过默认 ID 找不到时，遍历所有 bean 按类型名匹配
			bean = findBeanByStructName(container, ann.structName)
			if bean == nil {
				continue
			}
		}
		method := resolveBeanMethod(bean, ann.methodName)
		if !method.IsValid() {
			continue
		}
		taskName := ann.structName + "." + ann.methodName
		methodCopy := method
		task := NewTask(taskName, ann.cronExpr, func(ctx context.Context) error {
			results := methodCopy.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if len(results) == 0 {
				return nil
			}
			if err, ok := results[0].Interface().(error); ok {
				return err
			}
			return nil
		})
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// resolveBeanMethod 在 bean 上按名称查找方法，支持指针/值接收者
//
// 先尝试指针接收者方法查找，失败后再尝试值接收者方法查找
func resolveBeanMethod(bean any, methodName string) reflect.Value {
	beanVal := reflect.ValueOf(bean)
	method := beanVal.MethodByName(methodName)
	if method.IsValid() {
		return method
	}
	if beanVal.Kind() == reflect.Pointer {
		method = beanVal.Elem().MethodByName(methodName)
	}
	return method
}

// collectStructNames 从 AST 中收集所有结构体类型名称
func collectStructNames(f *ast.File) map[string]bool {
	names := make(map[string]bool)
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			_, isStruct := typeSpec.Type.(*ast.StructType)
			if !isStruct {
				continue
			}
			names[typeSpec.Name.Name] = true
		}
	}
	return names
}

// collectScheduledAnnotations 在 AST 文件中收集所有 @Scheduled 注解
//
// 匹配条件：方法是结构体的接收者方法，且有 @Scheduled(cron="...") 注释。
func collectScheduledAnnotations(f *ast.File, structNames map[string]bool) []annotationInfo {
	var result []annotationInfo

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}

		recvType := resolveRecvType(funcDecl.Recv)
		if !structNames[recvType] {
			continue
		}
		if funcDecl.Doc == nil {
			continue
		}

		for _, comment := range funcDecl.Doc.List {
			text := strings.TrimSpace(comment.Text)
			text = strings.TrimPrefix(text, "//")
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSpace(text)

			if !strings.HasPrefix(strings.ToLower(text), "@scheduled") {
				continue
			}
			cronExpr := extractCronFromAnnotation(text)
			if cronExpr == "" {
				continue
			}

			result = append(result, annotationInfo{
				structName: recvType,
				methodName: funcDecl.Name.Name,
				cronExpr:   cronExpr,
			})
		}
	}

	return result
}

// resolveRecvType 解析接收者类型名称（支持指针和值接收者）
func resolveRecvType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// extractCronFromAnnotation 从 @Scheduled 注解文本中提取 cron 表达式
//
// 支持格式:
//   - @Scheduled(cron="0/5 * * * * ?")
//   - @Scheduled(cron='0/5 * * * * ?', zone="UTC")
//
// 不正确的多参数格式之前会导致解析错误。
func extractCronFromAnnotation(text string) string {
	start := strings.Index(text, "(")
	if start < 0 {
		return ""
	}
	end := strings.LastIndex(text, ")")
	if end <= start {
		return ""
	}

	paramsStr := text[start+1 : end]

	cronPrefix := "cron="
	_, after, ok := strings.Cut(paramsStr, cronPrefix)
	if !ok {
		return ""
	}

	valPart := after
	if len(valPart) == 0 {
		return ""
	}

	quote := valPart[0]
	if quote != '"' && quote != '\'' {
		return ""
	}

	closeIdx := strings.IndexByte(valPart[1:], quote)
	if closeIdx < 0 {
		return ""
	}

	return valPart[1 : 1+closeIdx]
}

// findBeanByStructName 遍历容器按 structName 后缀匹配 bean
func findBeanByStructName(container core.Container, structName string) any {
	for _, info := range container.ListBeans() {
		if strings.HasSuffix(info.Type, "."+structName) || info.Type == structName {
			if bean, err := container.Get(info.ID); err == nil {
				return bean
			}
		}
	}
	return nil
}

// toFirstCharLower 将字符串首字母转为小写
//
// 用于将结构体类型名转为默认 bean ID（如 MyService → myService）
func toFirstCharLower(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
