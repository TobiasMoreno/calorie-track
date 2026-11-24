package domain

import "testing"

func TestEstimateItemFromPortion(t *testing.T) {
	food := Food{ID: 1, Name: "Milanesa", CaloriesPer100G: 250}
	portion := FoodPortion{ID: 2, FoodID: 1, Name: "Mediana", Grams: 140}
	portionID := int64(2)
	item, err := EstimateItem(food, &portion, CreateItemInput{FoodID: 1, Quantity: 2, Unit: "unidades", PortionID: &portionID})
	if err != nil {
		t.Fatal(err)
	}
	if item.EstimatedGrams != 280 {
		t.Fatalf("grams = %v, want 280", item.EstimatedGrams)
	}
	if item.CalculatedCalories != 700 {
		t.Fatalf("calories = %v, want 700", item.CalculatedCalories)
	}
}

func TestEstimateItemFromGrams(t *testing.T) {
	food := Food{ID: 1, Name: "Arroz", CaloriesPer100G: 130}
	item, err := EstimateItem(food, nil, CreateItemInput{FoodID: 1, Quantity: 175, Unit: "g"})
	if err != nil {
		t.Fatal(err)
	}
	if item.CalculatedCalories != 227.5 {
		t.Fatalf("calories = %v, want 227.5", item.CalculatedCalories)
	}
}

func TestEstimateItemRejectsPortionFromAnotherFood(t *testing.T) {
	food := Food{ID: 1, CaloriesPer100G: 100}
	portion := FoodPortion{ID: 2, FoodID: 9, Grams: 100}
	if _, err := EstimateItem(food, &portion, CreateItemInput{FoodID: 1, Quantity: 1, Unit: "unidad"}); err == nil {
		t.Fatal("expected validation error")
	}
}
