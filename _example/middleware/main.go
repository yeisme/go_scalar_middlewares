package main

import (
	"fmt"
	"net/http"

	"github.com/yeisme/go_scalar_middlewares/middleware"
)

// helloHandler 是一个简单的 HTTP 处理程序作为示例
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello-world" {
		http.NotFound(w, r)
		return
	}
	_, _ = fmt.Fprintln(w, "Hello, world from the main application!")
}

func main() {
	// 方式1：使用默认配置
	// http.ListenAndServe(":8080", middleware.ScalarAPIDocs(helloHandler))

	// 方式2：使用自定义配置
	config := middleware.Config{
		// 手动指定特定的OpenAPI文件(可选)
		// JSONSpecPath: "./custom/myapi.json",
		// YAMLSpecPath: "./custom/myapi.yaml",
		
		// 自定义搜索目录(可选)
		SearchDirs: []string{"./api-docs", "./schemas", "."},
		
		// 自定义文档路径(可选)
		DocsPath: "/api-docs",
	}

	fmt.Println("Starting server on :18081...")
	if err := http.ListenAndServe(":18081", middleware.WithConfig(config)(helloHandler)); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
