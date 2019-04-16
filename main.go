package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	TelegramBotToken string
	// BotanApiToken    string
}

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("1"),
		tgbotapi.NewKeyboardButton("2"),
		tgbotapi.NewKeyboardButton("3"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("4"),
		tgbotapi.NewKeyboardButton("5"),
		tgbotapi.NewKeyboardButton("6"),
	),
)

func main() {

	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}

	// socks5 := os.Getenv("SOCKS5_PROXY")

	// export SOCKS5_PROXY="socks5://url:443"

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		switch update.Message.Text {
		case "open":
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}

		bot.Send(msg)

		// if update.Message.IsCommand() {
		// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		// 	switch update.Message.Command() {
		// 	case "help":
		// 		msg.Text = "type /sayhi or /status."
		// 	case "sayhi":
		// 		msg.Text = "Hi :)"
		// 	case "status":
		// 		msg.Text = "I'm ok."
		// 	default:
		// 		msg.Text = "I don't know that command"
		// 	}
		// 	bot.Send(msg)
		// }

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// bot.Send(msg)
	}
}

func timerHandler(m *tgbotapi.Message) {
	// m.Vars contains all variables, parsed during routing
	secondsStr := m.Vars["seconds"]
	// Convert string variable to integer seconds value
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil {
		m.Reply("Invalid number of seconds")
		return
	}
	m.Replyf("Timer for %d seconds started", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
	m.Reply("Time out!")
}

func textHandler(m *tgbotapi.Message) {
	text := fmt.Sprintf(
		"*%s*\n"+
			"*Level* _%v_\n"+
			"*School* _%s_\n"+
			"*Time* _%s_\n"+
			"*Range* _%s_\n"+
			"*Components* _%s_\n"+
			"*Duration* _%s_\n"+
			"*Classes* _%s_\n"+
			"*Roll* _%s_\n"+
			"%s",
		spell.Name,
		spell.Level,
		spell.School,
		spell.Time,
		spell.Range,
		spell.Components,
		spell.Duration,
		spell.Classes,
		strings.Join(spell.Rolls, ", "),
		strings.Join(spell.Texts, "\n"))

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

func buttonHandelr(m *tgbotapi.Message) {
	switch command {
	case "setclass":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select your class")

		keyboard := tgbotapi.InlineKeyboardMarkup{}
		for _, class := range classes {
			var row []tgbotapi.InlineKeyboardButton
			btn := tgbotapi.NewInlineKeyboardButtonData(class, class)
			row = append(row, btn)
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}

		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}

var classesMap map[int]string

type Spells struct {
	XMLName xml.Name `xml:"compendium"`
	Spells  []Spell  `xml:"spell"`
}

type Spell struct {
	XMLName    xml.Name `xml:"spell"`
	Name       string   `xml:"name"`
	Level      int      `xml:"level"`
	School     string   `xml:"school"`
	Time       string   `xml:"time"`
	Range      string   `xml:"range"`
	Components string   `xml:"components"`
	Duration   string   `xml:"duration"`
	Classes    string   `xml:"classes"`
	Texts      []string `xml:"text"`
	Rolls      []string `xml:"roll"`
}

func parseSpells() (Spells, error) {
	file, err := os.Open("phb.xml")
	if err != nil {
		log.Panic(err)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Panic(err)
	}

	var data = make([]byte, fi.Size())
	_, err = file.Read(data)
	if err != nil {
		log.Panic(err)
	}

	var v Spells
	err = xml.Unmarshal(data, &v)

	if err != nil {
		log.Println(err)
		return v, err
	} else {
		log.Printf("Total spells found: %v", len(v.Spells))
		return v, err
	}
}

func Filter(spells []Spell, fn func(spell Spell) bool) []Spell {
	var filtered []Spell
	for _, spell := range spells {
		if fn(spell) {
			filtered = append(filtered, spell)
		}
	}
	return filtered
}
