package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/kn9ka/fundbot-go/services/alphaVantage"
	"github.com/kn9ka/fundbot-go/services/corona"
	"github.com/kn9ka/fundbot-go/services/sheets"
	"github.com/kn9ka/fundbot-go/services/unistream"
	"log"
	"os"
	"strconv"
	"strings"
)

func initialize(path string) {
	if path == "" {
		path = ".env-dev"
	}

	if err := godotenv.Load(path); err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Printf("%v config initialize!", path)
}

func main() {
	initialize(".env")
	tgApiKey := os.Getenv("BOT_TOKEN")
	sheetsClient := sheets.NewService()

	myCommands := map[string]tgbotapi.BotCommand{
		"start": {Command: "/start", Description: "Список доступных команд"},
		"list":  {Command: "/list", Description: "Список долгов"},
		"rates": {Command: "/list", Description: "Курсы валют"},
	}

	bot, err := tgbotapi.NewBotAPI(tgApiKey)

	if err != nil {
		log.Fatalf(" Unable to create telegram bot: %v", err)
	}

	tgbotapi.NewSetMyCommands(myCommands["start"], myCommands["list"], myCommands["rates"])

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		msgId := update.Message.MessageID
		chatId := update.Message.Chat.ID

		// ignore all replies
		if update.Message.ReplyToMessage != nil {
			continue
		}

		if !update.Message.IsCommand() {
			msgText := update.Message.Text
			parts := strings.Split(msgText, " ")

			var amount float64 = 0
			if len(parts) >= 1 {
				s := strings.Replace(parts[0], ",", ".", -1)
				amount, _ = strconv.ParseFloat(s, 64)
			} else {
				continue
			}

			var reason = ""
			if len(parts) >= 2 {
				reason = strings.Join(parts[1:], " ")
			}

			values := [][]interface{}{
				{
					msgId,
					amount,
					reason,
					"",
					update.Message.Date,
					update.Message.From.UserName,
					true,
				},
			}

			res := sheetsClient.Write(values)
			msg := tgbotapi.NewMessage(chatId, "")

			if res {
				msg.Text = fmt.Sprintf("Сохранил: %.5f %s", amount, reason)
			} else {
				msg.Text = "При сохранении возникла ошибка"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Fatalf("Unable to send bot message from basic input: %v", err)
			}
			continue
		}

		currentBotCommand := update.Message.Command()
		msg := tgbotapi.NewMessage(chatId, "")
		typingMsg := tgbotapi.NewChatAction(chatId, tgbotapi.ChatTyping)

		_, _ = bot.Send(typingMsg)

		switch currentBotCommand {
		case "start":
			msg.Text = "/list - for list active debts \n/rates - for exchange RUB => USD/EUR/GEL rates"

		case "rates":
			currencies := []string{"USD", "GEL", "EUR"}
			officialRates := alphaVantage.GetRates()
			unistreamRates := unistream.GetRates()
			coronaRates := corona.GetRates()

			msg.ParseMode = "HTML"
			msg.Text = ""

			for _, currency := range currencies {
				msg.Text += fmt.Sprintf("<b>[%s]</b>\n", currency)

				if officialRate, ok := officialRates[currency]; ok {
					msg.Text += fmt.Sprintf("  official rate: %s\n", officialRate)
				}

				if unistreamRate, ok := unistreamRates[currency]; ok {
					msg.Text += fmt.Sprintf("  unistream: %s\n", unistreamRate)
				}

				if coronaRate, ok := coronaRates[currency]; ok {
					msg.Text += fmt.Sprintf("  corona: %s\n", coronaRate)
				}

				msg.Text += "\n"
			}

		case "list":
			exp := sheetsClient.LoadTotalByUsers(true)
			str := "Ничего не найдено"

			for _, row := range exp {
				str = fmt.Sprintf("<b>@%s</b>: ", row.Name) + fmt.Sprintf("%.2f \n", row.Total)
			}

			msg.ParseMode = "HTML"
			msg.Text = str
		}

		if _, err := bot.Send(msg); err != nil {
			log.Fatalf("Unable to send bot message after command: %v", err)
		}
	}
}
