package domain

func AggregateDay(date string, target *float64, tasks []DailyMealTask, entries []MealEntry) DaySummary {
	calories := 0.0
	for _, entry := range entries {
		calories += entry.TotalCalories
	}
	pending := 0
	registered := 0
	skipped := 0
	for _, task := range tasks {
		switch task.Status {
		case TaskPending:
			pending++
		case TaskRegistered:
			registered++
		case TaskSkipped:
			skipped++
		}
	}
	return DaySummary{Date: date, Calories: round1(calories), Target: target, TaskCount: len(tasks), RegisteredTasks: registered, SkippedTasks: skipped, PendingTasks: pending, HasData: len(entries) > 0}
}

func SumCaloriesFromEntries(entries []MealEntry) float64 {
	total := 0.0
	for _, entry := range entries {
		total += entry.TotalCalories
	}
	return round1(total)
}

func RoundCalories(value float64) float64 { return round1(value) }
