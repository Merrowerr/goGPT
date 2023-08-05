package main

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Config struct {
	token    string
	api_keys []string
	admins   []int
	VIPs     []int
}

var config = Config{
	token:  "TELEGRAM_TOKEN",
	admins: []int{},
}

var message = "справочник по боту chatGPT от мерровичка:\nКоманды:\n/chat ВАШ_ЗАПРОС - отправляет Ваше сообщение на сервера openAI и генерирует ответ.\n/gpt ВАШ_ЗАПРОС - делает абсолютно то же самое, что и команда /chat.\n/ai ВАШ_ЗАПРОС - синоним команды /gpt\n/settings - настройки бота\n/savemessages - команда, переключающая режимы работы с ботом (будет ли бот запоминать сообщения)\n/info - Ваша персональная статистика.\n/stats - статистика сообщений от бота в группе.\n/help - Вы здесь!"

var stopMessage = []string{
	"ААААААААААА, НЕ ЛЕЗЬ, ОНО ТЕБЯ СОЖРЁТ!!!!!",
	"ОЙ, ОЙ, ОЙ, ТЫ ТУТ НЕ ПРОШЁЛ, ЛИМИТ ОШИБСЯ!",
	"АЛАРМ, АЛАРМ! АВАРИЙНАЯ ОСТАНОВКА!",
	"ЭТО НЕ ТВОЕ МЕСТО, ВЫШЕЛ ОТСЮДА, РОЗБiЙНИК!",
	"ОШИБКА 404: ЮМОР НЕ НАЙДЕН!",
	"ЭЙ, ТЫ ТАМ, ОСТАНОВИСЬ! ДОСТУП ЗАПРЕЩЁН!",
	"Ммм, что это за запах? ах, Вы случайно не пытаетесь достучаться до закрытого раздела?",
	"время остановиться и подумать: зачем тебе доступ сюда?",
	"упс, видимо, что-то пошло не так. попробуйте снова через... никогда.",
	"ОПА, ОПА, ОПА, КУДА ТЫ ПОЛЗЁШЬ? ДОСТУП ОТКАЗАН!",
	"слышь, ты тут не причём. продолжай смотреть свой контент.",
	"Эй, это не совсем то, что ты искал, правда?",
	"Подожди-ка, расскажи мне, что ты тут такого ищешь?",
	"Упс, кажется тут таинственные силы запрещают тебе доступ.",
	"Что-то мне подсказывает, что это не твоя зона комфорта.",
	"Чтобы получить доступ, нужно знать пароль. А пароль - \"не-по-хо-же\".",
	"К сожалению, тебе не хватает ключа, чтобы открыть этот раздел.",
	"Эээ, я бы на твоём месте подумал дважды о своих действиях.",
	"Эй! Тебе не разрешено сюда заходить!",
	"Доступ к этой части сайта возможен только по приглашению.",
	"Тут столько секретов, что даже влезать не хочется...",
}

func getVIP(user_id string) bool {
	chatsdb, err := sql.Open("sqlite3", "chats.db")
	errcheck(err, "config.go, getVIP")

	defer chatsdb.Close()

	tx, err := chatsdb.BeginTx(context.Background(), nil)
	errcheck(err, "config.go, getVIP")

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (id INT, username TEXT, isVIP BOOL, isAdmin BOOL, time INT NULL)") //создаём табличку
	errcheck(err, "config.go, getVIP")

	var result bool

	user_id_int, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		err = tx.QueryRow("SELECT isVIP FROM users WHERE username=?", user_id).Scan(&result) // Проверяем, есть ли уже запись с переданным id
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		}
	}

	err = tx.QueryRow("SELECT isVIP FROM users WHERE id=?", user_id_int).Scan(&result) // Проверяем, есть ли уже запись с переданным id
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)

	}
	tx.Commit()
	chatsdb.Close()
	return result
}

