package utils

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// ExtractText は Gemini のレスポンスから文字列を取り出す。
// text: 取得した文字列、ok: 取得に成功したか
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

// SplitJSONString は JSON 文字列で渡された配列を n 件ずつに分割し、
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
