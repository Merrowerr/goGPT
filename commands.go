package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/garydevenay/go-chatgpt-client"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mazen160/go-random"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"
)

var botID int

func chatGPTtext(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	username := update.Message.From.Username
	if !checkBanUtil(username, strconv.Itoa(user_id)) {
		sendMessage(ctx, b, update, "Вы забанены в этом боте.")
		return
	}

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	checkChatID(int(update.Message.Chat.ID))

	cooldown := checkCooldown(user_id, 7)

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
		return
	}

	checkUserInGroup(ctx, b, update, user_id)

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
		} else if utf8.RuneCountInString(fullMessage) > 512 {
			sendMessage(ctx, b, update, "Максимальная длина запроса 512 символов.")
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
		if len(allMessages) > 6 {
			removeMessages(user_id)
			if checkGleb(user_id) {
				updateGleb(user_id)
			}
		}
	} else if !isAdmin && isVip {
		if len(allMessages) >= 16 {
			removeMessages(user_id)
			if checkGleb(user_id) {
				updateGleb(user_id)
			}
		}
	} else if isAdmin {
		if len(allMessages) >= 20 {
			removeMessages(user_id)
			if checkGleb(user_id) {
				updateGleb(user_id)
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

	for ; i < len(allMessages); i++ {
		if i%2 == 0 {
			messages = append(messages, chatgpt.Message{Role: "user", Content: allMessages[i]})
		} else {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[i]})
		}
	}
	messages = append(messages, chatgpt.Message{Role: "user", Content: fullMessage})

	for {
		randomNumber, _ := random.GetInt(len(config.api_keys))

		client := chatgpt.NewClient(config.api_keys[randomNumber])

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: "typing",
		})

		response, err := client.SendMessage("gpt-3.5-turbo", messages)
		if err != nil {
			if slices.Contains([]string(strings.Split(err.Error(), " ")), "400") {
				sendMessage(ctx, b, update, "Произошла ошибка при генерации ответа!\n"+string(err.Error()))
				return
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

		return
	}
}

func img(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	checkChatID(int(update.Message.Chat.ID))

	cooldown := checkCooldown(user_id, 20)

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
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
		randomNumber, _ := random.GetInt(len(config.api_keys))
		client := openai.NewClient(config.api_keys[randomNumber])

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
			}
			sendMessage(ctx, b, update, "Произошла ошибка при генерации изображения. Пожалуйста, попробуйте ещё раз.\n"+err.Error())
			return
		}
		respUrl2 := respUrl.Data[0]
		sendImage(ctx, b, update, fullMessage, respUrl2.URL)
		break
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

	user_id := int(update.Message.From.ID)

	var ReplyMessageFromBot string

	if update.Message.ReplyToMessage != nil {
		ReplyMessageFromBot = update.Message.ReplyToMessage.Text
		//fmt.Println(ReplyMessageFromBot)
	}

	if !checkBanUtil(update.Message.From.Username, strconv.Itoa(user_id)) {
		sendMessage(ctx, b, update, "Вы забанены в этом боте.")
		return
	}

	isAdmin := checkUserInAdminsOrVIP(user_id, "admins")
	isVip := checkUserInAdminsOrVIP(user_id, "VIP")

	checkChatID(int(update.Message.Chat.ID))

	cooldown := checkCooldown(user_id, 7)

	if cooldown != 0 && !isVip && !isAdmin {
		sendMessage(ctx, b, update, "Пожалуйста, подождите ещё "+strconv.Itoa(cooldown)+" секунд.")
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
		} else if utf8.RuneCountInString(fullMessage) > 512 {
			sendMessage(ctx, b, update, "Максимальная длина запроса 512 символов.")
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

	//fmt.Println(update.Message.From.Username, allMessages)

	if !isAdmin && !isVip {
		if len(allMessages) > 6 {
			removeMessages(user_id)
		}
	} else if !isAdmin && isVip {
		if len(allMessages) >= 16 {
			removeMessages(user_id)
		}
	} else if isAdmin {
		if len(allMessages) >= 20 {
			removeMessages(user_id)
		}
	}

	i := 0
	if checkGleb(int(update.Message.From.ID)) {
		if len(messages) == 1 {
			messages = append(messages, chatgpt.Message{Role: "system", Content: allMessages[0]})
			i++
		}
	}

	if utf8.RuneCountInString(ReplyMessageFromBot) > 0 && !checkGleb(user_id) {
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

	for {
		randomNumber, _ := random.GetInt(len(config.api_keys))

		client := chatgpt.NewClient(config.api_keys[randomNumber])

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: "typing",
		})

		response, err := client.SendMessage("gpt-3.5-turbo", messages)
		if err != nil {
			if slices.Contains([]string(strings.Split(err.Error(), " ")), "400") {
				sendMessage(ctx, b, update, "Произошла ошибка при генерации ответа!\n"+string(err.Error()))
				return
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

		return
	}
}

func saveMessagesHandlerForUser(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := update.Message.From.ID
	if setSaveMessages(int(user_id)) {
		sendMessage(ctx, b, update, "Теперь я сохраняю Ваши сообщения в истории.")
	} else {
		sendMessage(ctx, b, update, "Я больше не буду сохранять Ваши сообщения.")
	}
}

func broadcast(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	checkChatID(int(update.Message.Chat.ID))

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		sendMessage(ctx, b, update, "Не, ты че-то попутал, дружище. Тебе сюда нельзя.")
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
	checkChatID(int(update.Message.Chat.ID))

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		sendMessage(ctx, b, update, "Не, ты че-то попутал, дружище. Тебе сюда нельзя.")
		return
	}
	getStatsUtil(ctx, b, update)
}

