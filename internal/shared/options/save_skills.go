package options

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveSkills(db *gorm.DB, skills []string) error {
	// 引数のスライスが空の場合は何もせずに正常終了
	if len(skills) == 0 {
		return nil
	}

	// skillsに重複があった場合を想定して、一意なスキル名のリストを作成
	// 重複除去に必要な作業用マップと結果スライス
	processed := make(map[string]bool)
	uniqueSkills := []string{}

	// 重複除去
	for _, skill := range skills {
		if !processed[skill] {
			processed[skill] = true                    // 処理済みとしてマップに記録
			uniqueSkills = append(uniqueSkills, skill) // 新しいスライスに追加
		}
	}

	// トランザクションを開始
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// 関数内で予期せぬエラー(panic)が起きてもロールバックする
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, name := range uniqueSkills {
		// スキル名称が衝突(重複)しているかの確認をClausesのOnConflictで行う
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			// 存在した場合、Countを1だけインクリメント
			DoUpdates: clause.Assignments(map[string]interface{}{"count": gorm.Expr("count + 1")}),
		}).Create(&Skills{
			// 存在しない場合、Count: 1で新規作成
			Label: name,
			Count: 1,
		}).Error

		if err != nil {
			tx.Rollback()
			return err
		}

	}

	// 全て成功したらコミット
	if err := tx.Commit().Error; err != nil {
		// コミット自体が失敗する可能性も考慮してロールバック
		tx.Rollback()
		return err
	}

	return nil
}
