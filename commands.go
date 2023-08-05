package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/garydevenay/go-chatgpt-client"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"
)

var botID int

func chatGPTtext(ctx context.Context, b *bot.Bot, update *models.Update) {

	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	user_id := int(update.Message.From.ID)
	username := update.Message.From.Username
	if !checkBanUtil(username, strconv.Itoa(user_id)) {
		sendMessage(ctx, b, update, "Вы забанены в этом боте.")
		return
	}

	//isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isAdmin := getAdmin(strconv.Itoa(int(user_id)))
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	if !isAdmin {
		isAdmin = getAdmin(update.Message.From.Username)
		if !isAdmin {
			isAdmin = getAdmin(strconv.Itoa(int(update.Message.From.ID)))
		}
	}
	if !isVip {
		isVip = getVIP(update.Message.From.Username)
		if !isVip {
			isVip = getVIP(strconv.Itoa(int(update.Message.From.ID)))
		}
	}

	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), username)
	}

	updateUserVIPorAdmin(update.Message.From.Username, int(update.Message.From.ID))

	cooldown := checkCooldown(user_id, 7)

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
		return
	}

	if !checkUserInGroup(ctx, b, update, user_id) {
		sendMessage(ctx, b, update, "Чтобы пользоваться ботом, пожалуйста, подпишитесь на наш канал @MerrowTalks")
		return
	}

	messageText := strings.Split(update.Message.Text, " ") //get full text with command
	messageText = messageText[1:]                          //split text, remove command
	//replace symbols from string
	oldMessage := ""

	for i := 0; i < len(messageText); i++ {
		oldMessage = oldMessage + " " + messageText[i] //append all words to fullMessage
	}
	fullMessage := strings.Replace(oldMessage, "\u00ad", "", -1)
	fullMessage = strings.Replace(fullMessage, "\u202f", "", -1)

	if !isAdmin && !isVip {
		if utf8.RuneCountInString(fullMessage) < 5 {
			sendMessage(ctx, b, update, "Минимальная длина запроса 5 символов.")
			return
		} else if utf8.RuneCountInString(fullMessage) > 768 {
			sendMessage(ctx, b, update, "Максимальная длина запроса 768 символов.")
			return
		}
	} else if !isAdmin && isVip {
		if utf8.RuneCountInString(fullMessage) < 5 {
			sendMessage(ctx, b, update, "Минимальная длина запроса 5 символов.")
			return
		} else if utf8.RuneCountInString(fullMessage) > 1536 {
			sendMessage(ctx, b, update, "Максимальная длина запроса 1536 символов.")
			return
		}
	}

	var messages []chatgpt.Message

	allMessages := getMessages(user_id)

	if !isAdmin && !isVip {
		if isTotalLengthGreaterThan(allMessages, 4096) { //len(allMessages)/2 > 6 {
			removeMessages(user_id)
			if checkGleb(user_id) {
				updateGleb(user_id)
			}
			if checkDarkGPT(user_id) {
				updateDarkGPT(user_id)
			}
		}
	} else if !isAdmin && isVip {
		if len(allMessages)/2 >= 32 {
			if isTotalLengthGreaterThan(allMessages, 16384) {
				removeMessages(user_id)
				if checkGleb(user_id) {
					updateGleb(user_id)
				}
				if checkDarkGPT(user_id) {
					updateDarkGPT(user_id)
				}
			}
		}
	} else if isAdmin {
		if len(allMessages)/2 >= 48 {
			if isTotalLengthGreaterThan(allMessages, 16384) {
				removeMessages(user_id)
				if checkGleb(user_id) {
					updateGleb(user_id)
				}
				if checkDarkGPT(user_id) {
					updateDarkGPT(user_id)
				}
			}
		}
	}

	//fmt.Println(update.Message.From.Username, allMessages)

	i := 0
	if checkGleb(int(update.Message.From.ID)) {
		if len(messages) == 1 {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[0]})
			i++
		}
	}

	if checkDarkGPT(user_id) {
		if len(messages) == 1 {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[0]})
			i++
		}
	}

	for ; i < len(allMessages); i++ {
		if i%2 == 0 {
			messages = append(messages, chatgpt.Message{Role: "user", Content: allMessages[i]})
		} else {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[i]})
		}
	}
	messages = append(messages, chatgpt.Message{Role: "user", Content: fullMessage})

	for {
		//randomNumber, _ := random.GetInt(len(config.api_keys))

		tempToken := getToken()
		if len(tempToken) == 0 {
			sendMessage(ctx, b, update, "Закончились ключи openai!")
			return
		}
		client := chatgpt.NewClient(tempToken) //(config.api_keys[randomNumber])

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: "typing",
		})

		response, err := client.SendMessage("gpt-3.5-turbo", messages)
		if err != nil {
			if slices.Contains([]string(strings.Split(err.Error(), " ")), "400") {
				sendMessage(ctx, b, update, "Произошла ошибка при генерации ответа!\n"+string(err.Error()))
				return
			} else if slices.Contains([]string(strings.Split(err.Error(), " ")), "429") {
				fmt.Println(err)
				go checkTokenValid(tempToken, ctx, b)
				time.Sleep(3 * time.Second)
			}
		}
		errcheck(err, "commands.go")

		if !sendMessage(ctx, b, update, response) {
			time.Sleep(1500)
			continue
		}
		if checkSavingMessages(user_id) {
			var messagesToSave []string

			messagesToSave = append(messagesToSave, fullMessage)
			messagesToSave = append(messagesToSave, response)

			saveMessages(user_id, messagesToSave)
		}

		if checkGleb(user_id) && !checkSavingMessages(user_id) {
			var messagesToSave []string

			updateGleb(user_id)

			messagesToSave = append(messagesToSave, fullMessage)
			messagesToSave = append(messagesToSave, response)

			saveMessages(user_id, messagesToSave)
		}

		updateTextCompletions(user_id)

		if update.Message.Chat.ID < 0 {
			updateTextCompletions(int(update.Message.Chat.ID))
		}

		return
	}
}

