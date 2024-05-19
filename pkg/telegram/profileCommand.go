package telegram

import (
	"fmt"

	"github.com/ankogit/wwc_social_rating/pkg/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) GetUserProfile(message *tgbotapi.Message) error {
	var userData models.User
	var err error
	if message.CommandArguments() != "" {
		userData, err = b.services.Repositories.Users.GetByUsername(message.CommandArguments())
		if err != nil {
			return err
		}
	} else {
		userData, err = b.getOrCreateUserByMessage(message.From)
		if err != nil {
			return err
		}
	}

	if generatedFile, err := b.GenerateImageUserCard(userData); err == nil {
		photoFileBytes := tgbotapi.FileBytes{
			Name:  "picture",
			Bytes: generatedFile,
		}
		b.bot.Send(
			tgbotapi.PhotoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID: message.Chat.ID,
						// ReplyMarkup: InlineKeyboardButtonMarkup(message.From.ID),
					},
					File: photoFileBytes,
				},
				//Caption: "Test",
			},
		)
	}

	return nil
}

func InlineKeyboardButtonMarkup(userId int64) tgbotapi.InlineKeyboardMarkup {
	var rows []tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("➕", fmt.Sprintf("user:like:%v", userId)))
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("➖", fmt.Sprintf("user:dislike:%v", userId)))

	return tgbotapi.NewInlineKeyboardMarkup(rows)

}

func (b *Bot) getOrCreateUserByMessage(tgUser *tgbotapi.User) (models.User, error) {
	if tgUser == nil {
		return models.User{}, nil
	}
	userData, err := b.services.Repositories.Users.Get(tgUser.ID)

	if err != nil {
		if "not found" == err.Error() {
			var user models.User
			user.ID = tgUser.ID
			user.FirstName = tgUser.FirstName
			user.LastName = tgUser.LastName
			user.UserName = tgUser.UserName
			user.Score = 10
			b.services.Repositories.Users.Save(user)

			userData = user
			return userData, nil
		} else {
			return models.User{}, err
		}
	}
	return userData, nil
}
