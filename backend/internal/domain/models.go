package domain

import "time"

const DefaultTimezone = "America/Argentina/Cordoba"

var DefaultMeals = []string{"Desayuno", "Almuerzo", "Merienda", "Cena"}

type UserSettings struct {
	DailyCalorieTarget *float64  `json:"dailyCalorieTarget"`
	Timezone           string    `json:"timezone"`
	DefaultMeals       []string  `json:"defaultMeals"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskRegistered TaskStatus = "registered"
	TaskSkipped    TaskStatus = "skipped"
)

type DailyMealTask struct {
	ID        int64      `json:"id"`
	Date      string     `json:"date"`
	Name      string     `json:"name"`
	SlotKey   string     `json:"slotKey"`
	Position  int        `json:"position"`
	Status    TaskStatus `json:"status"`
	EntryID   *int64     `json:"entryId,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type Food struct {
	ID              int64         `json:"id"`
	Name            string        `json:"name"`
	Category        string        `json:"category"`
	CaloriesPer100G float64       `json:"caloriesPer100g"`
	ReferenceSource string        `json:"referenceSource"`
	ReferenceDetail string        `json:"referenceDetail"`
	Editable        bool          `json:"editable"`
	Portions        []FoodPortion `json:"portions"`
}

type FoodPortion struct {
	ID        int64   `json:"id"`
	FoodID    int64   `json:"foodId"`
	Name      string  `json:"name"`
	Unit      string  `json:"unit"`
	Size      string  `json:"size"`
	Grams     float64 `json:"grams"`
	IsDefault bool    `json:"isDefault"`
	IsCustom  bool    `json:"isCustom"`
}

type FoodPortionInput struct {
	Name      string  `json:"name"`
	Unit      string  `json:"unit"`
	Size      string  `json:"size,omitempty"`
	Grams     float64 `json:"grams"`
	IsDefault bool    `json:"isDefault"`
}

type FoodReferenceUpdate struct {
	CaloriesPer100G float64 `json:"caloriesPer100g"`
	ReferenceDetail string  `json:"referenceDetail"`
}

type MealEntryItem struct {
	ID                      int64   `json:"id"`
	FoodID                  int64   `json:"foodId"`
	FoodNameSnapshot        string  `json:"foodName"`
	Quantity                float64 `json:"quantity"`
	Unit                    string  `json:"unit"`
	PortionID               *int64  `json:"portionId,omitempty"`
	PortionNameSnapshot     string  `json:"portionName,omitempty"`
	EstimatedGrams          float64 `json:"estimatedGrams"`
	CaloriesPer100GSnapshot float64 `json:"caloriesPer100g"`
	CalculatedCalories      float64 `json:"calories"`
	CookingMethod           string  `json:"cookingMethod,omitempty"`
	Notes                   string  `json:"notes,omitempty"`
}

type MealEntry struct {
	ID                 int64           `json:"id"`
	Date               string          `json:"date"`
	TaskID             int64           `json:"taskId"`
	MealName           string          `json:"mealName"`
	RecipeID           *int64          `json:"recipeId,omitempty"`
	RecipeNameSnapshot string          `json:"recipeName,omitempty"`
	RecipeMultiplier   *float64        `json:"recipeMultiplier,omitempty"`
	TotalCalories      float64         `json:"totalCalories"`
	Items              []MealEntryItem `json:"items"`
	CreatedAt          time.Time       `json:"createdAt"`
}

type RecipeIngredient struct {
	ID                 int64   `json:"id"`
	RecipeID           int64   `json:"recipeId"`
	FoodID             int64   `json:"foodId"`
	FoodName           string  `json:"foodName"`
	Quantity           float64 `json:"quantity"`
	Unit               string  `json:"unit"`
	PortionID          *int64  `json:"portionId,omitempty"`
	PortionName        string  `json:"portionName,omitempty"`
	EstimatedGrams     float64 `json:"estimatedGrams"`
	CaloriesPer100G    float64 `json:"caloriesPer100g"`
	CalculatedCalories float64 `json:"calories"`
	CookingMethod      string  `json:"cookingMethod,omitempty"`
	Notes              string  `json:"notes,omitempty"`
}

type Recipe struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	Servings           float64            `json:"servings"`
	Notes              string             `json:"notes,omitempty"`
	Ingredients        []RecipeIngredient `json:"ingredients"`
	TotalCalories      float64            `json:"totalCalories"`
	CaloriesPerServing float64            `json:"caloriesPerServing"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
}

type RecipeInput struct {
	Name     string            `json:"name"`
	Servings float64           `json:"servings"`
	Notes    string            `json:"notes,omitempty"`
	Items    []CreateItemInput `json:"items"`
}

type RecipeMealEntryInput struct {
	Date       string  `json:"date"`
	TaskID     int64   `json:"taskId"`
	RecipeID   int64   `json:"recipeId"`
	Multiplier float64 `json:"multiplier"`
}

type DaySummary struct {
	Date            string   `json:"date"`
	Calories        float64  `json:"calories"`
	Target          *float64 `json:"target"`
	TaskCount       int      `json:"taskCount"`
	RegisteredTasks int      `json:"registeredTasks"`
	SkippedTasks    int      `json:"skippedTasks"`
	PendingTasks    int      `json:"pendingTasks"`
	HasData         bool     `json:"hasData"`
}

type CalendarMonth struct {
	Month string       `json:"month"`
	Days  []DaySummary `json:"days"`
}

type Dashboard struct {
	Date              string          `json:"date"`
	TargetCalories    *float64        `json:"targetCalories"`
	ConsumedCalories  float64         `json:"consumedCalories"`
	RemainingCalories *float64        `json:"remainingCalories"`
	Tasks             []DailyMealTask `json:"tasks"`
	Entries           []MealEntry     `json:"entries"`
	LastSevenDays     []DaySummary    `json:"lastSevenDays"`
}

type CreateItemInput struct {
	FoodID        int64   `json:"foodId"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	PortionID     *int64  `json:"portionId,omitempty"`
	CookingMethod string  `json:"cookingMethod,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

type CreateMealEntryInput struct {
	Date             string            `json:"date"`
	TaskID           int64             `json:"taskId"`
	Items            []CreateItemInput `json:"items"`
	RecipeID         *int64            `json:"-"`
	RecipeName       string            `json:"-"`
	RecipeMultiplier *float64          `json:"-"`
}

type UpdateMealEntryInput struct {
	Items []CreateItemInput `json:"items"`
}
