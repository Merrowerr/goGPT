// какую же я херню написал... но оно хотя бы работает.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
)

func main() {
	fmt.Println("Starting...")
	time.Sleep(1 * time.Second)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithCallbackQueryDataHandler("usingGleb", bot.MatchTypePrefix, mainSettings),
		bot.WithCallbackQueryDataHandler("saving", bot.MatchTypePrefix, mainSettings),
		bot.WithCallbackQueryDataHandler("darkgptmode", bot.MatchTypePrefix, mainSettings),
		bot.WithCallbackQueryDataHandler("settingsbota", bot.MatchTypePrefix, mainSettings),
		bot.WithCallbackQueryDataHandler("helpmeplease", bot.MatchTypePrefix, mainSettings),
		bot.WithDefaultHandler(chatGPTtextUser),
	}

	botinok, err := bot.New(config.token, opts...)
	errcheck(err, "bot.go")

	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/chat", bot.MatchTypePrefix, chatGPTtext)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/ai", bot.MatchTypePrefix, chatGPTtext)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/gpt", bot.MatchTypePrefix, chatGPTtext)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/img", bot.MatchTypePrefix, img)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/savemessages", bot.MatchTypePrefix, saveMessagesHandlerForUser)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/settings", bot.MatchTypePrefix, settings)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypePrefix, help)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/info", bot.MatchTypePrefix, info)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/stats", bot.MatchTypePrefix, stats)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, start)

	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/broadcast", bot.MatchTypePrefix, broadcast)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/statistic", bot.MatchTypePrefix, getStats)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/ban", bot.MatchTypePrefix, ban)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/unban", bot.MatchTypePrefix, unban)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/givevip", bot.MatchTypePrefix, givevip)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/giveadmin", bot.MatchTypePrefix, giveadmin)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/remvip", bot.MatchTypePrefix, remvip)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/token", bot.MatchTypePrefix, addtoken)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/sd", bot.MatchTypePrefix, stableDiffusionCommand)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/asd", bot.MatchTypePrefix, stableDiffusionAnimeCommand)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/givetempvip", bot.MatchTypePrefix, tempVIP)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/deletetableaeae", bot.MatchTypePrefix, deletetable)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/checktokensaeae", bot.MatchTypePrefix, checktokens)
	botinok.RegisterHandler(bot.HandlerTypeMessageText, "/exit", bot.MatchTypePrefix, exit)

	// botinok.RegisterHandler(bot.HandlerTypeMessageText, "/sd", bot.MatchTypePrefix, stablediffusion)
	// botinok.RegisterHandler(bot.HandlerTypeMessageText, "/tokensd", bot.MatchTypePrefix, addtokensd)
	// botinok.RegisterHandler(bot.HandlerTypeMessageText, "/nsfw", bot.MatchTypePrefix, nsfw)

	botuser, _ := botinok.GetMe(ctx)
	botID = int(botuser.ID)
	fmt.Println("Started Succesfully!")
	botinok.Start(ctx)
}
