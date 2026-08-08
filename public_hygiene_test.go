package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestTrackedFilesContainNoPrivateDeploymentMarkers(t *testing.T) {
	output, err := exec.Command("git", "ls-files", "-z").Output()
	if err != nil {
		t.Skipf("tracked-file hygiene check requires a Git checkout: %v", err)
	}

	markers := [][]byte{
		{109, 111, 98, 105, 108, 101, 121, 101},
		{47, 117, 115, 101, 114, 115, 47, 109, 100, 102, 47},
		{47, 117, 115, 101, 114, 115, 47, 109, 100, 47},
		{115, 103, 117, 105, 98, 111, 114, 64, 103, 109, 97, 105, 108, 46, 99, 111, 109},
		{109, 100, 102, 45, 116, 101, 97, 109, 115, 45, 97, 103, 101, 110, 116, 45, 115, 104, 101, 108, 108},
		{115, 112, 97, 99, 101, 99, 108, 105, 101, 110, 116},
		{64, 115, 112, 97, 99, 101, 109, 97, 99, 115},
		{109, 100, 46, 111, 98, 115, 105, 100, 105, 97, 110},
		{109, 101, 46, 111, 98, 115, 105, 100, 105, 97, 110},
		{100, 97, 118, 109, 97, 105, 108},
	}

	for _, rawPath := range bytes.Split(output, []byte{0}) {
		if len(rawPath) == 0 {
			continue
		}
		path := string(rawPath)
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read tracked file %q: %v", path, err)
		}
		lower := bytes.ToLower(contents)
		for _, marker := range markers {
			if bytes.Contains(lower, marker) {
				t.Errorf("tracked file %q contains a private deployment marker", path)
			}
		}
	}
}
