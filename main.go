package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	TelegramBotToken string
	BotanApiToken    string
}

type BotanMessage struct {
	Text   string
	ChatId int
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("Hi there! I'm DndSpellsBot!"))
}

// var numericKeyboard = tgbotapi.NewReplyKeyboard(
// 	tgbotapi.NewKeyboardButtonRow(
// 		tgbotapi.NewKeyboardButton("1"),
// 		tgbotapi.NewKeyboardButton("2"),
// 		tgbotapi.NewKeyboardButton("3"),
// 	),
// 	tgbotapi.NewKeyboardButtonRow(
// 		tgbotapi.NewKeyboardButton("4"),
// 		tgbotapi.NewKeyboardButton("5"),
// 		tgbotapi.NewKeyboardButton("6"),
// 	),
// )

func main() {
	//copy from https://bitbucket.org/feeeper/dndspellsbot/overview
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	classes := []string{
		"Barbarian",
		"Bard",
		"Cleric",
		"Druid",
		"Fighter",
		"Monk",
		"Paladin",
		"Ranger",
		"Rogue",
		"Sorcerer",
		"Warlock",
		"Wizard",
	}
	classesMap = make(map[int]string)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	// updates := bot.ListenForWebhook("/" + bot.Token)

	if err != nil {
		log.Panic(err)
	}

	spells, err := parseSpells()
	if err != nil {
		log.Panic(err)
	}

	http.HandleFunc("/", MainHandler)
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	for update := range updates {
		if update.Message == nil && update.InlineQuery != nil {
			query := update.InlineQuery.Query
			filteredSpells := Filter(spells.Spells, func(spell Spell) bool {
				classCond := true
				class, ok := classesMap[update.InlineQuery.From.ID]
				if ok {
					classCond = strings.Index(strings.ToLower(spell.Classes), strings.ToLower(class)) >= 0
				}
				return strings.Index(strings.ToLower(spell.Name), strings.ToLower(query)) >= 0 && classCond
			})

			var articles []interface{}
			if len(filteredSpells) == 0 {
				msg := tgbotapi.NewInlineQueryResultArticleMarkdown(update.InlineQuery.ID, "No one spells matches", "No one spells matches")
				articles = append(articles, msg)
			} else {
				var i = 0
				for _, spell := range filteredSpells {
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

					msg := tgbotapi.NewInlineQueryResultArticleMarkdown(spell.Name, spell.Name, text)
					articles = append(articles, msg)
					if i >= 10 {
						break
					}
				}
			}

			inlineConfig := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				IsPersonal:    true,
				CacheTime:     0,
				Results:       articles,
			}
			_, err := bot.AnswerInlineQuery(inlineConfig)
			if err != nil {
				log.Println(err)
			}
		} else {
			var command = ""
			if update.Message != nil {
				command = update.Message.Command()
				log.Println(command)

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

				if command == "" {
					query := update.Message.Text
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					bot.Send(msg)

					filteredSpells := Filter(spells.Spells, func(spell Spell) bool {
						class, ok := classesMap[update.Message.From.ID]
						classCond := true
						if ok {
							classCond = strings.Index(strings.ToLower(spell.Classes), strings.ToLower(class)) >= 0
						}
						return strings.Index(strings.ToLower(spell.Name), strings.ToLower(query)) >= 0 && classCond
					})

					if len(filteredSpells) == 0 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "No one spells matches")
						bot.Send(msg)
					}

					for _, spell := range filteredSpells {
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
				} else {
					switch command {
					case "setclass":
						log.Println("setclass case")
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
					case "removeclass":
						class, ok := classesMap[update.Message.From.ID]
						if ok {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
							bot.Send(msg)
							delete(classesMap, update.Message.From.ID)
							bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Now you're not a "+class))
						} else {
							bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You don't have any saved class"))
						}
					}
				}
			} else {
				if update.CallbackQuery != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					bot.Send(msg)

					class := update.CallbackQuery.Data
					classesMap[update.CallbackQuery.From.ID] = class
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ok, I remember"))
				}
			}

		}
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

// func timerHandler(m *tgbotapi.Message) {
// 	// m.Vars contains all variables, parsed during routing
// 	secondsStr := m.Vars["seconds"]
// 	// Convert string variable to integer seconds value
// 	seconds, err := strconv.Atoi(secondsStr)
// 	if err != nil {
// 		m.Reply("Invalid number of seconds")
// 		return
// 	}
// 	m.Replyf("Timer for %d seconds started", seconds)
// 	time.Sleep(time.Duration(seconds) * time.Second)
// 	m.Reply("Time out!")
// }

// func textHandler(m *tgbotapi.Message) {
// 	text := fmt.Sprintf(
// 		"*%s*\n"+
// 			"*Level* _%v_\n"+
// 			"*School* _%s_\n"+
// 			"*Time* _%s_\n"+
// 			"*Range* _%s_\n"+
// 			"*Components* _%s_\n"+
// 			"*Duration* _%s_\n"+
// 			"*Classes* _%s_\n"+
// 			"*Roll* _%s_\n"+
// 			"%s",
// 		spell.Name,
// 		spell.Level,
// 		spell.School,
// 		spell.Time,
// 		spell.Range,
// 		spell.Components,
// 		spell.Duration,
// 		spell.Classes,
// 		strings.Join(spell.Rolls, ", "),
// 		strings.Join(spell.Texts, "\n"))

// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
// 	msg.ParseMode = "markdown"
// 	bot.Send(msg)
// }

// func buttonHandelr(m *tgbotapi.Message) {
// 	switch command {
// 	case "setclass":
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select your class")

// 		keyboard := tgbotapi.InlineKeyboardMarkup{}
// 		for _, class := range classes {
// 			var row []tgbotapi.InlineKeyboardButton
// 			btn := tgbotapi.NewInlineKeyboardButtonData(class, class)
// 			row = append(row, btn)
// 			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
// 		}

// 		msg.ReplyMarkup = keyboard
// 		bot.Send(msg)
// 	}
// }
