package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
)

// ComponentScanner 组件扫描器
//
// 用于扫描指定包路径下的结构体,根据注释自动注册为组件
//
// 功能说明:
//   - 解析Go源码文件,识别结构体类型
//   - 根据结构体的注释识别组件类型(Component/Configuration/Service/Repository)
//   - 自动创建bean定义并注册到容器中
//
// 使用示例:
//
//	scanner := core.NewComponentScanner("./internal")
//	if err := scanner.Scan(container); err != nil {
//	    log.Fatal(err)
//	}
type ComponentScanner struct {
	basePath string
}

// NewComponentScanner 创建组件扫描器
//
// 参数:
//   - basePath: 要扫描的基础路径
//
// 返回值:
//   - *ComponentScanner: 组件扫描器实例
func NewComponentScanner(basePath string) *ComponentScanner {
	absPath, _ := filepath.Abs(basePath)
	return &ComponentScanner{basePath: absPath}
}

// Scan 扫描包路径并将所有组件注册到容器中
//
// 参数:
//   - container: 依赖注入容器
//
// 返回值:
//   - error: 扫描失败时返回错误
func (s *ComponentScanner) Scan(container Container) error {
	fset := token.NewFileSet()

	var files []string
	err := filepath.WalkDir(s.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, "_") && strings.HasSuffix(name, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", s.basePath, err)
	}

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

func (s *ComponentScanner) scanFile(container Container, file *ast.File) error {
	// 收集文件中所有结构体类型
	typeRegistry := s.collectStructTypes(file)

	// 构建注释到类型的映射
	commentMap := s.buildCommentMap(file)

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

			// 优先使用类型自身的注释
			var doc *ast.CommentGroup
			if typeSpec.Doc != nil {
				doc = typeSpec.Doc
			} else {
				// 从注释映射中获取
				comments := commentMap[typeSpec.Name.Name]
				if len(comments) > 0 {
					doc = &ast.CommentGroup{List: comments}
				}
			}

			beanID, compType := s.getComponentInfo(doc, typeSpec.Name.Name)
			if compType == ComponentTypeNone {
				continue
			}

			t := typeRegistry[typeSpec.Name.Name]

			opts := []BuilderOption{
				Factory(func(c Container) (interface{}, error) {
					instance := reflect.New(t).Interface()
					if err := c.Inject(instance); err != nil {
						return nil, err
					}
					return instance, nil
				}, t),
				Singleton(),
			}

			if err := container.Register(beanID, opts...); err != nil {
				fmt.Printf("warning: failed to register component %s: %v\n", beanID, err)
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
			return reflect.TypeOf((*interface{})(nil)).Elem()
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
		return reflect.TypeOf(int(0))
	case "int8":
		return reflect.TypeOf(int8(0))
	case "int16":
		return reflect.TypeOf(int16(0))
	case "int32":
		return reflect.TypeOf(int32(0))
	case "int64":
		return reflect.TypeOf(int64(0))
	case "uint":
		return reflect.TypeOf(uint(0))
	case "uint8":
		return reflect.TypeOf(uint8(0))
	case "uint16":
		return reflect.TypeOf(uint16(0))
	case "uint32":
		return reflect.TypeOf(uint32(0))
	case "uint64":
		return reflect.TypeOf(uint64(0))
	case "float32":
		return reflect.TypeOf(float32(0))
	case "float64":
		return reflect.TypeOf(float64(0))
	case "bool":
		return reflect.TypeOf(false)
	case "string":
		return reflect.TypeOf("")
	case "byte":
		return reflect.TypeOf(byte(0))
	case "rune":
		return reflect.TypeOf(rune(0))
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

		lowerText := strings.ToLower(text)

		if strings.HasPrefix(lowerText, "@component") {
			if strings.Contains(text, "(") {
				start := strings.Index(text, "(")
				end := strings.LastIndex(text, ")")
				if start > 0 && end > start {
					params := text[start+1 : end]
					name := strings.Trim(params, `"`)
					if name != "" {
						return name, ComponentTypeComponent
					}
				}
			}
			return toFirstCharLower(defaultName), ComponentTypeComponent
		}

		if text == "Component" {
			return toFirstCharLower(defaultName), ComponentTypeComponent
		}

		for _, tag := range []ComponentType{ComponentTypeConfiguration, ComponentTypeService, ComponentTypeRepository} {
			if strings.HasPrefix(lowerText, "@"+string(tag)) {
				var name string
				if strings.Contains(text, "(") {
					start := strings.Index(text, "(")
					end := strings.LastIndex(text, ")")
					if start > 0 && end > start {
						params := text[start+1 : end]
						name = strings.Trim(params, `"`)
					}
				}
				if name == "" {
					name = toFirstCharLower(defaultName)
				}
				return name, tag
			}

			if text == string(tag) {
				return toFirstCharLower(defaultName), tag
			}
		}
	}

	return "", ComponentTypeNone
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