func img(ctx context.Context, b *bot.Bot, update *models.Update) {

	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	cooldown := checkCooldown(user_id, 20)

	updateUserVIPorAdmin(update.Message.From.Username, int(update.Message.From.ID))

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
		return
	}

	if !checkUserInGroup(ctx, b, update, user_id) {
		sendMessage(ctx, b, update, "Чтобы пользоваться ботом, пожалуйста, подпишитесь на наш канал @MerrowTalks")
		return
	}

	messageText := strings.Split(update.Message.Text, " ") //get full text with command
	messageText = messageText[1:]                          //split text, remove command
	fullMessage := ""
	for i := 0; i < len(messageText); i++ {
		fullMessage = fullMessage + " " + messageText[i] //append all words to fullMessage
	}

	if utf8.RuneCountInString(fullMessage) < 4 {
		sendMessage(ctx, b, update, "Минимальная длина запроса 4 символа.")
		return
	} else if isVip && utf8.RuneCountInString(fullMessage) > 128 {
		sendMessage(ctx, b, update, "Максимальная длина запроса для генерации изображения 128 символов.")
		return
	} else if utf8.RuneCountInString(fullMessage) > 64 && !isAdmin {
		sendMessage(ctx, b, update, "Максимальная длина запроса для генерации изображения 64 символа.")
		return
	}

	for {
		// randomNumber, _ := random.GetInt(len(config.api_keys))
		// client := openai.NewClient(config.api_keys[randomNumber])
		tempToken := getToken()
		if len(tempToken) == 0 {
			sendMessage(ctx, b, update, "Закончились ключи openai!")
			return
		}
		client := openai.NewClient(tempToken)

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: "upload_photo",
		})

		reqUrl := openai.ImageRequest{
			Prompt:         fullMessage,
			Size:           openai.CreateImageSize1024x1024,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			N:              1,
		}
		respUrl, err := client.CreateImage(context.Background(), reqUrl)
		if err != nil {
			if string(err.Error()) == "error, status code: 400, message: Billing hard limit has been reached" {
				continue
			} else if slices.Contains([]string(strings.Split(err.Error(), " ")), "rejected") {
				sendMessage(ctx, b, update, "Я не буду это генерировать.")
				return
			} else if slices.Contains([]string(strings.Split(err.Error(), " ")), "429") {
				fmt.Println(err)
				go checkTokenValid(tempToken, ctx, b)
				time.Sleep(3 * time.Second)
			}
			sendMessage(ctx, b, update, "Произошла ошибка при генерации изображения. Пожалуйста, попробуйте ещё раз.\n"+err.Error())
			return
		}
		respUrl2 := respUrl.Data[0]
		sendImage(ctx, b, update, fullMessage, respUrl2.URL)
		break
	}

	updateImg(user_id)

	if update.Message.Chat.ID < 0 {
		updateImg(int(update.Message.Chat.ID))
	}
}

