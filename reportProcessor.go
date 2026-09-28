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

var counter int = 0

func ProcessReport() string {
	reportText := ""

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	db, err := NewDatabase(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	reports, err := GetLastWeekReports(ctx, db)
	if err != nil {
		log.Println(err)
	}

	for _, report := range reports {
		reportText = reportText + "\n" + parseReport(report)
	}

	return reportText
}

func parseReport(report Report) string {
	messageText := report.MessageText
	text := ""

	if strings.Contains(messageText, "Статус") {
		text = parseWork(report.MessageDate, messageText)
	} else {
		text = parseDelivery(report.MessageDate, messageText)
	}

	counter = counter + 1

	text = fmt.Sprintf("%d", counter) + "; " + text

	return text
}

func parseDelivery(date time.Time, messageText string) string {
	text := date.Format("02.01.2006")
	text = text + "; " + wordAfter(messageText, "FPV", 2)
	text = text + "; Логистика"

	if strings.Contains(text, "Не доставлено") {
		text = text + "; Не выполнено; " + wordsAfter(messageText, "Не доставлено")
	} else {
		text = text + "; Выполнено;"
	}

	return text
}

func parseWork(date time.Time, messageText string) string {
	text := date.Format("02.01.2006")
	text = text + "; "
	text = text + wordAfter(messageText, "FPV", 2)
	text = text + wordAfter(messageText, "Изделие", 1)
	text = text + "; Боевая"

	if strings.Contains(messageText, "Статус: Попадание") {
		text = text + "; Выполнено"
	} else {
		text = text + "; Не выполнено"
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
