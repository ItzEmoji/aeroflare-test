package ci

import (
	"strings"
	"testing"
)

func TestSummaryLine_Success(t *testing.T) {
	got := summaryLine(2, 2, 4, 4, 12, 3)
	if !strings.Contains(got, "builds 2/2") || !strings.Contains(got, "pushes 4/4") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "uploaded 12") || !strings.Contains(got, "skipped-upstream 3") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "OK") {
		t.Errorf("expected OK marker, got %q", got)
	}
}

func TestSummaryLine_Failure(t *testing.T) {
	got := summaryLine(2, 1, 2, 1, 5, 0)
	if !strings.Contains(got, "FAILED") {
		t.Errorf("expected FAILED marker, got %q", got)
	}
}
