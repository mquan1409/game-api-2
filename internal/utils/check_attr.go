package utils

import (
	"github.com/mquan1409/game-api/internal/models"
)

func AttributesPositive(attributes_map models.AttributesStatsMap) bool{
	for _, attr := range attributes_map {
		if attr < 0{
			return false	
		}
	}
	return true
}

func AttributePositive(attr models.AttributeStat) bool{
	return attr > 0
}