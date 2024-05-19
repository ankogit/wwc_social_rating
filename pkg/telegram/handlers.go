package telegram

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ankogit/wwc_social_rating/pkg/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleInlineQuery(query *tgbotapi.InlineQuery) {
	var resources []interface{}

	resources = append(resources,
		tgbotapi.InlineQueryResultArticle{
			Type:        "article",
			ID:          query.ID,
			Title:       b.messages.InlineContentTitle,
			Description: b.messages.InlineContentDescription,
			//URL:         "https://mandarinshow.ru/assets/img/main_iconv2_op.png",
			//ThumbURL:    "https://mandarinshow.ru/assets/img/main_iconv2_op.png",
			InputMessageContent: tgbotapi.InputTextMessageContent{
				Text: "/profile",
				//ParseMode: "markdown",
			},
		})

	if _, err := b.bot.Request(tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		CacheTime:     0,
		IsPersonal:    true,
		Results:       resources}); err != nil {
		log.Println(err)
	}

}

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case "start":
		b.SendWelcomeMessage(message.Chat.ID)
		return nil
	case "help":
		b.SendWelcomeMessage(message.Chat.ID)
		return nil
	case "profile":
		if err := b.handleCommandProfile(message); err != nil {
			return err
		}
		return nil
	case "rate":
		if err := b.handleCommandRate(message); err != nil {
			return err
		}
		return nil
	case "award":
		if err := b.handleCommandAward(message); err != nil {
			return err
		}
		return nil
	case "stoppoll":
		if err := b.handleCommandStopPoll(message); err != nil {
			return err
		}
		return nil
	default:
		return errUnknownCommand
	}
}
func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {

	split := strings.Split(query.Data, ":")
	if split[0] == "user" {
		b.handleUserCallbackQuery(query, split[1:]...)
		return
	}
}

func (b *Bot) handleUserCallbackQuery(query *tgbotapi.CallbackQuery, data ...string) {
	pagerType := data[0]
	userId, _ := strconv.ParseInt(data[1], 10, 64)
	user, _ := b.services.Repositories.Users.Get(userId)

	if pagerType == "like" {
		user, _ = b.AddScore(user, 1)
		b.SendCallbackAnswer(query, "Вы повысили рейтинг")
	}
	if pagerType == "dislike" {
		user, _ = b.ReduceScore(user, 1)
		b.SendCallbackAnswer(query, "Вы понизили рейтинг")
	}

	//Update image message
	generatedFile, err := b.GenerateImageUserCard(user)
	if err != nil {
		log.Panicln(err)
	}
	t := time.Now()

	keyboard := InlineKeyboardButtonMarkup(userId)
	b.bot.Send(tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      query.Message.Chat.ID,
			MessageID:   query.Message.MessageID,
			ReplyMarkup: &keyboard,
		},

		Media: tgbotapi.InputMediaPhoto{
			BaseInputMedia: tgbotapi.BaseInputMedia{
				Type: "photo",
				Media: tgbotapi.FileBytes{
					Name:  fmt.Sprintf("image_%v.png", t.GoString()),
					Bytes: generatedFile,
				},
			},
		},
	})
}

func (b *Bot) SendCallbackAnswer(query *tgbotapi.CallbackQuery, text string) {
	b.bot.Send(tgbotapi.CallbackConfig{
		Text:            text,
		CallbackQueryID: query.ID,
	})
}
func (b *Bot) SendWelcomeMessage(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, "Добро пожаловать в WWC Social Rating Bot!\n\n"+
		"Этот уникальный бот создан для формирования общего социального рейтинга в вашем Telegram-сообществе. С его помощью вы сможете:\n\n"+
		"- Отслеживать достойных людей и тех, кто подводит. Не оставляйте говнюков без внимания!\n"+
		"- Наказывать за невыполненные обещания. Сделаем вашу группу более ответственным и надежным местом.\n"+
		"- Вознаграждать за вклад в open-source проекты. Цените тех, кто вносит значительный вклад в развитие.\n"+
		"- Красиво делиться своим профилем. Покажите всем свои достижения и рейтинг.\n"+
		"- Геймифицировать взаимодействие в группе. Превратите рутинные задачи в увлекательную игру!\n\n"+
		"Возможности:\n"+
		"- /rate <USERNAME> - Оценка пользователя (рейтинг). Дайте обратную связь каждому участнику сообщества.\n"+
		"- /award <USERNAME> - Запуск голосования на награду пользователя. Позвольте сообществу выбрать лучших из лучших.\n"+
		"- /stoppoll <MESSAGE ID> - Остановка голосования, подведение итогов.\n"+
		"- /profile - Получить карточку своего профиля.\n"+
		"- /profile <USERNAME> - Получить карточку профиля пользователя.\n\n"+
		"Присоединяйтесь к WWC Social Rating Bot и сделайте ваше сообщество более активным, справедливым и увлекательным!")
	b.bot.Send(msg)
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	if message == nil || message.From == nil {
		return nil
	}
	_, err := b.getOrCreateUserByMessage(message.From)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) handleCommandProfile(message *tgbotapi.Message) error {
	return b.GetUserProfile(message)
}

func (b *Bot) handleCommandRate(message *tgbotapi.Message) error {
	return b.CreateRatePoll(message)
}
func (b *Bot) handleCommandAward(message *tgbotapi.Message) error {
	return b.CreateAchievementPoll(message)
}

func (b *Bot) handleCommandStopPoll(message *tgbotapi.Message) error {
	return b.StopRatePoll(message)
}

func (b *Bot) handleNotificationEnable(message *tgbotapi.Message) error {
	if message.Chat == nil || message.Chat.Title == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Notifications are allowed to conduct group chats")
		b.bot.Send(msg)
		return nil
	}
	cParts := strings.SplitAfterN(message.Text, " ", 2)
	if len(cParts) == 1 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "To set notification you need to pass the cron param")
		b.bot.Send(msg)
		return nil
	}
	cronParam := cParts[1]

	var validID = regexp.MustCompile(`(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*) ?){5,7})`)
	if !validID.MatchString(cronParam) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Cron string is not correct :(")
		b.bot.Send(msg)
		return nil
	}

	var chat = models.Chat{
		ID:               message.Chat.ID,
		Title:            message.Chat.Title,
		NotificationCron: cronParam,
		EntryId:          0,
	}
	if err := b.chatRepository.Save(chat); err != nil {
		return err
	}

	entryId, err := b.cronService.SetJob(&chat, cronParam)
	if err != nil {
		return err
	}
	chat.EntryId = entryId

	if err := b.chatRepository.Save(chat); err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Notifications are installed for this chat 🔔")
	b.bot.Send(msg)
	return nil
}

func (b *Bot) handleNotificationDisable(message *tgbotapi.Message) error {
	if message.Chat == nil || message.Chat.Title == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Notifications are allowed to conduct group chats")
		b.bot.Send(msg)
		return nil
	}

	chat, _ := b.chatRepository.Get(message.Chat.ID)

	if (chat == (models.Chat{})) || (chat.NotificationCron == "" && chat.EntryId == 0) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Notifications are not already setting for this chat")
		b.bot.Send(msg)
		return nil
	}

	b.cronService.RemoveJob(&chat)
	chat.NotificationCron = ""
	chat.EntryId = 0

	if err := b.chatRepository.Save(chat); err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Notifications are disabled for this chat 🔕")
	b.bot.Send(msg)
	return nil
}
