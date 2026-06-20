package aop

import (
	"reflect"
	"regexp"
	"strings"
)

// PointCut 切点接口
//
// 定义AOP中用于匹配目标方法的规则.
// 切点决定了哪些类或方法需要被拦截.
//
// # 匹配流程
//
// 当一个方法被调用时,代理会通过以下步骤判断是否拦截:
//  1. 调用 MatchClass 检查目标类型是否匹配(如果类匹配器存在)
//  2. 调用 MatchMethod 检查目标方法是否匹配(如果方法匹配器存在)
//  3. 如果都匹配(或对应匹配器为nil),则该方法会被拦截
//
// # 组合切点
//
// 可以使用 MatchClassMethod 同时指定类和方法匹配条件,
// 也可以使用多个切点组合(通过多个Advisor)实现复杂的匹配逻辑.
type PointCut interface {
	// MatchClass 匹配类
	//
	// 检查给定类型是否匹配切点条件.
	// 如果返回true,表示该类型的所有方法都可能被拦截(还需通过方法匹配).
	// 如果返回false,则该类型的所有方法都不会被拦截.
	//
	// 参数:
	//   - c: 要检查的反射类型
	//
	// 返回值:
	//   - bool: 匹配返回true,否则返回false
	MatchClass(c reflect.Type) bool

	// MatchMethod 匹配方法
	//
	// 检查给定方法是否匹配切点条件.
	// 只有匹配的方法才会被代理拦截.
	//
	// 参数:
	//   - m: 要检查的反射方法
	//
	// 返回值:
	//   - bool: 匹配返回true,否则返回false
	MatchMethod(m reflect.Method) bool
}

// pointCutImpl 切点实现
type pointCutImpl struct {
	classMatcher  ClassMatcher
	methodMatcher MethodMatcher
	regexPattern  string
	regex         *regexp.Regexp
	interfaceType reflect.Type
}

func (p *pointCutImpl) MatchClass(c reflect.Type) bool {
	if p.classMatcher == nil {
		return true
	}
	return p.classMatcher(c)
}

func (p *pointCutImpl) MatchMethod(m reflect.Method) bool {
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
//   - PointCut: 匹配所有目标的切点
//
// 示例:
//
//	// 拦截所有方法
//	aop.MatchAll()
func MatchAll() PointCut {
	return &pointCutImpl{
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
//   - PointCut: 匹配给定类的切点
//
// 示例:
//
//	aop.MatchClass(func(t reflect.Type) bool {
//	    return t.Name() == "UserService"
//	})
func MatchClass(matcher ClassMatcher) PointCut {
	return &pointCutImpl{
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
//   - PointCut: 匹配给定方法的切点
//
// 示例:
//
//	aop.MatchMethod(func(m reflect.Method) bool {
//	    return m.Name == "DoSomething"
//	})
func MatchMethod(matcher MethodMatcher) PointCut {
	return &pointCutImpl{
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
//   - PointCut: 组合切点
func MatchClassMethod(classMatcher ClassMatcher, methodMatcher MethodMatcher) PointCut {
	return &pointCutImpl{
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
//   - PointCut: 匹配指定方法名的切点
//
// 示例:
//
//	// 只拦截 DoSomething 方法
//	aop.MatchByName("DoSomething")
func MatchByName(name string) PointCut {
	return &pointCutImpl{
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
//   - PointCut: 匹配指定前缀的切点
//
// 示例:
//
//	// 拦截所有以 Do 开头的方法
//	aop.MatchByNamePrefix("Do")
func MatchByNamePrefix(prefix string) PointCut {
	return &pointCutImpl{
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
//   - PointCut: 匹配正则表达式的切点
//
// 示例:
//
//	// 拦截所有以 do 或 Do 开头的方法
//	aop.MatchByRegex("(?i)^do.*")
func MatchByRegex(pattern string) PointCut {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return &pointCutImpl{
			methodMatcher: func(m reflect.Method) bool { return false },
			regexPattern:  pattern,
		}
	}
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return re.MatchString(m.Name)
		},
		regexPattern: pattern,
		regex:        re,
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
//   - PointCut: 匹配带注解方法的切点
//
// 注意:
//   - 此方法通过方法名前缀来匹配
func MatchByAnnotation(annotationType reflect.Type) PointCut {
	annotationName := annotationType.Name()
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return strings.HasPrefix(m.Name, annotationName+"_") ||
				strings.Contains(m.Name, "_"+annotationName+"_") ||
				strings.HasSuffix(m.Name, "_"+annotationName)
		},
	}
}

// MatchInterface 按接口类型匹配
//
// 匹配实现了指定接口的类型.
//
// 参数:
//   - y: 接口类型,传入接口变量即可
//
// 返回值:
//   - PointCut: 匹配实现接口的类的切点
//
// 示例:
//
//	// 拦截所有实现 ServiceInterface 接口的类
//	aop.MatchInterface((*ServiceInterface)(nil))
func MatchInterface(y any) PointCut {
	yType := reflect.TypeOf(y)
	if yType == nil {
		return &pointCutImpl{
			classMatcher:  func(t reflect.Type) bool { return false },
			methodMatcher: func(m reflect.Method) bool { return false },
		}
	}
	if yType.Kind() == reflect.Pointer {
		yType = yType.Elem()
	}

	if yType.Kind() != reflect.Interface {
		return &pointCutImpl{
			classMatcher: func(t reflect.Type) bool { return false },
		}
	}

	return &pointCutImpl{
		interfaceType: yType,
		classMatcher: func(t reflect.Type) bool {
			if t == nil {
				return false
			}
			for t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			return t.Implements(yType)
		},
	}
}
