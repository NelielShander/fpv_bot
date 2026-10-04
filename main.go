package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Запуск телеграм бота...")

	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	log.Println("База данных подключена")

	bot, err := NewTelegramBot(cfg.BotToken)
	if err != nil {
		log.Fatalf("Ошибка создания Telegram-бота: %v", err)
	}

	if err := SetBotCommands(bot); err != nil {
		log.Printf("ошибка установки команд: %v", err)
	}

	log.Printf("Бот запущен: @%s", bot.Self.UserName)

	scheduler, err := NewScheduler(bot, cfg)
	if err != nil {
		log.Fatalf("Ошибка создания планировщика: %v", err)
	}

	go scheduler.Run(ctx)

	log.Println("Ежедневная отправка сообщений включена")

	RunUpdates(ctx, bot, db)

	log.Println("Бот остановлен")
}
