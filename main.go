package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type MessageData struct {
	TelegramMessageID int
	Username          string
	OriginalText      string
	MessageDate       time.Time
}

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не задан")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	// Подключение к PostgreSQL
	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(ctx); err != nil {
		log.Fatal("Ошибка подключения к PostgreSQL:", err)
	}

	log.Println("PostgreSQL подключен")

	// Подключение к Telegram
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Ошибка подключения к Telegram:", err)
	}

	log.Printf("Бот запущен: @%s", bot.Self.UserName)

	//запуск отправки ежедневного сообщения
	go sendDailyMessage(bot)

	log.Println("Ежедневная отправка сообщений включена.")

	// Получаем обновления
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		message := update.Message

		// Нам нужен только текст
		if message.Text == "" {
			continue
		}

		// Обработать команды
		if update.Message.IsCommand() && update.Message.Command() == "nedelny" {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Нет еще отчетов, отвали",
			)

			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			continue
		}

		// текст больше 10 строк
		if hasLessLines(message.Text, 10) {
			continue
		}

		log.Printf(
			"Новое сообщение от %s: %s",
			message.From.UserName,
			message.Text,
		)

		// Ставим эмодзи
		if err := addReaction(bot, update.Message, "👀"); err != nil {
			log.Println(err)
		}

		//Обрабатываем сообщение
		data, ok := processMessage(message)
		if !ok {
			log.Println("Сообщение не удалось обработать")
			continue
		}

		// Сохраняем в БД
		if err := saveMessage(ctx, db, data); err != nil {
			log.Println("Ошибка сохранения:", err)
			continue
		}

		log.Println("Сообщение сохранено в БД")
	}

}

// Обработка сообщения
func processMessage(message *tgbotapi.Message) (MessageData, bool) {
	text := strings.TrimSpace(message.Text)

	if text == "" {
		return MessageData{}, false
	}

	username := ""

	if message.From != nil {
		username = message.From.UserName
	}

	// Здесь находится ваша бизнес-логика.
	//
	// Например:
	// "товар яблоки количество 10"
	//
	// можно превратить в:
	// "товар=яблоки; количество=10"

	return MessageData{
		TelegramMessageID: message.MessageID,
		Username:          username,
		OriginalText:      text,
		MessageDate:       time.Unix(int64(message.Date), 0),
	}, true
}

// Сохранение в PostgreSQL
func saveMessage(ctx context.Context, db *pgxpool.Pool, data MessageData) error {

	query := `
		INSERT INTO messages (
			telegram_message_id,
			username,
			original_text,
			message_date
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(
		ctx,
		query,
		data.TelegramMessageID,
		data.Username,
		data.OriginalText,
		data.MessageDate,
	)

	return err
}

// Отправка ежедневного сообщения
func sendDailyMessage(bot *tgbotapi.BotAPI) {

	userIDstr := os.Getenv("NOTIFY_USER_ID")
	if userIDstr == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не задан")
	}

	userID, err := strconv.ParseInt(userIDstr, 10, 64)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	groupIDstr := os.Getenv("GROUP_ID")
	if groupIDstr == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не задан")
	}

	groupID, err := strconv.ParseInt(groupIDstr, 10, 64)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	for {
		now := time.Now()

		location, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			log.Println(err)
		}

		// Следующее наступление 19:00
		next := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			19,
			0,
			0,
			0,
			location,
		)

		// Если сегодня 09:00 уже прошло — отправляем завтра
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		time.Sleep(time.Until(next))

		text := fmt.Sprintf(
			"<a href=\"tg://user?id=%d\">Шаман</a> что там по расчётам КТр по\nШтату- ?\nСписку- ?\nНа лбс- ?\nВ работе- ?\nБлижний- ?\nДальний - ?",
			userID,
		)

		msg := tgbotapi.NewMessage(groupID, text)
		msg.ParseMode = tgbotapi.ModeHTML

		if _, err := bot.Send(msg); err != nil {
			log.Printf("ошибка отправки ежедневного сообщения: %v", err)
		}
	}
}

// Проверка количества строк
func hasLessLines(text string, lines int) bool {
	count := 0

	for _, r := range text {
		if r == '\n' {
			count++
		}

		if count > lines {
			return false
		}
	}

	return true
}

// Добавление реакции к сообщению
func addReaction(bot *tgbotapi.BotAPI, message *tgbotapi.Message, emoji string) error {
	params := tgbotapi.Params{}
	params["chat_id"] = fmt.Sprintf("%v", message.Chat.ID)
	params["message_id"] = fmt.Sprintf("%v", message.MessageID)
	params["reaction"] = `[{"type":"emoji","emoji":"` + emoji + `"}]`

	_, err := bot.MakeRequest("setMessageReaction", params)
	return err
}

func init() {
	log.Println("Starting Telegram bot...")
}
