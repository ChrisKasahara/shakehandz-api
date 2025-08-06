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
)

// Service: Gmailメッセージ取得→Gemini解析→（将来）DB保存
type Service struct {
	Fetcher msg.MessageFetcher
	Gemini  *Client
}

func NewService(f msg.MessageFetcher, g *Client) *Service {
	return &Service{Fetcher: f, Gemini: g}
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

	// 最初のチャンクをGeminiに送信
	if len(chunkedMsgs) > 0 {
		structuredResp, err := chat.SendMessage(ctx, genai.Text(chunkedMsgs[0]))
		structuredRespStr, _ := ExtractText(structuredResp)
		if err != nil {
			log.Printf("Gemini API 呼び出し失敗: %v", err)
			return nil, fmt.Errorf("Gemini API 呼び出し失敗")
		}

		fmt.Printf("Geminiからの構造化応答: %s\n", structuredRespStr)

		// JSON文字列をHumanResourceオブジェクトに変換
		humanResources, err := parseAndSaveHumanResources(structuredRespStr)
		if err != nil {
			log.Printf("HumanResource解析・保存エラー: %v", err)
			return nil, fmt.Errorf("HumanResource解析・保存エラー: %v", err)
		}

		return humanResources, nil
	}

	return nil, nil
}

// parseAndSaveHumanResources はGeminiからのJSON応答を解析してHumanResourceオブジェクトに変換・保存する
func parseAndSaveHumanResources(jsonStr string) ([]humanresource.HumanResource, error) {
	// パターン1: HumanResourceオブジェクトの配列として解析を試行
	var humanResources []humanresource.HumanResource
	err := json.Unmarshal([]byte(jsonStr), &humanResources)
	if err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %v", err)
	}

	// 解析成功時の処理
	fmt.Printf("解析成功: %d件のHumanResourceオブジェクトを取得\n", len(humanResources))
	for i, hr := range humanResources {
		fmt.Printf("HumanResource %d: ID=%s, Age=%v, CandidateInitial=%v\n",
			i+1, hr.MessageID, hr.Age, hr.CandidateInitial)
	}

	// TODO: ここでデータベースへの保存処理を実装
	// 例: db.Create(&humanResources)

	return humanResources, nil
}

// JSON変換ユーティリティ関数

// ObjectToJSON はオブジェクトをJSON文字列に変換する
func ObjectToJSON(obj interface{}) (string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("JSON変換エラー: %v", err)
	}
	return string(jsonBytes), nil
}

// ObjectToJSONIndent はオブジェクトをインデント付きJSON文字列に変換する
func ObjectToJSONIndent(obj interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON変換エラー: %v", err)
	}
	return string(jsonBytes), nil
}

// JSONToObject はJSON文字列を指定された型のオブジェクトに変換する
// 使用例:
//
//	var hr humanresource.HumanResource
//	err := JSONToObject(jsonStr, &hr)
//
//	var hrList []humanresource.HumanResource
//	err := JSONToObject(jsonStr, &hrList)
func JSONToObject(jsonStr string, target interface{}) error {
	err := json.Unmarshal([]byte(jsonStr), target)
	if err != nil {
		return fmt.Errorf("JSON解析エラー: %v", err)
	}
	return nil
}
