package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) SendTestMessage(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Test message")
	msg.ParseMode = "MARKDOWN"
	msg.DisableNotification = true
	b.bot.Send(msg)
	return nil
}
