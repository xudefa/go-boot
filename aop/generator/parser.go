package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Parser Go 源码注解解析器
//
// 扫描 Go 源码文件，解析 @Aspect、@AopProxy、@Before 等注解，
// 提取切面、代理和通知信息。
type Parser struct {
	fset    *token.FileSet         // 文件位置信息
	aspects map[string]*AspectInfo // 切面信息映射（按结构体名索引）
	proxies map[string]*ProxyInfo  // 代理信息映射（按结构体名索引）
	funcs   map[string]*AdviceInfo // 独立函数通知映射（按函数名索引）
}

// NewParser 创建源码注解解析器
func NewParser() *Parser {
	return &Parser{
		fset:    token.NewFileSet(),
		aspects: make(map[string]*AspectInfo),
		proxies: make(map[string]*ProxyInfo),
		funcs:   make(map[string]*AdviceInfo),
	}
}

// ParseDir 递归扫描目录中的 Go 源码文件并解析注解
//
// 跳过 _test.go 和 _aop.go 文件。
func (p *Parser) ParseDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_aop.go") {
			return nil
		}
		return p.parseFile(path)
	})
}

// parseFile 解析单个 Go 源码文件
func (p *Parser) parseFile(filePath string) error {
	f, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := f.Name.Name

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			p.parseGenDecl(d, pkgName, filePath)
		case *ast.FuncDecl:
			p.parseFuncDecl(d, pkgName)
		}
	}

	return nil
}

// parseGenDecl 解析类型声明，提取 @Aspect 和 @AopProxy 注解
func (p *Parser) parseGenDecl(decl *ast.GenDecl, pkgName, filePath string) {
	if decl.Tok != token.TYPE {
		return
	}

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		if _, ok := typeSpec.Type.(*ast.StructType); !ok {
			continue
		}

		structName := typeSpec.Name.Name
		comments := decl.Doc

		if comments == nil {
			continue
		}

		for _, comment := range comments.List {
			text := strings.TrimSpace(comment.Text)
			text = strings.TrimPrefix(text, "//")
			text = strings.TrimSpace(text)

			if strings.HasPrefix(text, "@Aspect") {
				p.parseAspectAnnotation(text, structName, pkgName)
			} else if strings.HasPrefix(text, "@AopProxy") {
				p.parseProxyAnnotation(text, structName, pkgName, filePath)
			}
		}
	}
}

// parseAspectAnnotation 解析 @Aspect 注解，提取切面信息
func (p *Parser) parseAspectAnnotation(text, structName, pkgName string) {
	order := 0

	if idx := strings.Index(text, "order="); idx >= 0 {
		start := idx + 6
		end := strings.IndexAny(text[start:], ",)")
		if end >= 0 {
			if _, err := fmt.Sscanf(text[start:start+end], "%d", &order); err != nil {
				fmt.Printf("[go-boot] failed to parse aspect order from annotation: %v\n", err)
			}
		}
	}

	p.aspects[structName] = &AspectInfo{
		Name:    structName,
		Order:   order,
		Package: pkgName,
		Advices: []AdviceInfo{},
	}
}

// parseProxyAnnotation 解析 @AopProxy 注解，提取代理信息
func (p *Parser) parseProxyAnnotation(text, structName, pkgName, filePath string) {
	beanID := strings.ToLower(string(structName[0])) + structName[1:]

	if idx := strings.Index(text, "beanId="); idx >= 0 {
		start := idx + 7
		end := strings.IndexAny(text[start:], ",)")
		if end >= 0 {
			beanID = strings.Trim(text[start:start+end], `"`)
		}
	}

	p.proxies[structName] = &ProxyInfo{
		Name:     structName,
		Package:  pkgName,
		FilePath: filePath,
		Target:   structName,
		Methods:  []MethodInfo{},
		Aspects:  []AspectInfo{},
		BeanID:   beanID,
	}
}

// parseFuncDecl 解析函数声明，提取代理方法或独立通知函数
func (p *Parser) parseFuncDecl(decl *ast.FuncDecl, pkgName string) {
	if decl.Recv == nil {
		p.parseStandaloneFunc(decl, pkgName)
		return
	}

	recvType := p.resolveRecvType(decl.Recv)
	if recvType == "" {
		return
	}

	aspect, isAspect := p.aspects[recvType]
	proxy, isProxy := p.proxies[recvType]

	if !isAspect && !isProxy {
		return
	}

	methodName := decl.Name.Name

	if isProxy {
		method := p.parseMethodInfo(decl, methodName)
		proxy.Methods = append(proxy.Methods, method)
	}

	if isAspect && decl.Doc != nil {
		p.parseAspectMethod(decl, aspect, methodName, pkgName)
	}
}

