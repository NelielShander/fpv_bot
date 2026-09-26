package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	envBotToken    = "TELEGRAM_BOT_TOKEN"
	envDatabaseURL = "DATABASE_URL"
	envNotifyUser  = "NOTIFY_USER_ID"
	envGroupID     = "GROUP_ID"
)

type Config struct {
	BotToken     string
	DatabaseURL  string
	NotifyUserID int64
	GroupID      int64
}

func LoadConfig() (Config, error) {
	botToken, err := requiredEnv(envBotToken)
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := requiredEnv(envDatabaseURL)
	if err != nil {
		return Config{}, err
	}

	notifyUserID, err := envInt64(envNotifyUser)
	if err != nil {
		return Config{}, err
	}

	groupID, err := envInt64(envGroupID)
	if err != nil {
		return Config{}, err
	}

	return Config{
		BotToken:     botToken,
		DatabaseURL:  databaseURL,
		NotifyUserID: notifyUserID,
		GroupID:      groupID,
	}, nil
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)

	if value == "" {
		return "", fmt.Errorf("%s не задан", name)
	}

	return value, nil
}

func envInt64(name string) (int64, error) {
	value, err := requiredEnv(name)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"%s содержит некорректное число %q: %w",
			name,
			value,
			err,
		)
	}

	return result, nil
}
