// how to write telegram bot on golang with library github.com/go-telegram/bot?
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithCallbackQueryDataHandler("usingGleb", bot.MatchTypePrefix, mainSettings),
		bot.WithCallbackQueryDataHandler("saving", bot.MatchTypePrefix, mainSettings),
		bot.WithDefaultHandler(chatGPTtextUser),
	}

	botinok, err := bot.New(config.token, opts...)
	errcheck(err, "bot.go")

	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/chat", bot.MatchTypePrefix, chatGPTtext)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/ai", bot.MatchTypePrefix, chatGPTtext)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/gpt", bot.MatchTypePrefix, chatGPTtext)

	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/savemessages", bot.MatchTypePrefix, saveMessagesHandlerForUser)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/broadcast", bot.MatchTypePrefix, broadcast)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/stats", bot.MatchTypePrefix, getStats)

	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/img", bot.MatchTypePrefix, img)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/ban", bot.MatchTypePrefix, ban)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/unban", bot.MatchTypePrefix, unban)

	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/settings", bot.MatchTypePrefix, settings)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypePrefix, help)

	botuser, _ := botinok.GetMe(ctx)
	botID = int(botuser.ID)

	botinok.Start(ctx)
}
