package domain

import (
	"fmt"
	"strings"
)

type NewTask struct {
	Name     string
	SlotKey  string
	Position int
}

func BuildDefaultTasks(meals []string) ([]NewTask, error) {
	if len(meals) == 0 {
		return nil, fmt.Errorf("%w: at least one default meal is required", ErrValidation)
	}
	seen := map[string]bool{}
	result := make([]NewTask, 0, len(meals))
	for index, raw := range meals {
		name := strings.TrimSpace(raw)
		if name == "" {
			return nil, fmt.Errorf("%w: meal name cannot be empty", ErrValidation)
		}
		key := slug(name)
		if seen[key] {
			return nil, fmt.Errorf("%w: duplicate default meal", ErrValidation)
		}
		seen[key] = true
		result = append(result, NewTask{Name: name, SlotKey: key, Position: index})
	}
	return result, nil
}

func slug(value string) string {
	replacer := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", " ", "-")
	return strings.ToLower(replacer.Replace(strings.TrimSpace(value)))
}
