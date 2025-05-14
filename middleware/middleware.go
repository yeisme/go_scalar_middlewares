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
	DefaultScalarDocsPath = "/scalar"
)

// Config 定义了Scalar中间件的配置项
type Config struct {
	// 用于手动指定OpenAPI规范文件的路径
	JSONSpecPath string
	YAMLSpecPath string
	// 搜索目录，当未指定具体文件路径时使用
	SearchDirs []string
	// 文档UI的路径前缀
	DocsPath string
}

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
func initializeScalar(config Config) {
	var primarySpecError error

	// 使用配置的JSON规范文件（如果有）
	if config.JSONSpecPath != "" {
		jsonContent, jsonErr := os.ReadFile(config.JSONSpecPath)
		if jsonErr == nil {
			openapiJSONContent = jsonContent
			actualOpenapiJSONPath = "/" + filepath.Base(config.JSONSpecPath)
			actualDefaultSpecURLForUI = actualOpenapiJSONPath
		} else {
			primarySpecError = fmt.Errorf("无法读取指定的JSON规范文件: %v", jsonErr)
		}
	}

	// 使用配置的YAML规范文件（如果有）
	if config.YAMLSpecPath != "" {
		yamlContent, yamlErr := os.ReadFile(config.YAMLSpecPath)
		if yamlErr == nil {
			openapiYAMLContent = yamlContent
			actualOpenapiYAMLPath = "/" + filepath.Base(config.YAMLSpecPath)
			if actualDefaultSpecURLForUI == "" {
				actualDefaultSpecURLForUI = actualOpenapiYAMLPath
			}
		} else if primarySpecError == nil {
			primarySpecError = fmt.Errorf("无法读取指定的YAML规范文件: %v", yamlErr)
		}
	}

	// 如果没有配置具体文件路径，则尝试自动查找
	if config.JSONSpecPath == "" && config.YAMLSpecPath == "" {
		searchPaths := config.SearchDirs
		if len(searchPaths) == 0 {
			searchPaths = searchDirs
		}

		yamlContent, yamlPath, yamlErr := tryLoadFile(openapiYAMLFilename, searchPaths)
		if yamlErr == nil {
			openapiYAMLContent = yamlContent
			actualOpenapiYAMLPath = yamlPath
			actualDefaultSpecURLForUI = actualOpenapiYAMLPath
		} else if !os.IsNotExist(yamlErr) {
			primarySpecError = yamlErr
		}

		jsonContent, jsonPath, jsonErr := tryLoadFile(openapiJSONFilename, searchPaths)
		if jsonErr == nil {
			openapiJSONContent = jsonContent
			actualOpenapiJSONPath = jsonPath
			if actualDefaultSpecURLForUI == "" {
				actualDefaultSpecURLForUI = actualOpenapiJSONPath
			}
		} else if !os.IsNotExist(jsonErr) && primarySpecError == nil {
			primarySpecError = jsonErr
		}
	}

	if actualDefaultSpecURLForUI == "" {
		if primarySpecError != nil {
			loadError = primarySpecError
		} else {
			loadError = fmt.Errorf("未找到OpenAPI规范文件")
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

// WithConfig 使用给定的配置创建Scalar API文档中间件
func WithConfig(config Config) func(http.HandlerFunc) http.HandlerFunc {
	// 设置默认路径
	if config.DocsPath == "" {
		config.DocsPath = DefaultScalarDocsPath
	}

	// 重置全局变量，以便重新加载
	loadOnce = sync.Once{}
	openapiJSONContent = nil
	openapiYAMLContent = nil
	actualOpenapiJSONPath = ""
	actualOpenapiYAMLPath = ""
	actualDefaultSpecURLForUI = ""
	generatedScalarHTML = ""
	loadError = nil

	return func(next http.HandlerFunc) http.HandlerFunc {
		loadOnce.Do(func() { initializeScalar(config) })

		return func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, config.DocsPath) {
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
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
}

// ScalarAPIDocs 是符合 go-zero 中间件签名的函数.
// 它提供 Scalar API 文档 UI 以及 OpenAPI 规范文件.
func ScalarAPIDocs(next http.HandlerFunc) http.HandlerFunc {
	config := Config{
		DocsPath:   DefaultScalarDocsPath,
		SearchDirs: searchDirs,
	}
	return WithConfig(config)(next)
}