func chatGPTtextUser(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		if update.Message.Chat.ID < 0 {
			if update.Message.ReplyToMessage != nil {
				if update.Message.ReplyToMessage.From.ID != int64(botID) {
					return
				}
			} else {
				return
			}
		}
	} else {
		return
	}

	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	user_id := int(update.Message.From.ID)

	updateUserVIPorAdmin(update.Message.From.Username, int(update.Message.From.ID))

	var ReplyMessageFromBot string

	if update.Message.ReplyToMessage != nil {
		ReplyMessageFromBot = update.Message.ReplyToMessage.Text
	}

	if !checkBanUtil(update.Message.From.Username, strconv.Itoa(user_id)) {
		sendMessage(ctx, b, update, "Вы забанены в этом боте.")
		return
	}

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	if !isAdmin {
		isAdmin = getAdmin(update.Message.From.Username)
		if !isAdmin {
			isAdmin = getAdmin(strconv.Itoa(int(update.Message.From.ID)))
		}
	}
	if !isVip {
		isVip = getVIP(update.Message.From.Username)
		if !isVip {
			isVip = getVIP(strconv.Itoa(int(update.Message.From.ID)))
		}
	}

	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	cooldown := checkCooldown(user_id, 7)

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
		return
	}

	if !checkUserInGroup(ctx, b, update, user_id) {
		//sendMessage(ctx, b, update, "Чтобы пользоваться ботом, пожалуйста, подпишитесь на наш канал @MerrowTalks")
		return
	}

	fullMessage := update.Message.Text

	fullMessage = strings.Replace(fullMessage, "\\u00ad", "", -1)

	if !isAdmin && !isVip {
		if utf8.RuneCountInString(fullMessage) < 1 {
			return
		} else if utf8.RuneCountInString(fullMessage) < 5 {
			sendMessage(ctx, b, update, "Минимальная длина запроса 5 символов.")
			return
		} else if utf8.RuneCountInString(fullMessage) > 768 {
			sendMessage(ctx, b, update, "Максимальная длина запроса 768 символов.")
			return
		}
	} else if !isAdmin && isVip {
		if utf8.RuneCountInString(fullMessage) < 1 {
			return
		} else if utf8.RuneCountInString(fullMessage) < 5 {
			sendMessage(ctx, b, update, "Минимальная длина запроса 5 символов.")
			return
		} else if utf8.RuneCountInString(fullMessage) > 1536 {
			sendMessage(ctx, b, update, "Максимальная длина запроса 1536 символов.")
			return
		}
	}

	var messages []chatgpt.Message

	allMessages := getMessages(user_id)

	if !isAdmin && !isVip {
		if isTotalLengthGreaterThan(allMessages, 4096) { //len(allMessages)/2 > 6 {
			removeMessages(user_id)
			if checkGleb(user_id) {
				updateGleb(user_id)
			}
			if checkDarkGPT(user_id) {
				updateDarkGPT(user_id)
			}
		}
	} else if !isAdmin && isVip {
		if len(allMessages)/2 >= 32 {
			if isTotalLengthGreaterThan(allMessages, 16384) {
				removeMessages(user_id)
				if checkGleb(user_id) {
					updateGleb(user_id)
				}
				if checkDarkGPT(user_id) {
					updateDarkGPT(user_id)
				}
			}
		}
	} else if isAdmin {
		if len(allMessages)/2 >= 48 {
			if isTotalLengthGreaterThan(allMessages, 16384) {
				removeMessages(user_id)
				if checkGleb(user_id) {
					updateGleb(user_id)
				}
				if checkDarkGPT(user_id) {
					updateDarkGPT(user_id)
				}
			}
		}
	}

	i := 0
	if checkGleb(int(update.Message.From.ID)) {
		if len(messages) == 1 {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[0]})
			i++
		}
	}

	if checkDarkGPT(user_id) {
		if len(messages) == 1 {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[0]})
			i++
		}
	}

	if utf8.RuneCountInString(ReplyMessageFromBot) > 0 && !checkGleb(user_id) && !checkSavingMessages(user_id) {
		messages = append(messages, chatgpt.Message{Role: "system", Content: ReplyMessageFromBot})
		i++
	}

	for ; i < len(allMessages); i++ {
		if i%2 == 0 {
			messages = append(messages, chatgpt.Message{Role: "user", Content: allMessages[i]})
		} else {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[i]})
		}
	}
	messages = append(messages, chatgpt.Message{Role: "user", Content: fullMessage})

	temp := 0
	for {

		tempToken := getToken()
		if len(tempToken) == 0 {
			sendMessage(ctx, b, update, "Закончились ключи openai!")
			return
		}
		client := chatgpt.NewClient(tempToken)

		// randomNumber, _ := random.GetInt(len(config.api_keys))
		// client := chatgpt.NewClient(config.api_keys[randomNumber])

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: "typing",
		})

		response, err := client.SendMessage("gpt-3.5-turbo", messages)
		if err != nil {
			if slices.Contains([]string(strings.Split(err.Error(), " ")), "400") {
				sendMessage(ctx, b, update, "Произошла ошибка при генерации ответа!\n"+string(err.Error()))
				return
			} else if slices.Contains([]string(strings.Split(err.Error(), " ")), "billing") || slices.Contains([]string(strings.Split(err.Error(), " ")), "429") {
				fmt.Println(err)
				go checkTokenValid(tempToken, ctx, b)
				time.Sleep(3 * time.Second)
			}
		}
		errcheck(err, "commands.go")

		if update.Message.Chat.ID > 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      response,
				ParseMode: "MarkDown",
			})
			if err != nil {
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   response,
				})
				if err != nil {
					time.Sleep(time.Second * 2)
				}
				continue
			}
		} else {
			if !sendMessage(ctx, b, update, response) {
				time.Sleep(1500)
				temp++
				if temp > 3 {
					sendMessage(ctx, b, update, "Произошла ошибка при генерации ответа.")
				}
				continue
			}
		}

		if checkSavingMessages(user_id) {
			var messagesToSave []string

			messagesToSave = append(messagesToSave, fullMessage)
			messagesToSave = append(messagesToSave, response)

			saveMessages(user_id, messagesToSave)
		}

		if checkGleb(user_id) && !checkSavingMessages(user_id) {
			var messagesToSave []string

			updateGleb(user_id)

			messagesToSave = append(messagesToSave, fullMessage)
			messagesToSave = append(messagesToSave, response)

			//fmt.Println(getMessages(user_id))

			saveMessages(user_id, messagesToSave)

			//fmt.Println(getMessages(user_id))
		}
		updateTextCompletions(user_id)

		if update.Message.Chat.ID < 0 {
			updateTextCompletions(int(update.Message.Chat.ID))
		}

		return
	}
}

