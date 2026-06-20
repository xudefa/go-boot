// Package generator 提供 AOP 代理代码生成功能。
//
// 通过解析 Go 源码中的注解（@Aspect、@AopProxy、@Before 等），
// 自动生成代理类代码，实现面向切面编程。
//
// 核心组件：
//   - Generator: 代理代码生成器
//   - Parser: Go 源码注解解析器
//   - TemplateEngine: 代码模板引擎
//   - Registry: 代理注册表
//
// 支持的通知类型：
//   - @Before: 前置通知
//   - @After: 后置通知
//   - @Around: 环绕通知
//   - @AfterReturning: 返回后通知
//   - @AfterThrowing: 异常通知
package generator

// AdviceType 通知类型
type AdviceType string

const (
	AdviceBefore         AdviceType = "before"          // 前置通知
	AdviceAfter          AdviceType = "after"           // 后置通知
	AdviceAround         AdviceType = "around"          // 环绕通知
	AdviceAfterReturning AdviceType = "after_returning" // 返回后通知
	AdviceAfterThrowing  AdviceType = "after_throwing"  // 异常通知
)

// AdviceInfo 通知信息
type AdviceInfo struct {
	Type       AdviceType // 通知类型
	Method     string     // 通知方法名
	Targets    []string   // 目标方法列表（格式：StructName.MethodName）
	IsFunc     bool       // 是否为独立函数（非结构体方法）
	FuncName   string     // 函数名（独立函数时使用）
	Package    string     // 所属包名
	AspectName string     // 所属切面名称
}

// AspectInfo 切面信息
type AspectInfo struct {
	Name    string       // 切面名称
	Order   int          // 切面优先级（值越小优先级越高）
	Package string       // 所属包名
	Advices []AdviceInfo // 通知列表
}

// MethodInfo 方法信息
type MethodInfo struct {
	Name     string      // 方法名
	Receiver string      // 接收者类型名
	Params   []ParamInfo // 参数列表
	Results  []ParamInfo // 返回值列表
	Exported bool        // 是否为导出方法
}

// ParamInfo 参数信息
type ParamInfo struct {
	Name string // 参数名
	Type string // 参数类型
}

// ProxyInfo 代理信息
type ProxyInfo struct {
	Name     string       // 代理名称（即目标结构体名）
	Package  string       // 所属包名
	FilePath string       // 源文件路径
	Target   string       // 目标结构体名
	Methods  []MethodInfo // 方法列表
	Aspects  []AspectInfo // 关联的切面列表
	BeanID   string       // Bean 标识
}
