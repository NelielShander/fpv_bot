package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	reactionEmoji  = "👀"
	updatesTimeout = 60

	commandWeekly = "nedelny"
)

func NewTelegramBot(
	token string,
) (*tgbotapi.BotAPI, error) {
	return tgbotapi.NewBotAPI(token)
}

func RunUpdates(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	db *Database,
) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = updatesTimeout

	updates := bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			log.Println("Получение обновлений остановлено")
			return

		case update, ok := <-updates:
			if !ok {
				log.Println("Канал обновлений Telegram закрыт")
				return
			}

			handleUpdate(ctx, bot, db, update)
		}
	}
}

func handleUpdate(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	db *Database,
	update tgbotapi.Update,
) {
	message := update.Message

	if message == nil || message.Text == "" {
		return
	}

	if message.IsCommand() {
		handleCommand(bot, message)
		return
	}

	if hasLessLines(message.Text, 10) {
		return
	}

	username := messageUsername(message)

	log.Printf(
		"Новое сообщение от %s: %s",
		username,
		message.Text,
	)

	if err := addReaction(bot, message, reactionEmoji); err != nil {
		log.Printf(
			"ошибка добавления реакции: %v",
			err,
		)
	}

	data, ok := processMessage(message)
	if !ok {
		log.Println("сообщение не удалось обработать")
		return
	}

	if err := db.SaveMessage(ctx, data); err != nil {
		log.Printf(
			"ошибка сохранения сообщения: %v",
			err,
		)
		return
	}

	log.Println("Сообщение сохранено в БД")
}

func handleCommand(
	bot *tgbotapi.BotAPI,
	message *tgbotapi.Message,
) {
	switch message.Command() {
	case commandWeekly:
		sendWeeklyReportMessage(
			bot,
			message.Chat.ID,
		)
	}
}

func sendWeeklyReportMessage(
	bot *tgbotapi.BotAPI,
	chatID int64,
) {
	msg := tgbotapi.NewMessage(
		chatID,
		"Нет еще отчетов, отвали",
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf(
			"ошибка отправки ответа: %v",
			err,
		)
	}
}

func processMessage(
	message *tgbotapi.Message,
) (MessageData, bool) {
	text := strings.TrimSpace(message.Text)

	if text == "" {
		return MessageData{}, false
	}

	return MessageData{
		TelegramMessageID: message.MessageID,
		Username:          messageUsername(message),
		OriginalText:      text,
		MessageDate:       message.Time(),
	}, true
}

func messageUsername(
	message *tgbotapi.Message,
) string {
	if message == nil || message.From == nil {
		return ""
	}

	return message.From.UserName
}

func hasLessLines(
	text string,
	lines int,
) bool {
	return strings.Count(text, "\n") <= lines
}

func addReaction(
	bot *tgbotapi.BotAPI,
	message *tgbotapi.Message,
	emoji string,
) error {
	if message == nil || message.Chat == nil {
		return fmt.Errorf(
			"сообщение или чат отсутствует",
		)
	}

	params := tgbotapi.Params{
		"chat_id":    strconv.FormatInt(message.Chat.ID, 10),
		"message_id": strconv.Itoa(message.MessageID),
		"reaction":   `[{"type":"emoji","emoji":"` + emoji + `"}]`,
	}

	_, err := bot.MakeRequest(
		"setMessageReaction",
		params,
	)

	return err
}
