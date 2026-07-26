// Package pprof 提供了性能分析工具的 HTTP 服务支持
package pprof

import (
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"path/filepath"
	runtimepprof "runtime/pprof"
	"strconv"
	"strings"
	"time"
)

// PanicOnError 控制在遇到错误时是否触发 panic
// 默认为 true，表示遇到错误时会 panic
var PanicOnError = true

func init() {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", resolvePort(os.Getenv("PPROF_PORT"))))
	if err != nil {
		log.Println("pprof server start failed:", err)
		if PanicOnError {
			panic(err)
		}
		return
	}

	port := l.Addr().(*net.TCPAddr).Port
	pid := os.Getpid()
	cleanupStaleArtifacts(os.Args[0], pid)
	writeAddr(os.Args[0], pid, port)
	go genDumpScript(os.Args[0], pid, port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", pprof.Index)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	mux.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	mux.Handle("/debug/vars", expvar.Handler())
	registerHeapHandlers(mux)

	go func() {
		if err := http.Serve(l, mux); err != nil {
			log.Println("pprof server stopped:", err)
			if PanicOnError {
				panic(err)
			}
		}
	}()
}

// resolvePort 解析 PPROF_PORT 环境变量值，无效或未设置时返回 0（随机端口）
func resolvePort(envVal string) int {
	if envVal == "" {
		return 0
	}
	n, err := strconv.Atoi(envVal)
	if err != nil || n < 0 || n > 65535 {
		return 0
	}
	return n
}

// buildFilename 根据 binary 路径、pid 和 port 生成唯一的 .pprof 文件名
func buildFilename(binaryPath string, pid, port int) string {
	base := filepath.Base(binaryPath)
	return fmt.Sprintf("%s_%d_%d.pprof", base, pid, port)
}

// writeAddr 在 binary 所在目录写入 <name>_<pid>_<port>.pprof 文件，内容为 pprof 服务地址
func writeAddr(binaryPath string, pid, port int) {
	dir := filepath.Dir(binaryPath)
	filename := buildFilename(binaryPath, pid, port)
	fullPath := filepath.Join(dir, filename)

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "http://127.0.0.1:%d\n", port)
}

// cleanupStaleArtifacts 删除已停止实例遗留的地址文件和 dump 脚本。
func cleanupStaleArtifacts(binaryPath string, currentPID int) {
	dir := filepath.Dir(binaryPath)
	base := filepath.Base(binaryPath)
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return
	}

	activePIDs := map[int]bool{currentPID: true}
	for _, entry := range entries {
		pid, ok := parseAddrFilename(base, entry.Name())
		if !ok {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if addrFileAlive(path) {
			activePIDs[pid] = true
			continue
		}
		if err := os.Remove(path); err != nil {
			log.Println(err)
		}
	}

	for _, entry := range entries {
		pid, ok := parseDumpScriptFilename(base, entry.Name())
		if !ok || activePIDs[pid] {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			log.Println(err)
		}
	}
}

func parseAddrFilename(base, name string) (int, bool) {
	if !strings.HasPrefix(name, base+"_") || !strings.HasSuffix(name, ".pprof") {
		return 0, false
	}

	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(name, base+"_"), ".pprof"), "_")
	if len(parts) != 2 {
		return 0, false
	}
	pid, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return 0, false
	}
	return pid, true
}

func parseDumpScriptFilename(base, name string) (int, bool) {
	const suffix = "_profile_dump.sh"
	if !strings.HasPrefix(name, base+"_") || !strings.HasSuffix(name, suffix) {
		return 0, false
	}

	pid, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, base+"_"), suffix))
	if err != nil {
		return 0, false
	}
	return pid, true
}

func addrFileAlive(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(string(data)))
	if err != nil || u.Scheme != "http" || u.Host == "" {
		return false
	}

	conn, err := net.DialTimeout("tcp", u.Host, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// dumpScriptFilename 生成 dump 脚本文件名
func dumpScriptFilename(binaryPath string, pid int) string {
	return fmt.Sprintf("%s_%d_profile_dump.sh", filepath.Base(binaryPath), pid)
}

// genDumpScript 生成一键采集 pprof 数据的 shell 脚本
func genDumpScript(binaryPath string, pid, port int) {
	dir := filepath.Dir(binaryPath)
	filename := dumpScriptFilename(binaryPath, pid)
	fullPath := filepath.Join(dir, filename)

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	base := filepath.Base(binaryPath)
	ts := time.Now().Format("20060102_150405")
	prefix := fmt.Sprintf("%s_%d_%s", base, pid, ts)

	fmt.Fprintf(f, "#!/bin/sh\n")
	fmt.Fprintf(f, "set -x\n")
	fmt.Fprintf(f, "DIR=profile/%s\n", prefix)
	fmt.Fprintf(f, "mkdir -p \"${DIR}\"\n")
	fmt.Fprintf(f, "curl -sS '%s/debug/vars' -o \"${DIR}/%s_vars\"\n", addr, prefix)
	for _, p := range runtimepprof.Profiles() {
		name := p.Name()
		fmt.Fprintf(f, "curl -sS '%s/debug/pprof/%s' -o \"${DIR}/%s_%s\"\n", addr, name, prefix, name)
		if name == "goroutine" {
			fmt.Fprintf(f, "curl -sS '%s/debug/pprof/%s?debug=2' -o \"${DIR}/%s_%s_debug2\"\n", addr, name, prefix, name)
		}
	}
	fmt.Fprintf(f, "curl -sS '%s/debug/pprof/profile?seconds=5' -o \"${DIR}/%s_cpuprofile\"\n", addr, prefix)
	fmt.Fprintf(f, "curl -sS '%s/debug/pprof/trace?seconds=5' -o \"${DIR}/%s_trace\"\n", addr, prefix)
	fmt.Fprintf(f, "tar czf \"${DIR}.tar.gz\" \"${DIR}\"\n")
	fmt.Fprintf(f, "echo \"done: ${DIR}.tar.gz\"\n")
}
