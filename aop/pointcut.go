package aop

import (
	"reflect"
	"regexp"
	"strings"
)

// Pointcut 切点接口
//
// 定义AOP中用于匹配目标方法的规则.
// 切点决定了哪些类或方法需要被拦截.
type Pointcut interface {
	// MatchClass 匹配类
	//
	// 检查给定类型是否匹配切点条件
	MatchClass(c reflect.Type) bool
	// MatchMethod 匹配方法
	//
	// 检查给定方法是否匹配切点条件
	MatchMethod(m reflect.Method) bool
}

// pointcut 切点实现
type pointcut struct {
	classMatcher  ClassMatcher
	methodMatcher MethodMatcher
}

func (p *pointcut) MatchClass(c reflect.Type) bool {
	if p.classMatcher == nil {
		return true
	}
	return p.classMatcher(c)
}

func (p *pointcut) MatchMethod(m reflect.Method) bool {
	if p.methodMatcher == nil {
		return true
	}
	return p.methodMatcher(m)
}

// ClassMatcher 类匹配器
//
// 函数类型,接收一个反射类型,返回是否匹配
type ClassMatcher func(reflect.Type) bool

// MethodMatcher 方法匹配器
//
// 函数类型,接收一个反射方法,返回是否匹配
type MethodMatcher func(reflect.Method) bool

// MatchAll 匹配所有
//
// 返回匹配所有类和方法的切点.
//
// 返回值:
//   - Pointcut: 匹配所有目标的切点
//
// 示例:
//
//	// 拦截所有方法
//	aop.MatchAll()
func MatchAll() Pointcut {
	return &pointcut{
		classMatcher:  nil,
		methodMatcher: nil,
	}
}

// MatchClass 匹配类
//
// 返回只匹配类的切点,不匹配具体方法.
//
// 参数:
//   - matcher: 类匹配函数
//
// 返回值:
//   - Pointcut: 匹配给定类的切点
//
// 示例:
//
//	aop.MatchClass(func(t reflect.Type) bool {
//	    return t.Name() == "UserService"
//	})
func MatchClass(matcher ClassMatcher) Pointcut {
	return &pointcut{
		classMatcher:  matcher,
		methodMatcher: nil,
	}
}

// MatchMethod 匹配方法
//
// 返回只匹配方法的切点,不匹配具体类.
//
// 参数:
//   - matcher: 方法匹配函数
//
// 返回值:
//   - Pointcut: 匹配给定方法的切点
//
// 示例:
//
//	aop.MatchMethod(func(m reflect.Method) bool {
//	    return m.Name == "DoSomething"
//	})
func MatchMethod(matcher MethodMatcher) Pointcut {
	return &pointcut{
		classMatcher:  nil,
		methodMatcher: matcher,
	}
}

// MatchClassMethod 匹配类和方法的组合切点
//
// 同时指定类和方法匹配条件.
//
// 参数:
//   - classMatcher: 类匹配函数
//   - methodMatcher: 方法匹配函数
//
// 返回值:
//   - Pointcut: 组合切点
func MatchClassMethod(classMatcher ClassMatcher, methodMatcher MethodMatcher) Pointcut {
	return &pointcut{
		classMatcher:  classMatcher,
		methodMatcher: methodMatcher,
	}
}

// MatchByName 按方法名匹配
//
// 匹配指定名称的方法.
//
// 参数:
//   - name: 方法名
//
// 返回值:
//   - Pointcut: 匹配指定方法名的切点
//
// 示例:
//
//	// 只拦截 DoSomething 方法
//	aop.MatchByName("DoSomething")
func MatchByName(name string) Pointcut {
	return &pointcut{
		methodMatcher: func(m reflect.Method) bool {
			return m.Name == name
		},
	}
}

// MatchByNamePrefix 按方法名前缀匹配
//
// 匹配指定前缀的方法.
//
// 参数:
//   - prefix: 方法名前缀
//
// 返回值:
//   - Pointcut: 匹配指定前缀的切点
//
// 示例:
//
//	// 拦截所有以 Do 开头的方法
//	aop.MatchByNamePrefix("Do")
func MatchByNamePrefix(prefix string) Pointcut {
	return &pointcut{
		methodMatcher: func(m reflect.Method) bool {
			return strings.HasPrefix(m.Name, prefix)
		},
	}
}

// MatchByRegex 按正则表达式匹配
//
// 匹配符合正则表达式的方法名.
//
// 参数:
//   - pattern: 正则表达式
//
// 返回值:
//   - Pointcut: 匹配正则表达式的切点
//
// 示例:
//
//	// 拦截所有以 do 或 Do 开头的方法
//	aop.MatchByRegex("(?i)^do.*")
func MatchByRegex(pattern string) Pointcut {
	re := regexp.MustCompile(pattern)
	return &pointcut{
		methodMatcher: func(m reflect.Method) bool {
			return re.MatchString(m.Name)
		},
	}
}

// MatchByAnnotation 按注解类型匹配
//
// 匹配带有指定注解类型的方法.
//
// 参数:
//   - annotationType: 注解类型
//
// 返回值:
//   - Pointcut: 匹配带注解方法的切点
//
// 注意:
//   - 当前实现通过检查方法类型的字段来判断,可能需要调整
func MatchByAnnotation(annotationType reflect.Type) Pointcut {
	return &pointcut{
		methodMatcher: func(m reflect.Method) bool {
			if m.Type == nil {
				return false
			}
			_, has := m.Type.FieldByName(annotationType.Name())
			return has
		},
	}
}

// MatchInterface 按接口类型匹配
//
// 匹配实现了指定接口的类型.
//
// 参数:
//   - iface: 接口类型,传入接口变量即可
//
// 返回值:
//   - Pointcut: 匹配实现接口的类的切点
//
// 示例:
//
//	// 拦截所有实现 ServiceInterface 接口的类
//	aop.MatchInterface((*ServiceInterface)(nil))
func MatchInterface(iface interface{}) Pointcut {
	ifaceType := reflect.TypeOf(iface)
	if ifaceType.Kind() == reflect.Ptr {
		ifaceType = ifaceType.Elem()
	}
	return &pointcut{
		classMatcher: func(t reflect.Type) bool {
			for t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			return t.Implements(ifaceType)
		},
	}
}
