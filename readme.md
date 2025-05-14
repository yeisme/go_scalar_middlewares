# go_scalar_middlewares: Go 项目的现代化 API 文档中间件

<p align="center">
  <strong>让 OpenAPI 文档展示变得优雅简单</strong>
</p>

---

## 📖 项目简介

`go_scalar_middlewares` 是一个简洁高效的Go中间件，用于在你的Web应用中无缝集成 [Scalar](https://scalar.com/) API文档界面。Scalar是一个美观、现代的API文档解决方案，相比传统的Swagger UI，它提供了更好的用户体验和功能。

本项目提供两种集成方式：

- 📌 **中间件模式**：适用于需要精细控制的场景
- 🚀 **自动注册模式**：实现零代码集成

## ✨ 为什么选择 Scalar？

相比传统的Swagger UI，Scalar提供了：

- 🎨 更现代化、美观的界面设计
- 🔍 更强大的搜索和导航功能
- 💻 更友好的代码示例生成
- 📱 更好的响应式设计，支持移动设备访问
- ⚡ 更快的加载速度和更流畅的交互体验

## 📦 安装

```bash
go get github.com/yeisme/go_scalar_middlewares
```

## 🚀 快速开始

### 方法一：中间件模式

将 Scalar 集成到你现有的 HTTP 处理函数中：

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/yeisme/go_scalar_middlewares/middleware"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	// 处理你的 API 请求
	fmt.Fprintln(w, "API endpoint")
}

func main() {
	// 包装你的处理函数
	http.ListenAndServe(":8080", middleware.ScalarAPIDocs(myHandler))
}
```

### 方法二：自动注册模式

最简单的集成方式，只需导入 `auto` 包：

```go
package main

import (
	"fmt"
	"net/http"

	_ "github.com/yeisme/go_scalar_middlewares/auto" // 导入即启用
)

func main() {
	// 注册你的API路由
	http.HandleFunc("/api/users", usersHandler)
	
	// 启动服务
	http.ListenAndServe(":8080", nil)
}
```

通过导入 `auto` 包，Scalar文档会自动注册到根路径，并自动发现你项目中的OpenAPI规范文件。

## 🔧 高级配置

### 中间件配置选项

通过 `middleware.Config` 可以自定义多种配置项：

```go
config := middleware.Config{
	// 手动指定OpenAPI文件路径
	JSONSpecPath: "./api/custom-openapi.json",
	// 或者YAML格式
	// YAMLSpecPath: "./api/custom-openapi.yaml",
	
	// 自定义搜索目录
	SearchDirs: []string{"./api-docs", "./schemas", "."},
	
	// 自定义文档访问路径前缀（默认为 "/scalar"）
	DocsPath: "/api-docs",
}

http.ListenAndServe(":8080", middleware.WithConfig(config)(myHandler))
```

### 自动发现机制

`auto` 包具有智能识别机制，会扫描以下目录寻找有效的OpenAPI规范文件：

- `api/`
- `doc/`
- `docs/`
- `openapi/`
- 当前目录 `.`

支持的文件格式：

- `.json`
- `.yaml`
- `.yml`

第一个被找到的有效规范文件将被用于生成文档。

### 查看自动发现结果

```go
import "github.com/yeisme/go_scalar_middlewares/auto"

func main() {
	// 查看发现的文件
	if auto.IsInitialized() {
		fmt.Println("找到的OpenAPI文件:")
		for _, file := range auto.GetFoundSpecFiles() {
			fmt.Printf("  - %s\n", file)
		}
	}
	
	// 启动服务
	http.ListenAndServe(":8080", nil)
}
```

## 📋 特性列表

- ✅ **自动发现 OpenAPI 文件**：无需手动配置文件路径
- ✅ **支持多种文件格式**：同时支持 JSON 和 YAML 格式的规范文件
- ✅ **灵活的集成方式**：支持中间件模式和自动注册模式
- ✅ **干净的 URL 映射**：不会干扰你现有的路由系统
- ✅ **智能回退机制**：找不到规范文件时提供清晰的错误信息
- ✅ **零配置选项**：适用于快速原型和简单项目
- ✅ **高度可配置**：适用于复杂项目和特殊需求

## 🌐 访问文档

启动服务后，可通过以下URL访问：

- **Scalar UI**：`http://localhost:8080/scalar` (或你自定义的路径)
- **JSON 规范**：自动映射到发现的 JSON 文件路径
- **YAML 规范**：自动映射到发现的 YAML 文件路径

## 📝 示例项目

项目包含两个示例应用：

1. **middleware示例**：`_example/middleware/main.go`
   展示如何使用中间件模式集成

2. **auto示例**：`_example/auto/main.go`
   展示如何使用自动注册模式集成

运行示例：

```bash
cd _example/auto
go run main.go
```

然后在浏览器访问 `http://localhost:18081/scalar`

## 📄 许可证

MIT License | Copyright © 2025 Yeisme
