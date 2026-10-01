package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	dailyHour   = 19
	dailyMinute = 0
)

type Scheduler struct {
	bot *tgbotapi.BotAPI
	cfg Config
	loc *time.Location
}

func NewScheduler(bot *tgbotapi.BotAPI, cfg Config) (*Scheduler, error) {
	location, err := time.LoadLocation(cfg.TZ)
	if err != nil {
		return nil, fmt.Errorf(
			"загрузка часового пояса %s: %w",
			cfg.TZ,
			err,
		)
	}

	return &Scheduler{
		bot: bot,
		cfg: cfg,
		loc: location,
	}, nil
}

func (s *Scheduler) Run(ctx context.Context) {
	for {
		next := s.nextRun(time.Now())

		timer := time.NewTimer(
			time.Until(next),
		)

		select {
		case <-ctx.Done():
			stopTimer(timer)
			log.Println("Планировщик остановлен")
			return

		case <-timer.C:
			s.sendDailyMessage()
		}
	}
}

func (s *Scheduler) nextRun(now time.Time) time.Time {
	now = now.In(s.loc)

	next := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		dailyHour,
		dailyMinute,
		0,
		0,
		s.loc,
	)

	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}

func (s *Scheduler) sendDailyMessage() {
	text := fmt.Sprintf(
		"<a href=\"tg://user?id=%d\">Шаман</a> "+
			"что там по расчётам КТр по\n"+
			"Штату- ?\n"+
			"Списку- ?\n"+
			"На лбс- ?\n"+
			"В работе- ?\n"+
			"Ближний- ?\n"+
			"Дальний - ?",
		s.cfg.NotifyUserID,
	)

	msg := tgbotapi.NewMessage(
		s.cfg.GroupID,
		text,
	)

	msg.ParseMode = tgbotapi.ModeHTML

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf(
			"ошибка отправки ежедневного сообщения: %v",
			err,
		)
	}
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
