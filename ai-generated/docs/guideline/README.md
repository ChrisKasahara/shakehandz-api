# shakehandz-api ガイドライン（シンプル構成）

---

## 1. プロジェクト構成

- Go 言語（Go Modules）
- 主要ディレクトリ
  - [`main.go`](../../main.go): エントリーポイント
  - [`config/`](../../config/): DB 初期化等
  - [`model/`](../../model/): DB モデル
  - [`handler/`](../../handler/): API ハンドラ
  - [`router/`](../../router/): ルーティング
- DB: RDS（MySQL）

---

## 2. テーブル設計

### projects テーブル

| カラム名                   | 型        | NULL 許可 | 主キー |
| -------------------------- | --------- | --------- | ------ |
| id                         | string    | ×         | ○      |
| email_id                   | string    | ×         |        |
| email_subject              | string    | ○         |        |
| email_sender               | string    | ○         |        |
| email_received_at          | time.Time | ○         |        |
| project_start_month        | time.Time | ○         |        |
| prefecture                 | string    | ○         |        |
| work_location              | string    | ○         |        |
| remote_work_frequency      | string    | ○         |        |
| working_hours              | string    | ○         |        |
| required_skills            | string    | ○         |        |
| unit_price_min             | uint      | ○         |        |
| unit_price_max             | uint      | ○         |        |
| unit_price_unit            | string    | ○         |        |
| business_flow              | string    | ○         |        |
| business_flow_restrictions | string    | ○         |        |
| priority_talent            | string    | ○         |        |
| project_summary            | string    | ○         |        |
| registered_at              | time.Time | ○         |        |
| extraction_confidence      | float64   | ○         |        |
| extraction_notes           | string    | ○         |        |

---

## 3. 開発ルール

- Go 公式の[Effective Go](https://go.dev/doc/effective_go)に準拠
- ディレクトリ・ファイルはスネークケース
- 構造体・関数・変数はキャメルケース
- コメントは日本語で簡潔に
- .env で DB 接続情報を管理

---

## 4. API 設計

- RESTful 設計
- `/projects` でプロジェクト情報取得
- 必要に応じて `/todos` も利用

---

## 5. その他

- コードレビュー推奨
- ドキュメントは随時最新化

---

## 6. POSTMAN 用インポートファイル出力ルール

- プロジェクト内の全 API エンドポイントは、POSTMAN 用コレクション（JSON 形式）として `ai-generated/postman/` ディレクトリに出力すること
- POSTMAN への出力指示があった場合は `postman/` ディレクトリに出力すること
- コレクションは v2.1 形式で記述し、エンドポイント追加時は必ず反映すること
