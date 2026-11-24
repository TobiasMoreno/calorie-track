package domain

import "testing"

func TestAggregateDayKeepsIncompleteDistinctFromZero(t *testing.T) {
	target := 1800.0
	tasks := []DailyMealTask{{Status: TaskPending}, {Status: TaskSkipped}}
	summary := AggregateDay("2026-11-23", &target, tasks, nil)
	if summary.HasData {
		t.Fatal("day without entries must not report data")
	}
	if summary.Calories != 0 {
		t.Fatalf("calories = %v", summary.Calories)
	}
	if summary.PendingTasks != 1 {
		t.Fatalf("pending = %d, want 1", summary.PendingTasks)
	}
}

func TestAggregateWeekUsesOnlyRecordedEntries(t *testing.T) {
	days := [][]MealEntry{
		{{TotalCalories: 1000}}, nil, {{TotalCalories: 900}, {TotalCalories: 450}}, nil, nil, {{TotalCalories: 1200}}, {{TotalCalories: 800}},
	}
	total := 0.0
	recordedDays := 0
	for _, entries := range days {
		summary := AggregateDay("2026-11-23", nil, nil, entries)
		if summary.HasData {
			total += summary.Calories
			recordedDays++
		}
	}
	if recordedDays != 4 {
		t.Fatalf("recorded days = %d, want 4", recordedDays)
	}
	if total/float64(recordedDays) != 1087.5 {
		t.Fatalf("average = %v, want 1087.5", total/float64(recordedDays))
	}
}
