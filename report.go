package main

import (
	"regexp"
	"strings"
	"time"
)

type Report struct {
	Username    string
	MessageText string
	MessageDate time.Time
}

// Type Тип задачи
func (report *Report) Type() string {
	if strings.Contains(report.MessageText, "Статус:") {
		return "Боевая"
	} else {
		return "Логистика (Сброс)"
	}
}

// Date Дата отчета
func (report *Report) Date() string {
	return report.MessageDate.Format("02.01.2006")
}

// FpvType Тип БпЛА
func (report *Report) FpvType() string {
	if report.Type() == "Боевая" {
		return wordAfter(report.MessageText, "Изделие", 1)[1:]
	}

	return wordAfter(report.MessageText, "FPV", 2)
}

// Status Статус
func (report *Report) Status() string {
	if strings.Contains(report.MessageText, "Доставлено ✅") ||
		strings.Contains(report.MessageText, "Cтатус: поражено") ||
		strings.Contains(report.MessageText, "Статус:  поражено") ||
		strings.Contains(report.MessageText, "Cтатус: Попадание") {
		return "Выполнена"
	}

	return "Не выполнена"
}

// Note примечание
func (report *Report) Note() string {
	if report.Type() == "Боевая" {
		if strings.Index(report.MessageText, "не поражено") == -1 {
			return wordsAfter(report.MessageText, "Статус:")
		} else {
			return nextLine(report.MessageText, "не поражено")
		}
	} else {
		lastLine(report.MessageText, "Не доставлено")
	}
	return ""
}

func (report *Report) Team() string {
	if report.Type() == "Боевая" {
		return wordAfter(report.MessageText, "Расчёт:", 0)
	} else {
		return "Беркут"
	}
}

func (report *Report) WorkTime() string {
	text := report.MessageText
	re := regexp.MustCompile(`(?m)^\d{7}\r?\n\d{7}\r?\n\r?\n([^\r\n]+)`)

	match := re.FindStringSubmatch(text)
	if len(match) > 1 {

		if len(match[1]) >= 5 {
			return match[1][:5]
		}

		return match[1]
	} else {
		return wordsAfter(report.MessageText, "Время вылета :")
	}
}

func (report *Report) Coordinates() []string {
	re := regexp.MustCompile(`(?m)^(\d{7})\r?\n(\d{7})\r?\n\r?\n`)
	match := re.FindStringSubmatch(report.MessageText)

	return match
}

func lastLine(text string, key string) string {
	pos := strings.Index(text, key)
	if pos == -1 {
		return ""
	}

	text = text[pos+len(key):]

	lines := strings.Split(text, "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		if line != "" {
			return line
		}
	}

	return ""
}

func wordsAfter(text string, key string) string {
	pos := strings.Index(text, key)
	if pos == -1 {
		return ""
	}

	text = text[pos+len(key):]

	if end := strings.IndexAny(text, "\r\n"); end != -1 {
		text = text[:end]
	}

	return strings.TrimSpace(text)
}

func wordAfter(text string, pointer string, skip int) string {
	if skip < 0 {
		return ""
	}

	words := strings.Fields(text)

	for i, word := range words {
		if word == pointer && i+skip < len(words) {
			return words[i+skip]
		}
	}

	return ""
}

func nextLine(text, key string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if strings.Contains(line, key) && i+1 < len(lines) {
			return lines[i+1]
		}
	}

	return ""
}
