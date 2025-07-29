package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type GeminiHandler struct {
	Client *genai.Client
}

func NewGeminiHandler(client *genai.Client) *GeminiHandler {
	return &GeminiHandler{Client: client}
}

type ConvertRequest struct {
	Text string `"json:"text"`
}

type ConvertResponse struct {
	Converted string `"json:"converted"`
}

func (h *GeminiHandler) Greet(c *gin.Context) {
	ctx := context.Background()

	result, err := h.Client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text("こんにちは！"),
		nil,
	)
	if err != nil {
		log.Printf("GenerateContent error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gemini API 呼び出し失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"greeting": result.Text(),
	})
}

func (h *GeminiHandler) Convert(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "リクエスト形式が不正です"})
		return
	}

	prompt := `
送信されたメールの本文から要員情報をJSON形式で抽出してください。
不明な項目はnullにしてください。
また、出力すべき結果のみを返してください。Modelからのフィードバックなどは不要です。

type HumanResource struct {
	ID                   string     "gorm:"primaryKey" json:"id"
	EmailID              string     "gorm:"type:varchar(255)" json:"email_id"
	EmailSubject         *string    "json:"email_subject,omitempty"
	EmailSender          *string    "json:"email_sender,omitempty"
	EmailReceivedAt      *time.Time "json:"email_received_at,omitempty"
	AttachmentFilename   *string    "json:"attachment_filename,omitempty"
	CandidateInitial     *string    "json:"candidate_initial,omitempty"
	Age                  *uint8     "json:"age,omitempty"
	Prefecture           *string    "gorm:"type:varchar(255)" json:"prefecture,omitempty"
	NearestStation       *string    "json:"nearest_station,omitempty"
	WorkConditions       *string    "json:"work_conditions,omitempty"
	EmploymentType       *string    "json:"employment_type,omitempty"
	MainLangFW           *string    "json:"main_lang_fw,omitempty"
	MainSkills           *string    "json:"main_skills,omitempty"
	ExperiencePhases     *string    "json:"experience_phases,omitempty"
	AvailableStartMonth  *string    "json:"available_start_month,omitempty"
	HourlyRateMin        *uint      "json:"hourly_rate_min,omitempty"
	HourlyRateMax        *uint      "json:"hourly_rate_max,omitempty"
	HourlyRateUnit       *string    "json:"hourly_rate_unit,omitempty"
	AdditionalInfo       *string    "json:"additional_info,omitempty"
	ExtractionConfidence *float64   "json:"extraction_confidence,omitempty"
	ExtractionNotes      *string    "json:"extraction_notes,omitempty"
	CareerSummary        *string    "json:"career_summary,omitempty"
	RegisteredAt         *time.Time "json:"registered_at,omitempty"
}

` + req.Text

	ctx := context.Background()
	result, err := h.Client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.Printf("Gemini API 呼び出し失敗: %v", err)
		c.JSON(500, gin.H{"error": "Gemini API 呼び出し失敗"})
		return
	}

	c.JSON(200, ConvertResponse{
		Converted: result.Text(),
	})
}
