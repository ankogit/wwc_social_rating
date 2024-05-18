package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const PREFIX_UP_VOTE = "➕"
const PREFIX_DOWN_VOTE = "➖"
const PREFIX_CANCEL_VOTE = "🛑"

type RatePollResult struct {
	Positive int `json:"positive"`
	Negative int `json:"negative"`
	Canceled int `json:"canceled"`
}

type RatePollResultType uint

const (
	VOTE_UP     RatePollResultType = 0
	VOTE_DOWN   RatePollResultType = 1
	VOTE_CANCEL RatePollResultType = 2
)

func (b *Bot) CreateRatePoll(message *tgbotapi.Message) error {

	u, _ := b.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: message.Chat.ID,
			UserID: message.From.ID,
		},
	})
	if !(u.IsAdministrator() || u.IsCreator()) || u.User.IsBot {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Создавать опросы может только админ КАНАЛА/ГРУППЫ`)))
		return nil
	}

	if message.CommandArguments() == "" {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Введите username человека`)))
		return nil
	}

	userData, err := b.services.Repositories.Users.GetByUsername(message.CommandArguments())
	if err != nil {
		//TODO: добавить вывод, если человек не найден в боте
		return err
	}

	pollUser, _ := b.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: message.Chat.ID,
			UserID: userData.ID,
		},
	})

	if pollUser.Status == "left" {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Можно проводить опросы человека, который сейчас находится в этом чате."))
		return nil
	}

	if userData.ScoreUpdatedAt != nil && time.Since(*userData.ScoreUpdatedAt) < 24*time.Hour {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Изменять рейтинг пользователя можно раз в сутки"))
		return nil
	}

	poll := tgbotapi.SendPollConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
		},
		Question: fmt.Sprintf("Голосование за %s %s - @%s #%s", userData.FirstName, userData.LastName, userData.UserName, strconv.Itoa(int(userData.ID))),
		Type:     "regular",
		Options: []string{
			fmt.Sprintf(`%s Похвалить`, PREFIX_UP_VOTE),
			fmt.Sprintf(`%s Поругать`, PREFIX_DOWN_VOTE),
			fmt.Sprintf(`%s Отмена голосования`, PREFIX_CANCEL_VOTE),
		},
		// CorrectOptionID: 1,
		IsAnonymous: false,
		// OpenPeriod:      20,
	}
	pollMessage, err := b.bot.Send(poll)
	if err != nil {
		return nil
	}
	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Message ID: %s`, strconv.Itoa(pollMessage.MessageID))))
	return nil
}

func (b *Bot) StopRatePoll(message *tgbotapi.Message) error {
	if message.CommandArguments() == "" {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Введите message ID`)))
	}
	messageId, err := strconv.Atoi(message.CommandArguments())
	if err != nil {
		return err
	}

	poll, err := b.bot.StopPoll(tgbotapi.StopPollConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    message.Chat.ID,
			MessageID: messageId,
		},
	})
	if err != nil {
		return err
	}

	split := strings.Split(poll.Question, "#")
	userID := split[len(split)-1]
	userIDint, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return err
	}

	userData, err := b.services.Repositories.Users.Get(userIDint)
	if err != nil {
		//TODO: добавить вывод, если человек не найден в боте
		return err
	}

	if userData.ScoreUpdatedAt != nil && time.Since(*userData.ScoreUpdatedAt) < 24*time.Hour {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Изменять рейтинг пользователя можно раз в сутки"))
		return nil
	}

	ratePollResult := b.parseRatePollResult(poll.Options)

	chatMemberCount, err := b.bot.GetChatMembersCount(tgbotapi.ChatMemberCountConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: message.Chat.ID,
		},
	})
	if err != nil {
		return err
	}

	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Результаты: \nПользователь: @%s \nПохвалили: %s \nПоругали: %s \nЗа отмену: %s \nВсего людей в чате: %s",
		userData.UserName, strconv.Itoa(ratePollResult.Positive), strconv.Itoa(ratePollResult.Negative), strconv.Itoa(ratePollResult.Canceled), strconv.Itoa(chatMemberCount))))

	switch b.selectRatePollResultOptions(ratePollResult, chatMemberCount) {
	case VOTE_UP:
		b.AddScore(userData, int64(ratePollResult.Positive))
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Итог: Ранг увеличен`)))

	case VOTE_DOWN:
		b.ReduceScore(userData, int64(ratePollResult.Negative))
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Итог: Ранг понижен`)))

	case VOTE_CANCEL:
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Итог: Голосование отменено`)))
	}
	return nil
}

func (b *Bot) parseRatePollResult(options []tgbotapi.PollOption) RatePollResult {
	ratePollResults := RatePollResult{}
	for _, option := range options {
		if strings.HasPrefix(option.Text, PREFIX_UP_VOTE) {
			ratePollResults.Positive = option.VoterCount
		}
		if strings.HasPrefix(option.Text, PREFIX_DOWN_VOTE) {
			ratePollResults.Negative = option.VoterCount
		}
		if strings.HasPrefix(option.Text, PREFIX_CANCEL_VOTE) {
			ratePollResults.Canceled = option.VoterCount
		}
	}
	return ratePollResults
}

func (b *Bot) selectRatePollResultOptions(pollResults RatePollResult, userCharCount int) RatePollResultType {
	if userCharCount <= 2 || (pollResults.Canceled+pollResults.Negative+pollResults.Positive) < (userCharCount/2) {
		return VOTE_CANCEL
	}

	if pollResults.Canceled >= pollResults.Positive+pollResults.Negative {
		return VOTE_CANCEL
	} else {
		if pollResults.Positive > pollResults.Negative {
			return VOTE_UP
		} else if pollResults.Negative > pollResults.Positive {
			return VOTE_DOWN
		} else {
			return VOTE_CANCEL
		}
	}
}
