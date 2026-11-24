package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"calorietrack/backend/internal/domain"
)

func TestEnsureDayIsIdempotentAndSnapshotsTarget(t *testing.T) {
	databaseURL := os.Getenv("CALORIETRACK_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CALORIETRACK_TEST_DATABASE_URL is not configured")
	}
	db, err := Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	tasks, _ := domain.BuildDefaultTasks(domain.DefaultMeals)
	date := time.Now().UTC().AddDate(20, 0, int(time.Now().UnixNano()%20)).Format("2006-01-02")
	defer db.Exec(`DELETE FROM daily_days WHERE local_date=$1::date`, date)
	first := 1800.0
	second := 2200.0
	if err := repo.EnsureDay(context.Background(), date, &first, tasks); err != nil {
		t.Fatal(err)
	}
	if err := repo.EnsureDay(context.Background(), date, &second, tasks); err != nil {
		t.Fatal(err)
	}
	gotTasks, err := repo.GetTasks(context.Background(), date)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotTasks) != 4 {
		t.Fatalf("tasks = %d, want 4 (%s)", len(gotTasks), fmt.Sprint(gotTasks))
	}
	target, err := repo.GetDayTarget(context.Background(), date)
	if err != nil {
		t.Fatal(err)
	}
	if target == nil || *target != first {
		t.Fatalf("target = %v, want %v", target, first)
	}
}
