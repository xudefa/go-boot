package boot

import (
	"errors"
	"testing"
)

type testAnalyzer struct{}

func (t *testAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, ErrDatabaseConnection)
}

func (t *testAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Database Connection Failed",
		Description: "Cannot connect to database",
		Action:      "Check if database is running",
		Cause:       err.Error(),
	}
}

var ErrDatabaseConnection = errors.New("dial tcp localhost:3306: connect: connection refused")

func TestFailureAnalyzer_Match(t *testing.T) {
	t.Parallel()
	registry := NewFailureAnalyzerRegistry()
	registry.Register(&testAnalyzer{})

	report := registry.Analyze(ErrDatabaseConnection)
	if report == nil {
		t.Fatal("expected failure report")
	}
	if report.Description != "Cannot connect to database" {
		t.Fatalf("expected 'Cannot connect to database', got %s", report.Description)
	}
}

func TestFailureAnalyzer_NoMatch(t *testing.T) {
	t.Parallel()
	registry := NewFailureAnalyzerRegistry()
	registry.Register(&testAnalyzer{})

	report := registry.Analyze(errors.New("unknown error"))
	if report != nil {
		t.Fatal("expected nil for unknown errors")
	}
}

func TestRegisterFailureAnalyzer(t *testing.T) {
	RegisterFailureAnalyzer(&testAnalyzer{})

	if GlobalAnalyzerRegistry() == nil {
		t.Fatal("expected global registry to exist")
	}
}
