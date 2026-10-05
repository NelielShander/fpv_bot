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

	return header + reportText.String()
}

func ProcessDailyReport(handler *Handler, period []string) string {
	reports, err := GetDailyReports(handler.ctx, handler.db, period)
	if err != nil {
		log.Println(err)
		return ""
	}

	header := "№; Дата; Расчет; Время; Вид БпЛА; Тип задачи; Координата Х; Координата У; Статус\n"
	var reportText strings.Builder

	for i, report := range reports {
		reportText.WriteString(fmt.Sprintf("%d", i+1))
		reportText.WriteString("; ")
		reportText.WriteString(report.MessageDate.Format("02.01.2006"))
		reportText.WriteString("\n")

	}

	return header + reportText.String()
}
