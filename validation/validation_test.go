package validation

import (
	"net/http"
	"strings"
	"testing"
)

// TestBasicValidation 测试基本验证功能
func TestBasicValidation(t *testing.T) {
	validator := NewTagValidator()

	// 测试结构体验证
	type TestStruct struct {
		Field string `validate:"required,min=3"`
	}

	testObj := TestStruct{Field: "test"}
	errs := validator.Validate(testObj)
	if errs != nil {
		t.Errorf("预期没有验证错误，但得到: %v", errs)
	}

	testObj.Field = "ab" // 长度太短
	errs = validator.Validate(testObj)
	if errs == nil {
		t.Error("预期因字段太短而出现验证错误，但没有得到错误")
	}
}

// TestRequiredValidation 测试必需字段验证
func TestRequiredValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		Name string `validate:"required"`
		Age  int    `validate:"required"`
	}

	// 测试空字符串
	obj1 := TestStruct{Name: "", Age: 25}
	err := validator.Validate(obj1)
	if err == nil {
		t.Error("预期因Name为空而出现验证错误")
	}

	// 测试零值整数
	obj2 := TestStruct{Name: "张三", Age: 0}
	err = validator.Validate(obj2)
	if err == nil {
		t.Error("预期因Age为零而出现验证错误")
	}

	// 测试正常值
	obj3 := TestStruct{Name: "张三", Age: 25}
	err = validator.Validate(obj3)
	if err != nil {
		t.Errorf("预期验证成功，但得到错误: %v", err)
	}
}

// TestMinValidation 测试最小值验证
func TestMinValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		Name string `validate:"min=3"`
		Age  int    `validate:"min=18"`
	}

	// 测试字符串长度不足
	obj1 := TestStruct{Name: "ab", Age: 25}
	err := validator.Validate(obj1)
	if err == nil {
		t.Error("预期因Name长度不足而出现验证错误")
	}

	// 测试数值过小
	obj2 := TestStruct{Name: "张三丰", Age: 15}
	err = validator.Validate(obj2)
	if err == nil {
		t.Error("预期因Age过小而出现验证错误")
	}

	// 测试正常值
	obj3 := TestStruct{Name: "张三丰", Age: 25}
	err = validator.Validate(obj3)
	if err != nil {
		t.Errorf("预期验证成功，但得到错误: %v", err)
	}
}

// TestMaxValidation 测试最大值验证
func TestMaxValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		Name string `validate:"max=10"`
		Age  int    `validate:"max=100"`
	}

	// 测试字符串长度超出
	obj1 := TestStruct{Name: "这是一个非常长的名字超过了限制", Age: 25}
	err := validator.Validate(obj1)
	if err == nil {
		t.Error("预期因Name长度超出而出现验证错误")
	}

	// 测试数值过大
	obj2 := TestStruct{Name: "张三", Age: 150}
	err = validator.Validate(obj2)
	if err == nil {
		t.Error("预期因Age过大而出现验证错误")
	}

	// 测试正常值
	obj3 := TestStruct{Name: "张三丰", Age: 25}
	err = validator.Validate(obj3)
	if err != nil {
		t.Errorf("预期验证成功，但得到错误: %v", err)
	}
}

// TestEmailValidation 测试邮箱验证
func TestEmailValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		Email string `validate:"email"`
	}

	validEmails := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
	}

	for _, email := range validEmails {
		obj := TestStruct{Email: email}
		err := validator.Validate(obj)
		if err != nil {
			t.Errorf("预期邮箱 %s 验证成功，但得到错误: %v", email, err)
		}
	}

	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"user@",
		"user.example.com",
	}

	for _, email := range invalidEmails {
		obj := TestStruct{Email: email}
		err := validator.Validate(obj)
		if err == nil {
			t.Errorf("预期邮箱 %s 验证失败，但没有得到错误", email)
		}
	}
}

// TestURLValidation 测试URL验证
func TestURLValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		URL string `validate:"url"`
	}

	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"https://example.com/path",
		"http://example.com:8080/path?query=value",
	}

	for _, url := range validURLs {
		obj := TestStruct{URL: url}
		err := validator.Validate(obj)
		if err != nil {
			t.Errorf("预期URL %s 验证成功，但得到错误: %v", url, err)
		}
	}

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com", // 不支持FTP协议
		"http://",
		"example.com",
	}

	for _, url := range invalidURLs {
		obj := TestStruct{URL: url}
		err := validator.Validate(obj)
		if err == nil {
			t.Errorf("预期URL %s 验证失败，但没有得到错误", url)
		}
	}
}

// TestIPValidation 测试IP地址验证
func TestIPValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		IP string `validate:"ip"`
	}

	validIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"127.0.0.1",
		"255.255.255.255",
	}

	for _, ip := range validIPs {
		obj := TestStruct{IP: ip}
		err := validator.Validate(obj)
		if err != nil {
			t.Errorf("预期IP %s 验证成功，但得到错误: %v", ip, err)
		}
	}

	invalidIPs := []string{
		"256.1.1.1",
		"192.168.1",
		"192.168.1.1.1",
		"not-an-ip",
		"999.999.999.999",
	}

	for _, ip := range invalidIPs {
		obj := TestStruct{IP: ip}
		err := validator.Validate(obj)
		if err == nil {
			t.Errorf("预期IP %s 验证失败，但没有得到错误", ip)
		}
	}
}

