package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func ProcessWeeklyReport() string {
	reports, err := GetLastWeekReports(loadCtxDb())
	if err != nil {
		log.Println(err)
		return ""
	}

	header := "№; Дата; Наименование; Тип задачи; Статус; Примечание\n"
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

func ProcessDailyReport() string {

	reports, err := GetDailyReports(loadCtxDb())
	if err != nil {
		log.Println(err)
		return ""
	}

	header := "№; Дата\n"
	var reportText strings.Builder

	for i, report := range reports {
		reportText.WriteString(fmt.Sprintf("%d", i+1))
		reportText.WriteString("; ")
		reportText.WriteString(report.MessageDate.Format("02.01.2006"))
		reportText.WriteString("\n")
	}

	return header + reportText.String()
}

func loadCtxDb() (context.Context, *Database) {
	cfg, err := LoadConfig()
	if err != nil {
		log.Println(err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	db, err := NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	return ctx, db
}

func parseReport(report Report) string {
	messageText := report.MessageText

	if strings.Contains(messageText, "Статус:") {
		return parseWork(messageText)
	}

	return parseDelivery(messageText)
}

func parseDelivery(messageText string) string {
	var text strings.Builder

	text.WriteString(wordAfter(messageText, "FPV", 2))
	text.WriteString("; Логистика; Не выполнена; ")
	text.WriteString(lineAfter(messageText, "Не доставлено"))

	return text.String()
}

func parseWork(messageText string) string {
	var text strings.Builder

	text.WriteString(wordAfter(messageText, "Изделие", 1)[1:])
	text.WriteString("; Боевая; Не выполнена; ")
	text.WriteString(wordsAfter(messageText, "Статус:"))

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
