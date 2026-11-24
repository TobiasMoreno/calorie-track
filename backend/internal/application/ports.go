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
	CreateMealEntry(context.Context, domain.CreateMealEntryInput, []domain.MealEntryItem) (domain.MealEntry, error)
	DeleteMealEntry(context.Context, int64) error
	UpdateTaskStatus(context.Context, int64, domain.TaskStatus) error
	CreateTask(context.Context, string, string) (domain.DailyMealTask, error)
}
