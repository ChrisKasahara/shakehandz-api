package gemini

import (
	"encoding/json"
	"fmt"
	"log"
	"shakehandz-api/internal/humanresource"
	msg "shakehandz-api/internal/shared/message"
	"shakehandz-api/prompts"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"gorm.io/gorm"
)

// 　Gmailメッセージ取得→Gemini解析→（将来）DB保存
type Service struct {
	Fetcher msg.MessageFetcher
	Gemini  *Client
	DB      *gorm.DB
}

func NewService(f msg.MessageFetcher, g *Client, db *gorm.DB) *Service {
	return &Service{Fetcher: f, Gemini: g, DB: db}
}

func (s *Service) Run(c *gin.Context, token string) ([]humanresource.HumanResource, error) {
	ctx := c.Request.Context()
	msgs, err := s.Fetcher.FetchMsg(ctx, token, "has:attachment", 10)
	if err != nil {
		return nil, err
	}

	// chunkArrayで分割（JSON文字列の配列として）
	chunkedMsgs := chunkArray(msgs, 3)

	chat := s.Gemini.Model.StartChat()

	readyResp, err := chat.SendMessage(ctx, genai.Text(prompts.HRInstruction))

	// Geminiに変換用の指示プロンプトを記憶させる
	readyStr, ok := ExtractText(readyResp)
	fmt.Printf("Geminiからの応答: %s\n", readyStr)
	if err != nil {
		log.Printf("Gemini API 呼び出し失敗: %v", err)
		c.JSON(500, gin.H{"error": "Gemini API 呼び出し失敗"})
		return nil, fmt.Errorf("Gemini API 呼び出し失敗")
	}

	if !ok || strings.ToLower(strings.TrimSpace(readyStr)) != "ready" {
		c.JSON(500, gin.H{"error": "'ready' が返ってきませんでした"})
		return nil, fmt.Errorf("'ready' が返ってきませんでした: %s", readyStr)
	}

	// Geminiから「ready」が返ってきたことを確認
	fmt.Println("Geminiは準備を終えているようです。続いて変換処理を行います。")

	var humanResources []humanresource.HumanResource
	// 最初のチャンクをGeminiに送信
	if len(chunkedMsgs) > 0 {
		structuredResp, err := chat.SendMessage(ctx, genai.Text(chunkedMsgs[0]))
		structuredRespStr, _ := ExtractText(structuredResp)
		if err != nil {
			log.Printf("Gemini API 呼び出し失敗: %v", err)
			return nil, fmt.Errorf("Gemini API 呼びTrimPrefixAndSuffixGeminiResponse出し失敗")
		}

		fmt.Printf("Geminiからの構造化応答: %s\n", structuredRespStr)

		trimmedResponse := TrimPrefixAndSuffixGeminiResponse(structuredRespStr)

		humanResources = []humanresource.HumanResource{}

		if err := json.Unmarshal([]byte(trimmedResponse), &humanResources); err != nil {
			return nil, err
		}
	}
	// DB保存処理
	if len(humanResources) > 0 {
		if err := s.DB.Create(&humanResources).Error; err != nil {
			return nil, fmt.Errorf("DB保存失敗: %w", err)
		}
	}
	return humanResources, nil
}
