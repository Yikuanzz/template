package file

import (
	"context"

	fileModel "backend/app/model/file"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type FileRepoParams struct {
	fx.In

	DB *gorm.DB
}

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(params FileRepoParams) *FileRepo {
	return &FileRepo{
		db: params.DB,
	}
}

func (r *FileRepo) CreateFile(ctx context.Context, file *fileModel.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *FileRepo) GetFileByID(ctx context.Context, fileID uint) (*fileModel.File, error) {
	var file fileModel.File
	if err := r.db.WithContext(ctx).Where("id = ?", fileID).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepo) GetFileByHash(ctx context.Context, hash string) (*fileModel.File, error) {
	var file fileModel.File
	if err := r.db.WithContext(ctx).Where("file_hash = ?", hash).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepo) DeleteFile(ctx context.Context, fileID uint) error {
	return r.db.WithContext(ctx).Delete(&fileModel.File{}, fileID).Error
}
