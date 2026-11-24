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

func (s *Service) Calendar(ctx context.Context, month string) (domain.CalendarMonth, error) {
	first, err := time.Parse("2006-01", month)
	if err != nil || first.Format("2006-01") != month {
		return domain.CalendarMonth{}, fmt.Errorf("%w: el mes debe usar YYYY-MM", domain.ErrValidation)
	}
	last := first.AddDate(0, 1, -1)
	startDate := first.Format("2006-01-02")
	endDate := last.Format("2006-01-02")
	taskMap, entryMap, targetMap, err := s.repo.GetDateRange(ctx, startDate, endDate)
	if err != nil {
		return domain.CalendarMonth{}, err
	}
	days := make([]domain.DaySummary, 0, last.Day())
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		days = append(days, domain.AggregateDay(date, targetMap[date], taskMap[date], entryMap[date]))
	}
	return domain.CalendarMonth{Month: month, Days: days}, nil
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

func (s *Service) CreatePortion(ctx context.Context, foodID int64, input domain.FoodPortionInput) (domain.FoodPortion, error) {
	if foodID <= 0 {
		return domain.FoodPortion{}, fmt.Errorf("%w: alimento inválido", domain.ErrValidation)
	}
	if err := validatePortionInput(&input); err != nil {
		return domain.FoodPortion{}, err
	}
	if _, err := s.repo.GetFood(ctx, foodID); err != nil {
		return domain.FoodPortion{}, err
	}
	return s.repo.CreatePortion(ctx, foodID, input)
}

func (s *Service) UpdatePortion(ctx context.Context, id int64, input domain.FoodPortionInput) (domain.FoodPortion, error) {
	if id <= 0 {
		return domain.FoodPortion{}, fmt.Errorf("%w: porción inválida", domain.ErrValidation)
	}
	if err := validatePortionInput(&input); err != nil {
		return domain.FoodPortion{}, err
	}
	return s.repo.UpdatePortion(ctx, id, input)
}

func (s *Service) SetDefaultPortion(ctx context.Context, id int64) (domain.FoodPortion, error) {
	if id <= 0 {
		return domain.FoodPortion{}, fmt.Errorf("%w: porción inválida", domain.ErrValidation)
	}
	return s.repo.SetDefaultPortion(ctx, id)
}

func (s *Service) DeletePortion(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: porción inválida", domain.ErrValidation)
	}
	return s.repo.DeletePortion(ctx, id)
}

func validatePortionInput(input *domain.FoodPortionInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Unit = strings.TrimSpace(input.Unit)
	input.Size = strings.TrimSpace(input.Size)
	if input.Size == "" {
		input.Size = "personal"
	}
	if len(input.Name) < 2 || len(input.Name) > 80 {
		return fmt.Errorf("%w: el nombre de la porción debe tener entre 2 y 80 caracteres", domain.ErrValidation)
	}
	if input.Unit != "unidad" && input.Unit != "porción" && input.Unit != "cucharada" {
		return fmt.Errorf("%w: unidad de porción inválida", domain.ErrValidation)
	}
	if len(input.Size) > 40 {
		return fmt.Errorf("%w: tamaño de porción inválido", domain.ErrValidation)
	}
	if input.Grams < 0.1 || input.Grams > 5000 {
		return fmt.Errorf("%w: el gramaje debe estar entre 0,1 y 5000 g", domain.ErrValidation)
	}
	return nil
}

func (s *Service) CreateMealEntry(ctx context.Context, input domain.CreateMealEntryInput) (domain.MealEntry, error) {
	if _, err := parseDate(input.Date); err != nil {
		return domain.MealEntry{}, err
	}
	if input.TaskID <= 0 || len(input.Items) == 0 || len(input.Items) > 20 {
		return domain.MealEntry{}, fmt.Errorf("%w: la comida debe tener al menos un alimento", domain.ErrValidation)
	}
	items, err := s.estimateItems(ctx, input.Items)
	if err != nil {
		return domain.MealEntry{}, err
	}
	return s.repo.CreateMealEntry(ctx, input, items)
}

func (s *Service) UpdateMealEntry(ctx context.Context, id int64, input domain.UpdateMealEntryInput) (domain.MealEntry, error) {
	if id <= 0 || len(input.Items) == 0 || len(input.Items) > 20 {
		return domain.MealEntry{}, fmt.Errorf("%w: la comida debe tener al menos un alimento", domain.ErrValidation)
	}
	items, err := s.estimateItems(ctx, input.Items)
	if err != nil {
		return domain.MealEntry{}, err
	}
	return s.repo.UpdateMealEntry(ctx, id, items)
}

