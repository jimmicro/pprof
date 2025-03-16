// Package pprof 提供了性能分析工具的 HTTP 服务支持
package pprof

import (
	"expvar"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
)

var (
	// PanicOnError 控制在遇到错误时是否触发 panic
	// 默认为 true，表示遇到错误时会 panic
	PanicOnError = true
)

func init() {
	// 在本地随机端口启动 TCP 监听器
	// 为了安全起见，只允许在本地监听
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if PanicOnError {
			panic(err)
		}
		return
	}

	// 创建新的 HTTP 路由复用器
	mux := http.NewServeMux()
	// 注册各种 pprof 处理器
	// Index 页面会显示所有可用的 profile 列表
	mux.HandleFunc("/", pprof.Index)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	// 显示程序的命令行参数
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	// CPU profile 信息
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	// 程序中的 symbol 信息
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	// 程序执行追踪信息
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	// 内存分配采样信息
	mux.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	// goroutine 阻塞事件的采样信息
	mux.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
	// 当前所有 goroutine 的堆栈信息
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	// 堆内存分配情况的采样信息
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	// 互斥锁的竞争情况的采样信息
	mux.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	// 系统线程创建情况的采样信息
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	// 导出包中公开的变量
	mux.Handle("/debug/vars", expvar.Handler())

	// 在新的 goroutine 中启动 HTTP 服务
	go func() {
		genPprof(l)
		_ = http.Serve(l, mux)
	}()
}

// genPprof 生成包含 pprof 服务地址的文件
// l 为监听器实例，用于获取服务的实际地址
func genPprof(l net.Listener) {
	// 获取当前运行的程序名
	binaryName := os.Args[0]
	// 创建或打开.pprof 文件
	f, err := os.OpenFile(binaryName+".pprof", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	// 将服务地址写入文件
	content := l.Addr().String() + "\n"
	_, _ = f.Write([]byte(content))
}