func getAdmin(user_id string) bool {
	chatsdb, err := sql.Open("sqlite3", "chats.db")
	errcheck(err, "config.go, getAdmin")

	defer chatsdb.Close()

	tx, err := chatsdb.BeginTx(context.Background(), nil)
	errcheck(err, "config.go, getAdmin")

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (id INT, username TEXT, isVIP BOOL, isAdmin BOOL, time INT NULL)") //создаём табличку
	errcheck(err, "config.go, getAdmin")

	var result bool

	user_id_int, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		err = tx.QueryRow("SELECT isAdmin FROM users WHERE username=?", user_id).Scan(&result) // Проверяем, есть ли уже запись с переданным id
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		}
	}

	err = tx.QueryRow("SELECT isAdmin FROM users WHERE id=?", user_id_int).Scan(&result) // Проверяем, есть ли уже запись с переданным id
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	tx.Commit()
	chatsdb.Close()
	return result
}

func giveVIPUtil(user_id string) error {
	chatsdb, err := sql.Open("sqlite3", "chats.db")
	if err != nil {
		return err
	}
	defer chatsdb.Close()

	tx, _ := chatsdb.BeginTx(context.Background(), nil)

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (id INT NULL, username TEXT NULL, isVIP BOOL, isAdmin BOOL, time INT NULL)")
	if err != nil {
		tx.Commit()
		chatsdb.Close()
		return err
	}

	var result bool
	user_id_int, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		err = tx.QueryRow("SELECT isVIP FROM users WHERE username=?", user_id).Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = tx.Exec("INSERT INTO users (id, username, isVIP, isAdmin, time) VALUES (?, ?, ?, ?, ?)", nil, user_id, true, false, -1)
				if err != nil {
					tx.Commit()
					chatsdb.Close()
					return err
				}
			} else {
				tx.Commit()
				chatsdb.Close()
				return err
			}
		} else {
			_, err = tx.Exec("UPDATE users SET isVIP=? WHERE username=?", true, user_id)
			if err != nil {
				return err
			}
			_, _ = tx.Exec("UPDATE users SET time=? WHERE username=?", -1, user_id)
		}
	} else {
		err = tx.QueryRow("SELECT isVIP FROM users WHERE id=?", user_id_int).Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = tx.Exec("INSERT INTO users (id, username, isVIP, isAdmin, time) VALUES (?, ?, ?, ?, ?)", user_id_int, "", true, false, -1)
				if err != nil {
					tx.Commit()
					chatsdb.Close()
					return err
				}
			} else {
				tx.Commit()
				chatsdb.Close()
				return err
			}
		} else {
			_, err = tx.Exec("UPDATE users SET isVIP=? WHERE id=?", true, user_id_int)
			if err != nil {
				tx.Commit()
				chatsdb.Close()
				return err
			}
			_, _ = tx.Exec("UPDATE users SET time=? WHERE id=?", -1, user_id_int)
		}
	}
	tx.Commit()
	chatsdb.Close()
	return nil
}

func giveAdminUtil(user_id string) error {
	chatsdb, err := sql.Open("sqlite3", "chats.db")
	if err != nil {
		return err
	}
	defer chatsdb.Close()

	tx, err := chatsdb.BeginTx(context.Background(), nil)

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (id INT NULL, username TEXT NULL, isVIP BOOL, isAdmin BOOL, time INT NULL)")
	if err != nil {
		tx.Commit()
		chatsdb.Close()
		return err
	}

	var result bool
	user_id_int, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		err = tx.QueryRow("SELECT isVIP FROM users WHERE username=?", user_id).Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = tx.Exec("INSERT INTO users (id, username, isVIP, isAdmin) VALUES (?, ?, ?, ?)", nil, user_id, true, true)
				if err != nil {
					tx.Commit()
					chatsdb.Close()
					return err
				}
				tx.Commit()
				chatsdb.Close()
			} else {
				tx.Commit()
				chatsdb.Close()
				return err
			}
		} else {
			_, err = tx.Exec("UPDATE users SET isAdmin=? WHERE username=?", true, user_id)
			if err != nil {
				tx.Commit()
				chatsdb.Close()
				return err
			}
			tx.Commit()
			chatsdb.Close()
		}
	} else {
		err = tx.QueryRow("SELECT isVIP FROM users WHERE id=?", user_id_int).Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = tx.Exec("INSERT INTO users (id, username, isVIP, isAdmin) VALUES (?, ?, ?, ?)", user_id_int, "", true, true)
				if err != nil {
					tx.Commit()
					chatsdb.Close()
					return err
				}
			} else {
				tx.Commit()
				chatsdb.Close()
				return err
			}
		} else {
			_, err = tx.Exec("UPDATE users SET isAdmin=? WHERE id=?", true, user_id_int)
			if err != nil {
				tx.Commit()
				chatsdb.Close()
				return err
			}
			tx.Commit()
			chatsdb.Close()
		}
		tx.Commit()
		chatsdb.Close()
	}
	tx.Commit()
	chatsdb.Close()
	return nil
}

