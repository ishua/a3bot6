package main

import (
	"context"
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ishua/a3bot6/mcore/pkg/mcoreclient"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type MyConfig struct {
	Token      string `required:"true" env:"TELEGRAMBOTTOKEN" usage:"token for your telegram bot"`
	Debug      bool   `default:"false" usage:"turn on debug mode"`
	MCoreAddr  string `default:"http://127.0.0.1:8080" usage:"host and port for mcore"`
	TBotSecret string `default:"test" usage:"secret key for api"`
}

var (
	cfg MyConfig
)

func main() {

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"conf/tbot_config.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})
	if err := loader.Load(); err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatalf("tg init bot %s", err.Error())
	}

	mcore := mcoreclient.NewClient(cfg.MCoreAddr, cfg.TBotSecret)
	tgClient := newTgClient(bot, mcore)

	ctx, cancel := context.WithCancel(context.Background())

	mcore.ListeningTasks(ctx, schema.TaskTypeMsg, tgClient, time.Duration(1*time.Second))
	log.Println("listen mcore")
	tgClient.ListeningTg(ctx)
	log.Println("listen tgClient")

	// stop service here
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// waiting signal for stop
	sig := <-sigChan
	log.Printf("Received signal: %s. Stopping...\n", sig)
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Program has stopped.")

}

type tgClient struct {
	bot   *tgbotapi.BotAPI
	mcore *mcoreclient.Client
}

func newTgClient(bot *tgbotapi.BotAPI, mcore *mcoreclient.Client) *tgClient {
	return &tgClient{bot: bot, mcore: mcore}
}

func (tg *tgClient) DoTask(task schema.Task) (string, error) {
	if task.Type != schema.TaskTypeMsg {
		return "", fmt.Errorf("dotask only for TaskTypeMsg")
	}
	if task.TaskData.Msg.ChatId == 0 {
		return "", fmt.Errorf("chatId in taskType Msg is 0")
	}
	if len(task.TaskData.Msg.Text) == 0 {
		return "", fmt.Errorf("text in taskType Msg is 0")
	}
	msg := tgbotapi.NewMessage(task.TaskData.Msg.ChatId, task.TaskData.Msg.Text)
	msg.ParseMode = "html"
	msg.ReplyToMessageID = task.TaskData.Msg.ReplyMessageId

	_, err := tg.bot.Send(msg)
	if err != nil {
		return "", fmt.Errorf("something went wrong when tg sending messages %s", err.Error())
	}
	return "", nil
}

func (tg *tgClient) ListeningTg(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)

	go func() {
		for {
			select {
			case update := <-updates:
				{
					if update.Message == nil {
						continue
					}
					if update.Message.Chat == nil {
						continue
					}
					var fileUrl string
					var err error
					if update.Message.Document != nil {
						fileUrl, err = tg.bot.GetFileDirectURL(update.Message.Document.FileID)
						if err != nil {
							log.Printf("tg createMessage can't get file url: %s", err.Error())
							continue
						}
					}
					quickMsg, err := tg.mcore.AddMsg(schema.Message{
						UserName:         update.Message.Chat.UserName,
						MessageId:        update.Message.MessageID,
						ReplyToMessageID: getReplyId(update),
						ChatId:           update.Message.Chat.ID,
						Text:             update.Message.Text,
						Caption:          update.Message.Caption,
						FileUrl:          fileUrl,
						Type:             0,
					})

					if quickMsg.Error != "" {
						_, err := tg.DoTask(schema.Task{
							Type: schema.TaskTypeMsg,
							TaskData: schema.TaskData{
								Msg: schema.TaskMsg{
									ChatId:         update.Message.Chat.ID,
									Text:           fmt.Sprintf("something went wrong %s", err.Error()),
									ReplyMessageId: getReplyId(update),
								},
							},
						})
						if err != nil {
							log.Printf("quickMsg error: %s", err.Error())
						}
						continue
					}
					if quickMsg.Data.ChatId != 0 {
						_, err := tg.DoTask(schema.Task{
							Type: schema.TaskTypeMsg,
							TaskData: schema.TaskData{
								Msg: quickMsg.Data,
							},
						})
						if err != nil {
							log.Printf("quickMsg error: %s", err.Error())
						}
					}
				}

			case <-ctx.Done():
				{
					log.Println("stopping listen telegram")
					return
				}

			}
		}
	}()
}

func getReplyId(update tgbotapi.Update) int {
	var replyMsgId int
	if update.Message != nil && update.Message.ReplyToMessage != nil {
		replyMsgId = update.Message.ReplyToMessage.MessageID
	}
	return replyMsgId
}
