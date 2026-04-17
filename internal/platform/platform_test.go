package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectKey(t *testing.T) {
	i := Detect()
	if i.OS != runtime.GOOS || i.Arch != runtime.GOARCH {
		t.Fatalf("Detect mismatch: %+v", i)
	}
	key := i.Key()
	if !strings.Contains(key, "/") {
		t.Fatalf("Key should contain /: %q", key)
	}
}

func TestExecSuffix(t *testing.T) {
	if got := (Info{OS: "windows"}).ExecSuffix(); got != ".exe" {
		t.Fatalf("windows suffix = %q", got)
	}
	if got := (Info{OS: "linux"}).ExecSuffix(); got != "" {
		t.Fatalf("linux suffix = %q", got)
	}
}
