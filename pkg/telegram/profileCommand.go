package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

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
		resp, err := uploadImage(userData.UserName, generatedFile)
		if err == nil && resp != "" {
			userData.ProfileURL = resp
			b.services.Repositories.Users.Save(userData)
		}
	}

	return nil
}

// Определяем структуру для JSON-ответа
type ImageData struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URLViewer  string `json:"url_viewer"`
	URL        string `json:"url"`
	DisplayURL string `json:"display_url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int    `json:"size"`
	Time       int    `json:"time"`
	Expiration int    `json:"expiration"`
	Image      struct {
		Filename  string `json:"filename"`
		Name      string `json:"name"`
		Mime      string `json:"mime"`
		Extension string `json:"extension"`
		URL       string `json:"url"`
	} `json:"image"`
	Thumb struct {
		Filename  string `json:"filename"`
		Name      string `json:"name"`
		Mime      string `json:"mime"`
		Extension string `json:"extension"`
		URL       string `json:"url"`
	} `json:"thumb"`
	DeleteURL string `json:"delete_url"`
}

type ImgbbResponse struct {
	Data    ImageData `json:"data"`
	Success bool      `json:"success"`
	Status  int       `json:"status"`
}

func uploadImage(name string, image []byte) (string, error) {
	apiKey := "6c4883a310e0c81da5d9d11d04fac2a4"
	url := fmt.Sprintf("https://api.imgbb.com/1/upload?name=%s&key=%s", name, apiKey)

	// Создаем буфер и мультипарт-вайтер
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем изображение в мультипарт-форму
	part, err := writer.CreateFormFile("image", "image.jpg")
	if err != nil {
		return "", err
	}
	_, err = part.Write(image)
	if err != nil {
		return "", err
	}

	// Закрываем writer, чтобы завершить мультипарт-сообщение
	err = writer.Close()
	if err != nil {
		return "", err
	}

	// Создаем новый POST-запрос
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", err
	}

	// Устанавливаем заголовок Content-Type для мультипарт-формы
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response ImgbbResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("upload failed, status: %d", response.Status)
	}

	return response.Data.URL, nil
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
			user.Score = 0
			b.services.Repositories.Users.Save(user)

			userData = user
			return userData, nil
		} else {
			return models.User{}, err
		}
	}

	userData.FirstName = tgUser.FirstName
	userData.LastName = tgUser.LastName
	b.services.Repositories.Users.Save(userData)

	return userData, nil
}
