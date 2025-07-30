package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"

	"shakehandz-api/prompts"
)

type GeminiHandler struct {
	Client *genai.GenerativeModel
}

func NewGeminiHandler(client *genai.GenerativeModel) *GeminiHandler {
	return &GeminiHandler{Client: client}
}

type ConvertRequest struct {
	Text string `json:"text"`
}
type ConvertResponse struct {
	Converted string `json:"converted"`
}

func (h *GeminiHandler) Convert(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "リクエスト形式が不正です"})
		return
	}

	ctx := context.Background()

	chat := h.Client.StartChat()

	readyText, err := chat.SendMessage(ctx, genai.Text(prompts.HRInstruction))

	isReady := false
	if readyText != nil && len(readyText.Candidates) > 0 {
		var sb strings.Builder
		for _, part := range readyText.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				sb.WriteString(string(txt))
			}
		}
		isReady = sb.String() == "ready"
	} else {
		fmt.Println("Instruction Response にテキスト候補がありませんでした")
	}

	// “ready” の確認（大小文字無視）
	if !isReady {
		c.JSON(500, gin.H{"error": "Gemini から 'ready' が返ってきませんでした"})
		return
	}

	fmt.Println("Geminiは準備を終えているようです。続いて変換処理を行います。")

	result, err := chat.SendMessage(ctx, genai.Text(req.Text))

	if err != nil {
		log.Printf("Gemini API 呼び出し失敗: %v", err)
		c.JSON(500, gin.H{"error": "Gemini API 呼び出し失敗"})
		return
	}

	// 生成されたテキストを抽出する
	var generatedText string
	if result != nil && len(result.Candidates) > 0 {
		var partsBuilder strings.Builder
		for _, part := range result.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				partsBuilder.WriteString(string(txt))
			}
		}
		generatedText = partsBuilder.String()
	} else {
		// 候補がない場合やエラーの場合の処理
		log.Println("Gemini API から有効なレスポンス候補がありませんでした。")
		c.JSON(500, gin.H{"error": "Gemini API から有効なレスポンスがありません"})
		return
	}

	c.JSON(200, ConvertResponse{
		Converted: generatedText,
	})
}
