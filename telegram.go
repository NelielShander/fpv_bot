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

	commandWeekly  = "nedelny"
	commandMorning = "utro"
	commandEvening = "vecher"
)

type Handler struct {
	ctx context.Context
	bot *tgbotapi.BotAPI
	db  *Database
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

func RunUpdates(ctx context.Context, bot *tgbotapi.BotAPI, db *Database) {
	handler := &Handler{
		ctx: ctx,
		bot: bot,
		db:  db,
	}

	handler.runUpdates()
}

func SetBotCommands(bot *tgbotapi.BotAPI) error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     commandWeekly,
			Description: "отчет по расходу за неделю",
		},
		{
			Command:     commandMorning,
			Description: "отчет по применению за ночь",
		},
		{
			Command:     commandEvening,
			Description: "Применение за день",
		},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Request(config)

	return err
}

func (h *Handler) sendCSV(chatID int64, csvText, reportName string) error {
	bot := h.bot
	// UTF-8 BOM для корректного определения кодировки Excel.
	data := append(
		[]byte{0xEF, 0xBB, 0xBF},
		[]byte(csvText)...,
	)

	file := tgbotapi.FileBytes{
		Name:  reportName,
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

	local, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Println(err)
	}

	data := MessageData{
		TelegramMessageID: message.MessageID,
		Username:          username,
		OriginalText:      text,
		MessageDate:       message.Time().In(local),
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
	var reportName string

	switch message.Command() {
	case "start":
		processStart(h.bot, message.Chat.ID)
		return
	case "help":
		processHelp(h.bot, message.Chat.ID)
		return
	case commandWeekly:
		csvText, reportName = ProcessWeeklyReport(h)
	case commandMorning:
		csvText, reportName = ProcessDailyReport(h, []string{"-2 hours", "10 hours"})
	case commandEvening:
		csvText, reportName = ProcessDailyReport(h, []string{"10 hours", "22 hours"})
	default:
		return
	}

	if err := h.sendCSV(message.Chat.ID, csvText, reportName); err != nil {
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
	return len(strings.Split(text, "\n")) <= lines
}

func trimText(text string, limit int) string {
	text = strings.ReplaceAll(text, "\n", " ")

	runes := []rune(text)

	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit])
}

func processStart(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(
		chatID,
		"Бот запущен.\n"+
			"Для вывода команд наберите /help",
	)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("ошибка отправки сообщения: %v", err)
	}
}

func processHelp(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(
		chatID,
		"Доступные команды:\n"+
			"/start — Запуск бота\n"+
			"/help — помощь\n"+
			"/nedelny — Расход за неделю\n"+
			"/utro — Применение за ночь\n"+
			"/vecher — Применение за день",
	)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("ошибка отправки сообщения: %v", err)
	}
}
