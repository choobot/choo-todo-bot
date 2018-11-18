package bot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/choobot/choo-todo-bot/app/model"
	"github.com/line/line-bot-sdk-go/linebot"
)

type TodoBot struct {
	Client    *linebot.Client
	TodoModel model.TodoModel
}

func (this *TodoBot) Remind() error {
	userTodos, err := this.TodoModel.Remind()
	if err != nil {
		return err
	}
	for userID, todos := range userTodos {
		message := "Hi there,\n"
		showDone := false
		remaining := 0
		for i, todo := range todos {
			if i == 0 && todo.Done {
				message += "Well done, you have no remaining tasks to be done :)\n"
			} else if i == 0 {
				message += "Tasks to be done:\n"
			}
			if !todo.Done {
				remaining++
			} else if todo.Done && !showDone {
				message += "Tasks completed:\n"
				showDone = true
			}
			if todo.Pin {
				message += "*** "
			} else {
				message += "    "
			}
			due := this.FormatDate(time.Now(), todo.Due)
			if !todo.Done && time.Now().After(todo.Due) {
				due += " (overdue)"
			}
			message += fmt.Sprintf("%v : %v\n", todo.Task, due)

		}
		if remaining != 0 {
			message += fmt.Sprintf("%d of %d remaining, just do it!", remaining, len(todos))
		}
		//Fork for massive API calls
		go this.PushMessage(userID, message)
	}
	return nil
}

func (this *TodoBot) FormatDate(now time.Time, date time.Time) string {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	dateText := date.Format("2006-01-02")
	timeText := date.Format("15:04")
	_, todayWeek := now.ISOWeek()
	_, dueWeek := date.ISOWeek()
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	if dateText == today {
		// Today
		return "Today at " + timeText
	} else if dateText == tomorrow {
		// Tomorrow
		return "Tomorrow at " + timeText
	} else if dateText == yesterday {
		// Yesterday
		return "Yesterday at " + timeText
	} else if todayWeek == dueWeek && now.After(date) {
		// This week
		return "Last " + date.Format("Mon at 15:04")
	} else if todayWeek == dueWeek {
		// This week in the past
		return date.Format("Mon at 15:04")
	} else if dueWeek-todayWeek == 1 {
		// Next week
		return "Next " + date.Format("Mon at 15:04")
	}
	return date.Format("Mon 2 Jan 06 at 15:04")
}

func (this *TodoBot) PushMessage(userID string, message string) {
	if _, err := this.Client.PushMessage(userID, linebot.NewTextMessage(message)).Do(); err != nil {
		log.Println(err)
	}
}

// 1) Go shopping : 2/5/18 : 13:00
// 2) Go shopping : 2/5/18
// 3) Go shopping : today : 15:30
// 4) Go shopping : today
// 5) Go shopping : tomorrow : 18:00
// 6) Go shopping : tomorrow
func (this *TodoBot) ParseUserMessage(msg string) (model.Todo, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	getDate := func(word string) string {
		format := "2/1/06"
		if strings.ToLower(word) == "today" {
			return time.Now().In(loc).Format(format)
		} else if strings.ToLower(word) == "tomorrow" {
			return time.Now().In(loc).AddDate(0, 0, 1).Format(format)
		}
		return word
	}
	layout := "2/1/06 15:04"
	words := strings.Split(msg, " : ")
	task := ""
	var due time.Time
	var err error
	if len(words) == 2 {
		task = words[0]
		due, err = time.ParseInLocation(layout, getDate(words[1])+" 12:00", loc)
		if err != nil {
			return model.Todo{}, errors.New("Wrong format")
		}
	} else if len(words) == 3 {
		task = words[0]
		due, err = time.ParseInLocation(layout, getDate(words[1])+" "+words[2], loc)
		if err != nil {
			return model.Todo{}, errors.New("Wrong format")
		}
	} else {
		return model.Todo{}, errors.New("Wrong format")
	}
	todo := model.Todo{
		Task: task,
		Due:  due,
	}
	return todo, nil
}

func (this *TodoBot) Response(events []*linebot.Event) error {
	howto := `You can create todo list by using these formats:
	1) Go shopping : 25/5/18 : 13:00
	2) Go shopping : 25/5/18
	3) Go shopping : today : 15:30
	4) Go shopping : today
	5) Go shopping : tomorrow : 18:00
	6) Go shopping : tomorrow
You can edit todo list by input word "edit"`

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msg := message.Text
				if strings.ToLower(msg) == "edit" {
					reply := "Please go to " + os.Getenv("EDIT_URL")
					if _, err := this.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
						return err
					}
				} else {
					todo, err := this.ParseUserMessage(msg)
					if err != nil {
						if _, err = this.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(howto)).Do(); err != nil {
							return err
						}
					} else {
						todo.UserID = event.Source.UserID
						if err := this.TodoModel.Create(todo); err != nil {
							reply := err.Error()
							if _, err = this.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
								return err
							}
						} else {
							reply := "Task has been created."
							if _, err = this.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
								return err
							}
						}
					}
				}

			}
		} else if event.Type == linebot.EventTypeJoin {
			replyMessage := "Thanks for adding me. I'm Choo Todo Bot, I'm here to help you to manage your tasks.\n" + howto
			if _, err := this.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
				return err
			}
		}
	}
	return nil
}
