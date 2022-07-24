package telegram

import (
	config "github.com/ankogit/wwc_social_rating/configs"
	"github.com/ankogit/wwc_social_rating/pkg/service"
	"github.com/ankogit/wwc_social_rating/pkg/storage"
	"github.com/ankogit/wwc_social_rating/pkg/telegram/jobs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"log"
)

type Bot struct {
	bot            *tgbotapi.BotAPI
	config         *config.IniConf
	version        string
	messages       config.Messages
	chatRepository storage.ChatRepository
	cronService    *service.CronService
	services       *service.Services
	Services       *service.Services
}

func NewBot(
	bot *tgbotapi.BotAPI,
	con *config.IniConf,
	version string,
	messages config.Messages,
	chatRepository storage.ChatRepository,
	services *service.Services,
) *Bot {
	return &Bot{bot: bot, config: con, version: version, messages: messages, chatRepository: chatRepository, services: services, Services: services}
}

func (b *Bot) CronInit(scheduler *cron.Cron) {
	telegramNotifications := jobs.NewTelegramNotifications(b)
	b.cronService = service.NewCronService(scheduler, b.chatRepository, telegramNotifications)
	b.cronService.Init()
}
func (b *Bot) CronStart() {
	b.cronService.Start()
}

func (b *Bot) Start() error {

	// Авторизация бота
	log.Printf("Authorized on account %s", b.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	//updates := bot.ListenForWebhook("/" + bot.Token)
	b.handleUpdates(updates)

	return nil
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	// Бесконечно ждем апдейтов от сервера
	for update := range updates {
		switch {
		// Пришло обычное сообщение
		case update.Message != nil && update.Message.ViaBot == nil && !update.Message.IsCommand() && update.Message.ReplyToMessage == nil && update.Message.Chat.Type == "private":
			b.SendWelcomeMessage(update.Message.Chat.ID)
			break

		case update.Message != nil && update.Message.IsCommand():
			if err := b.handleCommand(update.Message); err != nil {
				b.handleError(update.Message.Chat.ID, err)
			}
			break

		// Пришел inline запрос
		case update.InlineQuery != nil:
			b.handleInlineQuery(update.InlineQuery)
			break

		// Пришел callback запрос
		case update.CallbackQuery != nil:
			b.handleCallback(update.CallbackQuery)
			break
		}
	}
}
