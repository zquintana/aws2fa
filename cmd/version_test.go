package cmd

import (
	"testing"
)

func TestIsLatest(t *testing.T) {
	v := "0.2-abc123"

	if isNewerVersion(parseVersion(v), "0.2") {
		t.Fatal("Version is already latest")
	}

	if false == isNewerVersion(parseVersion(v), "0.2.1") {
		t.Fatal("There is a new version available")
	}

	if isNewerVersion(parseVersion(v), "0.1") {
		t.Fatal("The current version is newer")
	}
}
