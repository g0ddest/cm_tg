package bot

import (
	"cm_tg/internal/config"
	"cm_tg/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func StartBot(cfg config.Config) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Error creating Telegram bot: %s", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Error getting updates channel: %s", err)
	}

	for update := range updates {
		if update.Message != nil {
			handlers.HandleMessage(cfg, update.Message)
		}
	}
}
