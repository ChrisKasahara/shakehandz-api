// Package gemini: 汎用ユーティリティ関数
package gemini

import (
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// ExtractText は Gemini のレスポンスから文字列を取り出す。
func ExtractText(resp *genai.GenerateContentResponse) (text string, ok bool) {
	if resp == nil || len(resp.Candidates) == 0 {
		return "", false
	}
	var sb strings.Builder
	for _, p := range resp.Candidates[0].Content.Parts {
		// テキストパーツからテキストを抽出
		if textPart, ok := p.(genai.Text); ok {
			sb.WriteString(string(textPart))
		}
	}
	if sb.Len() == 0 {
		return "", false
	}
	return sb.String(), true
}

func TrimPrefixAndSuffixGeminiResponse(target string) string {
	cleaned := strings.TrimPrefix(target, "```json\n")
	cleaned = strings.TrimSuffix(cleaned, "```")

	return cleaned
}
