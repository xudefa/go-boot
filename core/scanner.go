package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
)

// ComponentScanner 组件扫描器
//
// 用于扫描指定包路径下的 Go 源文件，根据结构体注释中的组件标签自动注册为 Bean。
//
// # 功能说明
//
//   - 解析 Go 源码文件 (AST)，识别结构体类型
//   - 根据结构体的注释识别组件类型:
//     @Component("name") 或 @Component -> Component 类型
//     @Configuration("name") 或 @Configuration -> Configuration 类型
//     @Service("name") 或 @Service -> Service 类型
//     @Repository("name") 或 @Repository -> Repository 类型
//   - 自动创建 Bean 定义并注册到容器中
//   - 支持通过 inject 标签自动注入依赖
//
// # 组件命名规则
//
//   - 如果注释中指定了名称（如 @Service("userService")），则使用该名称作为 BeanID
//   - 如果未指定名称，则使用结构体名称首字母小写作为 BeanID
//   - 例如: UserService -> userService
//
// # 使用示例
//
//	// 在结构体注释中添加组件标签
//	// @Service("userService")
//	type UserService struct {
//	    DB *sql.DB `inject:"db"`
//	}
//
//	// 创建扫描器并扫描
//	scanner := core.NewComponentScanner("./internal")
//	if err := scanner.Scan(container); err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取 Bean
//	svc, _ := container.Get("userService")
type ComponentScanner struct {
	basePath string // 扫描的基础路径
}

// NewComponentScanner 创建组件扫描器
//
// 参数:
//   - basePath: 要扫描的基础路径（目录或文件路径）
//
// 返回值:
//   - *ComponentScanner: 组件扫描器实例
//
// 注意:
//   - basePath 会被转换为绝对路径
//   - 如果路径无效，会保留原路径并继续尝试解析
func NewComponentScanner(basePath string) *ComponentScanner {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		absPath = basePath
	}
	return &ComponentScanner{basePath: absPath}
}

// Scan 扫描包路径并将所有组件注册到容器中
//
// 扫描流程:
//  1. 遍历基础路径下的所有 .go 文件（排除 _test.go 和以 _ 开头的文件）
//  2. 解析每个文件为 AST，提取结构体定义和注释
//  3. 根据注释识别组件类型，构建 Bean 定义
//  4. 将组件注册到容器中
//
// 参数:
//   - container: 依赖注入容器
//
// 返回值:
//   - error: 扫描失败时返回错误
func (s *ComponentScanner) Scan(container Container) error {
	fset := token.NewFileSet()

	// 收集所有要扫描的 Go 源文件
	var files []string
	err := filepath.WalkDir(s.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		// 排除测试文件、以 _ 开头的文件和非法文件名
		if !strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, "_") && strings.HasSuffix(name, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", s.basePath, err)
	}

	// 解析并扫描每个文件
	for _, filePath := range files {
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", filePath, err)
		}

		if err := s.scanFile(container, file); err != nil {
			return err
		}
	}

	return nil
}

// scanFile 扫描单个文件并注册组件
//
// 参数:
//   - container: 依赖注入容器
//   - file: AST 文件节点
func (s *ComponentScanner) scanFile(container Container, file *ast.File) error {
	// 收集文件中所有结构体反射类型（有限支持：仅基本类型字段）
	typeRegistry := s.collectStructTypes(file)

	// 构建注释到类型的映射
	commentMap := s.buildCommentMap(file)

	// 遍历所有类型声明
	for _, decl := range file.Decls {
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

			// 优先使用类型自身的注释，否则从注释映射中获取
			var doc *ast.CommentGroup
			if typeSpec.Doc != nil {
				doc = typeSpec.Doc
			} else {
				comments := commentMap[typeSpec.Name.Name]
				if len(comments) > 0 {
					doc = &ast.CommentGroup{List: comments}
				}
			}

			// 解析组件信息（ID 和类型）
			beanID, compType := s.getComponentInfo(doc, typeSpec.Name.Name)
			if compType == ComponentTypeNone {
				continue
			}

			// 获取结构体的反射类型（仅 AST 重建，有限支持）
			t, ok := typeRegistry[typeSpec.Name.Name]

			// 注册 bean：优先使用反射类型创建实例，否则注册查找委托
			if ok && t != nil {
				// 有限支持：仅基本类型字段的简单结构体可通过 AST 重建
				opts := []BuilderOption{
					Factory(func(c Container) (any, error) {
						instance := reflect.New(t).Interface()
						if err := c.Inject(instance); err != nil {
							return nil, err
						}
						return instance, nil
					}, t),
					Singleton(),
				}
				if err := container.Register(beanID, opts...); err != nil {
					return fmt.Errorf("failed to register component %s: %w", beanID, err)
				}
			} else {
				// 注册懒加载定义：运行时从容器中查找已注册的实际类型实例
				structName := typeSpec.Name.Name
				if err := container.Register(beanID,
					Factory(func(c Container) (any, error) {
						for _, info := range c.ListBeans() {
							if strings.HasSuffix(info.Type, "."+structName) || info.Type == structName {
								return c.Get(info.ID)
							}
						}
						return nil, fmt.Errorf("component %s (%s): actual type not registered; register the bean instance before scanning",
							beanID, structName)
					}, nil),
					Singleton(),
				); err != nil {
					return fmt.Errorf("failed to register component %s: %w", beanID, err)
				}
				slog.Warn("component scanner: type info unavailable; uses lazy lookup",
					"beanID", beanID,
					"message", "types with non-basic fields (pointers, interfaces, external types, embedded fields) "+
						"cannot be reconstructed from AST. register the bean instance separately")
			}
		}
	}

	return nil
}

