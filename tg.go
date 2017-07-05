package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nanobox-io/scribble"
	"github.com/sasbury/mini"
)

func main() {

	type UserInfo struct {
		Name     string
		Username string
		ID       int
		Bio      string
	}

	cache, err := scribble.New("cache", nil)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	config, err := mini.LoadConfiguration("config.ini")
	if err != nil {
		log.Print("Config file does not exist")
		return
	}

	bot, err := tgbotapi.NewBotAPI(config.String("token", ""))

	if err != nil {
		log.Panicf("panic in bot %s", err)
		return
	}

	bot.Debug = false

	var ucfg = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 10

	updates, err := bot.GetUpdatesChan(ucfg)

	if err != nil {
		log.Panicf("panic in updates %s", err)
		return
	}

	for update := range updates {
		var message *tgbotapi.Message

		if update.Message != nil {
			message = update.Message
		}

		if message == nil {
			continue
		}

		if message.Text == "" {
			continue
		}

		cmd := strings.SplitN(message.Text, " ", 2)[0]

		if cmd == "/bio" && strings.Count(message.Text, " ") > 0 {
			userinfo := UserInfo{message.From.FirstName, message.From.UserName, message.From.ID, strings.SplitN(message.Text, " ", 2)[1]}
			if err := cache.Write(message.From.UserName, "bio", userinfo); err != nil {
				fmt.Printf("cache err %s\n", err)
				os.Exit(1)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Bio has been updated!")
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		} else if cmd == "/getbio" && strings.Count(message.Text, " ") > 0 {
			records, err := cache.ReadAll(strings.SplitN(message.Text, " ", 2)[1])
			if err != nil {
				log.Printf("ReadAll err: %s\n", err)
				continue
			}

			userinfo := UserInfo{}
			err = json.Unmarshal([]byte(records[0]), &userinfo)
			if err != nil {
				log.Printf("Unmarshal err: %s\n", err)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, userinfo.Bio)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = "Markdown"

			bot.Send(msg)
		}
	}
}
