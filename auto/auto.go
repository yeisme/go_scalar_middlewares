package auto

import (
	"net/http"

	"github.com/yeisme/go_scalar_middlewares/middleware"
)

func init() {
	// 定义当 ScalarAPIDocs 不处理请求时调用的 'next' 处理器。
	// 对于自动注册场景，如果请求的路径不是 ScalarAPIDocs 设计处理的路径，
	// 并且该请求匹配了根路径 "/" 的注册，则应返回 404。
	// http.ServeMux 的最长匹配规则意味着，
	// 只有当请求路径与用户注册的更具体路由不匹配时，此处理器才会被调用。
	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// 获取 Scalar API 文档的处理器。
	// ScalarAPIDocs 内部会在首次被调用时执行初始化逻辑（如加载 OpenAPI 文件）。
	scalarDocsHandler := middleware.ScalarAPIDocs(notFoundHandler)

	// 将 Scalar 文档处理器注册到 http.DefaultServeMux 的根路径 "/"。
	// 当请求到达时，http.DefaultServeMux 会将请求路由到此处理器。
	// 然后，scalarDocsHandler (即 ScalarAPIDocs 返回的包装器)
	// 会检查请求路径是否为 /scalar 或其已知的 OpenAPI 规范文件路径。
	// 用户在 main 中为 http.DefaultServeMux 注册的更具体的路由
	// (例如 http.HandleFunc("/my-specific-route", ...))
	// 将因最长路径匹配规则而优先于此 "/" 注册。
	http.HandleFunc("/", scalarDocsHandler)
}
