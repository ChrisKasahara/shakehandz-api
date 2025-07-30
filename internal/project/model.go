package project

import "time"

type Project struct {
	ID                       string     `gorm:"primaryKey" json:"id"`
	EmailID                  string     `gorm:"type:varchar(255)" json:"email_id"`
	EmailSubject             *string    `json:"email_subject,omitempty"`
	EmailSender              *string    `json:"email_sender,omitempty"`
	EmailReceivedAt          *time.Time `json:"email_received_at,omitempty"`
	ProjectStartMonth        *time.Time `json:"project_start_month,omitempty"`
	Prefecture               *string    `gorm:"type:varchar(255)" json:"prefecture,omitempty"`
	WorkLocation             *string    `json:"work_location,omitempty"`
	RemoteWorkFrequency      *string    `json:"remote_work_frequency,omitempty"`
	WorkingHours             *string    `json:"working_hours,omitempty"`
	RequiredSkills           *string    `json:"required_skills,omitempty"`
	UnitPriceMin             *uint      `json:"unit_price_min,omitempty"`
	UnitPriceMax             *uint      `json:"unit_price_max,omitempty"`
	UnitPriceUnit            *string    `json:"unit_price_unit,omitempty"`
	BusinessFlow             *string    `json:"business_flow,omitempty"`
	BusinessFlowRestrictions *string    `json:"business_flow_restrictions,omitempty"`
	PriorityTalent           *string    `json:"priority_talent,omitempty"`
	ProjectSummary           *string    `json:"project_summary,omitempty"`
	RegisteredAt             *time.Time `json:"registered_at,omitempty"`
	ExtractionConfidence     *float64   `json:"extraction_confidence,omitempty"`
	ExtractionNotes          *string    `json:"extraction_notes,omitempty"`
}
