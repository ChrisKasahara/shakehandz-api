# システム構成ドキュメント

## ディレクトリ構成（抜粋）

- `cmd/server/` ... エントリーポイント
- `internal/auth/` ... 認証関連
- `internal/gemini/` ... Gemini 連携
- `internal/gmail/` ... Gmail 連携・処理
- `internal/humanresource/` ... 要員情報モデル・ハンドラ
- `internal/project/` ... プロジェクト情報モデル・ハンドラ
- `internal/shared/` ... 共通処理（DB, メール IF 等）
- `prompts/` ... プロンプト・指示文
- `utils/` ... 汎用ユーティリティ

## 主要モデル

### HumanResource

- ID, MessageID（ユニーク）
- 添付ファイル情報
- 基本情報（氏名イニシャル・年齢・国籍）
- 経験領域・役割（JSON 配列）
- スキル（JSON 配列）
- 雇用体系・勤務スタイル
- 直請けフラグ

### Project

- ID（主キー）
- メール情報（ID, 件名, 送信者, 受信日時）
- プロジェクト開始月
- 勤務地・都道府県
- リモート頻度・勤務時間
- 必須スキル・単価
- 業務内容・制約・優先人材
- 抽出信頼度・備考

## DB 構成・マイグレーション

- `internal/shared/db.go` で GORM による DB 初期化・マイグレーション
- `Project`, `HumanResource` モデルを自動マイグレーション

## ER 図

ER 図は [`ai-generated/docs/ER/er_diagram.md`](ai-generated/docs/ER/er_diagram.md) を参照
