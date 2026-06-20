package config

import (
	"path/filepath"
	"testing"
)

func TestConfig_SetAndGet(t *testing.T) {
	c := NewConfig()

	c.Set("name", "test")
	if c.GetString("name") != "test" {
		t.Errorf("expected 'test', got %s", c.GetString("name"))
	}

	c.Set("count", 42)
	if c.GetInt("count") != 42 {
		t.Errorf("expected 42, got %d", c.GetInt("count"))
	}

	c.Set("enabled", true)
	if !c.GetBool("enabled") {
		t.Error("expected true")
	}
}

func TestConfig_GetDefaults(t *testing.T) {
	c := NewConfig()

	if c.GetString("missing") != "" {
		t.Error("expected empty string for missing key")
	}
	if c.GetInt("missing") != 0 {
		t.Error("expected 0 for missing key")
	}
	if c.GetBool("missing") {
		t.Error("expected false for missing key")
	}
}

func TestConfig_LoadAndSave(t *testing.T) {
	c := NewConfig()
	c.Set("key1", "value1")
	c.Set("key2", 123)

	// 创建临时文件
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	// 保存
	err := c.Save(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 加载到新配置
	c2 := NewConfig()
	err = c2.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c2.GetString("key1") != "value1" {
		t.Errorf("expected 'value1', got %s", c2.GetString("key1"))
	}
	// JSON 将整数解析为 float64
	if int(c2.Get("key2").(float64)) != 123 {
		t.Errorf("expected 123, got %v", c2.Get("key2"))
	}
}

func TestConfig_LoadNonExistent(t *testing.T) {
	c := NewConfig()
	err := c.Load("/non/existent/path.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
