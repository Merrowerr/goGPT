package main

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func mainSettings(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery.Data == "saving_Messages" {
		if update.CallbackQuery.Message.Chat.ID < 0 {
			if update.CallbackQuery.Sender.ID != update.CallbackQuery.Message.ReplyToMessage.From.ID {

				b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "❌ Эта кнопка не для тебя.",
				})
				return
			}
		}
		editMessageCallback(ctx, b, update, "saving_Messages")
		return
	}

	if update.CallbackQuery.Data == "usingGleb" {
		if update.CallbackQuery.Message.Chat.ID < 0 {
			if update.CallbackQuery.Sender.ID != update.CallbackQuery.Message.ReplyToMessage.From.ID {

				b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "❌ Эта кнопка не для тебя.",
				})
				return
			}
		}
		editMessageCallback(ctx, b, update, "setGleb")
		return
	}
}
