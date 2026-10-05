package main

import (
	"testing"
	"time"
)

func TestWordsAfter(t *testing.T) {
	tests := []struct {
		name string
		text string
		key  string
		want string
	}{
		{
			name: "обычный текст",
			text: "Статус: Выполнено",
			key:  "Статус:",
			want: "Выполнено",
		},
		{
			name: "текст с переносом строки",
			text: "Статус: Выполнено\nПримечание",
			key:  "Статус:",
			want: "Выполнено",
		},
		{
			name: "перенос CRLF",
			text: "Статус: Выполнено\r\nПримечание",
			key:  "Статус:",
			want: "Выполнено",
		},
		{
			name: "ключ отсутствует",
			text: "Выполнено",
			key:  "Статус:",
			want: "",
		},
		{
			name: "лишние пробелы",
			text: "Статус:     Выполнено   ",
			key:  "Статус:",
			want: "Выполнено",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordsAfter(tt.text, tt.key)

			if got != tt.want {
				t.Errorf(
					"wordsAfter() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestReportDateFormatting(t *testing.T) {
	report := Report{
		MessageDate: time.Date(
			2026,
			time.September,
			27,
			13,
			44,
			0,
			0,
			time.UTC,
		),
	}

	got := report.MessageDate.Format("02.01.2006")
	want := "27.09.2026"

	if got != want {
		t.Errorf(
			"date format = %q, want %q",
			got,
			want,
		)
	}
}