func saveMessagesHandlerForUser(ctx context.Context, b *bot.Bot, update *models.Update) {

	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	user_id := update.Message.From.ID
	if setSaveMessages(int(user_id)) {
		sendMessage(ctx, b, update, "Теперь я сохраняю Ваши сообщения в истории.")
	} else {
		sendMessage(ctx, b, update, "Я больше не буду сохранять Ваши сообщения.")
	}
}

func broadcast(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		rand.Seed(time.Now().UnixNano())
		message := stopMessage[rand.Intn(len(stopMessage))]
		sendMessage(ctx, b, update, message)
		return
	}

	messageText := strings.Split(update.Message.Text, " ") //get full text with command
	messageText = messageText[1:]                          //split text, remove command
	fullMessage := ""
	for i := 0; i < len(messageText); i++ {
		fullMessage = fullMessage + " " + messageText[i] //append all words to fullMessage
	}

	if utf8.RuneCountInString(fullMessage) < 10 {
		sendMessage(ctx, b, update, "Слишком маленькое сообщение.")
		return
	}

	broadcastUtil(fullMessage, ctx, b)
}

func getStats(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		rand.Seed(time.Now().UnixNano())
		message := stopMessage[rand.Intn(len(stopMessage))]
		sendMessage(ctx, b, update, message)
		return
	}
	getStatsUtil(ctx, b, update)
}

func ban(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		rand.Seed(time.Now().UnixNano())
		message := stopMessage[rand.Intn(len(stopMessage))]
		sendMessage(ctx, b, update, message)
		return
	}

	message := strings.Split(update.Message.Text, " ")

	if len(message) == 1 {
		sendMessage(ctx, b, update, "Айди-то укажи")
		return
	}
	err := banUtil(message[1])
	if err != nil {
		fmt.Println("пиздец")
	}

	errer := sendMessage(ctx, b, update, "Забанил пользователя "+message[1])
	if !errer {
		fmt.Println("пиздец")
	}

	fmt.Println("норм, забанил " + message[1])
}

