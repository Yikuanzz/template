package user

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var UserTableName = "user"

// User 用户
type User struct {
	ID           uint           `gorm:"column:id;type:uint;primarykey;comment:用户ID"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:datetime;default:current_timestamp;not null;index:idx_user_created_at;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:datetime;default:current_timestamp;on update:current_timestamp;not null;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;uniqueIndex:idx_username_deleted,idx_email_deleted,idx_phone_deleted;index:idx_user_deleted_at,idx_user_status;comment:删除时间"`
	Username     string         `gorm:"column:username;type:varchar(16);uniqueIndex:idx_username_deleted;comment:用户名"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(512);comment:密码哈希"`

	// 系统内详细用户信息
	NickName string `gorm:"column:nick_name;type:varchar(16);comment:昵称"`
	Avatar   string `gorm:"column:avatar;type:varchar(255);comment:头像"`

	// 扩展字段
	ExtraData datatypes.JSON `gorm:"column:extra_data;type:json;comment:扩展字段"`
}

func (User) TableName() string {
	return UserTableName
}