// buildCommentMap 构建类型名到注释的映射
//
// 参数:
//   - file: AST文件节点
//
// 返回值:
//   - map[string][]*ast.Comment: 类型名到注释列表的映射
func (s *ComponentScanner) buildCommentMap(file *ast.File) map[string][]*ast.Comment {
	result := make(map[string][]*ast.Comment)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// 获取该声明组的所有注释
		var allComments []*ast.Comment
		if genDecl.Doc != nil {
			allComments = genDecl.Doc.List
		}

		// 为每个类型分配注释
		typeSpecs := make([]*ast.TypeSpec, 0)
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			typeSpecs = append(typeSpecs, typeSpec)
		}

		// 简单策略：如果只有一个类型，注释归该类型
		// 如果有多个类型，假设注释顺序与类型顺序一致
		if len(typeSpecs) == 1 && typeSpecs[0].Doc == nil {
			if len(allComments) > 0 {
				result[typeSpecs[0].Name.Name] = allComments
			}
		} else if len(typeSpecs) > 1 {
			// 多个类型时，假设注释在对应类型之前
			// 简化处理：所有注释对所有类型可见，由 getComponentInfo 过滤
			for _, ts := range typeSpecs {
				if ts.Doc != nil {
					result[ts.Name.Name] = ts.Doc.List
				} else if len(allComments) > 0 {
					result[ts.Name.Name] = allComments
				}
			}
		}
	}

	return result
}

// collectStructTypes 收集文件中所有结构体类型及其反射类型
//
// 参数:
//   - file: AST文件节点
//
// 返回值:
//   - map[string]reflect.Type: 结构体名称到反射类型的映射
func (s *ComponentScanner) collectStructTypes(file *ast.File) map[string]reflect.Type {
	result := make(map[string]reflect.Type)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			t := s.buildStructType(structType)
			if t != nil {
				result[typeSpec.Name.Name] = t
			}
		}
	}

	return result
}

// buildStructType 根据AST结构体节点构建反射类型
//
// 参数:
//   - structType: AST结构体节点
//
// 返回值:
//   - reflect.Type: 反射类型,如果解析失败返回nil
func (s *ComponentScanner) buildStructType(structType *ast.StructType) reflect.Type {
	if structType == nil {
		return nil
	}

	var fields []reflect.StructField

	for _, f := range structType.Fields.List {
		if len(f.Names) == 0 {
			continue
		}

		fieldName := f.Names[0].Name

		fieldType := s.resolveFieldType(f.Type)

		if fieldType == nil {
			continue
		}

		var tag reflect.StructTag
		if f.Tag != nil {
			tag = reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
		}

		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Type: fieldType,
			Tag:  tag,
		})
	}

	if len(fields) == 0 {
		return reflect.StructOf(nil)
	}

	return reflect.StructOf(fields)
}

// resolveFieldType 解析字段类型表达式为反射类型
//
// 参数:
//   - expr: AST表达式节点
//
// 返回值:
//   - reflect.Type: 解析后的反射类型,不支持的类型返回nil
//
// 支持的类型:
//   - 基础类型(int, string, bool等)
//   - 指针类型(*T)
//   - 切片类型([]T)
//   - 映射类型(map[K]V)
//   - 接口类型(interface{})
func (s *ComponentScanner) resolveFieldType(expr ast.Expr) reflect.Type {
	switch t := expr.(type) {
	case *ast.Ident:
		return s.resolveBasicType(t.Name)
	case *ast.StarExpr:
		if innerType := s.resolveFieldType(t.X); innerType != nil {
			return reflect.PointerTo(innerType)
		}
	case *ast.ArrayType:
		if t.Len == nil {
			if elemType := s.resolveFieldType(t.Elt); elemType != nil {
				return reflect.SliceOf(elemType)
			}
		}
	case *ast.MapType:
		keyType := s.resolveFieldType(t.Key)
		valueType := s.resolveFieldType(t.Value)
		if keyType != nil && valueType != nil {
			return reflect.MapOf(keyType, valueType)
		}
	case *ast.InterfaceType:
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return reflect.TypeFor[any]()
		}
	}
	return nil
}

