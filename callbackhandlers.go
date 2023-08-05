package main

import (
	"context"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func mainSettings(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery.Data == "saving_Messages" {
		// if update.CallbackQuery.Message.Chat.ID < 0 {
		// 	if update.CallbackQuery.Sender.ID != update.CallbackQuery.Message.ReplyToMessage.From.ID {

		// 		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		// 			CallbackQueryID: update.CallbackQuery.ID,
		// 			Text:            "❌ Эта кнопка не для тебя.",
		// 		})
		// 		return
		// 	}
		// }
		editMessageCallback(ctx, b, update, "saving_Messages")
		return
	}

	if update.CallbackQuery.Data == "usingGleb" {
		// if update.CallbackQuery.Message.Chat.ID < 0 {
		// 	if update.CallbackQuery.Sender.ID != update.CallbackQuery.Message.ReplyToMessage.From.ID {
		// 		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		// 			CallbackQueryID: update.CallbackQuery.ID,
		// 			Text:            "❌ Эта кнопка не для тебя.",
		// 		})
		// 		return
		// 	}
		// }
		editMessageCallback(ctx, b, update, "setGleb")
		return
	}
	if update.CallbackQuery.Data == "darkgptmode" {
		vipID := getVIP(strconv.Itoa(int(update.CallbackQuery.Sender.ID)))
		var vipUsername bool
		var adminUsername bool
		if len(update.CallbackQuery.Sender.Username) > 0 {
			vipUsername = getVIP(update.CallbackQuery.Sender.Username)
			adminUsername = getAdmin(update.CallbackQuery.Sender.Username)
		} else {
			vipUsername = false
			adminUsername = false
		}
		adminID := getAdmin(strconv.Itoa(int(update.CallbackQuery.Sender.ID)))

		if !vipID && !vipUsername && !adminID && !adminUsername {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "❌ Недоступно. Купите VIP-подписку.",
			})
			return
		}
		// if update.CallbackQuery.Message.Chat.ID < 0 {
		// 	if update.CallbackQuery.Sender.ID != update.CallbackQuery.Message.ReplyToMessage.From.ID {

		// 		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		// 			CallbackQueryID: update.CallbackQuery.ID,
		// 			Text:            "❌ Эта кнопка не для тебя.",
		// 		})
		// 		return
		// 	}
		// }
		editMessageCallback(ctx, b, update, "setDarkGPT")
		return
	}

	if update.CallbackQuery.Data == "settingsbota" {
		if update.CallbackQuery.Message.Chat.ID < 0 {
			return
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.CallbackQuery.Message.Chat.ID,
				Text:        "Выберите необходимое действие.",
				ReplyMarkup: settingsUtil(update),
			})
			return
		}
	}

	if update.CallbackQuery.Data == "helpmeplease" {
		if update.CallbackQuery.Message.Chat.ID < 0 {
			return
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.CallbackQuery.Message.Chat.ID,
				Text:   message,
			})
			return
		}
	}
}

func settingsUtil(update *models.Update) models.InlineKeyboardMarkup {
	var saving string
	var gleb string
	var DarkGPT string

	if checkSavingMessages(int(update.CallbackQuery.Message.From.ID)) {
		saving = "✅ Сохранение истории сообщений"
	} else {
		saving = "❌ Сохранение истории сообщений"
	}

	if checkGleb(int(update.CallbackQuery.Message.From.ID)) {
		if checkDarkGPT(int(update.CallbackQuery.Message.From.ID)) {
			gleb = "✅ Режим Глеба"
			setDarkGPTMode(int(update.CallbackQuery.Message.From.ID))
		} else {
			gleb = "✅ Режим Глеба"
		}
	} else {
		gleb = "❌ Режим Глеба"
	}

	if checkDarkGPT(int(update.CallbackQuery.Message.From.ID)) {
		if checkGleb(int(update.CallbackQuery.Message.From.ID)) {
			DarkGPT = "✅ Режим DarkGPT"
			setGlebMode(int(update.CallbackQuery.Message.From.ID))
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
	return *kb
}
