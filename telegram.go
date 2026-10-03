package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	reactionEmoji  = "👀"
	updatesTimeout = 60

	commandWeekly = "nedelny"
	commandDaily  = "za_den"
)

type Handler struct {
	ctx context.Context
	bot *tgbotapi.BotAPI
	db  *Database
	loc *time.Location
}

type MessageData struct {
	TelegramMessageID int
	Username          string
	OriginalText      string
	MessageDate       time.Time
}

func NewTelegramBot(token string) (*tgbotapi.BotAPI, error) {
	return tgbotapi.NewBotAPI(token)
}

func RunUpdates(ctx context.Context, bot *tgbotapi.BotAPI, db *Database, loc *time.Location) {
	handler := &Handler{
		ctx: ctx,
		bot: bot,
		db:  db,
		loc: loc,
	}

	handler.runUpdates()
}

func (h *Handler) SendCSV(bot *tgbotapi.BotAPI, chatID int64, csvText string) error {
	// UTF-8 BOM для корректного определения кодировки Excel.
	data := append(
		[]byte{0xEF, 0xBB, 0xBF},
		[]byte(csvText)...,
	)

	file := tgbotapi.FileBytes{
		Name:  reportName(h.loc),
		Bytes: data,
	}

	msg := tgbotapi.NewDocument(chatID, file)

	_, err := bot.Send(msg)

	return err
}

func (h *Handler) runUpdates() {
	config := tgbotapi.NewUpdate(0)
	config.Timeout = updatesTimeout

	updates := h.bot.GetUpdatesChan(config)

	for {
		select {
		case <-h.ctx.Done():
			h.bot.StopReceivingUpdates()
			log.Println("Получение обновлений остановлено")
			return

		case update, ok := <-updates:
			if !ok {
				log.Println("Канал обновлений Telegram закрыт")
				return
			}

			h.handleUpdate(update)
		}
	}
}

func (h *Handler) handleUpdate(update tgbotapi.Update) {
	message := update.Message

	if message == nil {
		return
	}

	if message.IsCommand() {
		h.handleCommand(message)
		return
	}

	originalText := messageText(message)

	if originalText == "" || hasLessLines(originalText, 10) {
		return
	}

	text := normalizeMessageText(originalText)

	username := messageUsername(message)

	log.Printf(
		"Новое сообщение от '%s: %s...'",
		username,
		trimText(text, 50),
	)

	data := MessageData{
		TelegramMessageID: message.MessageID,
		Username:          username,
		OriginalText:      text,
		MessageDate:       message.Time(),
	}

	if err := h.db.SaveMessage(h.ctx, data); err != nil {
		log.Printf("ошибка сохранения сообщения: %v", err)
		return
	}

	log.Println("Сообщение сохранено в БД")

	if err := addReaction(h.bot, message); err != nil {
		log.Printf("ошибка добавления реакции: %v", err)
	}
}

func (h *Handler) handleCommand(message *tgbotapi.Message) {
	var csvText string

	switch message.Command() {
	case commandWeekly:
		csvText = ProcessWeeklyReport(h)
	case commandDaily:
		csvText = ProcessDailyReport(h)
	default:
		return
	}

	if err := h.SendCSV(h.bot, message.Chat.ID, csvText); err != nil {
		log.Printf("ошибка отправки ответа: %v", err)
	}
}

func addReaction(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if message == nil || message.Chat == nil {
		return fmt.Errorf("сообщение или чат отсутствует")
	}

	params := tgbotapi.Params{
		"chat_id":    strconv.FormatInt(message.Chat.ID, 10),
		"message_id": strconv.Itoa(message.MessageID),
		"reaction":   `[{"type":"emoji","emoji":"` + reactionEmoji + `"}]`,
	}

	_, err := bot.MakeRequest("setMessageReaction", params)

	return err
}

func messageText(message *tgbotapi.Message) string {
	if message == nil {
		return ""
	}

	text := message.Text

	if text == "" {
		text = message.Caption
	}

	return strings.TrimSpace(text)
}

func messageUsername(message *tgbotapi.Message) string {
	if message == nil || message.From == nil {
		return ""
	}

	return message.From.UserName
}

func normalizeMessageText(text string) string {
	text = strings.ReplaceAll(text, "ПВХ - ", "ПВХ-")
	text = strings.ReplaceAll(text, "ПВХ1", "ПВХ-1")

	return text
}

func hasLessLines(text string, lines int) bool {
	return strings.Count(text, "\n") <= lines
}

func trimText(text string, limit int) string {
	text = strings.ReplaceAll(text, "\n", " ")

	runes := []rune(text)

	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit])
}

func reportName(loc *time.Location) string {
	now := time.Now().In(loc).Add(-7 * 24 * time.Hour)

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	startOfWeek := now.AddDate(0, 0, 1-weekday)

	return fmt.Sprintf(
		"Отчет по применению от %s.csv",
		startOfWeek.Format("02.01.2006"),
	)
}
