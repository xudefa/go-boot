package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdviceInfo(t *testing.T) {
	advice := AdviceInfo{
		Type:    AdviceBefore,
		Method:  "Log",
		Targets: []string{"UserService.GetUser"},
		IsFunc:  false,
	}

	if advice.Type != AdviceBefore {
		t.Errorf("Expected AdviceBefore, got %v", advice.Type)
	}

	if len(advice.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(advice.Targets))
	}
}

func TestAspectInfo(t *testing.T) {
	aspect := AspectInfo{
		Name:    "LoggingAspect",
		Order:   1,
		Package: "service",
		Advices: []AdviceInfo{
			{Type: AdviceBefore, Method: "Log"},
		},
	}

	if aspect.Order != 1 {
		t.Errorf("Expected order 1, got %d", aspect.Order)
	}

	if len(aspect.Advices) != 1 {
		t.Errorf("Expected 1 advice, got %d", len(aspect.Advices))
	}
}

func TestMethodInfo(t *testing.T) {
	method := MethodInfo{
		Name:     "GetUser",
		Receiver: "*UserService",
		Params:   []ParamInfo{{Name: "id", Type: "int"}},
		Results:  []ParamInfo{{Name: "", Type: "*User"}, {Name: "", Type: "error"}},
		Exported: true,
	}

	if method.Name != "GetUser" {
		t.Errorf("Expected GetUser, got %s", method.Name)
	}

	if !method.Exported {
		t.Error("Expected exported method")
	}
}

func TestParser_ParseAspectAnnotation(t *testing.T) {
	parser := NewParser()

	text := `@Aspect(order=1)`
	parser.parseAspectAnnotation(text, "LoggingAspect", "service")

	aspect, exists := parser.aspects["LoggingAspect"]
	if !exists {
		t.Fatal("Aspect not found")
	}

	if aspect.Name != "LoggingAspect" {
		t.Errorf("Expected LoggingAspect, got %s", aspect.Name)
	}

	if aspect.Order != 1 {
		t.Errorf("Expected order 1, got %d", aspect.Order)
	}
}

func TestParser_ParseProxyAnnotation(t *testing.T) {
	parser := NewParser()

	text := `@AopProxy`
	parser.parseProxyAnnotation(text, "UserService", "service", "/path/to/service.go")

	proxy, exists := parser.proxies["UserService"]
	if !exists {
		t.Fatal("Proxy not found")
	}

	if proxy.Name != "UserService" {
		t.Errorf("Expected UserService, got %s", proxy.Name)
	}

	if proxy.BeanID != "userService" {
		t.Errorf("Expected userService, got %s", proxy.BeanID)
	}
}

func TestParser_ParseProxyAnnotationWithCustomBeanID(t *testing.T) {
	parser := NewParser()

	text := `@AopProxy(beanId="customService")`
	parser.parseProxyAnnotation(text, "UserService", "service", "/path/to/service.go")

	proxy, exists := parser.proxies["UserService"]
	if !exists {
		t.Fatal("Proxy not found")
	}

	if proxy.BeanID != "customService" {
		t.Errorf("Expected customService, got %s", proxy.BeanID)
	}
}

func TestParser_ExtractAdviceType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		text     string
		expected AdviceType
	}{
		{"@Before(\"UserService.GetUser\")", AdviceBefore},
		{"@After(\"UserService.GetUser\")", AdviceAfter},
		{"@Around(\"UserService.GetUser\")", AdviceAround},
		{"@AfterReturning(\"UserService.GetUser\")", AdviceAfterReturning},
		{"@AfterThrowing(\"UserService.GetUser\")", AdviceAfterThrowing},
		{"@Invalid(\"UserService.GetUser\")", ""},
	}

	for _, tt := range tests {
		result := parser.extractAdviceType(tt.text)
		if result != tt.expected {
			t.Errorf("extractAdviceType(%q) = %v, expected %v", tt.text, result, tt.expected)
		}
	}
}

func TestParser_ExtractTargets(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		text     string
		expected []string
	}{
		{"(\"UserService.GetUser\")", []string{"UserService.GetUser"}},
		{"(\"UserService.GetUser\", \"UserService.CreateUser\")", []string{"UserService.GetUser", "UserService.CreateUser"}},
		{"()", nil},
	}

	for _, tt := range tests {
		result := parser.extractTargets(tt.text)
		if len(result) != len(tt.expected) {
			t.Errorf("extractTargets(%q) = %v, expected %v", tt.text, result, tt.expected)
			continue
		}
		for i, target := range result {
			if target != tt.expected[i] {
				t.Errorf("extractTargets(%q)[%d] = %v, expected %v", tt.text, i, target, tt.expected[i])
			}
		}
	}
}

func TestParser_ParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @AopProxy
type UserService struct {}

func (s *UserService) GetUser(id int) (*User, error) {
	return nil, nil
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	if err := parser.parseFile(testFile); err != nil {
		t.Fatal(err)
	}

	aspects := parser.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("Expected 1 aspect, got %d", len(aspects))
	}

	proxies := parser.GetProxies()
	if len(proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(proxies))
	}

	if len(proxies[0].Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(proxies[0].Methods))
	}
}

