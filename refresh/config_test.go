package refresh

import (
	"log/slog"
	"testing"
	"time"
)

func TestRefreshConfig_Default(t *testing.T) {
	config := DefaultRefreshConfig()

	if !config.Enabled {
		t.Error("Expected enabled to be true by default")
	}

	if config.RefreshDelay != 100*time.Millisecond {
		t.Errorf("Expected refresh delay 100ms, got %v", config.RefreshDelay)
	}

	if config.MaxRefreshAttempts != 3 {
		t.Errorf("Expected max refresh attempts 3, got %d", config.MaxRefreshAttempts)
	}

	if config.Logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestRefreshConfig_ApplyOptions(t *testing.T) {
	config := DefaultRefreshConfig()
	customLogger := slog.Default()

	opts := []RefreshOption{
		WithRefreshEnabled(false),
		WithRefreshDelay(200 * time.Millisecond),
		WithMaxRefreshAttempts(5),
		WithRefreshLogger(customLogger),
	}

	config.ApplyOptions(opts)

	if config.Enabled {
		t.Error("Expected enabled to be false")
	}

	if config.RefreshDelay != 200*time.Millisecond {
		t.Errorf("Expected refresh delay 200ms, got %v", config.RefreshDelay)
	}

	if config.MaxRefreshAttempts != 5 {
		t.Errorf("Expected max refresh attempts 5, got %d", config.MaxRefreshAttempts)
	}

	if config.Logger != customLogger {
		t.Error("Expected custom logger")
	}
}