func (s *Service) estimateItems(ctx context.Context, inputs []domain.CreateItemInput) ([]domain.MealEntryItem, error) {
	items := make([]domain.MealEntryItem, 0, len(inputs))
	for _, itemInput := range inputs {
		food, err := s.repo.GetFood(ctx, itemInput.FoodID)
		if err != nil {
			return nil, err
		}
		var portion *domain.FoodPortion
		if itemInput.PortionID != nil {
			value, err := s.repo.GetPortion(ctx, *itemInput.PortionID)
			if err != nil {
				return nil, err
			}
			if value.FoodID != food.ID {
				return nil, fmt.Errorf("%w: la porción no corresponde al alimento", domain.ErrValidation)
			}
			portion = &value
		}
		item, err := domain.EstimateItem(food, portion, itemInput)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) ListRecipes(ctx context.Context, query string) ([]domain.Recipe, error) {
	return s.repo.ListRecipes(ctx, strings.TrimSpace(query))
}

func (s *Service) CreateRecipe(ctx context.Context, input domain.RecipeInput) (domain.Recipe, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Notes = strings.TrimSpace(input.Notes)
	if err := validateRecipeInput(input); err != nil {
		return domain.Recipe{}, err
	}
	items, err := s.estimateItems(ctx, input.Items)
	if err != nil {
		return domain.Recipe{}, err
	}
	return s.repo.CreateRecipe(ctx, input, recipeIngredients(items))
}

func (s *Service) UpdateRecipe(ctx context.Context, id int64, input domain.RecipeInput) (domain.Recipe, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Notes = strings.TrimSpace(input.Notes)
	if id <= 0 {
		return domain.Recipe{}, fmt.Errorf("%w: receta inválida", domain.ErrValidation)
	}
	if err := validateRecipeInput(input); err != nil {
		return domain.Recipe{}, err
	}
	items, err := s.estimateItems(ctx, input.Items)
	if err != nil {
		return domain.Recipe{}, err
	}
	return s.repo.UpdateRecipe(ctx, id, input, recipeIngredients(items))
}

func (s *Service) DeleteRecipe(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: receta inválida", domain.ErrValidation)
	}
	return s.repo.DeleteRecipe(ctx, id)
}

func (s *Service) CreateRecipeMealEntry(ctx context.Context, input domain.RecipeMealEntryInput) (domain.MealEntry, error) {
	if _, err := parseDate(input.Date); err != nil {
		return domain.MealEntry{}, err
	}
	if input.TaskID <= 0 || input.RecipeID <= 0 || input.Multiplier < 0.1 || input.Multiplier > 20 {
		return domain.MealEntry{}, fmt.Errorf("%w: datos de registro frecuente inválidos", domain.ErrValidation)
	}
	recipe, err := s.repo.GetRecipe(ctx, input.RecipeID)
	if err != nil {
		return domain.MealEntry{}, err
	}
	items := make([]domain.MealEntryItem, 0, len(recipe.Ingredients))
	for _, ingredient := range recipe.Ingredients {
		grams := domain.RoundCalories(ingredient.EstimatedGrams * input.Multiplier)
		items = append(items, domain.MealEntryItem{
			FoodID:                  ingredient.FoodID,
			FoodNameSnapshot:        ingredient.FoodName,
			Quantity:                domain.RoundCalories(ingredient.Quantity * input.Multiplier),
			Unit:                    ingredient.Unit,
			PortionID:               ingredient.PortionID,
			PortionNameSnapshot:     ingredient.PortionName,
			EstimatedGrams:          grams,
			CaloriesPer100GSnapshot: ingredient.CaloriesPer100G,
			CalculatedCalories:      domain.RoundCalories(grams * ingredient.CaloriesPer100G / 100),
			CookingMethod:           ingredient.CookingMethod,
			Notes:                   ingredient.Notes,
		})
	}
	recipeID := recipe.ID
	multiplier := input.Multiplier
	return s.repo.CreateMealEntry(ctx, domain.CreateMealEntryInput{
		Date: input.Date, TaskID: input.TaskID, RecipeID: &recipeID, RecipeName: recipe.Name, RecipeMultiplier: &multiplier,
	}, items)
}

func validateRecipeInput(input domain.RecipeInput) error {
	if len(input.Name) < 2 || len(input.Name) > 100 {
		return fmt.Errorf("%w: el nombre de la receta debe tener entre 2 y 100 caracteres", domain.ErrValidation)
	}
	if input.Servings < 0.1 || input.Servings > 100 {
		return fmt.Errorf("%w: las porciones deben estar entre 0,1 y 100", domain.ErrValidation)
	}
	if len(input.Items) == 0 || len(input.Items) > 30 {
		return fmt.Errorf("%w: la receta debe tener entre 1 y 30 ingredientes", domain.ErrValidation)
	}
	if len(input.Notes) > 500 {
		return fmt.Errorf("%w: las notas no pueden superar 500 caracteres", domain.ErrValidation)
	}
	return nil
}

func recipeIngredients(items []domain.MealEntryItem) []domain.RecipeIngredient {
	result := make([]domain.RecipeIngredient, 0, len(items))
	for _, item := range items {
		result = append(result, domain.RecipeIngredient{
			FoodID: item.FoodID, FoodName: item.FoodNameSnapshot, Quantity: item.Quantity, Unit: item.Unit,
			PortionID: item.PortionID, PortionName: item.PortionNameSnapshot, EstimatedGrams: item.EstimatedGrams,
			CaloriesPer100G: item.CaloriesPer100GSnapshot, CalculatedCalories: item.CalculatedCalories,
			CookingMethod: item.CookingMethod, Notes: item.Notes,
		})
	}
	return result
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