func getUserInfo(ctx context.Context, b *bot.Bot, update *models.Update) error {
	var message string

	var user_who string

	var smilik string

	var isVip bool
	var isAdmin bool

	user_id := update.Message.From.ID
	username := update.Message.From.Username

	if utf8.RuneCountInString(username) > 1 {
		isVip = getVIP(username)
		isAdmin = getAdmin(username)

		if !isVip && !isAdmin {
			isVip = getVIP(strconv.Itoa(int(user_id)))
			isAdmin = getAdmin(strconv.Itoa(int(user_id)))
			if !isVip && !isAdmin {
				user_who = "user"
				smilik = "👨‍💼"
			} else if isAdmin {
				user_who = "Admin"
				smilik = "💡"
			} else if isVip {
				user_who = "VIP"
				smilik = "🌟"
			}
		} else if isAdmin {
			user_who = "Admin"
			smilik = "💡"
		} else if isVip {
			user_who = "VIP"
			smilik = "🌟"
		}
	} else {
		isVip = getVIP(strconv.Itoa(int(user_id)))
		isAdmin = getAdmin(strconv.Itoa(int(user_id)))
		if !isVip && !isAdmin {
			user_who = "user"
			smilik = "👨‍💼"
		} else if isAdmin {
			user_who = "Admin"
			smilik = "💡"
		} else if isVip {
			user_who = "VIP"
			smilik = "🌟"
		}
	}

	if isAdmin {
		message = smilik + " Ваш статус: *" + user_who +
			"*\n📩 Кол-во ваших сообщений: *" + strconv.Itoa(getCountTextCompletions(int(user_id))) +
			"*\n🌅 Сгенерировано изображений: *" + strconv.Itoa(getCountImgCompletions(int(user_id))) + "*"

		sendMessage(ctx, b, update, message)

		return nil
	}

	if isVip {
		time := time.Now().Unix()

		if checkNotTempVIP(username) || checkNotTempVIP(strconv.Itoa(int(user_id))) {
			message = smilik + " Ваш статус: *" + user_who + "* (истекает через: *НИКОГДА*)" +
				"\n📩 Кол-во ваших сообщений: *" + strconv.Itoa(getCountTextCompletions(int(user_id))) +
				"*\n🌅 Сгенерировано изображений: *" + strconv.Itoa(getCountImgCompletions(int(user_id))) + "*"
		} else {
			message = smilik + " Ваш статус: *" + user_who + "* (истекает через: *" + formatTime(int(durationVIP(username)-time)) + "*)" +
				"\n📩 Кол-во ваших сообщений: *" + strconv.Itoa(getCountTextCompletions(int(user_id))) +
				"*\n🌅 Сгенерировано изображений: *" + strconv.Itoa(getCountImgCompletions(int(user_id))) + "*"
		}

		sendMessage(ctx, b, update, message)
		return nil

	}

	message = smilik + " Ваш статус: *" + user_who +
		"*\n📩 Кол-во ваших сообщений: *" + strconv.Itoa(getCountTextCompletions(int(user_id))) +
		"*\n🌅 Сгенерировано изображений: *" + strconv.Itoa(getCountImgCompletions(int(user_id))) + "*"

	sendMessage(ctx, b, update, message)

	return nil

}

func statsUtil(ctx context.Context, b *bot.Bot, update *models.Update) error {
	var message string

	id := update.Message.Chat.ID

	message = "Чат: *" + update.Message.Chat.Title + "*" +
		"\n📩 Кол-во сообщений в чате: " +
		"*" + strconv.Itoa(getCountTextCompletions(int(id))) + "*" +
		"\n🌅 Сгенерировано изображений в чате: " +
		"*" + strconv.Itoa(getCountImgCompletions(int(id))) + "*"

	sendbroadcast(ctx, b, int(id), message)

	return nil
}
