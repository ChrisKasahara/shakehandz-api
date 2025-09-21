package humanresource

import (
	"shakehandz-api/internal/auth"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"gorm.io/gorm"
)

/* ---------- 列挙型 ---------- */

// 雇用体系
type EmploymentType string

const (
	EmploymentFulltime  EmploymentType = "fulltime"
	EmploymentFreelance EmploymentType = "freelance"
	EmploymentOther     EmploymentType = "other"
)

// 勤務スタイル
type WorkStyle string

const (
	WorkStyleFullRemote WorkStyle = "full_remote" // フルリモート希望
	WorkStyleCombined   WorkStyle = "combined"    // リモート併用希望
	WorkStyleOnSite     WorkStyle = "onSite"      // 常駐可能
)

// 経験領域
type ExperienceArea string

const (
	ExpDefinition     ExperienceArea = "definition"
	ExpBasicDesign    ExperienceArea = "basicDesign"
	ExpDetailedDesign ExperienceArea = "detailedDesign"
	ExpImplementation ExperienceArea = "implementation"
	ExpTesting        ExperienceArea = "testing"
	ExpMaintenance    ExperienceArea = "maintenance"
)

// 役割
type Role string

const (
	RoleDevelopment       Role = "development"
	RoleInfrastructure    Role = "infrastructure"
	RoleMobile            Role = "mobile"
	RoleTestAndQuality    Role = "testAndQuality"
	RoleDataAnalytics     Role = "dataAnalytics"
	RoleProjectManagement Role = "projectManagement"
	RoleHelpdesk          Role = "helpdesk"
	RoleSecurity          Role = "security"
	RoleDevOpsSRE         Role = "devOpsSRE"
	RoleProductOwner      Role = "productOwner"
	RoleConsulting        Role = "consulting"
	RoleLowSkill          Role = "lowSkill"
)

// 国籍
type Nationality string

const (
	NatJapan       Nationality = "japan"
	NatForeigner   Nationality = "foreigner"
	NatNaturalized Nationality = "naturalized"
)

/* ---------- モデル ---------- */

type HumanResource struct {
	/* 0. 一意キー */
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID string `gorm:"type:varchar(255);uniqueIndex" json:"message_id"`

	/* メールから直接抜ける情報 */
	AttachmentType     *string `gorm:"type:varchar(50)"  json:"attachment_type,omitempty"`
	AttachmentFilename *string `gorm:"type:varchar(255)" json:"attachment_filename,omitempty"`
	EmailReceivedAt    string  `json:"email_received_at"`
	ProviderCompany    *string `gorm:"type:varchar(255)" json:"provider_company,omitempty"`
	SalesPerson        *string `gorm:"type:varchar(255)" json:"sales_person,omitempty"`

	/* 要員の基本情報（年齢・氏名・国籍） */
	CandidateInitial *string      `gorm:"type:varchar(10)" json:"candidate_initial,omitempty"`
	Age              *uint8       `json:"age,omitempty"`
	Nationality      *Nationality `gorm:"type:enum('japan','foreigner','naturalized')" json:"nationality,omitempty"`

	/* 要員の領域に関する情報 */
	Roles           datatypes.JSONSlice[Role]           `gorm:"type:json" json:"roles,omitempty"`
	ExperienceAreas datatypes.JSONSlice[ExperienceArea] `gorm:"type:json" json:"experience_areas,omitempty"`

	/* 要員のスキルに関する情報 */
	MainSkills     datatypes.JSONSlice[string] `gorm:"type:json" json:"main_skills,omitempty"`
	SubSkills      datatypes.JSONSlice[string] `gorm:"type:json" json:"sub_skills,omitempty"`
	AdditionalInfo *string                     `gorm:"type:text" json:"additional_info,omitempty"`

	/* 要員の所属や働き方などの情報 */
	EmploymentType       *EmploymentType          `gorm:"type:enum('fulltime','freelance','other')" json:"employment_type,omitempty"`
	WorkStyle            *WorkStyle               `gorm:"type:enum('full_remote','combined','onSite')" json:"work_style,omitempty"`
	IsDirectlyUnder      bool                     `json:"is_directly_under"`
	Residence            *string                  `gorm:"type:varchar(255)" json:"residence,omitempty"`
	NearestStation       *string                  `gorm:"type:varchar(255)" json:"nearest_station,omitempty"`
	AvailableStartMonths datatypes.JSONSlice[int] `gorm:"type:json" json:"available_start_months,omitempty"`
	MonthlyRateMax       *uint                    `json:"monthly_rate_max,omitempty"`
	MonthlyRateMin       *uint                    `json:"monthly_rate_min,omitempty"`
	HourlyRateMax        *uint                    `json:"hourly_rate_max,omitempty"`
	HourlyRateMin        *uint                    `json:"hourly_rate_min,omitempty"`

	/* メタ情報 */
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedByID *uuid.UUID     `gorm:"type:char(36)" json:"created_by_id,omitempty"`
	UpdatedByID *uuid.UUID     `gorm:"type:char(36)" json:"updated_by_id,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Userモデルとのリレーションを定義
	// これにより、Preloadなどでユーザー情報を一緒に取得できる
	Creator *auth.User `gorm:"foreignKey:CreatedByID;references:ID" json:"creator,omitempty"`
	Updater *auth.User `gorm:"foreignKey:UpdatedByID;references:ID" json:"updater,omitempty"`
}

type HumanResourceFilter struct {
	// テキスト入力
	FreeWord string `form:"free_word" json:"free_word"`

	// 数値範囲（配列形式）
	Age       []int `form:"age" json:"age"`
	UnitPrice []int `form:"unit_price" json:"unit_price"`

	// チェックボックス（文字列配列）
	EmploymentType []string `form:"employmentType" json:"employment_type"`
	Nationality    []string `form:"nationality" json:"nationality"`
	WorkStyle      []string `form:"workStyle" json:"work_style"`

	// セレクト（オブジェクト形式）
	MainSkills []string `json:"main_skills"`
	SubSkills  []string `json:"sub_skills"`

	// スイッチ（真偽値）
	Affiliation *bool `form:"affiliation" json:"affiliation"`

	// 最もふるい受信日
	ReceiveAt string `form:"receive_at" json:"receive_at"`

	// ページング用
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

type HumanResourceResponse struct {
	Pagination         PaginationInfo       `json:"pagination"`
	AppliedFilters     *HumanResourceFilter `json:"appliedFilters"`
	HumanResourcesData []HumanResource      `json:"humanResourcesData"`
}

type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}
