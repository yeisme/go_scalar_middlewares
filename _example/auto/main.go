package main

import (
	"fmt"
	"net/http"

	"github.com/yeisme/go_scalar_middlewares/auto"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "这是API端点")
}

func main() {
	// 显示找到的规范文件
	if auto.IsInitialized() {
		fmt.Println("成功自动发现 OpenAPI 规范文件:")
		for i, file := range auto.GetFoundSpecFiles() {
			if i == 0 {
				fmt.Printf("  - %s (已选择)\n", file)
			} else {
				fmt.Printf("  - %s\n", file)
			}
		}
	} else if err := auto.GetInitError(); err != nil {
		fmt.Printf("初始化错误: %v\n", err)
	} else {
		fmt.Println("未找到有效的 OpenAPI 规范文件")
	}

	// 注册自定义处理器
	http.HandleFunc("/api", apiHandler)
	
	fmt.Println("服务器启动在 :18081 端口...")
	if err := http.ListenAndServe(":18081", nil); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
