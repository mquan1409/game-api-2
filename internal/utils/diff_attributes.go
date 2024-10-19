package utils

import (
	"github.com/mquan1409/game-api/internal/models"
)

func Minus(a []models.AttributeName, b []models.AttributeName) []models.AttributeName {
	result := make([]models.AttributeName, 0)
	bMap := make(map[models.AttributeName]bool)

	// Create a map of elements in b for faster lookup
	for _, item := range b {
		bMap[item] = true
	}

	// Add elements from a to result if they're not in b
	for _, item := range a {
		if !bMap[item] {
			result = append(result, item)
		}
	}

	return result
}