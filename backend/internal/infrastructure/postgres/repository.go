package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"calorietrack/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) GetSettings(ctx context.Context) (domain.UserSettings, error) {
	var settings domain.UserSettings
	var target sql.NullFloat64
	var mealsJSON string
	err := r.db.QueryRowContext(ctx, `SELECT daily_calorie_target, timezone, default_meals_json, updated_at FROM user_settings WHERE id=1`).Scan(&target, &settings.Timezone, &mealsJSON, &settings.UpdatedAt)
	if err != nil {
		return settings, err
	}
	if target.Valid {
		settings.DailyCalorieTarget = &target.Float64
	}
	if err := json.Unmarshal([]byte(mealsJSON), &settings.DefaultMeals); err != nil {
		return settings, err
	}
	return settings, nil
}

func (r *Repository) UpdateSettings(ctx context.Context, settings domain.UserSettings) (domain.UserSettings, error) {
	meals, err := json.Marshal(settings.DefaultMeals)
	if err != nil {
		return settings, err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE user_settings SET daily_calorie_target=$1, timezone=$2, default_meals_json=$3, updated_at=CURRENT_TIMESTAMP WHERE id=1`, settings.DailyCalorieTarget, settings.Timezone, string(meals))
	if err != nil {
		return settings, err
	}
	return r.GetSettings(ctx)
}

func (r *Repository) EnsureDay(ctx context.Context, date string, target *float64, tasks []domain.NewTask) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO daily_days(local_date, calorie_target_snapshot) VALUES ($1::date, $2) ON CONFLICT(local_date) DO UPDATE SET calorie_target_snapshot=COALESCE(daily_days.calorie_target_snapshot, EXCLUDED.calorie_target_snapshot)`, date, target)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		_, err = tx.ExecContext(ctx, `INSERT INTO daily_meal_tasks(local_date,name,slot_key,position) VALUES ($1::date,$2,$3,$4) ON CONFLICT(local_date,slot_key) DO NOTHING`, date, task.Name, task.SlotKey, task.Position)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) GetDayTarget(ctx context.Context, date string) (*float64, error) {
	var value sql.NullFloat64
	if err := r.db.QueryRowContext(ctx, `SELECT calorie_target_snapshot FROM daily_days WHERE local_date=$1::date`, date).Scan(&value); err != nil {
		return nil, mapError(err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Float64, nil
}

func (r *Repository) GetTasks(ctx context.Context, date string) ([]domain.DailyMealTask, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT t.id,t.local_date::text,t.name,t.slot_key,t.position,t.status,e.id,t.created_at FROM daily_meal_tasks t LEFT JOIN meal_entries e ON e.task_id=t.id WHERE t.local_date=$1::date ORDER BY t.position,t.id`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []domain.DailyMealTask{}
	for rows.Next() {
		var task domain.DailyMealTask
		var entryID sql.NullInt64
		if err := rows.Scan(&task.ID, &task.Date, &task.Name, &task.SlotKey, &task.Position, &task.Status, &entryID, &task.CreatedAt); err != nil {
			return nil, err
		}
		if entryID.Valid {
			task.EntryID = &entryID.Int64
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *Repository) GetEntries(ctx context.Context, date string) ([]domain.MealEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,local_date::text,task_id,meal_name_snapshot,recipe_id,recipe_name_snapshot,recipe_multiplier,total_calories,created_at FROM meal_entries WHERE local_date=$1::date ORDER BY created_at,id`, date)
	if err != nil {
		return nil, err
	}
	entries := []domain.MealEntry{}
	for rows.Next() {
		var entry domain.MealEntry
		var recipeID sql.NullInt64
		var multiplier sql.NullFloat64
		if err := rows.Scan(&entry.ID, &entry.Date, &entry.TaskID, &entry.MealName, &recipeID, &entry.RecipeNameSnapshot, &multiplier, &entry.TotalCalories, &entry.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		if recipeID.Valid {
			entry.RecipeID = &recipeID.Int64
		}
		if multiplier.Valid {
			entry.RecipeMultiplier = &multiplier.Float64
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range entries {
		items, err := r.getEntryItems(ctx, entries[index].ID)
		if err != nil {
			return nil, err
		}
		entries[index].Items = items
	}
	return entries, nil
}

func (r *Repository) getEntryItems(ctx context.Context, entryID int64) ([]domain.MealEntryItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,food_id,food_name_snapshot,quantity,unit,portion_id,portion_name_snapshot,estimated_grams,calories_per_100g_snapshot,calculated_calories,cooking_method,notes FROM meal_entry_items WHERE meal_entry_id=$1 ORDER BY position,id`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.MealEntryItem{}
	for rows.Next() {
		var item domain.MealEntryItem
		var portionID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.FoodID, &item.FoodNameSnapshot, &item.Quantity, &item.Unit, &portionID, &item.PortionNameSnapshot, &item.EstimatedGrams, &item.CaloriesPer100GSnapshot, &item.CalculatedCalories, &item.CookingMethod, &item.Notes); err != nil {
			return nil, err
		}
		if portionID.Valid {
			item.PortionID = &portionID.Int64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetDateRange(ctx context.Context, start, end string) (map[string][]domain.DailyMealTask, map[string][]domain.MealEntry, map[string]*float64, error) {
	tasks := map[string][]domain.DailyMealTask{}
	entries := map[string][]domain.MealEntry{}
	targets := map[string]*float64{}

	dayRows, err := r.db.QueryContext(ctx, `SELECT local_date::text,calorie_target_snapshot FROM daily_days WHERE local_date BETWEEN $1::date AND $2::date`, start, end)
	if err != nil {
		return nil, nil, nil, err
	}
	for dayRows.Next() {
		var date string
		var target sql.NullFloat64
		if err := dayRows.Scan(&date, &target); err != nil {
			dayRows.Close()
			return nil, nil, nil, err
		}
		if target.Valid {
			value := target.Float64
			targets[date] = &value
		}
	}
	if err := dayRows.Err(); err != nil {
		dayRows.Close()
		return nil, nil, nil, err
	}
	dayRows.Close()

	rows, err := r.db.QueryContext(ctx, `SELECT id,local_date::text,name,slot_key,position,status,created_at FROM daily_meal_tasks WHERE local_date BETWEEN $1::date AND $2::date ORDER BY local_date,position`, start, end)
	if err != nil {
		return nil, nil, nil, err
	}
	for rows.Next() {
		var task domain.DailyMealTask
		if err := rows.Scan(&task.ID, &task.Date, &task.Name, &task.SlotKey, &task.Position, &task.Status, &task.CreatedAt); err != nil {
			rows.Close()
			return nil, nil, nil, err
		}
		tasks[task.Date] = append(tasks[task.Date], task)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, nil, err
	}
	rows.Close()

	entryRows, err := r.db.QueryContext(ctx, `SELECT id,local_date::text,task_id,meal_name_snapshot,total_calories,created_at FROM meal_entries WHERE local_date BETWEEN $1::date AND $2::date ORDER BY local_date,created_at`, start, end)
	if err != nil {
		return nil, nil, nil, err
	}
	for entryRows.Next() {
		var entry domain.MealEntry
		if err := entryRows.Scan(&entry.ID, &entry.Date, &entry.TaskID, &entry.MealName, &entry.TotalCalories, &entry.CreatedAt); err != nil {
			entryRows.Close()
			return nil, nil, nil, err
		}
		entries[entry.Date] = append(entries[entry.Date], entry)
	}
	if err := entryRows.Err(); err != nil {
		entryRows.Close()
		return nil, nil, nil, err
	}
	entryRows.Close()
	return tasks, entries, targets, nil
}

func (r *Repository) ListFoods(ctx context.Context, query string) ([]domain.Food, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,category,calories_per_100g,reference_source,reference_detail,editable FROM foods WHERE lower(name) LIKE $1 ORDER BY CASE WHEN lower(name)=lower($2) THEN 0 ELSE 1 END,name LIMIT 50`, pattern, query)
	if err != nil {
		return nil, err
	}
	foods := []domain.Food{}
	for rows.Next() {
		var food domain.Food
		if err := rows.Scan(&food.ID, &food.Name, &food.Category, &food.CaloriesPer100G, &food.ReferenceSource, &food.ReferenceDetail, &food.Editable); err != nil {
			rows.Close()
			return nil, err
		}
		foods = append(foods, food)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range foods {
		portions, err := r.listPortions(ctx, foods[index].ID)
		if err != nil {
			return nil, err
		}
		foods[index].Portions = portions
	}
	return foods, nil
}

func (r *Repository) GetFood(ctx context.Context, id int64) (domain.Food, error) {
	var food domain.Food
	err := r.db.QueryRowContext(ctx, `SELECT id,name,category,calories_per_100g,reference_source,reference_detail,editable FROM foods WHERE id=$1`, id).Scan(&food.ID, &food.Name, &food.Category, &food.CaloriesPer100G, &food.ReferenceSource, &food.ReferenceDetail, &food.Editable)
	if err != nil {
		return food, mapError(err)
	}
	food.Portions, err = r.listPortions(ctx, id)
	return food, err
}

func (r *Repository) UpdateFood(ctx context.Context, id int64, input domain.FoodReferenceUpdate) (domain.Food, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE foods SET calories_per_100g=$1, reference_source='Ajuste personal', reference_detail=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3 AND editable=TRUE`, input.CaloriesPer100G, input.ReferenceDetail, id)
	if err != nil {
		return domain.Food{}, err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.Food{}, domain.ErrNotFound
	}
	return r.GetFood(ctx, id)
}

func (r *Repository) listPortions(ctx context.Context, foodID int64) ([]domain.FoodPortion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,food_id,name,unit,size,grams,is_default,is_custom FROM food_portions WHERE food_id=$1 ORDER BY is_default DESC,grams,name`, foodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.FoodPortion{}
	for rows.Next() {
		var portion domain.FoodPortion
		if err := rows.Scan(&portion.ID, &portion.FoodID, &portion.Name, &portion.Unit, &portion.Size, &portion.Grams, &portion.IsDefault, &portion.IsCustom); err != nil {
			return nil, err
		}
		result = append(result, portion)
	}
	return result, rows.Err()
}

func (r *Repository) GetPortion(ctx context.Context, id int64) (domain.FoodPortion, error) {
	var portion domain.FoodPortion
	err := r.db.QueryRowContext(ctx, `SELECT id,food_id,name,unit,size,grams,is_default,is_custom FROM food_portions WHERE id=$1`, id).Scan(&portion.ID, &portion.FoodID, &portion.Name, &portion.Unit, &portion.Size, &portion.Grams, &portion.IsDefault, &portion.IsCustom)
	return portion, mapError(err)
}

func (r *Repository) CreatePortion(ctx context.Context, foodID int64, input domain.FoodPortionInput) (domain.FoodPortion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.FoodPortion{}, err
	}
	defer tx.Rollback()
	if input.IsDefault {
		if _, err = tx.ExecContext(ctx, `UPDATE food_portions SET is_default=FALSE WHERE food_id=$1`, foodID); err != nil {
			return domain.FoodPortion{}, err
		}
	}
	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO food_portions(food_id,name,unit,size,grams,is_default,is_custom) VALUES ($1,$2,$3,$4,$5,$6,TRUE) RETURNING id`, foodID, input.Name, input.Unit, input.Size, input.Grams, input.IsDefault).Scan(&id)
	if err != nil {
		return domain.FoodPortion{}, mapError(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.FoodPortion{}, err
	}
	return r.GetPortion(ctx, id)
}

func (r *Repository) UpdatePortion(ctx context.Context, id int64, input domain.FoodPortionInput) (domain.FoodPortion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.FoodPortion{}, err
	}
	defer tx.Rollback()
	var foodID int64
	var wasDefault, isCustom bool
	if err = tx.QueryRowContext(ctx, `SELECT food_id,is_default,is_custom FROM food_portions WHERE id=$1 FOR UPDATE`, id).Scan(&foodID, &wasDefault, &isCustom); err != nil {
		return domain.FoodPortion{}, mapError(err)
	}
	if !isCustom {
		return domain.FoodPortion{}, fmt.Errorf("%w: sólo se pueden editar porciones personales", domain.ErrValidation)
	}
	if input.IsDefault {
		if _, err = tx.ExecContext(ctx, `UPDATE food_portions SET is_default=FALSE WHERE food_id=$1`, foodID); err != nil {
			return domain.FoodPortion{}, err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE food_portions SET name=$1,unit=$2,size=$3,grams=$4,is_default=$5 WHERE id=$6`, input.Name, input.Unit, input.Size, input.Grams, input.IsDefault, id)
	if err != nil {
		return domain.FoodPortion{}, mapError(err)
	}
	if wasDefault && !input.IsDefault {
		if _, err = tx.ExecContext(ctx, `UPDATE food_portions SET is_default=TRUE WHERE id=(SELECT id FROM food_portions WHERE food_id=$1 AND id<>$2 ORDER BY grams,id LIMIT 1)`, foodID, id); err != nil {
			return domain.FoodPortion{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.FoodPortion{}, err
	}
	return r.GetPortion(ctx, id)
}

func (r *Repository) SetDefaultPortion(ctx context.Context, id int64) (domain.FoodPortion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.FoodPortion{}, err
	}
	defer tx.Rollback()
	var foodID int64
	if err = tx.QueryRowContext(ctx, `SELECT food_id FROM food_portions WHERE id=$1 FOR UPDATE`, id).Scan(&foodID); err != nil {
		return domain.FoodPortion{}, mapError(err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE food_portions SET is_default=(id=$1) WHERE food_id=$2`, id, foodID); err != nil {
		return domain.FoodPortion{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.FoodPortion{}, err
	}
	return r.GetPortion(ctx, id)
}

func (r *Repository) DeletePortion(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var foodID int64
	var isDefault, isCustom bool
	if err = tx.QueryRowContext(ctx, `SELECT food_id,is_default,is_custom FROM food_portions WHERE id=$1 FOR UPDATE`, id).Scan(&foodID, &isDefault, &isCustom); err != nil {
		return mapError(err)
	}
	if !isCustom {
		return fmt.Errorf("%w: sólo se pueden eliminar porciones personales", domain.ErrValidation)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM food_portions WHERE id=$1`, id); err != nil {
		return err
	}
	if isDefault {
		if _, err = tx.ExecContext(ctx, `UPDATE food_portions SET is_default=TRUE WHERE id=(SELECT id FROM food_portions WHERE food_id=$1 ORDER BY grams,id LIMIT 1)`, foodID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ListRecipes(ctx context.Context, query string) ([]domain.Recipe, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,servings,notes,created_at,updated_at FROM recipes WHERE lower(name) LIKE $1 ORDER BY updated_at DESC,name LIMIT 100`, pattern)
	if err != nil {
		return nil, err
	}
	recipes := []domain.Recipe{}
	for rows.Next() {
		var recipe domain.Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Servings, &recipe.Notes, &recipe.CreatedAt, &recipe.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range recipes {
		ingredients, err := r.getRecipeIngredients(ctx, recipes[index].ID)
		if err != nil {
			return nil, err
		}
		recipes[index].Ingredients = ingredients
		completeRecipeTotals(&recipes[index])
	}
	return recipes, nil
}

func (r *Repository) GetRecipe(ctx context.Context, id int64) (domain.Recipe, error) {
	var recipe domain.Recipe
	err := r.db.QueryRowContext(ctx, `SELECT id,name,servings,notes,created_at,updated_at FROM recipes WHERE id=$1`, id).Scan(&recipe.ID, &recipe.Name, &recipe.Servings, &recipe.Notes, &recipe.CreatedAt, &recipe.UpdatedAt)
	if err != nil {
		return recipe, mapError(err)
	}
	recipe.Ingredients, err = r.getRecipeIngredients(ctx, id)
	if err != nil {
		return recipe, err
	}
	completeRecipeTotals(&recipe)
	return recipe, nil
}

func (r *Repository) getRecipeIngredients(ctx context.Context, recipeID int64) ([]domain.RecipeIngredient, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT ri.id,ri.recipe_id,ri.food_id,f.name,ri.quantity,ri.unit,ri.portion_id,ri.portion_name_snapshot,ri.estimated_grams,f.calories_per_100g,ri.cooking_method,ri.notes FROM recipe_ingredients ri JOIN foods f ON f.id=ri.food_id WHERE ri.recipe_id=$1 ORDER BY ri.position,ri.id`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.RecipeIngredient{}
	for rows.Next() {
		var ingredient domain.RecipeIngredient
		var portionID sql.NullInt64
		if err := rows.Scan(&ingredient.ID, &ingredient.RecipeID, &ingredient.FoodID, &ingredient.FoodName, &ingredient.Quantity, &ingredient.Unit, &portionID, &ingredient.PortionName, &ingredient.EstimatedGrams, &ingredient.CaloriesPer100G, &ingredient.CookingMethod, &ingredient.Notes); err != nil {
			return nil, err
		}
		if portionID.Valid {
			ingredient.PortionID = &portionID.Int64
		}
		ingredient.CalculatedCalories = domain.RoundCalories(ingredient.EstimatedGrams * ingredient.CaloriesPer100G / 100)
		result = append(result, ingredient)
	}
	return result, rows.Err()
}

func completeRecipeTotals(recipe *domain.Recipe) {
	total := 0.0
	for _, ingredient := range recipe.Ingredients {
		total += ingredient.CalculatedCalories
	}
	recipe.TotalCalories = domain.RoundCalories(total)
	if recipe.Servings > 0 {
		recipe.CaloriesPerServing = domain.RoundCalories(total / recipe.Servings)
	}
}

func (r *Repository) CreateRecipe(ctx context.Context, input domain.RecipeInput, ingredients []domain.RecipeIngredient) (domain.Recipe, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Recipe{}, err
	}
	defer tx.Rollback()
	var id int64
	if err = tx.QueryRowContext(ctx, `INSERT INTO recipes(name,servings,notes) VALUES ($1,$2,$3) RETURNING id`, input.Name, input.Servings, input.Notes).Scan(&id); err != nil {
		return domain.Recipe{}, mapError(err)
	}
	if err := insertRecipeIngredients(ctx, tx, id, ingredients); err != nil {
		return domain.Recipe{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Recipe{}, err
	}
	return r.GetRecipe(ctx, id)
}

func (r *Repository) UpdateRecipe(ctx context.Context, id int64, input domain.RecipeInput, ingredients []domain.RecipeIngredient) (domain.Recipe, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Recipe{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE recipes SET name=$1,servings=$2,notes=$3,updated_at=CURRENT_TIMESTAMP WHERE id=$4`, input.Name, input.Servings, input.Notes, id)
	if err != nil {
		return domain.Recipe{}, mapError(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.Recipe{}, domain.ErrNotFound
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM recipe_ingredients WHERE recipe_id=$1`, id); err != nil {
		return domain.Recipe{}, err
	}
	if err := insertRecipeIngredients(ctx, tx, id, ingredients); err != nil {
		return domain.Recipe{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Recipe{}, err
	}
	return r.GetRecipe(ctx, id)
}

func insertRecipeIngredients(ctx context.Context, tx *sql.Tx, recipeID int64, ingredients []domain.RecipeIngredient) error {
	for index, ingredient := range ingredients {
		_, err := tx.ExecContext(ctx, `INSERT INTO recipe_ingredients(recipe_id,food_id,quantity,unit,estimated_grams,position,portion_id,portion_name_snapshot,cooking_method,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, recipeID, ingredient.FoodID, ingredient.Quantity, ingredient.Unit, ingredient.EstimatedGrams, index, ingredient.PortionID, ingredient.PortionName, ingredient.CookingMethod, ingredient.Notes)
		if err != nil {
			return mapError(err)
		}
	}
	return nil
}

func (r *Repository) DeleteRecipe(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM recipes WHERE id=$1`, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) CreateMealEntry(ctx context.Context, input domain.CreateMealEntryInput, items []domain.MealEntryItem) (domain.MealEntry, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.MealEntry{}, err
	}
	defer tx.Rollback()
	var taskDate, taskName, status string
	if err = tx.QueryRowContext(ctx, `SELECT local_date::text,name,status FROM daily_meal_tasks WHERE id=$1 FOR UPDATE`, input.TaskID).Scan(&taskDate, &taskName, &status); err != nil {
		return domain.MealEntry{}, mapError(err)
	}
	if taskDate != input.Date {
		return domain.MealEntry{}, fmt.Errorf("%w: la tarea no pertenece a la fecha", domain.ErrValidation)
	}
	if status == string(domain.TaskRegistered) {
		return domain.MealEntry{}, fmt.Errorf("%w: la comida ya está registrada", domain.ErrConflict)
	}
	total := domain.SumCalories(items)
	var entryID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO meal_entries(local_date,task_id,meal_name_snapshot,recipe_id,recipe_name_snapshot,recipe_multiplier,total_calories) VALUES ($1::date,$2,$3,$4,$5,$6,$7) RETURNING id`, input.Date, input.TaskID, taskName, input.RecipeID, input.RecipeName, input.RecipeMultiplier, total).Scan(&entryID)
	if err != nil {
		return domain.MealEntry{}, mapError(err)
	}
	for index, item := range items {
		_, err = tx.ExecContext(ctx, `INSERT INTO meal_entry_items(meal_entry_id,food_id,food_name_snapshot,quantity,unit,portion_id,portion_name_snapshot,estimated_grams,calories_per_100g_snapshot,calculated_calories,cooking_method,notes,position) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, entryID, item.FoodID, item.FoodNameSnapshot, item.Quantity, item.Unit, item.PortionID, item.PortionNameSnapshot, item.EstimatedGrams, item.CaloriesPer100GSnapshot, item.CalculatedCalories, item.CookingMethod, item.Notes, index)
		if err != nil {
			return domain.MealEntry{}, mapError(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE daily_meal_tasks SET status='registered',updated_at=CURRENT_TIMESTAMP WHERE id=$1`, input.TaskID); err != nil {
		return domain.MealEntry{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.MealEntry{}, err
	}
	return domain.MealEntry{ID: entryID, Date: input.Date, TaskID: input.TaskID, MealName: taskName, RecipeID: input.RecipeID, RecipeNameSnapshot: input.RecipeName, RecipeMultiplier: input.RecipeMultiplier, TotalCalories: total, Items: items, CreatedAt: time.Now()}, nil
}

func (r *Repository) UpdateMealEntry(ctx context.Context, id int64, items []domain.MealEntryItem) (domain.MealEntry, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.MealEntry{}, err
	}
	defer tx.Rollback()
	var entry domain.MealEntry
	var recipeID sql.NullInt64
	var multiplier sql.NullFloat64
	err = tx.QueryRowContext(ctx, `SELECT id,local_date::text,task_id,meal_name_snapshot,recipe_id,recipe_name_snapshot,recipe_multiplier,created_at FROM meal_entries WHERE id=$1 FOR UPDATE`, id).Scan(&entry.ID, &entry.Date, &entry.TaskID, &entry.MealName, &recipeID, &entry.RecipeNameSnapshot, &multiplier, &entry.CreatedAt)
	if err != nil {
		return domain.MealEntry{}, mapError(err)
	}
	if recipeID.Valid {
		entry.RecipeID = &recipeID.Int64
	}
	if multiplier.Valid {
		entry.RecipeMultiplier = &multiplier.Float64
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM meal_entry_items WHERE meal_entry_id=$1`, id); err != nil {
		return domain.MealEntry{}, err
	}
	for index, item := range items {
		_, err = tx.ExecContext(ctx, `INSERT INTO meal_entry_items(meal_entry_id,food_id,food_name_snapshot,quantity,unit,portion_id,portion_name_snapshot,estimated_grams,calories_per_100g_snapshot,calculated_calories,cooking_method,notes,position) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, id, item.FoodID, item.FoodNameSnapshot, item.Quantity, item.Unit, item.PortionID, item.PortionNameSnapshot, item.EstimatedGrams, item.CaloriesPer100GSnapshot, item.CalculatedCalories, item.CookingMethod, item.Notes, index)
		if err != nil {
			return domain.MealEntry{}, mapError(err)
		}
	}
	entry.TotalCalories = domain.SumCalories(items)
	if _, err = tx.ExecContext(ctx, `UPDATE meal_entries SET total_calories=$1,updated_at=CURRENT_TIMESTAMP WHERE id=$2`, entry.TotalCalories, id); err != nil {
		return domain.MealEntry{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.MealEntry{}, err
	}
	entry.Items = items
	return entry, nil
}

func (r *Repository) DeleteMealEntry(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var taskID int64
	if err = tx.QueryRowContext(ctx, `SELECT task_id FROM meal_entries WHERE id=$1 FOR UPDATE`, id).Scan(&taskID); err != nil {
		return mapError(err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM meal_entries WHERE id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE daily_meal_tasks SET status='pending',updated_at=CURRENT_TIMESTAMP WHERE id=$1`, taskID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, id int64, status domain.TaskStatus) error {
	result, err := r.db.ExecContext(ctx, `UPDATE daily_meal_tasks SET status=$1,updated_at=CURRENT_TIMESTAMP WHERE id=$2 AND status!='registered'`, status, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) CreateTask(ctx context.Context, date, name string) (domain.DailyMealTask, error) {
	var next int
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(position),-1)+1 FROM daily_meal_tasks WHERE local_date=$1::date`, date).Scan(&next); err != nil {
		return domain.DailyMealTask{}, err
	}
	key := fmt.Sprintf("extra-%d", time.Now().UnixNano())
	var id int64
	if err := r.db.QueryRowContext(ctx, `INSERT INTO daily_meal_tasks(local_date,name,slot_key,position) VALUES ($1::date,$2,$3,$4) RETURNING id`, date, name, key, next).Scan(&id); err != nil {
		return domain.DailyMealTask{}, mapError(err)
	}
	return domain.DailyMealTask{ID: id, Date: date, Name: name, SlotKey: key, Position: next, Status: domain.TaskPending, CreatedAt: time.Now()}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: el registro ya existe", domain.ErrConflict)
		case "23503", "23514", "22P02":
			return fmt.Errorf("%w: datos inválidos", domain.ErrValidation)
		}
	}
	return err
}