// TestOneOfValidation 测试选项验证
func TestOneOfValidation(t *testing.T) {
	validator := NewTagValidator()

	type TestStruct struct {
		Role string `validate:"oneof=admin user guest"`
	}

	validRoles := []string{"admin", "user", "guest"}

	for _, role := range validRoles {
		obj := TestStruct{Role: role}
		err := validator.Validate(obj)
		if err != nil {
			t.Errorf("预期角色 %s 验证成功，但得到错误: %v", role, err)
		}
	}

	invalidRoles := []string{"superadmin", "moderator", "owner"}

	for _, role := range invalidRoles {
		obj := TestStruct{Role: role}
		err := validator.Validate(obj)
		if err == nil {
			t.Errorf("预期角色 %s 验证失败，但没有得到错误", role)
		}
	}
}

// TestBindingPackage 测试绑定功能
func TestBindingPackage(t *testing.T) {
	validator := NewTagValidator()
	binder := NewDefaultBinder(validator)

	// 创建带有查询参数的简单请求
	req, _ := http.NewRequest("GET", "/test?field=value&number=42", nil)

	// 测试要绑定到的结构体
	type TestStruct struct {
		Field  string `json:"field" form:"field"`
		Number int    `json:"number" form:"number"`
	}

	obj := &TestStruct{}
	err := binder.Bind(req, obj)
	if err != nil {
		t.Errorf("预期没有绑定错误，但得到: %v", err)
	}

	if obj.Field != "value" {
		t.Errorf("预期Field为'value'，但得到: '%s'", obj.Field)
	}

	if obj.Number != 42 {
		t.Errorf("预期Number为42，但得到: %d", obj.Number)
	}
}

// TestBindAndValidate 测试绑定和验证功能
func TestBindAndValidate(t *testing.T) {
	// 创建带有JSON主体的简单请求
	jsonData := `{"field":"test","age":25}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 测试要绑定和验证的结构体
	type TestStruct struct {
		Field string `json:"field" validate:"required,min=3"`
		Age   int    `json:"age" validate:"required,min=1,max=120"`
	}

	obj := &TestStruct{}
	err := BindAndValidate(req, obj)
	if err != nil {
		t.Errorf("预期没有绑定/验证错误，但得到: %v", err)
	}

	if obj.Field != "test" {
		t.Errorf("预期Field为'test'，但得到: '%s'", obj.Field)
	}

	if obj.Age != 25 {
		t.Errorf("预期Age为25，但得到: %d", obj.Age)
	}
}

// TestValidateStruct 测试结构体验证便捷函数
func TestValidateStruct(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required,min=2"`
		Age  int    `validate:"required,min=1,max=120"`
	}

	// 测试有效数据
	validObj := TestStruct{Name: "张三", Age: 25}
	err := ValidateStruct(validObj)
	if err != nil {
		t.Errorf("预期验证成功，但得到错误: %v", err)
	}

	// 测试无效数据
	invalidObj := TestStruct{Name: "张", Age: 25} // 名字太短
	err = ValidateStruct(invalidObj)
	if err == nil {
		t.Error("预期验证失败，但没有得到错误")
	}
}

// TestValidateSingleValue 测试单值验证
func TestValidateSingleValue(t *testing.T) {
	// 测试邮箱验证
	err := Validate("test@example.com", "email")
	if err != nil {
		t.Errorf("预期邮箱验证成功，但得到错误: %v", err)
	}

	err = Validate("invalid-email", "email")
	if err == nil {
		t.Error("预期邮箱验证失败，但没有得到错误")
	}

	// 测试范围验证
	err = Validate(25, "min=1,max=100")
	if err != nil {
		t.Errorf("预期数字验证成功，但得到错误: %v", err)
	}

	err = Validate(-5, "min=1,max=100")
	if err == nil {
		t.Error("预期数字验证失败，但没有得到错误")
	}
}

// TestJSONBinder 测试JSON绑定器
func TestJSONBinder(t *testing.T) {
	jsonData := `{"name":"张三","email":"zhangsan@example.com"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	type User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
	}

	user := &User{}
	binder := NewJSONBinder(nil)
	err := binder.BindJSON(req, user)
	if err != nil {
		t.Errorf("预期JSON绑定成功，但得到错误: %v", err)
	}

	if user.Name != "张三" {
		t.Errorf("预期Name为'张三'，但得到: '%s'", user.Name)
	}

	if user.Email != "zhangsan@example.com" {
		t.Errorf("预期Email为'zhangsan@example.com'，但得到: '%s'", user.Email)
	}
}

// TestFormBinder 测试表单绑定器
func TestFormBinder(t *testing.T) {
	formData := "name=李四&age=30"
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	type User struct {
		Name string `form:"name" validate:"required"`
		Age  int    `form:"age" validate:"min=1,max=120"`
	}

	user := &User{}
	binder := NewFormBinder(nil)
	err := binder.BindForm(req, user)
	if err != nil {
		t.Errorf("预期表单绑定成功，但得到错误: %v", err)
	}

	if user.Name != "李四" {
		t.Errorf("预期Name为'李四'，但得到: '%s'", user.Name)
	}

	if user.Age != 30 {
		t.Errorf("预期Age为30，但得到: %d", user.Age)
	}
}

// TestQueryBinder 测试查询参数绑定器
func TestQueryBinder(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test?username=admin&role=user", nil)

	type Params struct {
		Username string `form:"username" validate:"required"`
		Role     string `form:"role" validate:"oneof=admin user guest"`
	}

	params := &Params{}
	binder := NewQueryBinder(nil)
	err := binder.BindQuery(req, params)
	if err != nil {
		t.Errorf("预期查询参数绑定成功，但得到错误: %v", err)
	}

	if params.Username != "admin" {
		t.Errorf("预期Username为'admin'，但得到: '%s'", params.Username)
	}

	if params.Role != "user" {
		t.Errorf("预期Role为'user'，但得到: '%s'", params.Role)
	}
}
