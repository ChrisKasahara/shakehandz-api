package humanresource

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// フィルターパラメータを解析する関数
func (h *HumanResourcesHandler) parseFilterParams(c *gin.Context) (*HumanResourceFilter, error) {
	var filter HumanResourceFilter

	// Content-TypeがJSONの場合（POST/PUTリクエスト）
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&filter); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body: %v", err)
		}

	} else {
		// クエリパラメータから基本的な値を取得
		if err := c.ShouldBindQuery(&filter); err != nil {
			return nil, fmt.Errorf("failed to parse query params: %v", err)
		}

		// 配列パラメータを手動で解析
		if err := h.parseArrayParams(c, &filter); err != nil {
			return nil, err
		}

		// 複雑なオブジェクトパラメータを解析
		if err := h.parseSkillParams(c, &filter); err != nil {
			return nil, err
		}

		// デバッグログ（開発時のみ）
		filterJSON, _ := json.Marshal(filter)
		fmt.Printf("Parsed query filter: %s\n", string(filterJSON))
	}

	return &filter, nil
}

// 配列パラメータを解析する関数
func (h *HumanResourcesHandler) parseArrayParams(c *gin.Context, filter *HumanResourceFilter) error {
	// 年齢範囲の解析
	if ageParam := c.Query("age"); ageParam != "" {
		var ages []int
		if err := json.Unmarshal([]byte(ageParam), &ages); err == nil && len(ages) == 2 {
			filter.Age = ages
		}
	}

	// 単価範囲の解析
	if priceParam := c.Query("unit_price"); priceParam != "" {
		var prices []int
		if err := json.Unmarshal([]byte(priceParam), &prices); err == nil && len(prices) == 2 {
			filter.UnitPrice = prices
		}
	}

	// 雇用形態の解析
	if empTypes := c.QueryArray("employmentType"); len(empTypes) > 0 {
		filter.EmploymentType = empTypes
	} else if empParam := c.Query("employmentType"); empParam != "" {
		var empTypes []string
		if err := json.Unmarshal([]byte(empParam), &empTypes); err == nil {
			filter.EmploymentType = empTypes
		}
	}

	// 国籍の解析
	if nationalities := c.QueryArray("nationality"); len(nationalities) > 0 {
		filter.Nationality = nationalities
	} else if natParam := c.Query("nationality"); natParam != "" {
		var nationalities []string
		if err := json.Unmarshal([]byte(natParam), &nationalities); err == nil {
			filter.Nationality = nationalities
		}
	}

	// 働き方の解析
	if workStyles := c.QueryArray("workStyle"); len(workStyles) > 0 {
		filter.WorkStyle = workStyles
	} else if workParam := c.Query("workStyle"); workParam != "" {
		var workStyles []string
		if err := json.Unmarshal([]byte(workParam), &workStyles); err == nil {
			filter.WorkStyle = workStyles
		}
	}

	return nil
}

// スキルパラメータを解析する関数
func (h *HumanResourcesHandler) parseSkillParams(c *gin.Context, filter *HumanResourceFilter) error {
	// メインスキルの解析
	if mainSkillParam := c.Query("main_skills"); mainSkillParam != "" {
		var mainSkills []string
		if err := json.Unmarshal([]byte(mainSkillParam), &mainSkills); err == nil {
			filter.MainSkills = mainSkills
		}
	}

	// サブスキルの解析
	if subSkillParam := c.Query("sub_skills"); subSkillParam != "" {
		var subSkills []string
		if err := json.Unmarshal([]byte(subSkillParam), &subSkills); err == nil {
			filter.SubSkills = subSkills
		}
	}

	return nil
}

