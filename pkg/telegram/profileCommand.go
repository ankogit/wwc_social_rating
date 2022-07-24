package telegram

import (
	"fmt"
	"github.com/ankogit/wwc_social_rating/pkg/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) GetUserProfile(message *tgbotapi.Message) error {
	userData, err := b.services.Repositories.Users.Get(message.From.ID)

	if err != nil {
		if "not found" == err.Error() {
			var user models.User
			user.ID = message.From.ID
			user.FirstName = message.From.FirstName
			user.LastName = message.From.LastName
			user.UserName = message.From.UserName
			user.Score = 10
			b.services.Repositories.Users.Save(user)

			userData = user
		} else {
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
						ChatID:      message.Chat.ID,
						ReplyMarkup: InlineKeyboardButtonMarkup(message.From.ID),
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