func ban(ctx context.Context, b *bot.Bot, update *models.Update) {
	user_id := int(update.Message.From.ID)
	checkChatID(int(update.Message.Chat.ID))

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		sendMessage(ctx, b, update, "Не, ты че-то попутал, дружище. Тебе сюда нельзя.")
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
	checkChatID(int(update.Message.Chat.ID))

	flag := false
	for i := 0; i < len(config.admins); i++ {
		if user_id == config.admins[i] {
			flag = true
		}
	}
	if !flag {
		sendMessage(ctx, b, update, "Не, ты че-то попутал, дружище. Тебе сюда нельзя.")
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

	var saving string
	var gleb string

	if checkSavingMessages(int(update.Message.From.ID)) {
		saving = "✅ Сохранение истории сообщений"
	} else {
		saving = "❌ Сохранение истории сообщений"
	}

	if checkGleb(int(update.Message.From.ID)) {
		gleb = "✅ Режим Глеба"
	} else {
		gleb = "❌ Режим Глеба"
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: saving, CallbackData: "saving_Messages"},
			}, {
				{Text: gleb, CallbackData: "usingGleb"},
			},
		},
	}

	if update.Message.Chat.ID < 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           update.Message.Chat.ID,
			Text:             "Выберите необходимое действие.",
			ReplyToMessageID: update.Message.ID,
			ReplyMarkup:      kb,
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите необходимое действие.",
			ReplyMarkup: kb,
		})
	}
}

func help(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := "справочник по боту chatGPT от мерровичка:\nКоманды:\n/chat ВАШ_ЗАПРОС - отправляет Ваше сообщение на сервера openAI.\n/gpt ВАШ_ЗАПРОС - делает абсолютно то же самое, что и команда /chat.\n/ai ВАШ_ЗАПРОС - синоним команды /gpt\n/settings - настройки бота\n/help - Вы здесь!"
	sendMessage(ctx, b, update, message)
}

func start(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMessage(ctx, b, update, "Привет! Я - бот с ChatGPT! Используй команду /help, чтобы узнать подробнее о командах.")
}
