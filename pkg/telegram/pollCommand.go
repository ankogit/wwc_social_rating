package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ankogit/wwc_social_rating/pkg/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const PREFIX_UP_VOTE = "➕"
const PREFIX_DOWN_VOTE = "➖"
const PREFIX_CANCEL_VOTE = "🛑"

type RatePollResult struct {
	Positive int
	Negative int
	Canceled int
	Medal    int
	Clown    int
	Heart    int
	Like     int
	Fun      int
	Skull    int
	Hole     int
}

type RatePollResultType uint

const (
	VOTE_UP     RatePollResultType = 0
	VOTE_DOWN   RatePollResultType = 1
	VOTE_CANCEL RatePollResultType = 2

	ADD_MEDAL RatePollResultType = 3
	ADD_CLOWN RatePollResultType = 4
	ADD_HEART RatePollResultType = 5
	ADD_LIKE  RatePollResultType = 6
	ADD_FUN   RatePollResultType = 7
	ADD_SKULL RatePollResultType = 8
	ADD_HOLE  RatePollResultType = 9
)

func (b *Bot) CreateAchievementPoll(message *tgbotapi.Message) error {

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

	poll := tgbotapi.SendPollConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
		},
		Question: fmt.Sprintf("Голосование за присуждение внеочередного звания пользователю: %s %s - @%s #%s", userData.FirstName, userData.LastName, userData.UserName, strconv.Itoa(int(userData.ID))),
		Type:     "regular",
		Options: []string{
			fmt.Sprintf(`%s - медаль за заслуги`, AchievementsEmoji[AchievementMedal]),
			fmt.Sprintf(`%s - ну ты клоун`, AchievementsEmoji[AchievementClown]),
			fmt.Sprintf(`%s - это любовь`, AchievementsEmoji[AchievementHeart]),
			fmt.Sprintf(`%s - ну это лайк`, AchievementsEmoji[AchievementLike]),
			fmt.Sprintf(`%s - забавный чел`, AchievementsEmoji[AchievementFun]),
			fmt.Sprintf(`%s - дед инсайд`, AchievementsEmoji[AchievementSkull]),
			fmt.Sprintf(`%s - дырка`, AchievementsEmoji[AchievementHole]),
			fmt.Sprintf(`%s Отмена голосования`, PREFIX_CANCEL_VOTE),
		},
		// CorrectOptionID: 1,
		IsAnonymous: false,
		OpenPeriod:  3600,
	}
	pollMessage, err := b.bot.Send(poll)
	if err != nil {
		return nil
	}
	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Message ID: %s`, strconv.Itoa(pollMessage.MessageID))))
	return nil
}

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

	// if userData.ScoreUpdatedAt != nil && time.Since(*userData.ScoreUpdatedAt) < 24*time.Hour {
	// 	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Изменять рейтинг пользователя можно раз в сутки"))
	// 	return nil
	// }

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
		// OpenPeriod:  16,
		CloseDate: int(time.Now().Add(time.Second * 5).Unix()),
	}
	pollMessage, err := b.bot.Send(poll)
	if err != nil {
		return nil
	}
	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Message ID: %s`, strconv.Itoa(pollMessage.MessageID))))

	b.services.Repositories.Polls.Save(models.Poll{
		ID:        message.MessageID,
		PollID:    pollMessage.Poll.ID,
		UserID:    userData.ID,
		IsClosed:  false,
		CreatedAt: time.Now(),
	})
	fmt.Println(b.services.Repositories.Polls.AllActive())
	return nil
}

