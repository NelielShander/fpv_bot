package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var counter int

func ProcessReport() string {
	cfg, err := LoadConfig()
	if err != nil {
		log.Println(err)
		return ""
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
		return ""
	}
	defer db.Close()

	reports, err := GetLastWeekReports(ctx, db)
	if err != nil {
		log.Println(err)
		return ""
	}

	counter = 0

	var reportText strings.Builder

	for _, report := range reports {
		reportText.WriteString(parseReport(report))
		reportText.WriteString("\n")
	}

	header := "№; Дата; Наименование; Тип задачи; Статус; Примечание\n"
	header += strings.TrimSuffix(reportText.String(), "\n")

	return header
}

func parseReport(report Report) string {
	messageText := report.MessageText

	var text string

	if strings.Contains(messageText, "Статус") {
		text = parseWork(report.MessageDate, messageText)
	} else {
		text = parseDelivery(report.MessageDate, messageText)
	}

	counter++

	return fmt.Sprintf("%d; %s", counter, text)
}

func parseDelivery(date time.Time, messageText string) string {
	text := date.Format("02.01.2006")

	text += "; "
	text += wordAfter(messageText, "FPV", 2)

	text += "; Логистика"

	if strings.Contains(messageText, "Не доставлено") {
		text += "; Не выполнено; "
		text += wordsAfter(messageText, "Не доставлено")
	} else {
		text += "; Выполнено"
	}

	return text
}

func parseWork(date time.Time, messageText string) string {
	text := date.Format("02.01.2006")

	text += "; "
	text += wordAfter(messageText, "FPV", 2)
	text += wordAfter(messageText, "Изделие", 1)[1:]

	text += "; Боевая"

	if strings.Contains(messageText, "Статус: Попадание") {
		text += "; Выполнено"
	} else {
		text += "; Не выполнено"
	}

	return text
}

func wordsAfter(text, target string) string {
	words := strings.Fields(text)

	for i, word := range words {
		if word == target && i+1 < len(words) {
			return strings.Join(words[i+1:], " ")
		}
	}

	return ""
}

func wordAfter(text string, pointer string, skip int) string {
	target := ""

	if !strings.Contains(text, pointer) {
		return target
	}

	words := strings.Fields(text)

	for i, word := range words {
		if word == pointer && i+1 < len(words) {
			target = words[i+skip]
		}
	}

	return target
}
