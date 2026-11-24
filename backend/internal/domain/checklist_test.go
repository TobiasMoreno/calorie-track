package domain

import "testing"

func TestBuildDefaultTasks(t *testing.T) {
	tasks, err := BuildDefaultTasks([]string{"Desayuno", "Almuerzo", "Merienda", "Cena"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 4 {
		t.Fatalf("len = %d, want 4", len(tasks))
	}
	if tasks[2].SlotKey != "merienda" || tasks[2].Position != 2 {
		t.Fatalf("unexpected task: %#v", tasks[2])
	}
}

func TestBuildDefaultTasksRejectsDuplicate(t *testing.T) {
	if _, err := BuildDefaultTasks([]string{"Cena", " cena "}); err == nil {
		t.Fatal("expected duplicate error")
	}
}
