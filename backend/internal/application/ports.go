package application

import (
	"calorietrack/backend/internal/domain"
	"context"
)

type Repository interface {
	GetSettings(context.Context) (domain.UserSettings, error)
	UpdateSettings(context.Context, domain.UserSettings) (domain.UserSettings, error)
	EnsureDay(context.Context, string, *float64, []domain.NewTask) error
	GetDayTarget(context.Context, string) (*float64, error)
	GetTasks(context.Context, string) ([]domain.DailyMealTask, error)
	GetEntries(context.Context, string) ([]domain.MealEntry, error)
	GetDateRange(context.Context, string, string) (map[string][]domain.DailyMealTask, map[string][]domain.MealEntry, map[string]*float64, error)
	ListFoods(context.Context, string) ([]domain.Food, error)
	GetFood(context.Context, int64) (domain.Food, error)
	UpdateFood(context.Context, int64, domain.FoodReferenceUpdate) (domain.Food, error)
	GetPortion(context.Context, int64) (domain.FoodPortion, error)
	CreatePortion(context.Context, int64, domain.FoodPortionInput) (domain.FoodPortion, error)
	UpdatePortion(context.Context, int64, domain.FoodPortionInput) (domain.FoodPortion, error)
	SetDefaultPortion(context.Context, int64) (domain.FoodPortion, error)
	DeletePortion(context.Context, int64) error
	ListRecipes(context.Context, string) ([]domain.Recipe, error)
	GetRecipe(context.Context, int64) (domain.Recipe, error)
	CreateRecipe(context.Context, domain.RecipeInput, []domain.RecipeIngredient) (domain.Recipe, error)
	UpdateRecipe(context.Context, int64, domain.RecipeInput, []domain.RecipeIngredient) (domain.Recipe, error)
	DeleteRecipe(context.Context, int64) error
	CreateMealEntry(context.Context, domain.CreateMealEntryInput, []domain.MealEntryItem) (domain.MealEntry, error)
	UpdateMealEntry(context.Context, int64, []domain.MealEntryItem) (domain.MealEntry, error)
	DeleteMealEntry(context.Context, int64) error
	UpdateTaskStatus(context.Context, int64, domain.TaskStatus) error
	CreateTask(context.Context, string, string) (domain.DailyMealTask, error)
}
