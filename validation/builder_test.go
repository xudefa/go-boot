package validation

import (
	"testing"
)

func TestRuleBuilder_BasicRules(t *testing.T) {
	rules := NewRuleBuilder().
		Required().
		Min(2).
		Max(50).
		Build()

	expected := "required,min=2,max=50"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Email(t *testing.T) {
	rules := NewRuleBuilder().
		Required().
		Email().
		Build()

	expected := "required,email"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_OneOf(t *testing.T) {
	rules := NewRuleBuilder().
		Required().
		OneOf("admin", "user", "guest").
		Build()

	expected := "required,oneof=admin user guest"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Regexp(t *testing.T) {
	rules := NewRuleBuilder().
		Regexp(`^[A-Z]+$`).
		Build()

	expected := "regexp=^[A-Z]+$"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_CustomMessage(t *testing.T) {
	rules, messages := NewRuleBuilder().
		Required().
		Min(2).
		CustomMessage("字段不能为空且长度至少为2").
		BuildWithMessages()

	if rules != "required,min=2" {
		t.Errorf("expected rules 'required,min=2', got %s", rules)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestRuleBuilder_ComplexValidation(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
		Role  string `json:"role"`
	}

	user := User{
		Name:  "ab",
		Email: "test@example.com",
		Age:   25,
		Role:  "admin",
	}

	err := ValidateStruct(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRuleBuilder_ChainedValidation(t *testing.T) {
	type Address struct {
		City    string `validate:"required,min=2" json:"city"`
		Country string `validate:"required" json:"country"`
	}

	type User struct {
		Name string `validate:"required,min=2" json:"name"`
		Age  int    `validate:"required,gte=0" json:"age"`
	}

	user := User{Name: "John", Age: 30}
	address := Address{City: "NY", Country: "US"}

	chain := NewValidatorChain().
		AddStruct(user).
		AddStruct(address)

	err := chain.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatorChain_StopOnFirstError(t *testing.T) {
	type User1 struct {
		Name string `validate:"required" json:"name"`
	}

	type User2 struct {
		Email string `validate:"required,email" json:"email"`
	}

	user1 := User1{Name: ""}
	user2 := User2{Email: "invalid"}

	chain := NewValidatorChain().
		StopOnFirstError().
		AddStruct(user1).
		AddStruct(user2)

	err := chain.Validate()
	if err == nil {
		t.Error("expected validation error")
	}

	// 应该只包含第一个结构体的错误
	if validationErrs, ok := err.(ValidationErrors); ok {
		if len(validationErrs) != 1 {
			t.Errorf("expected only first struct errors (1), got %d errors", len(validationErrs))
		}
	}
}

func TestValidatorChain_MultipleErrors(t *testing.T) {
	type User struct {
		Name  string `validate:"required" json:"name"`
		Email string `validate:"required,email" json:"email"`
	}

	user := User{Name: "", Email: "invalid"}

	chain := NewValidatorChain().
		AddStruct(user)

	err := chain.Validate()
	if err == nil {
		t.Error("expected validation error")
	}

	if validationErrs, ok := err.(ValidationErrors); ok {
		if len(validationErrs) < 2 {
			t.Errorf("expected at least 2 errors, got %d", len(validationErrs))
		}
	}
}

func TestValidatorChain_ValueValidation(t *testing.T) {
	chain := NewValidatorChain().
		AddValue("test@example.com", "required,email").
		AddValue(25, "required,gte=0,lte=150")

	err := chain.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatorChain_ValueValidationFailure(t *testing.T) {
	chain := NewValidatorChain().
		AddValue("", "required").
		AddValue(-1, "gte=0")

	err := chain.Validate()
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestRegexCache_Get(t *testing.T) {
	cache := NewRegexCache()

	pattern := `^[a-z]+$`
	re1, err := cache.Get(pattern)
	if err != nil {
		t.Fatalf("failed to get regex: %v", err)
	}

	re2, err := cache.Get(pattern)
	if err != nil {
		t.Fatalf("failed to get regex second time: %v", err)
	}

	if re1 != re2 {
		t.Error("expected same regex instance from cache")
	}

	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}
}

func TestRegexCache_MustGet(t *testing.T) {
	cache := NewRegexCache()

	re := cache.MustGet(`^[a-z]+$`)
	if re == nil {
		t.Error("expected non-nil regex")
	}

	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}
}

func TestRegexCache_MustGet_Panic(t *testing.T) {
	cache := NewRegexCache()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid pattern")
		}
	}()

	cache.MustGet(`[invalid`)
}

func TestRegexCache_Clear(t *testing.T) {
	cache := NewRegexCache()

	cache.MustGet(`^[a-z]+$`)
	cache.MustGet(`^[0-9]+$`)

	if cache.Size() != 2 {
		t.Errorf("expected cache size 2, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", cache.Size())
	}
}

func TestRegexCache_InvalidPattern(t *testing.T) {
	cache := NewRegexCache()

	_, err := cache.Get(`[invalid`)
	if err == nil {
		t.Error("expected error for invalid pattern")
	}
}

func TestRuleBuilder_URL(t *testing.T) {
	rules := NewRuleBuilder().
		URL().
		Build()

	expected := "url"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_IP(t *testing.T) {
	rules := NewRuleBuilder().
		IP().
		Build()

	expected := "ip"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Gt(t *testing.T) {
	rules := NewRuleBuilder().
		Gt(0).
		Build()

	expected := "gt=0"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Gte(t *testing.T) {
	rules := NewRuleBuilder().
		Gte(0).
		Build()

	expected := "gte=0"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Lt(t *testing.T) {
	rules := NewRuleBuilder().
		Lt(100).
		Build()

	expected := "lt=100"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Lte(t *testing.T) {
	rules := NewRuleBuilder().
		Lte(100).
		Build()

	expected := "lte=100"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestRuleBuilder_Len(t *testing.T) {
	rules := NewRuleBuilder().
		Len(10).
		Build()

	expected := "len=10"
	if rules != expected {
		t.Errorf("expected %s, got %s", expected, rules)
	}
}

func TestValidatorChain_Empty(t *testing.T) {
	chain := NewValidatorChain()

	err := chain.Validate()
	if err != nil {
		t.Errorf("expected no error for empty chain, got %v", err)
	}
}

func TestGetRegexCache(t *testing.T) {
	cache := GetRegexCache()
	if cache == nil {
		t.Error("expected non-nil default cache")
	}
}
