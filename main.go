package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Запуск телеграм бота...")

	if err := godotenv.Load(); err != nil {
		log.Println(
			"Файл .env не найден, используются переменные окружения",
		)
	}

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

	log.Println(
		"База данных подключена",
	)

	bot, err := NewTelegramBot(
		cfg.BotToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"Бот запущен: @%s",
		bot.Self.UserName,
	)

	scheduler, err := NewScheduler(
		bot,
		cfg,
	)
	if err != nil {
		log.Fatal(err)
	}

	go scheduler.Run(ctx)

	log.Println(
		"Ежедневная отправка сообщений включена.",
	)

	RunUpdates(
		ctx,
		bot,
		db,
	)
}
