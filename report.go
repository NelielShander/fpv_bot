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
	ID          int64
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
		re := regexp.MustCompile(`(?m)^Изделие\s*:\s*(\S+)`)
		match := re.FindStringSubmatch(report.MessageText)

		if len(match) < 2 {
			return ""
		}

		return match[1]
	}

	return wordAfter(report.MessageText, "FPV", 2)
}

// Status Статус
func (report *Report) Status() string {
	if strings.Contains(report.MessageText, "Доставлено ✅") ||
		strings.Contains(report.MessageText, "Cтатус: поражено") ||
		strings.Contains(report.MessageText, "Статус:  поражено") ||
		strings.Contains(report.MessageText, "Статус: Поражено") ||
		strings.Contains(report.MessageText, "Cтатус: Попадание") {
		return "Выполнена"
	}

	return "Не выполнена"
}

// Note примечание
func (report *Report) Note() string {
	if report.Status() == "Не выполнена" {
		if report.Type() == "Боевая" {
			if strings.Index(report.MessageText, "не поражено") == -1 {
				return wordsAfter(report.MessageText, "Статус:")
			} else {
				return nextLine(report.MessageText, "не поражено")
			}
		} else {
			return linesAfter(report.MessageText, "Не доставлено ❌")
		}
	}

	return ""
}

// Team расчет
func (report *Report) Team() string {
	if report.Type() == "Боевая" {
		return wordAfter(report.MessageText, "Расчёт:", 1)
	} else {
		return "Беркут"
	}
}

// WorkTime время начала вылета
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

func (report *Report) CoordinateX() string {
	if report.Type() == "Боевая" {
		re := regexp.MustCompile(`(?m)^Х=(.*)$`)
		match := re.FindStringSubmatch(report.MessageText)

		if len(match) < 2 {
			return ""
		}

		return match[1]
	}

	return coordinates(report)[1]
}

func (report *Report) CoordinateY() string {
	if report.Type() == "Боевая" {
		re := regexp.MustCompile(`(?m)^У=(.*)$`)
		match := re.FindStringSubmatch(report.MessageText)

		if len(match) < 2 {
			return ""
		}

		return match[1]
	}

	return coordinates(report)[2]
}

// Coordinates координаты
func coordinates(report *Report) []string {
	re := regexp.MustCompile(`(?m)^(\d{7})\r?\n(\d{7})\r?\n\r?\n`)
	match := re.FindStringSubmatch(report.MessageText)

	if len(match) < 3 {
		return make([]string, 3)
	}

	return match
}

func linesAfter(text, key string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if strings.Contains(line, key) {
			var result []string

			for _, s := range lines[i+1:] {
				s = strings.TrimSpace(s)
				if s != "" && !strings.Contains(s, "Борт учтённый") {
					result = append(result, s)
				}
			}

			return strings.Join(result, " ")
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
