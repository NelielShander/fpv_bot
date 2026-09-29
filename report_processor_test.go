package main

import (
	"testing"
	"time"
)

// ---------------------------------------------------------
// wordsAfter
// ---------------------------------------------------------

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

// ---------------------------------------------------------
// lineAfter
// ---------------------------------------------------------

func TestLineAfter(t *testing.T) {
	tests := []struct {
		name string
		text string
		key  string
		want string
	}{
		{
			name: "последняя непустая строка",
			text: "Текст\nНе доставлено ❌\nПримечание",
			key:  "Не доставлено",
			want: "Примечание",
		},
		{
			name: "после ключа одна строка",
			text: "Текст\nНе доставлено ❌\n",
			key:  "Не доставлено",
			want: "❌",
		},
		{
			name: "несколько пустых строк",
			text: "Текст\nНе доставлено\n\n\nПримечание\n\n",
			key:  "Не доставлено",
			want: "Примечание",
		},
		{
			name: "ключ отсутствует",
			text: "Текст\nПримечание",
			key:  "Не доставлено",
			want: "",
		},
		{
			name: "ключ в конце",
			text: "Текст\nНе доставлено",
			key:  "Не доставлено",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lineAfter(tt.text, tt.key)

			if got != tt.want {
				t.Errorf(
					"lineAfter() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

// ---------------------------------------------------------
// wordAfter
// ---------------------------------------------------------

func TestWordAfter(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pointer string
		skip    int
		want    string
	}{
		{
			name:    "следующее слово",
			text:    "FPV KT ПВХ-1",
			pointer: "FPV",
			skip:    1,
			want:    "KT",
		},
		{
			name:    "через два слова",
			text:    "FPV KT ПВХ-1",
			pointer: "FPV",
			skip:    2,
			want:    "ПВХ-1",
		},
		{
			name:    "ключ в середине",
			text:    "Тип FPV KT ПВХ-1",
			pointer: "FPV",
			skip:    2,
			want:    "ПВХ-1",
		},
		{
			name:    "ключ отсутствует",
			text:    "KT ПВХ-1",
			pointer: "FPV",
			skip:    1,
			want:    "",
		},
		{
			name:    "некуда пропускать",
			text:    "FPV KT",
			pointer: "FPV",
			skip:    2,
			want:    "",
		},
		{
			name:    "skip равен нулю",
			text:    "FPV KT ПВХ-1",
			pointer: "FPV",
			skip:    0,
			want:    "",
		},
		{
			name:    "skip отрицательный",
			text:    "FPV KT ПВХ-1",
			pointer: "FPV",
			skip:    -1,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordAfter(
				tt.text,
				tt.pointer,
				tt.skip,
			)

			if got != tt.want {
				t.Errorf(
					"wordAfter() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

// ---------------------------------------------------------
// parseDelivery
// ---------------------------------------------------------

func TestParseDelivery(t *testing.T) {
	text := `27.09.2026г

5.✈️
Жиган
FPV KT ПВХ-1
(4.9/2.4)

ТП АКУЛА
10,15 км

5364927
7370681

13:44-13:54

2л Бензин
Клей Момент

Не доставлено ❌

Б`

	want := "ПВХ-1; Логистика; Не выполнена; Б"

	got := parseDelivery(text)

	if got != want {
		t.Errorf(
			"parseDelivery() = %q, want %q",
			got,
			want,
		)
	}
}

// ---------------------------------------------------------
// parseReport
// ---------------------------------------------------------

func TestParseReport(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "доставка",
			text: `FPV KT ПВХ-1
Не доставлено ❌
Б`,
			want: "ПВХ-1; Логистика; Не выполнена; Б",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseReport(Report{
				MessageText: tt.text,
			})

			if got != tt.want {
				t.Errorf(
					"parseReport() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

// ---------------------------------------------------------
// parseReport — выбор типа
// ---------------------------------------------------------

func TestParseReportSelectsDeliveryWithoutStatus(t *testing.T) {
	report := Report{
		MessageText: `FPV KT ПВХ-1
Не доставлено ❌
Б`,
	}

	got := parseReport(report)

	want := "ПВХ-1; Логистика; Не выполнена; Б"

	if got != want {
		t.Errorf(
			"parseReport() = %q, want %q",
			got,
			want,
		)
	}
}

// ---------------------------------------------------------
// проверка Report
// ---------------------------------------------------------

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