func unban(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		rand.Seed(time.Now().UnixNano())
		message := stopMessage[rand.Intn(len(stopMessage))]
		sendMessage(ctx, b, update, message)
		return
	}

	message := strings.Split(update.Message.Text, " ")

	if len(message) == 1 {
		sendMessage(ctx, b, update, "Айди-то укажи")
		return
	}
	unbanUtil(message[1])

	errer := sendMessage(ctx, b, update, "Разбанил пользователя "+message[1])
	if !errer {
		fmt.Println("пиздец")
	}
}

func settings(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	var saving string
	var gleb string
	var DarkGPT string

	if checkSavingMessages(int(update.Message.From.ID)) {
		saving = "✅ Сохранение истории сообщений"
	} else {
		saving = "❌ Сохранение истории сообщений"
	}

	if checkGleb(int(update.Message.From.ID)) {
		if checkDarkGPT(int(update.Message.From.ID)) {
			gleb = "✅ Режим Глеба"
			setDarkGPTMode(int(update.Message.From.ID))
		} else {
			gleb = "✅ Режим Глеба"
		}
	} else {
		gleb = "❌ Режим Глеба"
	}

	if checkDarkGPT(int(update.Message.From.ID)) {
		if checkGleb(int(update.Message.From.ID)) {
			DarkGPT = "✅ Режим DarkGPT"
			setGlebMode(int(update.Message.From.ID))
		} else {
			DarkGPT = "✅ Режим DarkGPT"
		}

	} else {
		DarkGPT = "❌ Режим DarkGPT"
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: saving, CallbackData: "saving_Messages"},
			}, {
				{Text: gleb, CallbackData: "usingGleb"},
			}, {
				{Text: DarkGPT, CallbackData: "darkgptmode"},
			},
		},
	}

	if update.Message.Chat.ID < 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.From.ID,
			Text:        "Выберите необходимое действие.",
			ReplyMarkup: kb,
		})

		sendMessage(ctx, b, update, "Отправил Вам в личку меню настроек.")
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите необходимое действие.",
			ReplyMarkup: kb,
		})
	}
}

func help(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))
	sendMessage(ctx, b, update, message)
}

func start(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Настройки", CallbackData: "settingsbota"},
			}, {
				{Text: "Команды бота", CallbackData: "helpmeplease"},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:        "Привет! Я - бот с ChatGPT от мерровичка! Выбери дальнейшее действие.",
		ChatID:      update.Message.Chat.ID,
		ReplyMarkup: kb,
	})
}

func givevip(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		rand.Seed(time.Now().UnixNano())
		message := stopMessage[rand.Intn(len(stopMessage))]
		sendMessage(ctx, b, update, message)
		return
	}

	if update.Message.ReplyToMessage != nil {
		//giveVIPUtil(strconv.Itoa(int(update.Message.ReplyToMessage.From.ID)))
		giveVIPUtil(update.Message.ReplyToMessage.From.Username)
		giveVIPUtil(strconv.Itoa(int(update.Message.ReplyToMessage.From.ID)))
		var tempid string
		tempid = update.Message.ReplyToMessage.From.Username
		if len(tempid) != 0 {
			sendMessage(ctx, b, update, "Выдал статус Vip пользователю @"+tempid)
			sendbroadcast(ctx, b, int(getIDByUsername(tempid)), "Вам выдан *вечный* статус \"*VIP*\" пользователем админ")
		} else {
			sendMessage(ctx, b, update, "Выдал статус Vip пользователю *"+update.Message.ReplyToMessage.From.FirstName+"*")
			sendbroadcast(ctx, b, int(update.Message.ReplyToMessage.From.ID), "Вам выдан *вечный* статус \"*VIP*\" пользователем админ")
		}
		return
	}

	message := strings.Split(update.Message.Text, " ")

	if len(message) == 1 {
		sendMessage(ctx, b, update, "Айди-то укажи")
		return
	}

	user_id_str := message[1]

	giveVIPUtil(user_id_str)

	sendMessage(ctx, b, update, "Выдал вип пользователю "+user_id_str)
	sendbroadcast(ctx, b, int(getIDByUsername(user_id_str)), "Вам выдан *вечный* статус \"*VIP*\" пользователем админ")

	return
}

func giveadmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		rand.Seed(time.Now().UnixNano())
		message := stopMessage[rand.Intn(len(stopMessage))]
		sendMessage(ctx, b, update, message)
		return
	}

	if update.Message.ReplyToMessage != nil {
		giveAdminUtil(strconv.Itoa(int(update.Message.ReplyToMessage.From.ID)))
		giveAdminUtil(update.Message.ReplyToMessage.From.Username)
		var tempid string
		tempid = update.Message.ReplyToMessage.From.Username
		if len(tempid) != 0 {
			sendMessage(ctx, b, update, "Выдал статус Admin пользователю @"+tempid)
			sendbroadcast(ctx, b, int(getIDByUsername(tempid)), "Вам выдан *вечный* статус \"*Admin*\" пользователем админ")
		} else {
			sendMessage(ctx, b, update, "Выдал статус Admin пользователю *"+update.Message.ReplyToMessage.From.FirstName+"*")
			sendbroadcast(ctx, b, int(update.Message.ReplyToMessage.From.ID), "Вам выдан *вечный* статус \"*Admin*\" пользователем админ")
		}
		return
	}

	message := strings.Split(update.Message.Text, " ")

	if len(message) == 1 {
		sendMessage(ctx, b, update, "Айди-то укажи")
		return
	}

	user_id_str := message[1]

	giveAdminUtil(user_id_str)
	sendMessage(ctx, b, update, "Выдал статус Admin пользователю @"+user_id_str)
	sendbroadcast(ctx, b, int(getIDByUsername(user_id_str)), "Вам выдан *вечный* статус \"*Admin*\" пользователем админ")

	return
}

func info(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))
	getUserInfo(ctx, b, update)
}

func stats(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))
	statsUtil(ctx, b, update)
}

func stableDiffusionCommand(ctx context.Context, b *bot.Bot, update *models.Update) {

	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	cooldown := checkCooldown(user_id, 35)

	updateUserVIPorAdmin(update.Message.From.Username, int(update.Message.From.ID))

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
		return
	}

	if !checkUserInGroup(ctx, b, update, user_id) {
		sendMessage(ctx, b, update, "Чтобы пользоваться ботом, пожалуйста, подпишитесь на наш канал @MerrowTalks")
		return
	}

	messageText := strings.Split(update.Message.Text, " ") //get full text with command
	messageText = messageText[1:]                          //split text, remove command
	//replace symbols from string
	oldMessage := ""

	for i := 0; i < len(messageText); i++ {
		oldMessage = oldMessage + " " + messageText[i] //append all words to fullMessage
	}
	fullMessage := strings.Replace(oldMessage, "\u00ad", "", -1)
	fullMessage = strings.Replace(fullMessage, "\u202f", "", -1)

	if utf8.RuneCountInString(fullMessage) < 4 {
		sendMessage(ctx, b, update, "Минимальная длина запроса 4 символа.")
		return
	} else if isVip && utf8.RuneCountInString(fullMessage) > 368 {
		sendMessage(ctx, b, update, "Максимальная длина запроса для генерации изображения 368 символов.")
		return
	} else if utf8.RuneCountInString(fullMessage) > 192 && !isAdmin {
		sendMessage(ctx, b, update, "Максимальная длина запроса для генерации изображения 192 символа.")
		return
	}
	ENMessage := translateToEn(fullMessage)

	//ioutil.WriteFile("test.jpg", stableDiffusionStable(ENMessage), fs.ModeAppend)

	sendImageByte(ctx, b, update, stableDiffusionStable(ENMessage), fullMessage)

	//sendMessage(ctx, b, update, translateToEn(fullMessage))
}

