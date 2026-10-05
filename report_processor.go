package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func ProcessWeeklyReport(handler *Handler) (string, string) {
	startOfWeek := time.Now().Truncate(24 * time.Hour)
	name := fmt.Sprintf("Недельный отчет %s.csv", startOfWeek.Format("02.01.2006"))

	reports, err := GetLastWeekReports(handler.ctx, handler.db)
	if err != nil {
		log.Println(err)
		return "", name
	}

	header := "№; Дата; Вид БпЛА; Тип задачи; Статус; Примечание\n"
	var reportText strings.Builder

	for i, report := range reports {
		reportText.WriteString(fmt.Sprintf("%d", i+1))
		reportText.WriteString("; ")
		reportText.WriteString(report.Date())
		reportText.WriteString("; ")
		reportText.WriteString(report.FpvType())
		reportText.WriteString("; ")
		reportText.WriteString(report.Type())
		reportText.WriteString("; ")
		reportText.WriteString(report.Status())
		reportText.WriteString("; ")
		reportText.WriteString(report.Note())
		reportText.WriteString("\n")
	}

	return header + reportText.String(), name
}

func ProcessDailyReport(handler *Handler, period []string) (string, string) {
	var timeOfDay string

	if period[0] == "-2 hours" {
		timeOfDay = "утро"
	} else {
		timeOfDay = "вечер"
	}

	name := fmt.Sprintf("Отчет за %s %s.csv", timeOfDay, time.Now().Format("02.01.2006"))

	reports, err := GetDailyReports(handler.ctx, handler.db, period)
	if err != nil {
		log.Println(err)
		return "", name
	}

	header := "№; Дата; Расчет; Время; Вид БпЛА; Тип задачи; Координата Х; Координата У; Статус\n"
	var reportText strings.Builder

	for i, report := range reports {
		reportText.WriteString(fmt.Sprintf("%d", i+1))
		reportText.WriteString("; ")
		reportText.WriteString(report.MessageDate.Format("02.01.2006"))
		reportText.WriteString("\n")

	}

	return header + reportText.String(), name
}
