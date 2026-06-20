package boot

import (
	"bytes"
	"strings"
	"testing"
)

func TestBanner_Print(t *testing.T) {
	var buf bytes.Buffer
	banner := NewBanner([]string{"Test Banner"})
	banner.Print(&buf, "my-app", "1.0.0", []string{"dev"})

	output := buf.String()
	if !strings.Contains(output, "my-app") {
		t.Fatal("expected app name in banner")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Fatal("expected version in banner")
	}
	if !strings.Contains(output, "dev") {
		t.Fatal("expected profile in banner")
	}
}

func TestDefaultBanner(t *testing.T) {
	var buf bytes.Buffer
	DefaultBanner.Print(&buf, "test-app", "2.0.0", []string{"prod", "test"})

	output := buf.String()
	if !strings.Contains(output, "test-app") {
		t.Fatal("expected app name in default banner")
	}
	if !strings.Contains(output, "2.0.0") {
		t.Fatal("expected version in default banner")
	}
	if !strings.Contains(output, "prod") {
		t.Fatal("expected prod profile in banner")
	}
	if !strings.Contains(output, "test") {
		t.Fatal("expected test profile in banner")
	}
}

func TestBanner_NoProfiles(t *testing.T) {
	var buf bytes.Buffer
	banner := NewBanner([]string{"Line 1", "Line 2"})
	banner.Print(&buf, "app", "1.0", []string{})

	output := buf.String()
	if !strings.Contains(output, "profiles(default)") {
		t.Fatal("expected 'profiles(default)' when no profiles provided")
	}
}
