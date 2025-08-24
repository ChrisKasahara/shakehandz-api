package extractor

import (
	"encoding/json"
	msg "shakehandz-api/internal/shared/message"
)

// chunkArray は flatArray を chunkSize ごとに分割し、各チャンクをJSON文字列に変換して返す
func chunkArray(flatArray []*msg.Message, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{}
	}
	n := len(flatArray)
	if chunkSize >= n {
		if jsonStr, err := json.Marshal(flatArray); err == nil {
			return []string{string(jsonStr)}
		}
		return []string{}
	}
	var result []string
	for i := 0; i < n; i += chunkSize {
		end := i + chunkSize
		if end > n {
			end = n
		}
		chunk := flatArray[i:end]
		if jsonStr, err := json.Marshal(chunk); err == nil {
			result = append(result, string(jsonStr))
		}
	}
	return result
}
