package singleinstance

import (
	"errors"
	"strings"
	"testing"
)

func TestCombineNotifyExistingResults(t *testing.T) {
	signalErr := errors.New("signal failed")
	nativeErr := errors.New("native failed")

	if err := combineNotifyExistingResults(nil, nil); err != nil {
		t.Fatalf("both nil: %v", err)
	}
	if err := combineNotifyExistingResults(nil, nativeErr); err != nil {
		t.Fatalf("signal ok: %v", err)
	}
	if err := combineNotifyExistingResults(signalErr, nil); err != nil {
		t.Fatalf("native ok: %v", err)
	}

	err := combineNotifyExistingResults(signalErr, nativeErr)
	if err == nil {
		t.Fatal("expected error when both paths fail")
	}
	if !strings.Contains(err.Error(), "signal failed") || !strings.Contains(err.Error(), "native failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
