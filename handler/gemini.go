package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"

	"shakehandz-api/prompts"
	"shakehandz-api/utils"
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

	readyResp, err := chat.SendMessage(ctx, genai.Text(prompts.HRInstruction))

	// Geminiに変換用の指示プロンプトを記憶させる
	readyStr, ok := utils.ExtractText(readyResp)
	if err != nil {
		log.Printf("Gemini API 呼び出し失敗: %v", err)
		c.JSON(500, gin.H{"error": "Gemini API 呼び出し失敗"})
		return
	}

	if !ok || strings.ToLower(strings.TrimSpace(readyStr)) != "ready" {
		c.JSON(500, gin.H{"error": "'ready' が返ってきませんでした"})
		return
	}

	// Geminiから「ready」が返ってきたことを確認
	fmt.Println("Geminiは準備を終えているようです。続いて変換処理を行います。")

	// 変換処理をGeminiに依頼
	result, err := chat.SendMessage(ctx, genai.Text(req.Text))

	// req.Textは配列であるためJSON形式に変換してから最大5つのオブジェクトに切り分けてGeminiに送信

	// 生成されたテキストを抽出する
	structuredEmailText, ok := utils.ExtractText(result)
	if err != nil {
		log.Printf("Gemini API 呼び出し失敗: %v", err)
		c.JSON(500, gin.H{"error": "Gemini API 呼び出し失敗"})
		return
	}

	if !ok {
		c.JSON(500, gin.H{"error": "Gemini API から有効なレスポンスがありません"})
		return
	}

	c.JSON(200, ConvertResponse{
		Converted: structuredEmailText,
	})
}
