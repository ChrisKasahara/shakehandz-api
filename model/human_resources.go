package model

import "time"

type HumanResource struct {
	ID                   string     `gorm:"primaryKey" json:"id"`
	EmailID              string     `json:"email_id"`
	EmailSubject         *string    `json:"email_subject,omitempty"`
	EmailSender          *string    `json:"email_sender,omitempty"`
	EmailReceivedAt      *time.Time `json:"email_received_at,omitempty"`
	AttachmentFilename   *string    `json:"attachment_filename,omitempty"`
	CandidateInitial     *string    `json:"candidate_initial,omitempty"`
	Age                  *uint8     `json:"age,omitempty"`
	Prefecture           *string    `json:"prefecture,omitempty"`
	NearestStation       *string    `json:"nearest_station,omitempty"`
	WorkConditions       *string    `json:"work_conditions,omitempty"`
	EmploymentType       *string    `json:"employment_type,omitempty"`
	MainLangFW           *string    `json:"main_lang_fw,omitempty"`
	MainSkills           *string    `json:"main_skills,omitempty"`
	ExperiencePhases     *string    `json:"experience_phases,omitempty"`
	AvailableStartMonth  *string    `json:"available_start_month,omitempty"`
	HourlyRateMin        *uint      `json:"hourly_rate_min,omitempty"`
	HourlyRateMax        *uint      `json:"hourly_rate_max,omitempty"`
	HourlyRateUnit       *string    `json:"hourly_rate_unit,omitempty"`
	AdditionalInfo       *string    `json:"additional_info,omitempty"`
	ExtractionConfidence *float64   `json:"extraction_confidence,omitempty"`
	ExtractionNotes      *string    `json:"extraction_notes,omitempty"`
	CareerSummary        *string    `json:"career_summary,omitempty"`
	RegisteredAt         *time.Time `json:"registered_at,omitempty"`
}
