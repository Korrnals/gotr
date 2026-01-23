package utils

import "testing"

type testStruct struct {
	Title       string
	PriorityID  int64
	hiddenField string // не должно находиться
}

func TestGetFieldValue(t *testing.T) {
	obj := testStruct{
		Title:       "Тестовый кейс",
		PriorityID:  5,
		hiddenField: "секрет",
	}

	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"точное имя", "Title", "Тестовый кейс"},
		{"case-insensitive", "title", "Тестовый кейс"},
		{"число", "PriorityID", "5"},
		{"неверное поле", "unknown", ""},
		{"пустое имя", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFieldValue(obj, tt.field)
			if got != tt.expected {
				t.Errorf("GetFieldValue(%q) = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}
}