func (h *HumanResourcesHandler) applyFilters(query *gorm.DB, filter *HumanResourceFilter) *gorm.DB {
	// フリーワード検索（名前、メール、説明文などを横断検索）
	if filter.FreeWord != "" {
		searchTerm := "%" + filter.FreeWord + "%"
		query = query.Where("email_received_at LIKE ? OR sales_person LIKE ? OR candidate_initial LIKE ? OR main_skills LIKE ? OR sub_skills LIKE ? OR additional_info LIKE ? OR nearest_station LIKE ? OR nearest_station LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// 年齢範囲検索
	if len(filter.Age) == 2 {
		query = query.Where("age BETWEEN ? AND ?", filter.Age[0], filter.Age[1])
	}

	// 単価範囲検索
	if len(filter.UnitPrice) == 2 {
		query = query.Where("monthly_rate_max BETWEEN ? AND ?", filter.UnitPrice[0], filter.UnitPrice[1])
	}

	// 雇用形態検索（IN句）
	if len(filter.EmploymentType) > 0 {
		query = query.Where("employment_type IN ?", filter.EmploymentType)
	}

	// 国籍検索（IN句）
	if len(filter.Nationality) > 0 {
		query = query.Where("nationality IN ?", filter.Nationality)
	}

	// 働き方検索（IN句）
	if len(filter.WorkStyle) > 0 {
		query = query.Where("work_style IN ?", filter.WorkStyle)
	}

	// メインスキル検索
	if filter.MainSkills != nil && len(filter.MainSkills) > 0 {
		query = h.applySkillFilter(query, "main_skills", filter.MainSkills)
	}

	// サブスキル検索
	if filter.SubSkills != nil && len(filter.SubSkills) > 0 {
		query = h.applySkillFilter(query, "sub_skills", filter.SubSkills)
	}

	// 所属フィルター（真偽値）
	if filter.Affiliation != nil && *filter.Affiliation {
		query = query.Where("is_directly_under = ?", *filter.Affiliation)
	}

	if filter.ReceiveAt != "" {
		query = query.Where("email_received_at >= ?", filter.ReceiveAt)
	} else {
		// デフォルトで過去1ヶ月以内のデータに絞る
		query = query.Where("email_received_at >= NOW() - INTERVAL 2 WEEK")
	}

	return query
}

// スキルフィルターを適用する関数
func (h *HumanResourcesHandler) applySkillFilter(query *gorm.DB, columnName string, skillFilter []string) *gorm.DB {
	if len(skillFilter) == 0 {
		return query
	}

	// AND検索：すべてのスキルを持っている人を検索
	for _, skill := range skillFilter {
		// PostgreSQLのJSONB演算子を使用
		query = query.Where("JSON_CONTAINS("+columnName+", ?)", fmt.Sprintf(`["%s"]`, skill))
	}

	// TODO: 将来的にOR検索もサポートする場合のコード例
	// // OR検索：いずれかのスキルを持っている人を検索
	// orConditions := make([]string, len(skillFilter))
	// args := make([]interface{}, len(skillFilter))

	// for i, skill := range skillFilter {
	// 	orConditions[i] = columnName + " @> ?"
	// 	args[i] = fmt.Sprintf(`["%s"]`, skill)
	// }

	// query = query.Where(strings.Join(orConditions, " OR "), args...)

	return query
}

// 使用例のヘルパー関数：フィルター条件のバリデーション
func (h *HumanResourcesHandler) validateFilter(filter *HumanResourceFilter) error {
	// 年齢範囲のバリデーション
	if len(filter.Age) == 2 && filter.Age[0] > filter.Age[1] {
		return errors.New("invalid age range: min age cannot be greater than max age")
	}

	// 単価範囲のバリデーション
	if len(filter.UnitPrice) == 2 && filter.UnitPrice[0] > filter.UnitPrice[1] {
		return errors.New("invalid unit price range: min price cannot be greater than max price")
	}

	// スキル選択のバリデーション
	if filter.MainSkills != nil && len(filter.MainSkills) > 10 {
		return errors.New("too many main skills selected (max: 10)")
	}

	if filter.SubSkills != nil && len(filter.SubSkills) > 10 {
		return errors.New("too many sub skills selected (max: 10)")
	}

	return nil
}