func stableDiffusionAnimeCommand(ctx context.Context, b *bot.Bot, update *models.Update) {

	if len(update.Message.From.Username) > 0 {
		checkTempVIP(update.Message.From.Username)
	}
	checkTempVIP(strconv.Itoa(int(update.Message.From.ID)))

	user_id := int(update.Message.From.ID)
	if int(update.Message.Chat.ID) < 0 {
		checkChatID(int(update.Message.Chat.ID), "")
	} else {
		checkChatID(int(update.Message.Chat.ID), update.Message.From.Username)
	}

	cooldown := checkCooldown(user_id, 35)

	updateUserVIPorAdmin(update.Message.From.Username, int(update.Message.From.ID))

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
		return
	}

	if !checkUserInGroup(ctx, b, update, user_id) {
		sendMessage(ctx, b, update, "Чтобы пользоваться ботом, пожалуйста, подпишитесь на наш канал @MerrowTalks")
		return
	}

	messageText := strings.Split(update.Message.Text, " ") //get full text with command
	messageText = messageText[1:]                          //split text, remove command
	//replace symbols from string
	oldMessage := ""

	for i := 0; i < len(messageText); i++ {
		oldMessage = oldMessage + " " + messageText[i] //append all words to fullMessage
	}
	fullMessage := strings.Replace(oldMessage, "\u00ad", "", -1)
	fullMessage = strings.Replace(fullMessage, "\u202f", "", -1)

	if utf8.RuneCountInString(fullMessage) < 4 {
		sendMessage(ctx, b, update, "Минимальная длина запроса 4 символа.")
		return
	} else if isVip && utf8.RuneCountInString(fullMessage) > 368 {
		sendMessage(ctx, b, update, "Максимальная длина запроса для генерации изображения 368 символов.")
		return
	} else if utf8.RuneCountInString(fullMessage) > 192 && !isAdmin {
		sendMessage(ctx, b, update, "Максимальная длина запроса для генерации изображения 192 символа.")
		return
	}
	ENMessage := translateToEn(fullMessage)

	sendImageByte(ctx, b, update, stableDiffusionAnime(ENMessage), fullMessage)
}

func remvip(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if !getAdmin(update.Message.From.Username) && !getAdmin(strconv.Itoa(int(update.Message.From.ID))) {
		flag := false
		for i := 0; i < len(config.admins); i++ {
			if user_id == config.admins[i] {
				flag = true
			}
		}
		if !flag {
			rand.Seed(time.Now().UnixNano())
			message := stopMessage[rand.Intn(len(stopMessage))]
			sendMessage(ctx, b, update, message)
			return
		}
	}

	if update.Message.ReplyToMessage != nil {
		removeVIPUtil(update.Message.ReplyToMessage.From.Username)
		removeVIPUtil(strconv.Itoa(int(update.Message.ReplyToMessage.From.ID)))
		if len(update.Message.ReplyToMessage.From.Username) != 0 {
			sendMessage(ctx, b, update, "Снял випку у @"+update.Message.ReplyToMessage.From.Username)
		} else {
			sendMessage(ctx, b, update, "Снял випку у "+update.Message.ReplyToMessage.From.FirstName)
		}
	} else {
		text := strings.Split(update.Message.Text, " ")
		if len(text) > 1 {
			removeVIPUtil(text[1])
			sendMessage(ctx, b, update, "Снял випку у "+text[1])
		} else {
			sendMessage(ctx, b, update, "Айди-то укажи, умник")
		}
	}
}

func addtoken(ctx context.Context, b *bot.Bot, update *models.Update) {

	var message []chatgpt.Message

	var tokens []string

	message = append(message, chatgpt.Message{Role: "user", Content: "Hello"})

	tempMessage := strings.Split(update.Message.Text, " ")

	var temptoken string

	for i := 1; i < len(tempMessage); i++ {
		temptoken = tempMessage[i]
		client := chatgpt.NewClient(temptoken)
		_, err := client.SendMessage("gpt-3.5-turbo", message)
		if err != nil {
			continue
		} else {
			tokens = append(tokens, temptoken)
		}
	}

	addedTokens := addToken(tokens)

	sendMessage(ctx, b, update, strconv.Itoa(addedTokens)+" токенов добавлено в базу данных.")
}

func deletetable(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := strings.Split(update.Message.Text, " ")

	chatsDB, err := sql.Open("sqlite3", "chats.db")
	if err != nil {
		log.Println(err)
		return
	}
	defer chatsDB.Close()

	tx, err := chatsDB.BeginTx(context.Background(), nil)

	_, err = tx.Exec("DROP TABLE " + message[1])
	tx.Commit()
	chatsDB.Close()
	return
}

