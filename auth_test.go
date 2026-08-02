package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func writeTokenCommand(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("external token command tests require a POSIX shell")
	}
	path := filepath.Join(t.TempDir(), "token-command")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o700); err != nil {
		t.Fatalf("write token command: %v", err)
	}
	return path
}

func TestGetAccessTokenUsesExternalProvider(t *testing.T) {
	resetExternalTokenCache()
	t.Cleanup(resetExternalTokenCache)
	expiresAt := time.Now().Add(time.Hour).Unix()
	command := writeTokenCommand(t, fmt.Sprintf(
		"printf '%%s\\n' '{\"access_token\":\"shared-token\",\"expires_at\":%d}'\n",
		expiresAt,
	))
	t.Setenv(externalTokenCommandEnv, command)

	token, err := GetAccessToken("unused-client-id")
	if err != nil {
		t.Fatalf("GetAccessToken returned an error: %v", err)
	}
	if token != "shared-token" {
		t.Fatalf("expected shared-token, got %q", token)
	}
}

func TestExternalProviderFailureNeverFallsBackToDeviceCode(t *testing.T) {
	resetExternalTokenCache()
	t.Cleanup(resetExternalTokenCache)
	var deviceRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		deviceRequests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalDeviceCodeURL := deviceCodeURL
	deviceCodeURL = server.URL
	t.Cleanup(func() { deviceCodeURL = originalDeviceCodeURL })
	t.Setenv(externalTokenCommandEnv, filepath.Join(t.TempDir(), "missing-command"))

	_, err := GetAccessToken("unused-client-id")
	if err == nil || !strings.Contains(err.Error(), "external authentication failed") {
		t.Fatalf("expected external authentication failure, got %v", err)
	}
	if got := deviceRequests.Load(); got != 0 {
		t.Fatalf("device flow received %d requests; expected none", got)
	}
}

func TestExternalOnlyBuildNeverReadsCacheOrStartsDeviceFlow(t *testing.T) {
	resetExternalTokenCache()
	t.Cleanup(resetExternalTokenCache)
	t.Setenv(externalTokenCommandEnv, "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	originalAuthMode := authMode
	authMode = externalOnlyAuthMode
	t.Cleanup(func() { authMode = originalAuthMode })

	var deviceRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		deviceRequests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	originalDeviceCodeURL := deviceCodeURL
	deviceCodeURL = server.URL
	t.Cleanup(func() { deviceCodeURL = originalDeviceCodeURL })

	_, err := GetAccessToken("unused-client-id")
	if err == nil || !strings.Contains(err.Error(), "external authentication required") {
		t.Fatalf("expected external-only authentication failure, got %v", err)
	}
	if got := deviceRequests.Load(); got != 0 {
		t.Fatalf("device flow received %d requests; expected none", got)
	}
}

func TestExternalProviderCachesFreshTokenInMemory(t *testing.T) {
	resetExternalTokenCache()
	t.Cleanup(resetExternalTokenCache)
	counterPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("TEAMS_TUI_GO_TEST_COUNTER", counterPath)
	expiresAt := time.Now().Add(time.Hour).Unix()
	command := writeTokenCommand(t, fmt.Sprintf(
		"printf x >> \"$TEAMS_TUI_GO_TEST_COUNTER\"\nprintf '%%s\\n' '{\"access_token\":\"cached-token\",\"expires_at\":%d}'\n",
		expiresAt,
	))
	t.Setenv(externalTokenCommandEnv, command)

	for range 2 {
		if _, err := GetValidTokenSilent("unused-client-id"); err != nil {
			t.Fatalf("GetValidTokenSilent returned an error: %v", err)
		}
	}
	calls, err := os.ReadFile(counterPath)
	if err != nil {
		t.Fatalf("read command counter: %v", err)
	}
	if string(calls) != "x" {
		t.Fatalf("expected one provider call, got %q", calls)
	}
}
