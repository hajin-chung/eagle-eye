package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nrednav/cuid2"
)

type Job struct {
	Url    string
	ChatID int64
	Id     string
}

type Result struct {
	ChatID  int64
	Id      string
	Success bool
}

var jobs chan Job
var results chan Result

func InitWorkers(bot *tgbotapi.BotAPI) {
	jobs = make(chan Job)
	results = make(chan Result)
	go Worker()
	go Sender(bot)
}

func Worker() {
	for job := range jobs {
		cmd := exec.Command("python3", "script.py", job.Url, job.Id)
		_, err := cmd.Output()
		if err != nil {
			results <- Result{
				ChatID:  job.ChatID,
				Id:      job.Id,
				Success: false,
			}
		} else {
			results <- Result{
				ChatID:  job.ChatID,
				Id:      job.Id,
				Success: true,
			}
		}
	}
}

func Sender(bot *tgbotapi.BotAPI) {
	for result := range results {
		errorMsg := tgbotapi.NewMessage(result.ChatID, "ERROR")
		if result.Success == false {
			bot.Send(errorMsg)
			continue
		}

		msg, err := GenerateResultMsg(result.Id, result.ChatID)
		if err != nil {
			msg := tgbotapi.NewMessage(result.ChatID, fmt.Sprintf("ERROR: %s", err))
			bot.Send(msg)
		} else {
			bot.Send(msg)
		}
	}
}

func GenerateResultMsg(id string, chatId int64) (tgbotapi.Chattable, error) {
	filename1 := fmt.Sprintf("./result/%s-1.bmp", id)
	filename2 := fmt.Sprintf("./result/%s-2.bmp", id)

	content1, err := os.ReadFile(filename1)
	if err != nil {
		return nil, err
	}
	content2, err := os.ReadFile(filename2)
	if err != nil {
		return nil, err
	}
	file1 := tgbotapi.FileBytes{Name: filename1, Bytes: content1}
	file2 := tgbotapi.FileBytes{Name: filename2, Bytes: content2}

	media1 := tgbotapi.NewInputMediaPhoto(file1)
	media2 := tgbotapi.NewInputMediaPhoto(file2)
	message := tgbotapi.NewMediaGroup(chatId, []interface{}{media1, media2})

	return message, nil
}

func main() {
	token, err := GetToken()
	if err != nil {
		log.Panicf("failed to get token: %s\n", err)
	}
	log.Println("Starting eagle-eye bot")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panicf("wrong token: %s\n", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	InitWorkers(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			url := update.Message.Text
			id := cuid2.Generate()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "job queued! it may take a while")
			bot.Send(msg)

			jobs <- Job{
				Url:    url,
				Id:     id,
				ChatID: update.Message.Chat.ID,
			}
		}
	}
}

func GetToken() (string, error) {
	token, err := os.ReadFile(".token")
	if err != nil {
		return "", err
	}
	return string(token[:len(token)-1]), nil
}
