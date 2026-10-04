package main

import (
	"fmt"
	"log"
	"strings"
)

func ProcessWeeklyReport(handler *Handler) string {
	reports, err := GetLastWeekReports(handler.ctx, handler.db)
	if err != nil {
		log.Println(err)
		return ""
	}

	header := "№; Дата; Вид БпЛА; Тип задачи; Статус; Примечание\n"
	var reportText strings.Builder

	for i, report := range reports {
		reportText.WriteString(fmt.Sprintf("%d", i+1))
		reportText.WriteString("; ")
		reportText.WriteString(report.MessageDate.Format("02.01.2006"))
		reportText.WriteString("; ")
		reportText.WriteString(parseReport(report))
		reportText.WriteString("\n")
	}

	return header + reportText.String()
}

func ProcessDailyReport(handler *Handler, period []string) string {
	reports, err := GetDailyReports(handler.ctx, handler.db, period)
	if err != nil {
		log.Println(err)
		return ""
	}

	header := "№; Дата; Позывной; Время; Вид БпЛА; Тип задачи; Координата Х; Координата У; Статус\n"
	var reportText strings.Builder

	for i, report := range reports {
		reportText.WriteString(fmt.Sprintf("%d", i+1))
		reportText.WriteString("; ")
		reportText.WriteString(report.MessageDate.Format("02.01.2006"))
		reportText.WriteString("\n")

	}

	return header + reportText.String()
}

func parseReport(report Report) string {
	text := report.MessageText
	reportType := report.Type()

	if reportType == "Боевая" {
		return parseWork(text, reportType)
	} else {
		return parseDelivery(text, reportType)
	}
}

func parseDelivery(reportText string, reportType string) string {
	var text strings.Builder

	text.WriteString(wordAfter(reportText, "FPV", 2))
	text.WriteString(fmt.Sprintf("; %s;", reportType))
	text.WriteString("Не выполнена; ")
	text.WriteString(lineAfter(reportText, "Не доставлено"))

	return text.String()
}

func parseWork(reportText string, reportType string) string {
	var text strings.Builder

	text.WriteString(wordAfter(reportText, "Изделие", 1)[1:])
	text.WriteString(fmt.Sprintf("; %s;", reportType))
	text.WriteString("Не выполнена; ")
	text.WriteString(wordsAfter(reportText, "Статус:"))

	return text.String()
}

func lineAfter(text string, key string) string {
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
	if skip < 1 {
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
