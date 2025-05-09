package main

import (
	"net/http"

	_ "github.com/yeisme/go_scalar_middlewares/auto"
)

func main() {
	http.ListenAndServe(":8080", nil)
}
