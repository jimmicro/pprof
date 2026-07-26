package pprof

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildFilename(t *testing.T) {
	cases := []struct {
		binaryPath string
		pid        int
		port       int
		want       string
	}{
		{"myservice", 1234, 8080, "myservice_1234_8080.pprof"},
		{"/usr/bin/myservice", 1234, 8080, "myservice_1234_8080.pprof"},
		{"./myservice", 5678, 9090, "myservice_5678_9090.pprof"},
		{"/some/deep/path/svc", 1, 12345, "svc_1_12345.pprof"},
	}
	for _, c := range cases {
		got := buildFilename(c.binaryPath, c.pid, c.port)
		if got != c.want {
			t.Errorf("buildFilename(%q, %d, %d) = %q, want %q",
				c.binaryPath, c.pid, c.port, got, c.want)
		}
	}
}

func TestWriteAddrCreatesFile(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "testsvc")
	pid := os.Getpid()
	port := 19876

	writeAddr(binary, pid, port)

	filename := buildFilename(binary, pid, port)
	fullPath := filepath.Join(filepath.Dir(binary), filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("expected file %q to exist", fullPath)
	}
}

func TestWriteAddrFileContent(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "testsvc")
	pid := os.Getpid()
	port := 19877

	writeAddr(binary, pid, port)

	filename := buildFilename(binary, pid, port)
	fullPath := filepath.Join(filepath.Dir(binary), filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	want := fmt.Sprintf("http://127.0.0.1:%d\n", port)
	if string(data) != want {
		t.Errorf("file content = %q, want %q", string(data), want)
	}
}

func TestResolvePort(t *testing.T) {
	cases := []struct {
		envVal string
		want   int
	}{
		{"6060", 6060},
		{"0", 0},
		{"65535", 65535},
		{"", 0},
		{"abc", 0},
		{"-1", 0},
		{"65536", 0},
	}
	for _, c := range cases {
		got := resolvePort(c.envVal)
		if got != c.want {
			t.Errorf("resolvePort(%q) = %d, want %d", c.envVal, got, c.want)
		}
	}
}

func TestDumpScriptFilename(t *testing.T) {
	cases := []struct {
		binaryPath string
		pid        int
		want       string
	}{
		{"myservice", 1234, "myservice_1234_profile_dump.sh"},
		{"/usr/bin/myservice", 5678, "myservice_5678_profile_dump.sh"},
	}
	for _, c := range cases {
		got := dumpScriptFilename(c.binaryPath, c.pid)
		if got != c.want {
			t.Errorf("dumpScriptFilename(%q, %d) = %q, want %q",
				c.binaryPath, c.pid, got, c.want)
		}
	}
}

func TestGenDumpScriptCreatesExecutableFile(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "testsvc")
	pid := os.Getpid()
	port := 19879

	genDumpScript(binary, pid, port)

	scriptPath := filepath.Join(dir, dumpScriptFilename(binary, pid))
	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		t.Fatalf("expected script %q to exist", scriptPath)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("script %q should be executable, mode=%v", scriptPath, info.Mode())
	}
}

func TestGenDumpScriptContent(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "testsvc")
	pid := os.Getpid()
	port := 19880

	genDumpScript(binary, pid, port)

	scriptPath := filepath.Join(dir, dumpScriptFilename(binary, pid))
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	content := string(data)

	checks := []string{
		"#!/bin/sh",
		fmt.Sprintf("127.0.0.1:%d", port),
		"/debug/pprof/goroutine",
		"/debug/pprof/heap",
		"/debug/pprof/profile",
		"/debug/pprof/trace",
		"/debug/vars",
		"/metrics",
		"--full-goroutine",
		`if [ "${FULL_GOROUTINE}" = "1" ]`,
		fmt.Sprintf("%d", pid),
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("script missing %q", want)
		}
	}
}

func TestMultipleInstancesGetDifferentFiles(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "svc")

	f1 := buildFilename(binary, 111, 8081)
	f2 := buildFilename(binary, 222, 8082)

	if f1 == f2 {
		t.Errorf("different instances should produce different filenames, both got %q", f1)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	newServeMux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /metrics = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "go_goroutines") {
		t.Error("GET /metrics should include default Go runtime metrics")
	}
}

func TestCleanupStaleArtifacts(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "svc")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	activePort := listener.Addr().(*net.TCPAddr).Port

	staleListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for stale port: %v", err)
	}
	stalePort := staleListener.Addr().(*net.TCPAddr).Port
	staleListener.Close()

	writeAddr(binary, 111, activePort)
	genDumpScript(binary, 111, activePort)
	writeAddr(binary, 222, stalePort)
	genDumpScript(binary, 222, stalePort)

	cleanupStaleArtifacts(binary, os.Getpid())

	activeFiles := []string{
		filepath.Join(dir, buildFilename(binary, 111, activePort)),
		filepath.Join(dir, dumpScriptFilename(binary, 111)),
	}
	for _, path := range activeFiles {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("active artifact %q should be preserved: %v", path, err)
		}
	}

	staleFiles := []string{
		filepath.Join(dir, buildFilename(binary, 222, stalePort)),
		filepath.Join(dir, dumpScriptFilename(binary, 222)),
	}
	for _, path := range staleFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("stale artifact %q should be removed, stat err=%v", path, err)
		}
	}
}
