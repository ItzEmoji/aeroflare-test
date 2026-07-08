package ci

import (
	"reflect"
	"testing"
)

func TestScrapeStorePaths(t *testing.T) {
	in := "" +
		"warning: something\n" +
		"/nix/store/aaa-foo\n" +
		"# a comment /nix/store/ignored\n" +
		"  /nix/store/bbb-bar  \n" +
		"not a store path\n" +
		"\n"
	got := scrapeStorePaths(in)
	want := []string{"/nix/store/aaa-foo", "/nix/store/bbb-bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scrapeStorePaths = %v, want %v", got, want)
	}
}

func TestScrapeStorePaths_None(t *testing.T) {
	if got := scrapeStorePaths("nothing here\n"); len(got) != 0 {
		t.Errorf("expected no paths, got %v", got)
	}
}
