// Gmailサービス層: fetcherの結果をGemini解析・DB保存
package gmail

import (
	"context"
	"errors"
	"shakehandz-api/internal/gemini"
)

// ProcessAndSaveMessages: fetcherで取得したメッセージをGeminiで解析しDB保存
func ProcessAndSaveMessages(ctx context.Context, messages []string, geminiClient *gemini.GeminiHandler, db any) error {
	if len(messages) == 0 {
		return errors.New("no messages to process")
	}
	// Geminiで解析（仮実装・未実装のためコメントアウト）
	/*
		for _, msg := range messages {
			result, err := geminiClient.Analyze(ctx, msg)
			if err != nil {
				return err
			}
			// DB保存（仮実装）
			if err := db.SaveAnalysisResult(ctx, result); err != nil {
				return err
			}
		}
	*/
	return nil
}
