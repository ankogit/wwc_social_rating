package handler

import (
	"bytes"
	"fmt"
	"github.com/ankogit/wwc_social_rating/pkg/helpers"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) getTestPage(c *gin.Context) {
	idParam := c.Param("id")

	id, _ := strconv.ParseInt(idParam, 10, 64)

	userAvatars, err := h.services.Bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: id,
	})
	if err != nil {
		log.Panicln(err)
	}
	imageFlag := false
	if userAvatars.TotalCount != 0 {
		imageFlag = true
	}

	var newSizeAvatarImage image.Image
	if imageFlag {
		avatarFile, err := h.services.Bot.GetFile(tgbotapi.FileConfig{
			FileID: userAvatars.Photos[0][0].FileID,
		})
		avatarFileSrc := fmt.Sprintf("https://api.telegram.org/file/bot%v/%v", h.services.Config.GetStringKey("", "telegramkey"), avatarFile.FilePath)
		imageRes, err := http.Get(avatarFileSrc)
		if err != nil || imageRes.StatusCode != 200 {
			log.Panicln(err)
		}
		defer imageRes.Body.Close()
		avatarImage, _, err := image.Decode(imageRes.Body)
		if err != nil {
			log.Panicln(err)
		}

		newSizeAvatarImage = resize.Resize(97, 97, avatarImage, resize.Lanczos3)

	} else {
		avatarImage, err := gg.LoadImage("./storage/images/test.png")
		if err != nil {
			log.Fatal(err)
		}
		newSizeAvatarImage = resize.Resize(96, 96, avatarImage, resize.Lanczos3)
	}

	const height = 128
	const wight = 512

	bg, err := gg.LoadImage("./storage/images/tmplate_bot.png")
	if err != nil {
		log.Fatal(err)
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
	dc.DrawStringAnchored(fmt.Sprintf("Фамилия: %v", helpers.TruncateText("Test", 10)), 144, 28, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Имя: %v", helpers.TruncateText("Test", 10)), 144, 54, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Telegram: @%v", helpers.TruncateText("test", 10)), 144, 79, 0, 0)

	if err := dc.LoadFontFace("./storage/fonts/Pixel.ttf", 46); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored(fmt.Sprintf("+12"), 410, 70, 0, 0)

	dc.DrawRoundedRectangle(0, 0, 512, 512, 0)
	dc.DrawImage(newSizeAvatarImage, 16, 16)
	dc.Clip()

	var buffImage bytes.Buffer
	foo := io.Writer(&buffImage)
	dc.EncodePNG(foo)

	//var buf bytes.Buffer
	//gif.Encode(&buf, image.Rect(0, 0, 16, 16), nil)

	c.Writer.Header().Set("Content-Type", "image/jpeg")
	c.Writer.Header().Set("Content-Length", strconv.Itoa(len(buffImage.Bytes())))
	c.Writer.Write(buffImage.Bytes())
	//c.JSON(http.StatusOK, response.DataResponse{Data: fmt.Sprintf("Hello %s", id)})
}
