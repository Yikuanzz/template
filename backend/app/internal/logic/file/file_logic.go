package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"

	fileModel "backend/app/model/file"
	"backend/app/types/consts"
	"backend/app/types/dto"
	fileErr "backend/app/types/errorn"
	"backend/utils/envx"
	"backend/utils/errorx"
	"backend/utils/lofile"
	"backend/utils/logs"

	"go.uber.org/fx"
)

type FileRepo interface {
	CreateFile(ctx context.Context, file *fileModel.File) error
	GetFileByID(ctx context.Context, fileID uint) (*fileModel.File, error)
	GetFileByHash(ctx context.Context, hash string) (*fileModel.File, error)
	DeleteFile(ctx context.Context, fileID uint) error
}

type FileLogicParams struct {
	fx.In

	FileRepo FileRepo
}

type FileLogic struct {
	fileRepo FileRepo
	storage  *lofile.LocalStorage
}

func NewFileLogic(params FileLogicParams) *FileLogic {

	storageLocalPath, err := envx.GetString(consts.StorageLocalPath)
	if err != nil {
		logs.Error("获取 StorageLocalPath 配置失败", "error", err.Error())
		panic(err)
	}
	storageLocalBaseURL, err := envx.GetString(consts.StorageLocalBaseURL)
	if err != nil {
		logs.Error("获取 StorageLocalBaseURL 配置失败", "error", err.Error())
		panic(err)
	}

	storage := lofile.NewLocalStorage(storageLocalPath, storageLocalBaseURL)

	return &FileLogic{
		fileRepo: params.FileRepo,
		storage:  storage,
	}
}

// UploadFile 上传文件
func (l *FileLogic) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader) (*dto.FileDTO, error) {
	// 打开文件并读取内容用于计算哈希
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errorx.Wrap(err, fileErr.FileErrInvalidFile, "打开文件失败")
	}
	defer file.Close()

	// 读取文件内容用于计算哈希
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, errorx.Wrap(err, fileErr.FileErrInvalidFile, "读取文件内容失败")
	}

	// 计算文件哈希（SHA256）
	hashStr := l.calculateFileHash(fileContent)

	// 检查文件是否已存在（通过哈希）
	existingFile, err := l.fileRepo.GetFileByHash(ctx, hashStr)
	if err == nil && existingFile != nil {
		// 文件已存在，返回已存在的文件信息
		logs.Info("文件已存在，返回已存在的文件", "file_id", existingFile.ID, "hash", hashStr)
		return l.buildFileDTO(ctx, existingFile)
	}

	// 获取文件MIME类型
	mimeType := l.getMimeType(fileHeader)

	// 使用已读取的文件内容创建新的 Reader 用于上传
	fileReader := bytes.NewReader(fileContent)

	// 上传文件到存储
	storagePath, err := l.storage.Upload(ctx, fileReader, fileHeader.Filename, mimeType)
	if err != nil {
		return nil, errorx.Wrap(err, fileErr.FileErrStorageError, errorx.K("reason", err.Error()))
	}

	// 创建文件记录
	fileRecord := &fileModel.File{
		FileName:        fileHeader.Filename,
		FileStorageType: l.storage.GetType(),
		FileStoragePath: storagePath,
		FileMimeType:    mimeType,
		FileSize:        fileHeader.Size,
		FileHash:        hashStr,
	}

	if err := l.fileRepo.CreateFile(ctx, fileRecord); err != nil {
		// 如果数据库保存失败，尝试删除已上传的文件
		if delErr := l.storage.Delete(ctx, storagePath); delErr != nil {
			logs.Error("删除已上传文件失败", "error", delErr.Error(), "path", storagePath)
		}
		return nil, errorx.Wrap(err, fileErr.FileErrDatabaseError, errorx.K("reason", err.Error()))
	}

	logs.Info("文件上传成功", "file_id", fileRecord.ID, "filename", fileHeader.Filename, "size", fileHeader.Size)
	return l.buildFileDTO(ctx, fileRecord)
}

// calculateFileHash 计算文件内容的 SHA256 哈希值
func (l *FileLogic) calculateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// getMimeType 获取文件的 MIME 类型
func (l *FileLogic) getMimeType(fileHeader *multipart.FileHeader) string {
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "" {
		return mimeType
	}
	// 根据文件扩展名推断MIME类型
	ext := filepath.Ext(fileHeader.Filename)
	mimeType = mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream" // 默认类型
	}
	return mimeType
}

// buildFileDTO 构建文件 DTO 对象
func (l *FileLogic) buildFileDTO(ctx context.Context, file *fileModel.File) (*dto.FileDTO, error) {
	fileURL, err := l.storage.GetURL(ctx, file.FileStoragePath)
	if err != nil {
		return nil, errorx.Wrap(err, fileErr.FileErrStorageError, errorx.K("reason", err.Error()))
	}
	return &dto.FileDTO{
		FileID:   file.ID,
		FileName: file.FileName,
		FileURL:  fileURL,
	}, nil
}
