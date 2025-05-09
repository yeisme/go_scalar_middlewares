# 让 API 文档更优雅：使用 `go_scalar_middlewares` 轻松集成 Scalar UI

## 🧰 什么是 Scalar？

[Scalar](https://scalar.com/) 是一个现代化的 API 文档解决方案，以其简洁美观的界面、强大的功能以及良好的可定制性受到越来越多开发者的喜爱。相比传统的 Swagger UI，Scalar 在视觉体验和交互逻辑上都有显著提升。

Scalar 支持从 OpenAPI（原 Swagger）规范文件自动生成文档，并提供搜索、过滤、代码片段生成等实用功能。通过 `go_scalar_middlewares`，我们可以将这一能力无缝整合进 Go 的 HTTP 服务中。

---

## 📦 核心特性一览

- **自动发现 OpenAPI 文件**：支持自动查找并加载 `openapi.json` 或 `openapi.yaml`，优先级目录为 `api`, `doc`, `.`
- **动态生成 Scalar 文档页面**：根据找到的 OpenAPI 文件路径动态生成 HTML 页面
- **开箱即用的中间件模式**：兼容标准 `http.HandlerFunc` 接口，便于集成到任何基于 `net/http` 的框架中
- **优雅的 fallback 处理机制**：当未找到 OpenAPI 文件时返回清晰错误信息，避免静默失败
- **支持自动注册路由**：通过 `_ "github.com/yeisme/go_scalar_middlewares/auto"` 方式实现零配置注册

---

## 🛠️ 快速开始

### 安装

```bash
go get github.com/yeisme/go_scalar_middlewares
```

### 基础用法

你可以选择两种方式集成 Scalar 中间件：

#### ✅ 方法一：手动注册处理器

适用于需要精细控制路由的场景：

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/yeisme/go_scalar_middlewares/middleware"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello-world" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintln(w, "Hello, world from the main application!")
}

func main() {
	http.ListenAndServe(":8080", middleware.ScalarAPIDocs(helloHandler))
}
```

#### ✅ 方法二：自动注册路由

如果你希望以最简方式启用 Scalar 文档，并且不介意根路径 `/` 被占用，可以使用自动注册方式：

```go
package main

import (
	"net/http"

	_ "github.com/yeisme/go_scalar_middlewares/auto"
)

func main() {
	http.ListenAndServe(":8080", nil)
}
```

该方式会在根路径下处理所有 Scalar 相关请求，并自动匹配已存在的 API 文档文件。

---

### ⚙️ 请求路由

| 请求路径            | 行为说明                      |
| ------------------- | ----------------------------- |
| `/scalar`           | 返回 Scalar UI 页面           |
| `/api/openapi.yaml` | 返回 YAML 格式的 OpenAPI 规范 |
| `/api/openapi.json` | 返回 JSON 格式的 OpenAPI 规范 |

---

📌 **GitHub 地址**：[https://github.com/yeisme/go_scalar_middlewares](https://github.com/yeisme/go_scalar_middlewares)

MIT License | Copyright © 2025 Yeisme
