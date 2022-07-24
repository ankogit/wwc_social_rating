package service

import (
	config "github.com/ankogit/wwc_social_rating/configs"
	"github.com/ankogit/wwc_social_rating/pkg/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Services struct {
	Bot          *tgbotapi.BotAPI
	Config       *config.IniConf
	Repositories *storage.Repositories
}
type Deps struct {
	Bot          *tgbotapi.BotAPI
	Config       *config.IniConf
	Repositories *storage.Repositories
}

func NewServices(deps Deps) *Services {

	return &Services{
		Bot:          deps.Bot,
		Config:       deps.Config,
		Repositories: deps.Repositories,
	}
}
