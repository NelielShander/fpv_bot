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

	// Загружаем .env.
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	// Загружаем конфигурацию.
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Контекст завершается при Ctrl+C или SIGTERM.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Подключение к PostgreSQL.
	db, err := NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	log.Println("База данных подключена")

	// Создание Telegram-бота.
	bot, err := NewTelegramBot(cfg.BotToken)
	if err != nil {
		log.Fatalf("Ошибка создания Telegram-бота: %v", err)
	}

	log.Printf(
		"Бот запущен: @%s",
		bot.Self.UserName,
	)

	// Создание планировщика.
	scheduler, err := NewScheduler(bot, cfg)
	if err != nil {
		log.Fatalf("Ошибка создания планировщика: %v", err)
	}

	// Запускаем планировщик в отдельной горутине.
	go scheduler.Run(ctx)

	log.Println("Ежедневная отправка сообщений включена")

	// Основной цикл обработки Telegram-сообщений.
	RunUpdates(
		ctx,
		bot,
		db,
	)

	log.Println("Бот остановлен")
}
