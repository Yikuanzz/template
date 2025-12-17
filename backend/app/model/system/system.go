package system

var SystemConfigTableName = "system_config"

// SystemConfig 系统配置
type SystemConfig struct {
	K string `gorm:"column:k;type:varchar(255);not null;comment:系统键"`
	V string `gorm:"column:v;type:varchar(255);not null;comment:系统值"`
}

func (SystemConfig) TableName() string {
	return SystemConfigTableName
}
