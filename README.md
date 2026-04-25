# pprof

[![Go Report Card](https://goreportcard.com/badge/github.com/jimmicro/pprof)](https://goreportcard.com/report/github.com/jimmicro/pprof)
[![codecov](https://codecov.io/gh/jimmicro/pprof/branch/main/graph/badge.svg)](https://codecov.io/gh/jimmicro/pprof)
[![License](https://img.shields.io/github/license/jimmicro/pprof)](LICENSE)
[![Release](https://img.shields.io/github/v/release/jimmicro/pprof)](https://github.com/jimmicro/pprof/releases)

一个零配置的 Go pprof HTTP 服务包装器，导入即启动，支持多实例并发运行。

受 [gofiber/fiber pprof middleware](https://github.com/gofiber/fiber/blob/main/middleware/pprof/pprof.go) 启发。

## 安装

```bash
go get github.com/jimmicro/pprof@latest
```

## 快速开始

只需匿名导入，`init()` 会自动在本地随机端口启动 pprof HTTP 服务：

```go
import _ "github.com/jimmicro/pprof"
```

## 服务发现

启动后会在 binary 所在目录生成两个文件：

```
<binary>_<pid>_<port>.pprof          # 内含 pprof 服务地址，如 http://127.0.0.1:34721
<binary>_<pid>_profile_dump.sh       # 一键采集脚本
```

文件名包含 pid 和 port，同一 binary 多实例并发运行时各自独立，互不覆盖。

collector 侧发现所有实例：

```bash
ls <dir>/<binary>_*_*.pprof
```

## 配置

### 环境变量

| 变量          | 说明                                   | 默认值     |
|---------------|----------------------------------------|------------|
| `PPROF_PORT`  | 指定监听端口，无效值回退为随机端口     | 0（随机）  |

```bash
PPROF_PORT=6060 ./myservice
```

### 代码配置

```go
// 启动失败时不 panic，仅打印日志（适合非关键服务）
pprof.PanicOnError = false
```

## 可用 Endpoints

### pprof 分析

| Endpoint                      | 说明              |
|-------------------------------|-------------------|
| `GET /debug/pprof/`           | 首页，列出所有分析选项 |
| `GET /debug/pprof/profile`    | CPU 分析（默认 30s） |
| `GET /debug/pprof/heap`       | 堆内存分析        |
| `GET /debug/pprof/goroutine`  | Goroutine 分析    |
| `GET /debug/pprof/allocs`     | 内存分配采样      |
| `GET /debug/pprof/block`      | 阻塞事件采样      |
| `GET /debug/pprof/mutex`      | 互斥锁竞争采样    |
| `GET /debug/pprof/threadcreate` | 线程创建采样   |
| `GET /debug/pprof/trace`      | 执行追踪          |

### 堆内存管理

| Endpoint                          | 说明                        |
|-----------------------------------|-----------------------------|
| `POST /debug/heap/gc`             | 触发 `runtime.GC()`         |
| `POST /debug/heap/free`           | 触发 `debug.FreeOSMemory()` |
| `POST /debug/heap/gcp?v=<int>`    | 设置 `debug.SetGCPercent(v)` |
| `POST /debug/heap/memlimit?v=<bytes>` | 设置 `debug.SetMemoryLimit(v)` |

### 其他

| Endpoint        | 说明                  |
|-----------------|-----------------------|
| `GET /debug/vars` | 导出的 expvar 变量  |

## 一键采集脚本

启动时会在 binary 同目录生成 `<binary>_<pid>_profile_dump.sh`，执行后自动采集全量 pprof 数据并打包：

```bash
bash <binary>_<pid>_profile_dump.sh
# 输出: profile/<binary>_<pid>_<timestamp>.tar.gz
```

脚本内含 goroutine、heap、allocs 等所有 profile，以及 5 秒 CPU profile 和 trace。

## 开发

```bash
# 安装工具
task deps

# 运行 linter
task lint

# 运行测试
task test
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Apache 2.0](LICENSE)
