package jobs

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TelegramNotifications struct {
	Bot TelegramBot
}

type TelegramBot interface {
	SendTestMessage(message *tgbotapi.Message) error
}

func NewTelegramNotifications(bot TelegramBot) *TelegramNotifications {
	return &TelegramNotifications{Bot: bot}
}

func (t *TelegramNotifications) NotifyStats(message *tgbotapi.Message) error {
	return t.Bot.SendTestMessage(message)
}
