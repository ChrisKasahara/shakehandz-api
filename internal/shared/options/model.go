package options

import "gorm.io/gorm"

// Skill はエンジニアのスキル情報を表すモデルです。
type Skills struct {
	gorm.Model
	Label       string `gorm:"uniqueIndex:idx_name;not null;type:varchar(255)" json:"label"`
	Count       int    `gorm:"default:0" json:"count"`
	SearchCount int    `gorm:"default:0" json:"search_count"`
}

// TableName はGORMにテーブル名を明示的に指定します。
func (Skills) TableName() string {
	return "skills"
}
