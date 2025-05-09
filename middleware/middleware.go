package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	ScalarDocsPath = "/scalar"
)

var (
	openapiJSONContent []byte
	openapiYAMLContent []byte

	actualOpenapiJSONPath     string
	actualOpenapiYAMLPath     string
	actualDefaultSpecURLForUI string
	generatedScalarHTML       string

	loadOnce  sync.Once
	loadError error

	openapiJSONFilename = "openapi.json"
	openapiYAMLFilename = "openapi.yaml"
	searchDirs          = []string{"api", "doc", "."}
)

// tryLoadFile 尝试在 searchDirs 中查找并读取 filename.
// 返回内容、服务器路径 (例如 /api/openapi.json) 和错误.
func tryLoadFile(filename string, searchDirs []string) (content []byte, servePath string, err error) {
	for _, dir := range searchDirs {
		fqPath := filepath.Join(dir, filename)
		data, readErr := os.ReadFile(fqPath)
		if readErr == nil {
			if dir == "." {
				servePath = "/" + filename
			} else {
				servePath = "/" + filepath.ToSlash(fqPath)
			}
			return data, servePath, nil
		}
		if !os.IsNotExist(readErr) {
			err = fmt.Errorf("error reading %s: %v", fqPath, readErr)
			return nil, "", err
		}
	}
	return nil, "", os.ErrNotExist
}

// initializeScalar 负责加载 OpenAPI 规范并生成 HTML.
// 此函数由 loadOnce.Do() 调用.
func initializeScalar() {
	var primarySpecError error

	yamlContent, yamlPath, yamlErr := tryLoadFile(openapiYAMLFilename, searchDirs)
	if yamlErr == nil {
		openapiYAMLContent = yamlContent
		actualOpenapiYAMLPath = yamlPath
		actualDefaultSpecURLForUI = actualOpenapiYAMLPath
	} else if !os.IsNotExist(yamlErr) {
		primarySpecError = yamlErr
	}

	jsonContent, jsonPath, jsonErr := tryLoadFile(openapiJSONFilename, searchDirs)
	if jsonErr == nil {
		openapiJSONContent = jsonContent
		actualOpenapiJSONPath = jsonPath
		if actualDefaultSpecURLForUI == "" {
			actualDefaultSpecURLForUI = actualOpenapiJSONPath
		}
	} else if !os.IsNotExist(jsonErr) {
		if primarySpecError == nil {
			primarySpecError = jsonErr
		}
	}

	if actualDefaultSpecURLForUI == "" {
		if primarySpecError != nil {
			loadError = primarySpecError
		} else {
			loadError = fmt.Errorf("no '%s' or '%s' found in search paths (%v)",
				openapiYAMLFilename, openapiJSONFilename, strings.Join(searchDirs, ", "))
		}
		return
	}

	generatedScalarHTML = fmt.Sprintf(`<!doctype html>
<html>
<head>
    <title>Scalar API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
        body { margin: 0; }
    </style>
</head>
<body>
    <script
        id="api-reference"
        data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`, actualDefaultSpecURLForUI)
}

// ScalarAPIDocs 是符合 go-zero 中间件签名的函数.
// 它提供 Scalar API 文档 UI 以及 OpenAPI 规范文件.
func ScalarAPIDocs(next http.HandlerFunc) http.HandlerFunc {
	loadOnce.Do(initializeScalar)

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, ScalarDocsPath) {
			if loadError != nil {
				http.Error(w, fmt.Sprintf("Scalar UI unavailable: %s", loadError.Error()), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(generatedScalarHTML))
			return
		}

		if actualOpenapiJSONPath != "" && r.URL.Path == actualOpenapiJSONPath {
			if len(openapiJSONContent) > 0 {
				w.Header().Set("Content-Type", "text/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(openapiJSONContent)
				return
			}
			http.NotFound(w, r)
			return
		}

		if actualOpenapiYAMLPath != "" && r.URL.Path == actualOpenapiYAMLPath {
			if len(openapiYAMLContent) > 0 {
				w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(openapiYAMLContent)
				return
			}
			http.NotFound(w, r)
			return
		}

		next(w, r)
	}
}
