package telegram

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ankogit/wwc_social_rating/pkg/helpers"
	"github.com/ankogit/wwc_social_rating/pkg/models"
	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nfnt/resize"
)

func (b *Bot) ReduceScore(user models.User, score int64) (models.User, error) {
	if !user.IsLastUp {
		user, _ = b.AddAchievement(user, AchievementDec)
	}

	user.IsLastUp = false
	user.Score -= score
	scoreUpdatedAt := time.Now()
	user.ScoreUpdatedAt = &scoreUpdatedAt
	err := b.services.Repositories.Users.Save(user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (b *Bot) AddAchievement(user models.User, achievement string) (models.User, error) {
	achievements := strings.Split(user.Achievements, ",")
	achievements = append(achievements, achievement)
	user.Achievements = strings.Join(achievements, ",")
	err := b.services.Repositories.Users.Save(user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (b *Bot) AddScore(user models.User, score int64) (models.User, error) {
	if user.IsLastUp {
		user, _ = b.AddAchievement(user, AchievementInc)
	}

	user.Score += score
	scoreUpdatedAt := time.Now()
	user.ScoreUpdatedAt = &scoreUpdatedAt
	user.IsLastUp = true
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

	const height = 160
	const wight = 544

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
	dc.DrawStringAnchored(fmt.Sprintf("Фамилия: %v", helpers.TruncateText(user.LastName, 10)), 164, 44, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Имя: %v", helpers.TruncateText(user.FirstName, 10)), 164, 76, 0, 0)
	dc.DrawStringAnchored(fmt.Sprintf("Telegram: @%v", helpers.TruncateText(user.UserName, 15)), 164, 108, 0, 0)

	//dc.DrawStringAnchored(fmt.Sprintf("PROGRESS: 🤡x5|🏅x2"), 144, 116, 0, 0)

	if err := dc.LoadFontFace("./storage/fonts/Pixel.ttf", 46); err != nil {
		panic(err)
	}
	if user.Score > 0 {
		dc.SetColor(color.RGBA{
			R: 98,
			G: 172,
			B: 76,
			A: 255,
		})
		dc.DrawStringAnchored(fmt.Sprintf("+%v", user.Score), 470, 84, 0.5, 0.5)
	} else if user.Score < 0 {
		dc.SetColor(color.RGBA{
			R: 208,
			G: 49,
			B: 67,
			A: 255,
		})
		dc.DrawStringAnchored(fmt.Sprintf("%v", user.Score), 470, 84, 0.5, 0.5)
	} else {
		dc.SetColor(color.RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		})
		dc.DrawStringAnchored(fmt.Sprintf("%v", user.Score), 470, 84, 0.5, 0.5)
	}

	//dc.DrawStringAnchored(fmt.Sprintf("Фамилия: %v", message.From.FirstName)message.Chat.Bio, 228, 80, 0, 0)
	//dc.DrawStringAnchored("Рейтинг: ", wight/2, height/2, 0.5, 0.5)

	dc.DrawRoundedRectangle(0, 0, 512, 512, 0)
	dc.DrawImage(newSizeAvatarImage, 32, 32)

	// Откройте файл с сеткой спрайтов
	file, err := os.Open("./storage/images/tmplate_bot-Sheet.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Декодируйте изображение
	spritesheet, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	achievements := generateAchievents(user.Achievements)

	dc, err = drawAchievements(dc, achievements, spritesheet)
	if err != nil {
		return nil, err
	}
	dc.Clip()

	var buffImage bytes.Buffer
	foo := io.Writer(&buffImage)
	err = dc.EncodePNG(foo)
	if err != nil {
		return nil, err
	}

	return buffImage.Bytes(), nil
}

func generateAchievents(achStr string) UserAchievements {
	achievements := NewUserAchievements()

	achievementsStr := strings.Split(achStr, ",")
	achievements.IncrementAchievement("heart")

	for _, achievement := range achievementsStr {
		if spite := getSpriteByName(achievement); spite != nil {
			// Добавляем достижения
			achievements.IncrementAchievement(achievement)
		}
	}
	return *achievements
}

func drawAchievements(dc *gg.Context, userAchievements UserAchievements, spritesheet image.Image) (*gg.Context, error) {
	if len(userAchievements.Achievements) == 0 {
		return dc, nil
	}
	fmt.Println("userAchievements.Achievements", userAchievements.Achievements)
	startPointX, startPointY := 166, 118
	dc.SetColor(color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	})
	if err := dc.LoadFontFace("./storage/fonts/Pixel.ttf", 16); err != nil {
		panic(err)
	}
	for achievementName, achievementCount := range userAchievements.Achievements {
		currSprite := getSpriteByName(achievementName)
		spriteRect := image.Rect(currSprite.X*currSprite.Width, currSprite.Y*currSprite.Height, (currSprite.X+1)*currSprite.Width, (currSprite.Y+1)*currSprite.Height)
		sprite := image.NewRGBA(image.Rect(0, 0, currSprite.Width, currSprite.Height))
		draw.Draw(sprite, sprite.Bounds(), spritesheet, spriteRect.Min, draw.Src)
		dc.DrawImage(sprite, startPointX, startPointY)

		if achievementCount > 1 {
			dc.DrawStringAnchored(fmt.Sprintf("%v", achievementCount), float64(startPointX)+16, float64(startPointY)+16, 0.5, 0.5)
		}

		startPointX += 24
	}

	return dc, nil
}

type Sprite struct {
	Name   string
	X, Y   int // Позиция спрайта по X, U (столбец, строка начиная с 0)
	Width  int
	Height int
}

func getSpriteByName(name string) *Sprite {
	for _, sprite := range sprites {
		if sprite.Name == name {
			return &sprite
		}
	}
	return nil
}

var sprites = []Sprite{
	{"clown", 0, 0, 16, 16},
	{"up", 3, 0, 16, 16},
	{"down", 4, 0, 16, 16},
	{"medal", 5, 0, 16, 16},
	{"heart", 6, 0, 16, 16},
	{"money", 7, 0, 16, 16},
	{"moneys", 8, 0, 16, 16},
	{"moneyOne", 9, 0, 16, 16},
	{"skull", 10, 0, 16, 16},
	{"inc", 11, 0, 16, 16},
	{"dec", 12, 0, 16, 16},
	{"hole", 13, 0, 16, 16},
	{"like", 14, 0, 16, 16},
	{"time", 15, 0, 16, 16},
	{"fun", 16, 0, 16, 16},
}

// Определение констант для названий достижений
const (
	AchievementClown    = "clown"
	AchievementMedal    = "medal"
	AchievementHeart    = "heart"
	AchievementMoney    = "money"
	AchievementMoneys   = "moneys"
	AchievementMoneyOne = "moneyOne"
	AchievementSkull    = "skull"
	AchievementInc      = "inc"
	AchievementDec      = "dec"
	AchievementHole     = "hole"
	AchievementLike     = "like"
	AchievementTime     = "time"
	AchievementFun      = "fun"
)

// Определение map для сопоставления названий достижений с emoji
var AchievementsEmoji = map[string]string{
	AchievementClown:    "🤡",
	AchievementMedal:    "🏅",
	AchievementHeart:    "❤️",
	AchievementMoney:    "💰",
	AchievementMoneys:   "💵",
	AchievementMoneyOne: "💲",
	AchievementSkull:    "💀",
	AchievementHole:     "🕳️",
	AchievementLike:     "👍",
	AchievementTime:     "⌚",
	AchievementFun:      "😄",
}

type UserAchievements struct {
	Achievements map[string]int
}

// NewUserAchievements создает новый экземпляр UserAchievements
func NewUserAchievements() *UserAchievements {
	return &UserAchievements{
		Achievements: make(map[string]int),
	}
}

// AddAchievement добавляет или обновляет достижение
func (ua *UserAchievements) AddAchievement(name string, count int) {
	ua.Achievements[name] = count
}

// GetAchievement возвращает количество достижений по названию
func (ua *UserAchievements) GetAchievement(name string) (int, bool) {
	count, exists := ua.Achievements[name]
	return count, exists
}

// IncrementAchievement увеличивает количество достижений на 1
func (ua *UserAchievements) IncrementAchievement(name string) {
	if count, exists := ua.Achievements[name]; exists {
		ua.Achievements[name] = count + 1
	} else {
		ua.Achievements[name] = 1
	}
}

// DecrementAchievement уменьшает количество достижений на 1
func (ua *UserAchievements) DecrementAchievement(name string) {
	if count, exists := ua.Achievements[name]; exists && count > 0 {
		ua.Achievements[name] = count - 1
	}
}