func TestTemplateEngine_GenerateProxy(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatal(err)
	}

	data := ProxyTemplateData{
		Package:    "service",
		ProxyName:  "UserServiceProxy",
		TargetName: "UserService",
		BeanID:     "userService",
		Imports:    []string{"github.com/xudefa/go-boot/aop", "github.com/xudefa/go-boot/core"},
		Aspects: []AspectTemplateData{
			{
				MethodName:       "GetUser",
				AdviceType:       "Before",
				AdviceFunc:       "loggingAspectBeforeLog",
				Order:            1,
				AspectName:       "LoggingAspect",
				AspectMethodName: "LogBefore",
			},
		},
		Methods: []MethodTemplateData{
			{
				Name:                  "GetUser",
				ParamsStr:             "id int",
				ResultsStr:            "(*User, error)",
				ArgsList:              "id",
				HasError:              true,
				HasMultipleReturns:    true,
				HasSingleReturn:       false,
				HasSingleErrorReturn:  false,
				NoReturn:              false,
				ReturnValuesWithError: "tuple[0].(*User), err",
				ReturnValuesFromTuple: "tuple[0].(*User), nil",
				ReturnValuesFallback:  "nil, nil",
			},
		},
	}

	output, err := engine.GenerateProxy(data, true)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "type UserServiceProxy struct") {
		t.Error("Generated code should contain UserServiceProxy struct")
	}

	if !strings.Contains(output, "func (p *UserServiceProxy) GetUser(id int) (*User, error)") {
		t.Error("Generated code should contain GetUser method")
		fmt.Println("Generated output:")
		fmt.Println(output)
	}

	if !strings.Contains(output, "func init()") {
		t.Error("Generated code should contain init function")
	}
}

func TestBuildMethodTemplateData(t *testing.T) {
	method := MethodInfo{
		Name: "GetUser",
		Params: []ParamInfo{
			{Name: "id", Type: "int"},
		},
		Results: []ParamInfo{
			{Name: "", Type: "*User"},
			{Name: "", Type: "error"},
		},
		Exported: true,
	}

	data := buildMethodTemplateData(method)

	if data.Name != "GetUser" {
		t.Errorf("Expected GetUser, got %s", data.Name)
	}

	if data.ParamsStr != "id int" {
		t.Errorf("Expected 'id int', got %s", data.ParamsStr)
	}

	if data.ResultsStr != "(*User, error)" {
		t.Errorf("Expected '(*User, error)', got %s", data.ResultsStr)
	}

	if !data.HasError {
		t.Error("Expected hasError to be true")
	}
}

func TestGenerator_Generate(t *testing.T) {
	tmpDir := t.TempDir()

	serviceFile := filepath.Join(tmpDir, "service.go")
	contentStr := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @AopProxy
type UserService struct {}

func (s *UserService) GetUser(id int) (*User, error) {
	return nil, nil
}
`

	if err := os.WriteFile(serviceFile, []byte(contentStr), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Generate(tmpDir, true); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(tmpDir, "service_proxy.go")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Generated file should exist")
	}

	contentBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(contentBytes), "UserServiceProxy") {
		t.Error("Generated code should contain UserServiceProxy")
	}
}

func TestGenerator_Clean(t *testing.T) {
	tmpDir := t.TempDir()

	aopFile := filepath.Join(tmpDir, "service_proxy.go")
	if err := os.WriteFile(aopFile, []byte("// generated"), 0644); err != nil {
		t.Fatal(err)
	}

	normalFile := filepath.Join(tmpDir, "service.go")
	if err := os.WriteFile(normalFile, []byte("// normal"), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Clean(tmpDir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(aopFile); !os.IsNotExist(err) {
		t.Error("Generated file should be deleted")
	}

	if _, err := os.Stat(normalFile); os.IsNotExist(err) {
		t.Error("Normal file should not be deleted")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	registry.Register("userService", "/path/to/service_proxy.go")

	path, exists := registry.Get("userService")
	if !exists {
		t.Error("Bean should be registered")
	}

	if path != "/path/to/service_proxy.go" {
		t.Errorf("Expected /path/to/service_proxy.go, got %s", path)
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	registry.Register("userService", "/path/to/service_proxy.go")
	registry.Register("orderService", "/path/to/order_proxy.go")

	list := registry.List()

	if len(list) != 2 {
		t.Errorf("Expected 2 beans, got %d", len(list))
	}

	if list["userService"] != "/path/to/service_proxy.go" {
		t.Error("userService path mismatch")
	}
}

func TestRegistry_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	registryFile := filepath.Join(tmpDir, "registry.json")

	registry := NewRegistry()
	registry.Register("userService", "/path/to/service_proxy.go")

	if err := registry.Save(registryFile); err != nil {
		t.Fatal(err)
	}

	newRegistry := NewRegistry()
	if err := newRegistry.Load(registryFile); err != nil {
		t.Fatal(err)
	}

	path, exists := newRegistry.Get("userService")
	if !exists {
		t.Error("Bean should be loaded")
	}

	if path != "/path/to/service_proxy.go" {
		t.Errorf("Expected /path/to/service_proxy.go, got %s", path)
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	registry.Register("userService", "/path/to/service_proxy.go")
	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 beans after clear, got %d", len(list))
	}
}
