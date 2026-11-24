package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"calorietrack/backend/internal/domain"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) GetSettings(ctx context.Context) (domain.UserSettings, error) {
	return s.repo.GetSettings(ctx)
}

func (s *Service) UpdateSettings(ctx context.Context, settings domain.UserSettings) (domain.UserSettings, error) {
	if settings.DailyCalorieTarget != nil && (*settings.DailyCalorieTarget < 500 || *settings.DailyCalorieTarget > 10000) {
		return domain.UserSettings{}, fmt.Errorf("%w: el objetivo debe estar entre 500 y 10000 kcal", domain.ErrValidation)
	}
	if settings.Timezone == "" {
		settings.Timezone = domain.DefaultTimezone
	}
	if _, err := time.LoadLocation(settings.Timezone); err != nil {
		return domain.UserSettings{}, fmt.Errorf("%w: zona horaria inválida", domain.ErrValidation)
	}
	if len(settings.DefaultMeals) == 0 {
		settings.DefaultMeals = domain.DefaultMeals
	}
	if _, err := domain.BuildDefaultTasks(settings.DefaultMeals); err != nil {
		return domain.UserSettings{}, err
	}
	return s.repo.UpdateSettings(ctx, settings)
}

func (s *Service) Dashboard(ctx context.Context, date string) (domain.Dashboard, error) {
	parsed, err := parseDate(date)
	if err != nil {
		return domain.Dashboard{}, err
	}
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return domain.Dashboard{}, err
	}
	defaults, err := domain.BuildDefaultTasks(settings.DefaultMeals)
	if err != nil {
		return domain.Dashboard{}, err
	}
	if err := s.repo.EnsureDay(ctx, date, settings.DailyCalorieTarget, defaults); err != nil {
		return domain.Dashboard{}, err
	}
	target, err := s.repo.GetDayTarget(ctx, date)
	if err != nil {
		return domain.Dashboard{}, err
	}
	tasks, err := s.repo.GetTasks(ctx, date)
	if err != nil {
		return domain.Dashboard{}, err
	}
	entries, err := s.repo.GetEntries(ctx, date)
	if err != nil {
		return domain.Dashboard{}, err
	}
	consumed := domain.SumCaloriesFromEntries(entries)
	var remaining *float64
	if target != nil {
		value := domain.RoundCalories(*target - consumed)
		remaining = &value
	}
	start := parsed.AddDate(0, 0, -6).Format("2006-01-02")
	taskMap, entryMap, targetMap, err := s.repo.GetDateRange(ctx, start, date)
	if err != nil {
		return domain.Dashboard{}, err
	}
	summaries := make([]domain.DaySummary, 0, 7)
	for i := 6; i >= 0; i-- {
		day := parsed.AddDate(0, 0, -i).Format("2006-01-02")
		summaries = append(summaries, domain.AggregateDay(day, targetMap[day], taskMap[day], entryMap[day]))
	}
	return domain.Dashboard{Date: date, TargetCalories: target, ConsumedCalories: consumed, RemainingCalories: remaining, Tasks: tasks, Entries: entries, LastSevenDays: summaries}, nil
}

func (s *Service) ListFoods(ctx context.Context, query string) ([]domain.Food, error) {
	return s.repo.ListFoods(ctx, strings.TrimSpace(query))
}

func (s *Service) UpdateFood(ctx context.Context, id int64, input domain.FoodReferenceUpdate) (domain.Food, error) {
	input.ReferenceDetail = strings.TrimSpace(input.ReferenceDetail)
	if id <= 0 || input.CaloriesPer100G < 0 || input.CaloriesPer100G > 2000 {
		return domain.Food{}, fmt.Errorf("%w: el valor debe estar entre 0 y 2000 kcal cada 100 g", domain.ErrValidation)
	}
	if len(input.ReferenceDetail) < 5 || len(input.ReferenceDetail) > 500 {
		return domain.Food{}, fmt.Errorf("%w: describí brevemente la fuente o el motivo del ajuste", domain.ErrValidation)
	}
	return s.repo.UpdateFood(ctx, id, input)
}

func (s *Service) CreateMealEntry(ctx context.Context, input domain.CreateMealEntryInput) (domain.MealEntry, error) {
	if _, err := parseDate(input.Date); err != nil {
		return domain.MealEntry{}, err
	}
	if input.TaskID <= 0 || len(input.Items) == 0 || len(input.Items) > 20 {
		return domain.MealEntry{}, fmt.Errorf("%w: la comida debe tener al menos un alimento", domain.ErrValidation)
	}
	items := make([]domain.MealEntryItem, 0, len(input.Items))
	for _, itemInput := range input.Items {
		food, err := s.repo.GetFood(ctx, itemInput.FoodID)
		if err != nil {
			return domain.MealEntry{}, err
		}
		var portion *domain.FoodPortion
		if itemInput.PortionID != nil {
			value, err := s.repo.GetPortion(ctx, *itemInput.PortionID)
			if err != nil {
				return domain.MealEntry{}, err
			}
			portion = &value
		}
		item, err := domain.EstimateItem(food, portion, itemInput)
		if err != nil {
			return domain.MealEntry{}, err
		}
		items = append(items, item)
	}
	return s.repo.CreateMealEntry(ctx, input, items)
}

func (s *Service) DeleteMealEntry(ctx context.Context, id int64) error {
	return s.repo.DeleteMealEntry(ctx, id)
}

func (s *Service) UpdateTaskStatus(ctx context.Context, id int64, status domain.TaskStatus) error {
	if status != domain.TaskPending && status != domain.TaskSkipped {
		return fmt.Errorf("%w: estado de tarea inválido", domain.ErrValidation)
	}
	return s.repo.UpdateTaskStatus(ctx, id, status)
}

func (s *Service) CreateTask(ctx context.Context, date, name string) (domain.DailyMealTask, error) {
	if _, err := parseDate(date); err != nil {
		return domain.DailyMealTask{}, err
	}
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 60 {
		return domain.DailyMealTask{}, fmt.Errorf("%w: nombre inválido", domain.ErrValidation)
	}
	return s.repo.CreateTask(ctx, date, name)
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, fmt.Errorf("%w: la fecha debe usar YYYY-MM-DD", domain.ErrValidation)
	}
	return parsed, nil
}
