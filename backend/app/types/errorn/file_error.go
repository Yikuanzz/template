package errorn

import (
	"backend/utils/errorx"
)

const (
	// 文件错误码 (3000000-3000099)
	FileErrUploadFailed        = int32(3000000) // 文件上传失败
	FileErrInvalidFile         = int32(3000001) // 无效的文件
	FileErrFileTooLarge        = int32(3000002) // 文件过大
	FileErrUnsupportedType     = int32(3000003) // 不支持的文件类型
	FileErrStorageError        = int32(3000004) // 存储错误
	FileErrFileNotFound        = int32(3000005) // 文件不存在
	FileErrDeleteFailed        = int32(3000006) // 删除文件失败
	FileErrHashCalculateFailed = int32(3000007) // 计算文件哈希失败
	FileErrDatabaseError       = int32(3000008) // 数据库错误
)

func init() {
	// 注册文件错误码
	errorx.RegisterBatch(map[int32]string{
		FileErrUploadFailed:        "文件上传失败: {reason}",
		FileErrInvalidFile:         "无效的文件",
		FileErrFileTooLarge:        "文件过大，最大允许: {max_size}",
		FileErrUnsupportedType:     "不支持的文件类型: {file_type}",
		FileErrStorageError:        "存储错误: {reason}",
		FileErrFileNotFound:        "文件不存在: {file_id}",
		FileErrDeleteFailed:        "删除文件失败: {reason}",
		FileErrHashCalculateFailed: "计算文件哈希失败: {reason}",
		FileErrDatabaseError:       "数据库错误: {reason}",
	})
}
