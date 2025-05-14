package auto

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yeisme/go_scalar_middlewares/middleware"
)

var (
	// 存储找到的有效 OpenAPI 规范文件
	validSpecFiles  []string
	specFileContent []byte
	specFilePath    string
	specFileURL     string

	// 确保只初始化一次
	initOnce    sync.Once
	initErr     error
	initialized bool

	// 默认搜索目录
	defaultSearchDirs = []string{"api", "doc", "docs", "openapi", "."}
)

// 检查文件是否是 JSON 格式的 OpenAPI 文件
func isOpenAPIJSON(content []byte) bool {
	var doc map[string]any
	if err := json.Unmarshal(content, &doc); err != nil {
		return false
	}

	// 检查关键字段
	_, hasOpenAPI := doc["openapi"]
	_, hasSwagger := doc["swagger"]
	_, hasInfo := doc["info"].(map[string]any)
	_, hasPaths := doc["paths"]

	// OpenAPI 必须有 openapi/swagger 版本，info 对象和 paths
	return (hasOpenAPI || hasSwagger) && hasInfo && hasPaths
}

// 检查文件是否是 YAML 格式的 OpenAPI 文件
// 注意：这个实现比较简单，仅通过文本匹配检查常见的 OpenAPI YAML 结构
func isOpenAPIYAML(content []byte) bool {
	contentStr := string(content)

	// 检查 YAML 格式中常见的 OpenAPI 标记
	hasOpenAPIorSwagger := strings.Contains(contentStr, "openapi:") || strings.Contains(contentStr, "swagger:")
	hasInfo := strings.Contains(contentStr, "info:")
	hasPaths := strings.Contains(contentStr, "paths:")

	return hasOpenAPIorSwagger && hasInfo && hasPaths
}

// 查找目录中的所有潜在 OpenAPI 文件
func findOpenAPIFiles(searchDirs []string) ([]string, error) {
	var potentialFiles []string

	for _, dir := range searchDirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // 跳过无法访问的目录
			}

			if d.IsDir() {
				return nil // 继续处理文件
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".json" || ext == ".yaml" || ext == ".yml" {
				potentialFiles = append(potentialFiles, path)
			}

			return nil
		})

		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("无法遍历目录 %s: %v", dir, err)
		}
	}

	return potentialFiles, nil
}

// 验证文件是否是有效的 OpenAPI 规范
func validateOpenAPIFile(filePath string) (bool, []byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, nil, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var isValid bool

	if ext == ".json" {
		isValid = isOpenAPIJSON(content)
	} else if ext == ".yaml" || ext == ".yml" {
		isValid = isOpenAPIYAML(content)
	}

	if isValid {
		return true, content, nil
	}
	return false, nil, nil
}

// 初始化函数 - 查找和验证 OpenAPI 文件
func initialize() {
	// 查找所有潜在的 OpenAPI 文件
	files, err := findOpenAPIFiles(defaultSearchDirs)
	if err != nil {
		initErr = err
		return
	}

	// 检查每个文件是否是有效的 OpenAPI 规范
	for _, file := range files {
		isValid, content, err := validateOpenAPIFile(file)
		if err != nil {
			continue // 跳过无法读取的文件
		}

		if isValid {
			validSpecFiles = append(validSpecFiles, file)

			// 找到第一个有效文件后，保存其内容和路径
			if len(specFileContent) == 0 {
				specFileContent = content
				specFilePath = file
				specFileURL = "/" + filepath.Base(file)
				break // 找到第一个有效文件后立即停止
			}
		}
	}

	if len(validSpecFiles) == 0 {
		initErr = fmt.Errorf("未找到有效的 OpenAPI 规范文件")
		return
	}

	initialized = true
}

func init() {
	// 初始化
	initOnce.Do(initialize)

	// 定义不处理请求时的回退处理器
	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// 创建中间件配置
	config := middleware.Config{
		DocsPath: middleware.DefaultScalarDocsPath,
	}

	// 如果我们找到了有效的 OpenAPI 文件，直接配置其路径
	if initialized && len(validSpecFiles) > 0 {
		// 根据文件扩展名决定使用哪个配置字段
		ext := strings.ToLower(filepath.Ext(specFilePath))
		if ext == ".json" {
			config.JSONSpecPath = specFilePath
		} else if ext == ".yaml" || ext == ".yml" {
			config.YAMLSpecPath = specFilePath
		}
	}

	// 使用带有自定义配置的中间件
	scalarDocsHandler := middleware.WithConfig(config)(notFoundHandler)

	// 将处理器注册到默认 ServeMux
	http.HandleFunc("/", scalarDocsHandler)
}

// GetFoundSpecFiles 返回所有找到的有效 OpenAPI 规范文件
func GetFoundSpecFiles() []string {
	initOnce.Do(initialize)
	return validSpecFiles
}

// GetInitError 返回初始化过程中的错误（如果有）
func GetInitError() error {
	initOnce.Do(initialize)
	return initErr
}

// IsInitialized 返回是否成功初始化并找到有效的规范文件
func IsInitialized() bool {
	initOnce.Do(initialize)
	return initialized
}
