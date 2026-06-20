package condition

import (
	"strings"
	"testing"
)

type mockConditionContext struct {
	envFn      func(string) (any, bool)
	hasBeanFn  func(string) bool
	hasClassFn func(string) bool
}

func (m *mockConditionContext) Environment() interface{ GetProperty(key string) (any, bool) } {
	return envGetter{m.envFn}
}

func (m *mockConditionContext) Container() interface{ Has(id string) bool } {
	return containerChecker{m.hasBeanFn}
}

func (m *mockConditionContext) ClassLoader() interface{ HasClass(name string) bool } {
	return classChecker{m.hasClassFn}
}

func (m *mockConditionContext) GetBean(beanID string) (any, bool) {
	return nil, m.hasBeanFn(beanID)
}

func (m *mockConditionContext) HasProperty(key string) bool {
	_, ok := m.envFn(key)
	return ok
}

func (m *mockConditionContext) GetProperty(key string) (any, bool) {
	return m.envFn(key)
}

type envGetter struct{ fn func(string) (any, bool) }

func (e envGetter) GetProperty(key string) (any, bool) { return e.fn(key) }

type containerChecker struct{ fn func(string) bool }

func (c containerChecker) Has(id string) bool { return c.fn(id) }

type classChecker struct{ fn func(string) bool }

func (c classChecker) HasClass(name string) bool { return c.fn(name) }

// TestOnProperty 验证 OnProperty 条件的行为：
//  1. 属性存在时匹配
//  2. 属性不存在时不匹配
//  3. 属性值与指定值相同时匹配
//  4. 属性值与指定值不同时不匹配
func TestOnProperty(t *testing.T) {
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "server.enabled" {
				return "true", true
			}
			return nil, false
		},
	}

	c := OnProperty("server.enabled")
	if !c.Matches(ctx) {
		t.Fatal("expected condition to match when property exists")
	}

	c2 := OnProperty("nonexistent")
	if c2.Matches(ctx) {
		t.Fatal("expected condition to not match when property missing")
	}

	c3 := OnProperty("server.enabled", "true")
	if !c3.Matches(ctx) {
		t.Fatal("expected condition to match when property equals 'true'")
	}

	c4 := OnProperty("server.enabled", "false")
	if c4.Matches(ctx) {
		t.Fatal("expected condition to not match when property doesn't equal")
	}
}

// TestOnMissingProperty 验证 OnMissingProperty 条件的行为：
//  1. 属性不存在时匹配
//  2. 属性存在时不匹配
func TestOnMissingProperty(t *testing.T) {
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "exists" {
				return "val", true
			}
			return nil, false
		},
	}

	if OnMissingProperty("missing").Matches(ctx) != true {
		t.Fatal("OnMissingProperty should match for missing key")
	}
	if OnMissingProperty("exists").Matches(ctx) != false {
		t.Fatal("OnMissingProperty should not match for existing key")
	}
}

// TestOnBean 验证 OnBean 和 OnMissingBean 条件的行为：
//  1. OnBean 在 Bean 存在时匹配，不存在时不匹配
//  2. OnMissingBean 在 Bean 不存在时匹配，存在时不匹配
func TestOnBean(t *testing.T) {
	ctx := &mockConditionContext{
		hasBeanFn: func(id string) bool {
			return id == "db"
		},
	}

	if OnBean("db").Matches(ctx) != true {
		t.Fatal("OnBean should match existing bean")
	}
	if OnBean("cache").Matches(ctx) != false {
		t.Fatal("OnBean should not match missing bean")
	}
	if OnMissingBean("cache").Matches(ctx) != true {
		t.Fatal("OnMissingBean should match missing bean")
	}
	if OnMissingBean("db").Matches(ctx) != false {
		t.Fatal("OnMissingBean should not match existing bean")
	}
}