// parseStandaloneFunc 解析独立函数上的通知注解
func (p *Parser) parseStandaloneFunc(decl *ast.FuncDecl, pkgName string) {
	if decl.Doc == nil {
		return
	}

	funcName := decl.Name.Name

	for _, comment := range decl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)

		adviceType := p.extractAdviceType(text)
		if adviceType == "" {
			continue
		}

		targets := p.extractTargets(text)
		if len(targets) == 0 {
			continue
		}

		p.funcs[funcName] = &AdviceInfo{
			Type:     adviceType,
			Method:   funcName,
			Targets:  targets,
			IsFunc:   true,
			FuncName: funcName,
			Package:  pkgName,
		}
	}
}

// parseAspectMethod 解析切面方法上的通知注解
func (p *Parser) parseAspectMethod(decl *ast.FuncDecl, aspect *AspectInfo, methodName, pkgName string) {
	for _, comment := range decl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)

		adviceType := p.extractAdviceType(text)
		if adviceType == "" {
			continue
		}

		targets := p.extractTargets(text)
		if len(targets) == 0 {
			continue
		}

		aspect.Advices = append(aspect.Advices, AdviceInfo{
			Type:       adviceType,
			Method:     methodName,
			Targets:    targets,
			IsFunc:     false,
			FuncName:   "",
			Package:    pkgName,
			AspectName: aspect.Name,
		})
	}
}

// parseMethodInfo 解析方法签名信息
func (p *Parser) parseMethodInfo(decl *ast.FuncDecl, methodName string) MethodInfo {
	method := MethodInfo{
		Name:     methodName,
		Receiver: p.resolveRecvType(decl.Recv),
		Exported: decl.Name.IsExported(),
	}

	if decl.Type.Params != nil {
		for _, param := range decl.Type.Params.List {
			paramType := p.exprToString(param.Type)
			for _, name := range param.Names {
				method.Params = append(method.Params, ParamInfo{
					Name: name.Name,
					Type: paramType,
				})
			}
			if len(param.Names) == 0 {
				method.Params = append(method.Params, ParamInfo{
					Name: "",
					Type: paramType,
				})
			}
		}
	}

	if decl.Type.Results != nil {
		for _, result := range decl.Type.Results.List {
			resultType := p.exprToString(result.Type)
			for _, name := range result.Names {
				method.Results = append(method.Results, ParamInfo{
					Name: name.Name,
					Type: resultType,
				})
			}
			if len(result.Names) == 0 {
				method.Results = append(method.Results, ParamInfo{
					Name: "",
					Type: resultType,
				})
			}
		}
	}

	return method
}

// resolveRecvType 解析方法接收者类型名
func (p *Parser) resolveRecvType(recv *ast.FieldList) string {
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

// exprToString 将 AST 表达式转换为类型字符串
func (p *Parser) exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + p.exprToString(t.Elt)
	case *ast.SelectorExpr:
		return p.exprToString(t.X) + "." + t.Sel.Name
	case *ast.FuncType:
		return "func"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return "map[" + p.exprToString(t.Key) + "]" + p.exprToString(t.Value)
	default:
		return ""
	}
}

// extractAdviceType 从注解文本中提取通知类型
func (p *Parser) extractAdviceType(text string) AdviceType {
	text = strings.ToLower(text)
	if strings.HasPrefix(text, "@afterreturning") {
		return AdviceAfterReturning
	}
	if strings.HasPrefix(text, "@afterthrowing") {
		return AdviceAfterThrowing
	}
	if strings.HasPrefix(text, "@before") {
		return AdviceBefore
	}
	if strings.HasPrefix(text, "@after") {
		return AdviceAfter
	}
	if strings.HasPrefix(text, "@around") {
		return AdviceAround
	}
	return ""
}

// extractTargets 从注解文本中提取目标方法列表
func (p *Parser) extractTargets(text string) []string {
	start := strings.Index(text, "(")
	if start < 0 {
		return nil
	}
	end := strings.LastIndex(text, ")")
	if end <= start {
		return nil
	}

	paramsStr := text[start+1 : end]
	targets := strings.Split(paramsStr, ",")

	var result []string
	for _, target := range targets {
		target = strings.TrimSpace(target)
		target = strings.Trim(target, `"`)
		if target != "" {
			result = append(result, target)
		}
	}

	return result
}

// GetAspects 获取所有解析到的切面信息
func (p *Parser) GetAspects() []AspectInfo {
	var result []AspectInfo
	for _, aspect := range p.aspects {
		result = append(result, *aspect)
	}
	return result
}

// GetProxies 获取所有解析到的代理信息
func (p *Parser) GetProxies() []ProxyInfo {
	var result []ProxyInfo
	for _, proxy := range p.proxies {
		result = append(result, *proxy)
	}
	return result
}

// GetFuncs 获取所有解析到的独立通知函数
func (p *Parser) GetFuncs() []AdviceInfo {
	var result []AdviceInfo
	for _, advice := range p.funcs {
		result = append(result, *advice)
	}
	return result
}
