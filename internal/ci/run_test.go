package ci

import (
	"strings"
	"testing"
)

func TestSummaryLine_Success(t *testing.T) {
	got := summaryLine(2, 2, 1, 1, 417)
	if !strings.Contains(got, "builds 2/2") || !strings.Contains(got, "pushes 1/1") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "paths 417") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "OK") {
		t.Errorf("expected OK marker, got %q", got)
	}
}

func TestSummaryLine_Failure(t *testing.T) {
	got := summaryLine(2, 1, 1, 0, 5)
	if !strings.Contains(got, "FAILED") {
		t.Errorf("expected FAILED marker, got %q", got)
	}
}
