package main

import (
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestTodoMySqlModelCreate(t *testing.T) {
	wantErr := errors.New("Dummy error")
	// No table
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnError(wantErr)
	mock.ExpectExec("CREATE TABLE todo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO todo").WithArgs("dummy user", "dummy task", AnyTime{}).WillReturnResult(sqlmock.NewResult(1, 1))
	model := TodoMySqlModel{
		db: db,
	}
	todo := Todo{
		UserID: "dummy user",
		Task:   "dummy task",
		Due:    time.Now(),
	}
	err = model.Create(todo)
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Create(%#v) == %#v, want %#v", todo, err, nil)
	}
	// No table but error when create table
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnError(wantErr)
	mock.ExpectExec("CREATE TABLE todo").WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Create(todo)
	if err == nil {
		t.Errorf("Result TodoMySqlModel.Create(%#v) == %#v, want %#v", todo, err, wantErr)
	}
	// Table already exist
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectExec("INSERT INTO todo").WithArgs("dummy user", "dummy task", AnyTime{}).WillReturnResult(sqlmock.NewResult(1, 1))
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Create(todo)
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Create(%#v) == %#v, want %#v", todo, err, nil)
	}
	// Table already exist but error when insert row
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectExec("INSERT INTO todo").WithArgs("dummy user", "dummy task", AnyTime{}).WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Create(todo)
	if err == nil {
		t.Errorf("Result TodoMySqlModel.Create(%#v) == %#v, want %#v", todo, err, wantErr)
	}
}

func TestNewTodoMySqlModel(t *testing.T) {
	model := NewTodoMySqlModel()
	if model.db == nil {
		t.Errorf("Result NewTodoMySqlModel() == %#v", model.db)
	}

}
