package file

import "time"

var FileTableName = "file"

type File struct {
	ID        uint      `gorm:"column:id;type:uint;primarykey;comment:文件ID"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:current_timestamp;not null;index:idx_file_created_at;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:current_timestamp;on update:current_timestamp;not null;comment:更新时间"`
	// 文件名
	FileName string `gorm:"column:file_name;type:varchar(255);not null;comment:文件名"`

	// 文件存储
	FileStorageType string `gorm:"column:file_storage_type;type:varchar(12);not null;comment:文件存储类型"`
	FileStoragePath string `gorm:"column:file_storage_path;type:varchar(512);not null;comment:文件存储路径"`
	FileMimeType    string `gorm:"column:file_mime_type;type:varchar(128);not null;comment:文件MIME类型"`

	// 文件大小
	FileSize int64  `gorm:"column:file_size;type:bigint;not null;comment:文件大小"`
	FileHash string `gorm:"column:file_hash;type:varchar(512);not null;comment:文件哈希"`
}

func (File) TableName() string {
	return FileTableName
}
