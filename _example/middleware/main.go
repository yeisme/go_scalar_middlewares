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

	http.ListenAndServe(":8080", middleware.ScalarAPIDocs(helloHandler))
}
