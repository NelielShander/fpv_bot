package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

const (
	envBotToken   = "BOT_TOKEN"
	envNotifyUser = "NOTIFY_USER_ID"
	envGroupID    = "GROUP_ID"
	envDBHost     = "DB_HOST"
	envDBPort     = "DB_PORT"
	envDBName     = "DB_NAME"
	envDBUser     = "DB_USER"
	envDBSSLMode  = "DB_SSLMODE"
	envTz         = "TZ"
)

type Config struct {
	BotToken     string
	DatabaseURL  string
	NotifyUserID int64
	GroupID      int64
	TZ           string
}

func LoadConfig() (Config, error) {
	botToken, err := requiredEnv(envBotToken)
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := DatabaseURL()
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

	tz, err := requiredEnv(envTz)
	if err != nil {
		return Config{}, err
	}

	return Config{
		BotToken:     botToken,
		DatabaseURL:  databaseURL,
		NotifyUserID: notifyUserID,
		GroupID:      groupID,
		TZ:           tz,
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

func DatabaseURL() (string, error) {
	host, err := requiredEnv(envDBHost)
	port, err := requiredEnv(envDBPort)
	name, err := requiredEnv(envDBName)
	user, err := requiredEnv(envDBUser)
	password := os.Getenv("DB_PASSWORD")
	sslMode, err := requiredEnv(envDBSSLMode)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		url.QueryEscape(name),
		sslMode,
	), err
}
