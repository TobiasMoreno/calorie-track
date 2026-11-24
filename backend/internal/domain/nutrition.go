package domain

import (
	"fmt"
	"math"
	"strings"
)

func EstimateItem(food Food, portion *FoodPortion, input CreateItemInput) (MealEntryItem, error) {
	if input.Quantity <= 0 || input.Quantity > 10000 {
		return MealEntryItem{}, fmt.Errorf("%w: quantity must be greater than zero", ErrValidation)
	}

	var grams float64
	var portionName string
	switch strings.ToLower(strings.TrimSpace(input.Unit)) {
	case "g", "gramos":
		grams = input.Quantity
	case "unidad", "unidades", "porción", "porcion", "porciones", "cucharada", "cucharadas":
		if portion == nil || portion.Grams <= 0 || portion.FoodID != food.ID {
			return MealEntryItem{}, fmt.Errorf("%w: a valid portion is required for this unit", ErrValidation)
		}
		grams = input.Quantity * portion.Grams
		portionName = portion.Name
	default:
		return MealEntryItem{}, fmt.Errorf("%w: unsupported unit", ErrValidation)
	}

	calories := round1(grams * food.CaloriesPer100G / 100)
	return MealEntryItem{
		FoodID:                  food.ID,
		FoodNameSnapshot:        food.Name,
		Quantity:                input.Quantity,
		Unit:                    input.Unit,
		PortionID:               input.PortionID,
		PortionNameSnapshot:     portionName,
		EstimatedGrams:          round1(grams),
		CaloriesPer100GSnapshot: food.CaloriesPer100G,
		CalculatedCalories:      calories,
		CookingMethod:           strings.TrimSpace(input.CookingMethod),
		Notes:                   strings.TrimSpace(input.Notes),
	}, nil
}

func SumCalories(items []MealEntryItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.CalculatedCalories
	}
	return round1(total)
}

func round1(value float64) float64 { return math.Round(value*10) / 10 }
