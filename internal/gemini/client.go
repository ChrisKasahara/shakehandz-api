package gemini

import (
	"context"
	"encoding/json"
	"errors"
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

	readyResp, err := chat.SendMessage(ctx, genai.Text(prompts.HRInstruction))

	// Geminiに変換用の指示プロンプトを記憶させる
	readyStr, ok := ExtractText(readyResp)
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
	structuredEmailText, ok := ExtractText(result)
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

// ExtractText は Gemini のレスポンスから文字列を取り出す。
func ExtractText(resp *genai.GenerateContentResponse) (text string, ok bool) {
	if resp == nil || len(resp.Candidates) == 0 {
		return "", false
	}
	var sb strings.Builder
	for _, p := range resp.Candidates[0].Content.Parts {
		if t, ok := p.(genai.Text); ok {
			sb.WriteString(string(t))
		}
	}
	if sb.Len() == 0 {
		return "", false
	}
	return sb.String(), true
}

// 各チャンクを再び JSON 文字列にエンコードして返す。
// 例: s = `[{"a":1},{"a":2},{"a":3}]`, n = 2
//
//	→ ["[{"a":1},{"a":2}]", "[{"a":3}]"]
func SplitJSONString(s string, n int) ([]string, error) {
	if n <= 0 {
		return nil, errors.New("chunk size must be > 0")
	}

	// ① 文字列 -> 任意型のスライスにデコード
	var arr []map[string]any
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, err
	}

	// ② n 件ずつチャンク化
	var chunks [][]map[string]any
	for i := 0; i < len(arr); i += n {
		end := i + n
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}

	// ③ 各チャンクを JSON 文字列に再エンコード
	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		b, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		out = append(out, string(b))
	}

	return out, nil
}
