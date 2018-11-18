package bot

import (
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/choobot/choo-todo-bot/app/model"
	"github.com/line/line-bot-sdk-go/linebot"
)

func TestTodoBotParseUserMessage(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	cases := []struct {
		in       string
		wantTask string
		wantDue  string
		wantErr  string
	}{
		{
			in:       "Go shopping : 2/1/06 : 15:04",
			wantTask: "Go shopping",
			wantDue:  "2006-01-02T15:04:00+07:00",
			wantErr:  "",
		},
		{
			in:       "Go shopping : 2/1/06",
			wantTask: "Go shopping",
			wantDue:  "2006-01-02T12:00:00+07:00",
			wantErr:  "",
		},
		{
			in:       "Go shopping : today : 15:04",
			wantTask: "Go shopping",
			wantDue:  time.Now().In(loc).Format("2006-01-02") + "T15:04:00+07:00",
			wantErr:  "",
		},
		{
			in:       "Go shopping : today",
			wantTask: "Go shopping",
			wantDue:  time.Now().In(loc).Format("2006-01-02") + "T12:00:00+07:00",
			wantErr:  "",
		},
		{
			in:       "Go shopping : tomorrow : 15:04",
			wantTask: "Go shopping",
			wantDue:  time.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02") + "T15:04:00+07:00",
			wantErr:  "",
		},
		{
			in:       "Go shopping : tomorrow",
			wantTask: "Go shopping",
			wantDue:  time.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02") + "T12:00:00+07:00",
			wantErr:  "",
		},
		{
			in:       "dummy",
			wantTask: "",
			wantDue:  "0001-01-01T06:42:04+06:42",
			wantErr:  "Wrong format",
		},
		{
			in:       "Go shopping : 2/1/06 : 25:04",
			wantTask: "",
			wantDue:  "0001-01-01T06:42:04+06:42",
			wantErr:  "Wrong format",
		},
		{
			in:       "Go shopping : 32/1/06",
			wantTask: "",
			wantDue:  "0001-01-01T06:42:04+06:42",
			wantErr:  "Wrong format",
		},
	}

	bot := TodoBot{}
	for _, c := range cases {
		got, err := bot.ParseUserMessage(c.in)
		if got.Task != c.wantTask || got.Due.In(loc).Format(time.RFC3339) != c.wantDue || (err != nil && c.wantErr != err.Error()) {
			t.Errorf("TodoBot.ParseUserMessage(%q) == %v, %v, %v, want %v, %v, %v", c.in, got.Task, got.Due.In(loc).Format(time.RFC3339), err, c.wantTask, c.wantDue, c.wantErr)
		}
	}
}

func TestTodoBotFormatDate(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	layout := "2006-01-02 15:04"
	now, _ := time.ParseInLocation("2006-01-02", "2018-11-15", loc)
	cases := []struct {
		in   string
		want string
	}{
		{
			in:   "2018-11-15 15:04",
			want: "Today at 15:04",
		},
		{
			in:   "2018-11-16 15:04",
			want: "Tomorrow at 15:04",
		},
		{
			in:   "2018-11-17 15:04",
			want: "Sat at 15:04",
		},
		{
			in:   "2018-11-14 15:04",
			want: "Yesterday at 15:04",
		},
		{
			in:   "2018-11-13 15:04",
			want: "Last Tue at 15:04",
		},
		{
			in:   "2018-11-20 15:04",
			want: "Next Tue at 15:04",
		},
		{
			in:   "2018-11-10 15:04",
			want: "Sat 10 Nov 18 at 15:04",
		},
	}

	bot := TodoBot{}
	for _, c := range cases {
		date, _ := time.ParseInLocation(layout, c.in, loc)
		got := bot.FormatDate(now, date)
		if got != c.want {
			t.Errorf("TodoBot.FormatDate(%q) == %q want %q", c.in, got, c.want)
		}
	}
}
func TestTodoBotPushMessage(t *testing.T) {
	client, _ := linebot.New(os.Getenv("LINE_BOT_SECRET"), os.Getenv("LINE_BOT_TOKEN"))
	bot := TodoBot{
		Client: client,
	}
	bot.PushMessage("dummy", "dummy")
}

type mockTodoModel struct {
	willError       bool
	willNoRemaining bool
}

