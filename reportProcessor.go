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
	header += reportText.String()

	return header
}

func parseReport(report Report) string {
	messageText := report.MessageText
	date := report.MessageDate

	text := date.Format("02.01.2006")
	text += "; "

	if strings.Contains(messageText, "Статус") {
		text += parseWork(text, messageText)
	} else {
		text += parseDelivery(text, messageText)
	}

	counter++

	return fmt.Sprintf("%d; %s", counter, text)
}

func parseDelivery(text string, messageText string) string {
	text += wordAfter(messageText, "FPV", 2)

	text += "; Логистика"

	if strings.Contains(messageText, "Не доставлено") {
		text += "; Не выполнено; "
		text += linesAfter(messageText, "Не доставлено ❌")
	} else {
		text += "; Выполнено;"
	}

	return text
}

func parseWork(text string, messageText string) string {
	text += wordAfter(messageText, "FPV", 2)
	text += wordAfter(messageText, "Изделие", 1)[1:]

	text += "; Боевая"

	if strings.Contains(messageText, "Статус: Попадание") {
		text += "; Выполнено;"
	} else {
		text += "; Не выполнено;"
		text += wordsAfter(messageText, "Статус:")
	}

	return text
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

func linesAfter(text string, key string) string {
	pos := strings.Index(text, key)
	if pos == -1 {
		return ""
	}

	text = text[pos+len(key):]

	lines := strings.Split(text, "\n")

	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, " ")
}

func wordAfter(text string, pointer string, skip int) string {
	if !strings.Contains(text, pointer) {
		return ""
	}

	words := strings.Fields(text)

	for i, word := range words {
		if word == pointer && i+1 < len(words) {
			return words[i+skip]
		}
	}

	return ""
}
