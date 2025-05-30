package service

import (
	"fmt"

	"github.com/ankogit/wwc_social_rating/pkg/models"
	"github.com/ankogit/wwc_social_rating/pkg/storage"
	"github.com/ankogit/wwc_social_rating/pkg/telegram/jobs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

type CronService struct {
	Scheduler             *cron.Cron
	Repositories          *storage.Repositories
	TelegramNotifications *jobs.TelegramNotifications
}

func NewCronService(sh *cron.Cron, r *storage.Repositories, tn *jobs.TelegramNotifications) *CronService {
	return &CronService{
		Scheduler:             sh,
		Repositories:          r,
		TelegramNotifications: tn,
	}
}
func (c *CronService) Start() {
	c.Scheduler.Start()
}
func (c *CronService) Init() {

	chats, err := c.Repositories.Chats.All()
	if err != nil {
		fmt.Println(err)
	}

	for _, chat := range chats {
		if chat.EntryId != 0 && chat.NotificationCron != "" {
			jobID, err := c.Scheduler.AddFunc(chat.NotificationCron, func() {
				c.TelegramNotifications.NotifyStats(&tgbotapi.Message{
					Chat: &tgbotapi.Chat{
						ID: chat.ID,
					},
				})
			})
			if err != nil {
				fmt.Println(err)
			}
			chat.EntryId = jobID
			if err := c.Repositories.Chats.Save(chat); err != nil {
				fmt.Println(err)
			}
		}

	}

	// c.Scheduler.AddFunc("1 * * * *", func() {
	// 	c.Repositories.Polls.AllActive()

	// 	c.TelegramNotifications.StopPoll(&tgbotapi.Message{
	// 		Chat: &tgbotapi.Chat{
	// 			ID: chat.ID,
	// 		},
	// 	})
	// })
}

func (c *CronService) SetJob(chat *models.Chat, notificationCron string) (cron.EntryID, error) {
	c.RemoveJob(chat)

	jobID, err := c.Scheduler.AddFunc(chat.NotificationCron, func() {
		c.TelegramNotifications.NotifyStats(&tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: chat.ID,
			},
		})
	})
	if err != nil {
		return 0, err
	}
	return jobID, nil
}

func (c *CronService) RemoveJob(chat *models.Chat) {
	if chat.EntryId != 0 {
		c.Scheduler.Remove(chat.EntryId)
	}
}
