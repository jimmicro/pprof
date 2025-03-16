# pprof for golang

受 <https://github.com/gofiber/fiber/blob/main/middleware/pprof/pprof.go> 启发的，一个简单易用的 Golang pprof 性能分析工具包装器。

## 功能特点

- 自动启动 pprof HTTP 服务
- 随机端口分配，避免端口冲突
- 自动生成服务地址配置文件
- 支持所有标准 pprof 分析功能
- 简单集成，零配置使用

## 快速开始

### 安装

```bash
go get github.com/jimmicro/pprof@latest
```

### 使用方法

只需要在你的项目中导入这个包：

```go
import _ "github.com/jimmicro/pprof"
 ```

导入后，pprof 服务器会自动启动在本地的随机端口上。pprof 地址会被写入到 {your_binary_name}.pprof 文件中。

### 可用的分析端点

- /debug/pprof/ - pprof 首页，列出所有可用的分析选项
- /debug/pprof/profile - CPU 分析
- /debug/pprof/heap - 堆内存分析
- /debug/pprof/goroutine - Goroutine 分析
- /debug/pprof/block - 阻塞分析
- /debug/pprof/mutex - 互斥锁分析
- /debug/pprof/threadcreate - 线程创建分析
- /debug/pprof/trace - 执行追踪
- /debug/vars - 导出的变量

### 配置选项

包提供了以下配置选项：

```go
pprof.PanicOnError = false // 设置为 false 可以禁用错误时的 panic
```

## 使用示例

```go
package main

import (
    "fmt"
    _ "github.com/jimmicro/pprof" // 仅需要导入包
)

func main() {
    // 你的应用代码
    fmt.Println("Application is running with pprof enabled")
    // ... 
}
```

## 注意事项

- pprof 服务默认只监听在 localhost (127.0.0.1)