// resolveBasicType 解析基础类型名称为反射类型
//
// 参数:
//   - name: 类型名称
//
// 返回值:
//   - reflect.Type: 对应的反射类型,未知类型返回nil
func (s *ComponentScanner) resolveBasicType(name string) reflect.Type {
	switch name {
	case "int":
		return reflect.TypeFor[int]()
	case "int8":
		return reflect.TypeFor[int8]()
	case "int16":
		return reflect.TypeFor[int16]()
	case "int32":
		return reflect.TypeFor[int32]()
	case "int64":
		return reflect.TypeFor[int64]()
	case "uint":
		return reflect.TypeFor[uint]()
	case "uint8":
		return reflect.TypeFor[uint8]()
	case "uint16":
		return reflect.TypeFor[uint16]()
	case "uint32":
		return reflect.TypeFor[uint32]()
	case "uint64":
		return reflect.TypeFor[uint64]()
	case "float32":
		return reflect.TypeFor[float32]()
	case "float64":
		return reflect.TypeFor[float64]()
	case "bool":
		return reflect.TypeFor[bool]()
	case "string":
		return reflect.TypeFor[string]()
	case "byte":
		return reflect.TypeFor[byte]()
	case "rune":
		return reflect.TypeFor[rune]()
	default:
		return nil
	}
}

// ComponentType 组件类型枚举
//
// 用于标识结构体的组件类型,根据注释自动识别
type ComponentType string

const (
	ComponentTypeNone          ComponentType = "" // 无组件类型
	ComponentTypeComponent     ComponentType = "component"
	ComponentTypeConfiguration ComponentType = "configuration"
	ComponentTypeService       ComponentType = "service"
	ComponentTypeRepository    ComponentType = "repository"
)

// getComponentInfo 从结构体注释中获取组件信息
//
// 参数:
//   - doc: 结构体的注释组
//   - defaultName: 默认名称(结构体名)
//
// 返回值:
//   - string: 组件ID,如果无组件类型则为空
//   - ComponentType: 组件类型,如果无组件类型则返回ComponentTypeNone
//
// 注释识别规则:
//   - @Component("name") 或 @Component -> Component类型
//   - @Configuration("name") 或 @Configuration -> Configuration类型
//   - @Service("name") 或 @Service -> Service类型
//   - @Repository("name") 或 @Repository -> Repository类型
func (s *ComponentScanner) getComponentInfo(doc *ast.CommentGroup, defaultName string) (string, ComponentType) {
	if doc == nil {
		return "", ComponentTypeNone
	}

	for _, comment := range doc.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)

		// 解析标签名和参数
		tagName, name := parseComponentTag(text)
		if tagName == "" {
			continue
		}

		switch ComponentType(tagName) {
		case ComponentTypeComponent:
			if name == "" {
				name = toFirstCharLower(defaultName)
			}
			return name, ComponentTypeComponent
		case ComponentTypeConfiguration:
			if name == "" {
				name = toFirstCharLower(defaultName)
			}
			return name, ComponentTypeConfiguration
		case ComponentTypeService:
			if name == "" {
				name = toFirstCharLower(defaultName)
			}
			return name, ComponentTypeService
		case ComponentTypeRepository:
			if name == "" {
				name = toFirstCharLower(defaultName)
			}
			return name, ComponentTypeRepository
		}
	}

	return "", ComponentTypeNone
}

// parseComponentTag 解析组件标签，返回标签名和组件名
func parseComponentTag(text string) (tagName string, name string) {
	lowerText := strings.ToLower(text)

	// 支持的组件类型
	knownTags := []string{
		string(ComponentTypeComponent),
		string(ComponentTypeConfiguration),
		string(ComponentTypeService),
		string(ComponentTypeRepository),
	}

	for _, tag := range knownTags {
		if strings.HasPrefix(lowerText, "@"+tag) {
			tagName = tag
			// 解析参数中的名称
			if strings.Contains(text, "(") {
				start := strings.Index(text, "(")
				end := strings.LastIndex(text, ")")
				if start > 0 && end > start {
					params := text[start+1 : end]
					name = strings.Trim(params, `"`)
				}
			}
			return
		}

		// 支持不带@前缀的形式（如 "Component"）
		if text == tag || strings.HasPrefix(text, tag) {
			tagName = tag
			return
		}
	}

	return
}

// toFirstCharLower 将字符串首字母转为小写
//
// 参数:
//   - s: 要转换的字符串
//
// 返回值:
//   - string: 首字母小写后的字符串
func toFirstCharLower(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