func tempVIP(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	if !getAdmin(update.Message.From.Username) && !getAdmin(strconv.Itoa(int(update.Message.From.ID))) {
		flag := false
		for i := 0; i < len(config.admins); i++ {
			if user_id == config.admins[i] {
				flag = true
			}
		}
		if !flag {
			rand.Seed(time.Now().UnixNano())
			message := stopMessage[rand.Intn(len(stopMessage))]
			sendMessage(ctx, b, update, message)
			return
		}
	}
	message := strings.Split(update.Message.Text, " ")

	if update.Message.ReplyToMessage != nil {
		if len(message) < 2 {
			sendMessage(ctx, b, update, "Введи секунды, идиот")
			return
		}
		seconds, err := strconv.Atoi(message[1])
		if err != nil {
			sendMessage(ctx, b, update, "Введи нормальное количество секунд, дурак")
			return
		}

		giveTempVIP(strconv.Itoa(int(update.Message.ReplyToMessage.From.ID)), seconds)

		if len(update.Message.ReplyToMessage.From.Username) != 0 {
			sendMessage(ctx, b, update, "Выдал випку пользователю @"+update.Message.ReplyToMessage.From.Username+" на "+formatTime(seconds))
			sendbroadcast(ctx, b, int(update.Message.ReplyToMessage.From.ID), "Вам выдан статус \"*VIP*\" пользователем админ на "+formatTime(seconds))
			return

		}
		if err != nil {

		} else {
			sendMessage(ctx, b, update, "Выдал випку пользователю с айди "+strconv.Itoa(int(update.Message.ReplyToMessage.From.ID))+" на "+formatTime(seconds))
			sendbroadcast(ctx, b, int(update.Message.ReplyToMessage.From.ID), "Вам выдан статус \"*VIP*\" пользователем админ на "+formatTime(seconds))
			return
		}
	}

	if len(message) < 3 {
		sendMessage(ctx, b, update, "Ты указал какие-то неверные данные...")
	}

	id := message[1]
	seconds, err := strconv.Atoi(message[2])
	if err != nil {
		sendMessage(ctx, b, update, "Введи нормальное количество секунд")
		return
	}

	giveTempVIP(id, seconds)

	aboba, err := strconv.Atoi(message[1])
	if err != nil {
		sendMessage(ctx, b, update, "Выдал випку пользователю с айди "+message[1]+" на "+formatTime(seconds))
		sendbroadcast(ctx, b, aboba, "Вам выдан статус \"*VIP*\" пользователем админ на "+formatTime(seconds))
	} else {
		sendMessage(ctx, b, update, "Выдал випку пользователю @"+message[1]+" на "+formatTime(seconds))
		sendbroadcast(ctx, b, int(getIDByUsername(message[1])), "Вам выдан статус \"*VIP*\" пользователем админ на "+formatTime(seconds))
	}
}

func checktokens(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMessage(ctx, b, update, "Количество токенов: "+strconv.Itoa(checkTokenscount()))
}

func exit(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !getAdmin(update.Message.From.Username) {
		return
	}
	b.LeaveChat(ctx, &bot.LeaveChatParams{
		ChatID: update.Message.Chat.ID,
	})
}

type DetectedLanguage struct {
	Language string  `json:"language"`
	Score    float64 `json:"score"`
}

type Translation struct {
	Text string `json:"text"`
	To   string `json:"to"`
}

type TranslationResponse struct {
	DetectedLanguage DetectedLanguage `json:"detectedLanguage"`
	Translations     []Translation    `json:"translations"`
}

func translateToEn(message string) string {
	url := "https://microsoft-translator-text.p.rapidapi.com/translate?to%5B0%5D=en&api-version=3.0&profanityAction=NoAction&textType=plain"
	fmt.Println(message)
	payload := strings.NewReader(fmt.Sprintf(`[{ "Text": "%v" }]`, message))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-RapidAPI-Key", "4cb882adf7msh2daff29f772b3b8p1c5802jsn171b9a879d21")
	req.Header.Add("X-RapidAPI-Host", "microsoft-translator-text.p.rapidapi.com")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var translations []TranslationResponse
	if err := json.Unmarshal(body, &translations); err != nil {
		// обработка ошибки
		return "zukati"
	}
	return string(translations[0].Translations[0].Text)
}

func stableDiffusionStable(prompt string) []byte {

	//temp := prompt

	url := `https://image-diffusion.p.rapidapi.com/image/stable/diffusion?prompt=` + prompt[1:]

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", "4cb882adf7msh2daff29f772b3b8p1c5802jsn171b9a879d21")
	req.Header.Add("X-RapidAPI-Host", "image-diffusion.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return body
}

func stableDiffusionAnime(prompt string) []byte {

	//temp := prompt

	url := "https://image-diffusion.p.rapidapi.com/image/anime/diffusion?prompt=" + prompt[1:]

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", "4cb882adf7msh2daff29f772b3b8p1c5802jsn171b9a879d21")
	req.Header.Add("X-RapidAPI-Host", "image-diffusion.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res.Header)

	return body
}
