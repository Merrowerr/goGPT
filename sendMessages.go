package main

import (
	"bytes"
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func sendMessage(ctx context.Context, b *bot.Bot, update *models.Update, message string) bool {
	flag := true
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:           update.Message.Chat.ID,
		Text:             message,
		ReplyToMessageID: update.Message.ID,
		ParseMode:        "MarkDown",
	})
	if err != nil {
		//log.Println(err)
		err := sendMessageaeae(ctx, b, update, message)
		if err != nil {
			flag = false
		}
	}
	return flag
}

func sendMessageaeae(ctx context.Context, b *bot.Bot, update *models.Update, message string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:           update.Message.Chat.ID,
		Text:             message,
		ReplyToMessageID: update.Message.ID,
	})
	return err
}

func sendbroadcast(ctx context.Context, b *bot.Bot, chatID int, message string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "MarkDown",
	})
	return err
}

func sendImage(ctx context.Context, b *bot.Bot, update *models.Update, message string, link string) {
	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.ID,
		Caption:          message,
		Photo:            &models.InputFileString{Data: link},
	})
}

func sendMessageCallback(ctx context.Context, b *bot.Bot, update *models.Update, message string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Sender.ID,
		Text:   message,
	})
	if err != nil {
		log.Println(err)
	}
}

func editMessageCallback(ctx context.Context, b *bot.Bot, update *models.Update, checker string) {
	var saving string
	var gleb string
	var DarkGPT string

	if checker == "setGleb" {
		if checkDarkGPT(int(update.CallbackQuery.Sender.ID)) {
			setDarkGPTMode(int(update.CallbackQuery.Sender.ID))
		}
		setGlebMode(int(update.CallbackQuery.Sender.ID))
	}
	if checker == "setDarkGPT" {
		if checkGleb(int(update.CallbackQuery.Sender.ID)) {
			setGlebMode(int(update.CallbackQuery.Sender.ID))
		}
		setDarkGPTMode(int(update.CallbackQuery.Sender.ID))
	}
	if checker == "saving_Messages" {
		setSaveMessages(int(update.CallbackQuery.Sender.ID))
		if checkDarkGPT(int(update.CallbackQuery.Sender.ID)) {
			updateDarkGPT(int(update.CallbackQuery.Sender.ID))
		}
		if checkGleb(int(update.CallbackQuery.Sender.ID)) {
			updateGleb(int(update.CallbackQuery.Sender.ID))
		}
	}

	if checkSavingMessages(int(update.CallbackQuery.Sender.ID)) {
		saving = "✅ Сохранение истории сообщений"
	} else {
		saving = "❌ Сохранение истории сообщений"
	}

	if checkGleb(int(update.CallbackQuery.Sender.ID)) {
		gleb = "✅ Режим Глеба"
	} else {
		gleb = "❌ Режим Глеба"
	}

	if checkDarkGPT(int(update.CallbackQuery.Sender.ID)) {
		DarkGPT = "✅ Режим DarkGPT"
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
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.ID,
		Text:        "Выберите необходимое действие.",
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println(err)
	}
}

func sendImageByte(ctx context.Context, b *bot.Bot, update *models.Update, bytess []byte, prompt string) error {
	_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:           update.Message.Chat.ID,
		Caption:          prompt,
		ReplyToMessageID: update.Message.ID,
		Photo:            &models.InputFileUpload{Filename: "facebook.png", Data: bytes.NewReader(bytess)},
	})
	return err
}
