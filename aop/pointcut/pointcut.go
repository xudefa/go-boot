package pointcut

import (
	"reflect"
	"regexp"
	"strings"
)

// Pointcut 切点接口
type Pointcut interface {
	Matches(method reflect.Method) bool
	String() string
}

// namePointcut 按名称匹配
type namePointcut struct {
	pattern string
}

// ByName 按名称匹配切点
func ByName(pattern string) Pointcut {
	return &namePointcut{pattern: pattern}
}

func (p *namePointcut) Matches(method reflect.Method) bool {
	matched, _ := regexp.MatchString(p.pattern, method.Name)
	return matched
}

func (p *namePointcut) String() string {
	return "ByName(" + p.pattern + ")"
}

// prefixPointcut 按前缀匹配
type prefixPointcut struct {
	prefix string
}

// ByPrefix 按前缀匹配切点
func ByPrefix(prefix string) Pointcut {
	return &prefixPointcut{prefix: prefix}
}

func (p *prefixPointcut) Matches(method reflect.Method) bool {
	return strings.HasPrefix(method.Name, p.prefix)
}

func (p *prefixPointcut) String() string {
	return "ByPrefix(" + p.prefix + ")"
}

// regexPointcut 按正则匹配
type regexPointcut struct {
	regex *regexp.Regexp
}

// ByRegex 按正则匹配切点
func ByRegex(pattern string) (Pointcut, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &regexPointcut{regex: re}, nil
}

func (p *regexPointcut) Matches(method reflect.Method) bool {
	return p.regex.MatchString(method.Name)
}

func (p *regexPointcut) String() string {
	return "ByRegex(" + p.regex.String() + ")"
}

// allPointcut 所有条件都匹配
type allPointcut struct {
	pointcuts []Pointcut
}

// All 所有切点都匹配
func All(pointcuts ...Pointcut) Pointcut {
	return &allPointcut{pointcuts: pointcuts}
}

func (p *allPointcut) Matches(method reflect.Method) bool {
	for _, pc := range p.pointcuts {
		if !pc.Matches(method) {
			return false
		}
	}
	return true
}

func (p *allPointcut) String() string {
	return "All(...)"
}

// anyPointcut 任一条件匹配
type anyPointcut struct {
	pointcuts []Pointcut
}

// Any 任一切点匹配
func Any(pointcuts ...Pointcut) Pointcut {
	return &anyPointcut{pointcuts: pointcuts}
}

func (p *anyPointcut) Matches(method reflect.Method) bool {
	for _, pc := range p.pointcuts {
		if pc.Matches(method) {
			return true
		}
	}
	return false
}

func (p *anyPointcut) String() string {
	return "Any(...)"
}
