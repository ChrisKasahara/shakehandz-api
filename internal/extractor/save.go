package extractor

import (
	"fmt"
	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/humanresource"
)

func SaveExtractedHumanResources(hrs []humanresource.HumanResource, user auth.User, s *Service) (bool, error) {
	if len(hrs) > 0 {
		for i := range hrs {
			hrs[i].CreatedByID = &user.ID
			hrs[i].UpdatedByID = &user.ID
		}

		// BeforeCreateでUserIDをセットされるので、ここではセット不要
		if err := s.DB.Create(&hrs).Error; err != nil {
			return false, fmt.Errorf("DB保存失敗: %w", err)
		}

		// 成功した場合はtrueを返す
		return true, nil
	}

	// 何も保存しなかった場合はfalseを返す
	return false, nil
}
