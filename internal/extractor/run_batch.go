package extractor

import (
	"context"
	"fmt"
	"shakehandz-api/internal/auth"

	"time"
)

func RunHumanResourceExtractionBatch(s *Service, user auth.User, maxExecutionDuration time.Duration, newBatchExecutionFromFront *ExtractorBatchExecution) error {
	// バッチを起動
	go func() {
		bgCtx := context.Background()
		data, exists := GetClientData(user.ID)

		if !exists {
			fmt.Println("クライアントデータが見つかりません")
			return
		}

		var retryCount int = 0

		for {
			// 前回の実行結果を取得
			var currentBatch ExtractorBatchExecution

			// バッチのステータスを確認
			if newBatchExecutionFromFront != nil {
				// 新しいバッチレコードが渡された場合は、画面側からのリクエストと判断
				// 新しいバッチレコードを現在のバッチとして設定
				currentBatch = *newBatchExecutionFromFront
				fmt.Printf("画面からのリクエスト。新しいバッチレコードが渡されました。バッチID: %d\n", currentBatch.ID)
			} else {
				// 新しいバッチレコードが渡されなかった場合は、バッチ内での実行と判断
				// DBから最新のバッチレコードを取得し、前回のバッチステータス確認後に新規でバッチレコードを作成
				// （バッチ内での実行の場合は、前回のバッチステータス確認後に新規でバッチレコードを作成）
				// これにより、バッチ内での連続実行が可能になる
				// ただし、バッチ内での連続実行は、前回のバッチが失敗または有効期限切れの場合はバッチを終了する
				res := s.DB.Where("user_id = ? AND trigger_from = ?", user.ID, TriggerFront).Order("execution_date desc").First(&currentBatch)

				if res.Error != nil {
					fmt.Printf("バッチレコード取得エラー: %v\n", res.Error)
					return
				}

				// クライアントが有効期限切れ、あるいは失敗ステータスの場合は終了
				isFailed := currentBatch.Status == StatusFailed
				isExpired := time.Since(currentBatch.ExecutionDate) > maxExecutionDuration
				isOnProgress := currentBatch.Status == StatusInProgress

				// 進行中ステータスが長時間続いている場合は失敗扱いとする
				if isOnProgress {
					s.DB.Model(&currentBatch).Where("id = ?", currentBatch.ID).Update("status", StatusFailed)
					fmt.Println("前回のバッチの進行中になんらかの問題が発生しました。失敗扱いとみなします。")
				}

				// 有効期限切れの場合はバッチを終了
				if isExpired {
					// バッチレコードを有効期限切れステータスに更新
					s.DB.Model(&currentBatch).Where("id = ?", currentBatch.ID).Update("status", StatusExpired)
					fmt.Printf("有効期限切れによりバッチ処理が終了しました。最終実行日時: %s\n", currentBatch.ExecutionDate)
					retryCount = 0
					// バッチを完全終了。画面側からのリクエストを待つのみ
					break
				}

				// 失敗ステータスの場合はバッチを終了し、リトライ回数を確認して必要に応じてリトライ
				if isFailed {
					s.DB.Model(&currentBatch).Where("id = ?", currentBatch.ID).Update("status", StatusFailed)

					// 失敗ステータスの場合はリトライ回数を確認して、最大リトライ回数に達していなければリトライ
					if retryCount < MaxRetryCount {
						fmt.Printf("前回のバッチが失敗しています。リトライを試みます。現在のリトライ回数: %d\n", retryCount+1)
						retryCount++
					}

					// 最大リトライ回数に達した場合の処理
					if retryCount >= MaxRetryCount {
						// バッチレコードを失敗ステータスに更新
						fmt.Printf("バッチ処理が%d回失敗しました。最終ステータス: %s\n", MaxRetryCount, currentBatch.Status)
						retryCount = 0
						// バッチを完全終了。画面側からのリクエストを待つのみ
						break
					}

				}

				// 新規のバッチレコードを作成
				currentBatch = ExtractorBatchExecution{
					UserID:        user.ID,
					ExtractorType: TypeHumanResource,
					TriggerFrom:   TriggerAuto,
					Status:        StatusInProgress,
					ExecutionDate: time.Now(),
				}

				r := s.DB.Create(&currentBatch)

				if r.Error != nil {
					fmt.Printf("バッチレコード作成エラー: %v\n", r.Error)
					return
				}
			}

			// 処理終了後はバッチループとなるため、newBatchExecutionはnilに戻す
			newBatchExecutionFromFront = nil

			success, err := Extract(bgCtx, user, data.gemini_cli, data.gmail_svc, s, currentBatch)

			if err != nil {
				fmt.Printf("Extract処理エラー: %v\n", err)
				s.DB.Model(&currentBatch).Where("id = ?", currentBatch.ID).Update("status", StatusFailed)
				break
			}

			if success {
				err := s.DB.Model(&currentBatch).Where("id = ?", currentBatch.ID).Update("status", StatusCompleted)
				retryCount = 0
				if err.Error != nil {
					fmt.Printf("バッチレコード更新エラー: %v\n", err.Error)
				}
				fmt.Println("次の処理を開始します")
			} else {
				// 処理対象がない場合は終了
				err := s.DB.Model(&currentBatch).Where("id = ?", currentBatch.ID).Update("status", StatusNoData)
				retryCount = 0

				if err.Error != nil {
					fmt.Printf("バッチレコード更新エラー: %v\n", err.Error)
				}

				fmt.Println("処理対象がないため終了")
				time.Sleep(MaxExecutionDuration)
			}
		}
		fmt.Println("バッチ処理が終了しました")
	}()
	return nil
}