func (b *Bot) StopRatePoll(message *tgbotapi.Message) error {
	var messageId int
	var err error

	if message.CommandArguments() != "" {
		messageId, err = strconv.Atoi(message.CommandArguments())
		if err != nil {
			return err
		}
	} else if message.ReplyToMessage != nil && message.ReplyToMessage.MessageID != 0 {
		messageId = message.ReplyToMessage.MessageID
	} else {
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Введите Message ID или отправте в Reply на опрос`)))
		return nil
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

	ratePollResult := b.parseRatePollResult(poll.Options)

	chatMemberCount, err := b.bot.GetChatMembersCount(tgbotapi.ChatMemberCountConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: message.Chat.ID,
		},
	})
	if err != nil {
		return err
	}

	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Результаты: \nПользователь: @%s \nПроголосовали: %s \nЗа отмену: %s \nВсего людей в чате: %s",
		userData.UserName, strconv.Itoa(poll.TotalVoterCount), strconv.Itoa(ratePollResult.Canceled), strconv.Itoa(chatMemberCount))))

	switch b.selectRatePollResultOptions(ratePollResult, chatMemberCount) {
	case VOTE_UP:

		if userData.ScoreUpdatedAt != nil && time.Since(*userData.ScoreUpdatedAt) < 24*time.Hour {
			b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Изменять рейтинг пользователя можно раз в сутки"))
			return nil
		}
		b.AddScore(userData, int64(ratePollResult.Positive))
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Итог: Ранг увеличен`)))

	case VOTE_DOWN:
		if userData.ScoreUpdatedAt != nil && time.Since(*userData.ScoreUpdatedAt) < 24*time.Hour {
			b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Изменять рейтинг пользователя можно раз в сутки"))
			return nil
		}
		b.ReduceScore(userData, int64(ratePollResult.Negative))
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Итог: Ранг понижен`)))

	case VOTE_CANCEL:
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Итог: Голосование отменено`)))

	case ADD_CLOWN:
		b.AddAchievement(userData, AchievementClown)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementMedal])))

	case ADD_MEDAL:
		b.AddAchievement(userData, AchievementMedal)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementMedal])))

	case ADD_HEART:
		b.AddAchievement(userData, AchievementHeart)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementHeart])))

	case ADD_LIKE:
		b.AddAchievement(userData, AchievementLike)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementLike])))

	case ADD_FUN:
		b.AddAchievement(userData, AchievementFun)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementFun])))

	case ADD_SKULL:
		b.AddAchievement(userData, AchievementSkull)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementSkull])))

	case ADD_HOLE:
		b.AddAchievement(userData, AchievementHole)
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`Присвоено внеочередное взание: %s`, AchievementsEmoji[AchievementHole])))

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

		//ачивки
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementMedal]) {
			ratePollResults.Medal = option.VoterCount
		}
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementClown]) {
			ratePollResults.Clown = option.VoterCount
		}
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementHeart]) {
			ratePollResults.Heart = option.VoterCount
		}
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementHole]) {
			ratePollResults.Hole = option.VoterCount
		}
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementLike]) {
			ratePollResults.Like = option.VoterCount
		}
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementFun]) {
			ratePollResults.Fun = option.VoterCount
		}
		if strings.HasPrefix(option.Text, AchievementsEmoji[AchievementSkull]) {
			ratePollResults.Skull = option.VoterCount
		}
	}
	return ratePollResults
}

func (b *Bot) selectRatePollResultOptions(pollResults RatePollResult, userCharCount int) RatePollResultType {
	// if userCharCount <= 2 || (pollResults.Canceled+pollResults.Negative+pollResults.Positive+
	// 	pollResults.Clown+
	// 	pollResults.Medal+
	// 	pollResults.Heart+
	// 	pollResults.Like+
	// 	pollResults.Fun+
	// 	pollResults.Skull+
	// 	pollResults.Hole) < (userCharCount/2) {
	// 	return VOTE_CANCEL
	// }

	// Создаем map с названиями полей и их значением
	fields := map[string]int{
		"Positive": pollResults.Positive,
		"Negative": pollResults.Negative,
		"Canceled": pollResults.Canceled,
		"Medal":    pollResults.Medal,
		"Clown":    pollResults.Clown,
		"Heart":    pollResults.Heart,
		"Like":     pollResults.Like,
		"Fun":      pollResults.Fun,
		"Skull":    pollResults.Skull,
		"Hole":     pollResults.Hole,
	}

	// Инициализируем переменные для наивысшего значения и его поле
	maxValue := -1
	maxField := ""

	// Находим поле с наивысшим значением
	for field, value := range fields {
		if value > maxValue {
			maxValue = value
			maxField = field
		}
	}

	if pollResults.Canceled >= pollResults.Negative+pollResults.Positive+
		pollResults.Clown+
		pollResults.Medal+
		pollResults.Heart+
		pollResults.Like+
		pollResults.Fun+
		pollResults.Skull+
		pollResults.Hole {
		return VOTE_CANCEL
	}

	// Возвращаем соответствующий тип действия в зависимости от поля с наивысшим значением
	switch maxField {
	case "Positive":
		return VOTE_UP
	case "Negative":
		return VOTE_DOWN
	case "Canceled":
		return VOTE_CANCEL
	case "Medal":
		return ADD_MEDAL
	case "Clown":
		return ADD_CLOWN
	case "Heart":
		return ADD_HEART
	case "Like":
		return ADD_LIKE
	case "Fun":
		return ADD_FUN
	case "Skull":
		return ADD_SKULL
	case "Hole":
		return ADD_HOLE
	default:
		return VOTE_CANCEL
	}
}