// TestOnClass 验证 OnClass 和 OnMissingClass 条件的行为：
//  1. OnClass 在类存在时匹配，不存在时不匹配
//  2. OnMissingClass 在类不存在时匹配
func TestOnClass(t *testing.T) {
	ctx := &mockConditionContext{
		hasClassFn: func(name string) bool {
			return name == "gin.Engine"
		},
	}

	if OnClass("gin.Engine").Matches(ctx) != true {
		t.Fatal("OnClass should match existing class")
	}
	if OnClass("hertz.Engine").Matches(ctx) != false {
		t.Fatal("OnClass should not match missing class")
	}
	if OnMissingClass("hertz.Engine").Matches(ctx) != true {
		t.Fatal("OnMissingClass should match missing class")
	}
}

// TestCompositeConditions 验证复合条件 All、Any、Not 的逻辑运算行为：
//  1. All: 所有子条件匹配时返回 true（与运算）
//  2. Any: 任一子条件匹配时返回 true（或运算）
//  3. Not: 对子条件结果取反（非运算）
func TestCompositeConditions(t *testing.T) {
	trueCtx := &mockConditionContext{
		envFn: func(key string) (any, bool) { return "true", true },
	}
	falseCtx := &mockConditionContext{
		envFn: func(key string) (any, bool) { return nil, false },
	}

	cond1 := OnProperty("key1")
	cond2 := OnProperty("key2")

	all := All(cond1, cond2)
	if !all.Matches(trueCtx) {
		t.Fatal("All should match when both match")
	}
	if all.Matches(falseCtx) {
		t.Fatal("All should not match when none match")
	}

	any := Any(cond1, cond2)
	if !any.Matches(trueCtx) {
		t.Fatal("Any should match when at least one matches")
	}

	not := Not(cond1)
	if not.Matches(falseCtx) != true {
		t.Fatal("Not should invert false to true")
	}
	if not.Matches(trueCtx) != false {
		t.Fatal("Not should invert true to false")
	}
}

type profileEnvGetter struct {
	fn func(string) (any, bool)
}

func (p profileEnvGetter) GetProperty(key string) (any, bool) {
	return p.fn(key)
}

func (p profileEnvGetter) AcceptsProfile(profile string) bool {
	negate := false
	check := profile
	if strings.HasPrefix(profile, "!") {
		negate = true
		check = profile[1:]
	}
	if check == "dev" {
		return !negate
	}
	return negate
}

type profileMockCtx struct {
	mockConditionContext
}

func (p *profileMockCtx) Environment() interface{ GetProperty(key string) (any, bool) } {
	return profileEnvGetter{fn: p.envFn}
}

// TestOnProfileWithAcceptor 验证 OnProfile 在支持 AcceptsProfile 的环境中的行为：
//  1. OnProfile("dev") 在环境接受 "dev" profile 时匹配
//  2. OnProfile("prod") 在环境不接受 "prod" 时不匹配
//  3. OnProfile("!dev") 在 "dev" 活跃时不应匹配（双重否定不应误判）
//  4. OnProfile("!prod") 在 "prod" 不活跃时匹配
func TestOnProfileWithAcceptor(t *testing.T) {
	ctx := &profileMockCtx{
		mockConditionContext: mockConditionContext{
			envFn: func(key string) (any, bool) { return nil, false },
		},
	}

	if !OnProfile("dev").Matches(ctx) {
		t.Fatal("expected OnProfile('dev') to match when env accepts 'dev'")
	}
	if OnProfile("prod").Matches(ctx) {
		t.Fatal("expected OnProfile('prod') to not match when env doesn't accept 'prod'")
	}
	if OnProfile("!dev").Matches(ctx) {
		t.Fatal("expected OnProfile('!dev') to NOT match when 'dev' IS active (double negation bug)")
	}
	if !OnProfile("!prod").Matches(ctx) {
		t.Fatal("expected OnProfile('!prod') to match when 'prod' is NOT active")
	}
}