func (this *mockTodoModel) List(userID string) ([]model.Todo, error) {
	return nil, nil
}
func (this *mockTodoModel) Create(todo model.Todo) error {
	if this.willError {
		this.willError = false
		return errors.New("dummy")
	}
	return nil
}
func (this *mockTodoModel) Pin(todo model.Todo) error {
	return nil
}
func (this *mockTodoModel) Done(todo model.Todo) error {
	return nil
}
func (this *mockTodoModel) Remind() (map[string][]model.Todo, error) {
	if this.willError {
		this.willError = false
		return nil, errors.New("dummy")
	}
	userTodos := map[string][]model.Todo{}
	todos := userTodos["dummy"]
	if this.willNoRemaining {
		this.willNoRemaining = false
		todo := model.Todo{
			Done: true,
		}
		todos = append(todos, todo)
	} else {
		todo := model.Todo{
			Done: false,
		}
		todos = append(todos, todo)
	}

	todo := model.Todo{
		Pin: true,
	}
	todos = append(todos, todo)
	userTodos["dummy"] = todos
	return userTodos, nil
}

func TestTodoBotRemind(t *testing.T) {
	client, _ := linebot.New(os.Getenv("LINE_BOT_SECRET"), os.Getenv("LINE_BOT_TOKEN"))
	todoModel := mockTodoModel{}
	bot := TodoBot{
		Client:    client,
		TodoModel: &todoModel,
	}

	todoModel.willError = true
	err := bot.Remind()
	if err == nil {
		t.Errorf("TodoBot.Remind() == %v want %v", nil, err)
	}

	todoModel.willNoRemaining = true
	err = bot.Remind()
	if err != nil {
		t.Errorf("TodoBot.Remind() == %v want %v", err, nil)
	}

	err = bot.Remind()
	if err != nil {
		t.Errorf("TodoBot.Remind() == %v want %v", err, nil)
	}
}

func TestTodoBotResponse(t *testing.T) {
	wantErr := errors.New("linebot: APIError 400 Invalid reply token")
	client, _ := linebot.New(os.Getenv("LINE_BOT_SECRET"), os.Getenv("LINE_BOT_TOKEN"))
	todoModel := mockTodoModel{}
	bot := &TodoBot{
		Client:    client,
		TodoModel: &todoModel,
	}

	events := []*linebot.Event{}
	err := bot.Response(events)
	if err != nil {
		t.Errorf("TodoBot.Response(%v) == %v, want %v", events, err, nil)
	}

	// Wrong word
	event := linebot.Event{
		Type: linebot.EventTypeMessage,
		Message: &linebot.TextMessage{
			Text: "dummy",
		},
		Source: &linebot.EventSource{
			UserID: "dummy",
		},
		ReplyToken: "dummy",
	}
	events = append(events, &event)
	bot.Response(events)
	err = bot.Response(events)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("TodoBot.Response(%v) == %v, want %v", events, err, wantErr)
	}

	// //Edit
	event = linebot.Event{
		Type: linebot.EventTypeMessage,
		Message: &linebot.TextMessage{
			Text: "edit",
		},
		Source: &linebot.EventSource{
			UserID: "dummy",
		},
		ReplyToken: "dummy",
	}
	events = append(events, &event)
	err = bot.Response(events)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("TodoBot.Response(%v) == %v, want %v", events, err, wantErr)
	}

	//Create Task
	event = linebot.Event{
		Type: linebot.EventTypeMessage,
		Message: &linebot.TextMessage{
			Text: "Go shopping : today : 13:00",
		},
		Source: &linebot.EventSource{
			UserID: "dummy",
		},
		ReplyToken: "dummy",
	}
	events = append(events, &event)
	err = bot.Response(events)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("TodoBot.Response(%v) == %v, want %v", events, err, wantErr)
	}

	// Join
	event = linebot.Event{
		Type: linebot.EventTypeJoin,
		Message: &linebot.TextMessage{
			Text: "error_word",
		},
		Source: &linebot.EventSource{
			UserID: "dummy",
		},
		ReplyToken: "dummy",
	}
	events = append(events, &event)
	err = bot.Response(events)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("TodoBot.Response(%v) == %v, want %v", events, err, wantErr)
	}

	// Error when create tasks
	event = linebot.Event{
		Type: linebot.EventTypeMessage,
		Message: &linebot.TextMessage{
			Text: "Go shopping : today : 13:00",
		},
		Source: &linebot.EventSource{
			UserID: "dummy",
		},
		ReplyToken: "dummy",
	}
	events = append(events, &event)
	todoModel.willError = true
	err = bot.Response(events)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("TodoBot.Response(%v) == %v, want %v", events, err, wantErr)
	}
}
