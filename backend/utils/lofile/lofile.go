package lofile

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"backend/utils/logs"
)

// LocalStorage 本地存储实现
//
// 路径处理策略（跨平台兼容）：
// - 数据库存储的路径统一使用 URL 格式（正斜杠 /），例如：2025/12/18/filename.pdf
// - 文件系统操作时，使用 filepath.FromSlash() 将 URL 路径转换为系统路径格式
// - 在 Windows 上，系统路径使用反斜杠（\）；在 Linux 上，系统路径使用正斜杠（/）
// - 这样确保在 Windows 开发环境和 Linux 生产环境之间可以无缝迁移
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath string, baseURL string) *LocalStorage {
	// 确保目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		logs.Error("创建本地存储目录失败", "error", err.Error(), "path", basePath)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Upload 上传文件到本地
// 返回的路径是 URL 格式（正斜杠），用于存储在数据库中
func (s *LocalStorage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	// 生成唯一文件名：时间戳 + 原始文件名
	timestamp := time.Now().Format("20060102150405")
	ext := filepath.Ext(filename)
	name := filepath.Base(filename)
	nameWithoutExt := name[:len(name)-len(ext)]
	uniqueFilename := fmt.Sprintf("%s_%d_%s%s", nameWithoutExt, time.Now().UnixNano(), timestamp, ext)

	// 按日期组织目录结构：年/月/日（使用正斜杠，因为这是 URL 格式）
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	// 将 URL 格式路径转换为系统路径格式用于文件系统操作
	systemDatePath := filepath.FromSlash(datePath)
	fullDir := filepath.Join(s.basePath, systemDatePath)

	// 创建目录
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 完整文件路径
	fullPath := filepath.Join(fullDir, uniqueFilename)

	// 创建文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(fullPath) // 如果复制失败，删除已创建的文件
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回相对路径（相对于basePath），使用正斜杠作为URL路径分隔符
	// 这样存储在数据库中的路径格式统一，在 Windows 和 Linux 上都能正常工作
	relativePath := fmt.Sprintf("%s/%s", datePath, uniqueFilename)
	// 确保使用正斜杠（URL格式），filepath.ToSlash 在 Windows 上会转换，在 Linux 上保持不变
	relativePath = filepath.ToSlash(relativePath)
	return relativePath, nil
}

// GetURL 获取文件访问URL
func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	// 确保路径使用正斜杠（URL格式）
	path = filepath.ToSlash(path)

	if s.baseURL == "" {
		// 如果没有配置baseURL，返回相对路径
		return path, nil
	}

	// 确保baseURL不以/结尾，path不以/开头
	baseURL := s.baseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return fmt.Sprintf("%s/%s", baseURL, path), nil
}

// Delete 删除文件
// path 参数应该是 URL 格式的路径（正斜杠），会转换为系统路径格式
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	// 将 URL 格式的路径（正斜杠）转换为系统路径格式
	// 在 Windows 上会转换为反斜杠，在 Linux 上保持不变
	systemPath := filepath.FromSlash(path)
	fullPath := filepath.Join(s.basePath, systemPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，认为删除成功
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// GetType 获取存储类型
func (s *LocalStorage) GetType() string {
	return "local"
}
