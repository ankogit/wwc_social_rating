package telegram

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"time"

	"github.com/ankogit/wwc_social_rating/pkg/helpers"
	"github.com/ankogit/wwc_social_rating/pkg/models"
	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nfnt/resize"
)

func (b *Bot) ReduceScore(user models.User, score int64) (models.User, error) {
	user.Score -= score
	scoreUpdatedAt := time.Now()
	user.ScoreUpdatedAt = &scoreUpdatedAt
	err := b.services.Repositories.Users.Save(user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (b *Bot) AddScore(user models.User, score int64) (models.User, error) {
	user.Score += score
	scoreUpdatedAt := time.Now()
	user.ScoreUpdatedAt = &scoreUpdatedAt
	err := b.services.Repositories.Users.Save(user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (b *Bot) GenerateImageUserCard(user models.User) ([]byte, error) {
	userAvatars, err := b.bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: user.ID,
	})
	if err != nil {
		return nil, err
	}
	imageFlag := false
	if userAvatars.TotalCount != 0 {
		imageFlag = true
	}

	var newSizeAvatarImage image.Image
	if imageFlag {
		avatarFile, err := b.bot.GetFile(tgbotapi.FileConfig{
			FileID: userAvatars.Photos[0][0].FileID,
		})
		avatarFileSrc := fmt.Sprintf("https://api.telegram.org/file/bot%v/%v", b.config.GetStringKey("", "telegramkey"), avatarFile.FilePath)
		imageRes, err := http.Get(avatarFileSrc)
		if err != nil || imageRes.StatusCode != 200 {
			return nil, err
		}
		defer imageRes.Body.Close()
		avatarImage, _, err := image.Decode(imageRes.Body)
		if err != nil {
			return nil, err
		}

		newSizeAvatarImage = resize.Resize(97, 97, avatarImage, resize.Lanczos3)

	} else {
		avatarImage, err := gg.LoadImage("./storage/images/default.png")
		if err != nil {
			return nil, err
		}
		newSizeAvatarImage = resize.Resize(97, 97, avatarImage, resize.Lanczos3)
	}

	const height = 128
	const wight = 512

	bg, err := gg.LoadImage("./storage/images/tmplate_bot.png")
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(wight, height)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("./storage/fonts/pixel-font.ttf", 23); err != nil {
		panic(err)
	}
	dc.SetColor(color.RGBA{
		R: 255,
		B: 255,
		G: 255,
		A: 255,
	})
	dc.DrawImage(bg, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Фамилия: %v", helpers.TruncateText(user.LastName, 10)), 144, 28, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Имя: %v", helpers.TruncateText(user.FirstName, 10)), 144, 54, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Telegram: @%v", helpers.TruncateText(user.UserName, 15)), 144, 79, 0, 0)

	//dc.DrawStringAnchored(fmt.Sprintf("PROGRESS: 🤡x5|🏅x2"), 144, 116, 0, 0)

	if err := dc.LoadFontFace("./storage/fonts/Pixel.ttf", 46); err != nil {
		panic(err)
	}
	if user.Score >= 0 {
		dc.SetColor(color.RGBA{
			R: 98,
			G: 172,
			B: 76,

			A: 255,
		})
		dc.DrawStringAnchored(fmt.Sprintf("+%v", user.Score), 440, 53, 0.5, 0.5)
	} else {
		dc.SetColor(color.RGBA{
			R: 208,
			G: 49,
			B: 67,
			A: 255,
		})
		dc.DrawStringAnchored(fmt.Sprintf("%v", user.Score), 440, 53, 0.5, 0.5)
	}

	//dc.DrawStringAnchored(fmt.Sprintf("Фамилия: %v", message.From.FirstName)message.Chat.Bio, 228, 80, 0, 0)
	//dc.DrawStringAnchored("Рейтинг: ", wight/2, height/2, 0.5, 0.5)

	dc.DrawRoundedRectangle(0, 0, 512, 512, 0)
	dc.DrawImage(newSizeAvatarImage, 16, 16)
	//dc.DrawStringAnchored("Hello, world!", S/2, S/2, 0.5, 0.5)
	dc.Clip()

	var buffImage bytes.Buffer
	foo := io.Writer(&buffImage)
	err = dc.EncodePNG(foo)
	if err != nil {
		return nil, err
	}

	return buffImage.Bytes(), nil
}
