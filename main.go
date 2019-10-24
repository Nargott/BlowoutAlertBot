package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
)

// Config describes main bot configurations
type Config struct {
	TelegramToken string `json:"telegram_token"`
	TargetChatID  int64  `json:"target_chat_id"`

	EmissionAlertTime string `json:"emission_alert_time"`
	EmissionBeginTime string `json:"emission_begin_time"`
	EmissionEndTime   string `json:"emission_end_time"`
}

func readConfigs(filename string) (*Config, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var cfg Config
	err = json.Unmarshal([]byte(byteValue), &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

var (
	cfg *Config
	err error
	bot *tgbotapi.BotAPI
)

func emissionAlert() {
	logrus.Infof("Emission alert")

	msg := tgbotapi.NewAudioUpload(cfg.TargetChatID, "./audio/emissionAlert.mp3")
	msg.Caption = "Сталкеры, внимание! Выброс начнется с минуты на минуту! Ищите глубокую нору, если жить охота."

	bot.Send(msg)
}

func emissionBegin() {
	logrus.Infof("Emission started")

	msg := tgbotapi.NewAudioUpload(cfg.TargetChatID, "./audio/emission.mp3")
	msg.Caption = "Начался ВЫБРОС!"

	bot.Send(msg)
}

func emissionEnd() {
	logrus.Infof("Emission end")

	msg := tgbotapi.NewAudioUpload(cfg.TargetChatID, "./audio/emissionEnd.mp3")
	msg.Caption = "Ух. Все, ребята, выброс, слава Богу, закончился. Надеюсь, никто не пострадал?"

	bot.Send(msg)
}

func main() {
	//get configs
	cfg, err = readConfigs("./config.json")
	if err != nil {
		logrus.Fatal(err)
	}

	bot, err = tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		logrus.Fatal(err)
	}

	bot.Debug = false
	logrus.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logrus.Fatalf("cannot GetUpdatesChan: %s", err.Error())
	}

	gocron.Every(1).Friday().At(cfg.EmissionAlertTime).Do(emissionAlert)
	gocron.Every(1).Friday().At(cfg.EmissionBeginTime).Do(emissionBegin)
	gocron.Every(1).Friday().At(cfg.EmissionEndTime).Do(emissionEnd)

	go func() {
		<-gocron.Start()
	}()

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		logrus.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

		logrus.Warnf("chat id: %v", update.Message.Chat.ID)
	}
}
