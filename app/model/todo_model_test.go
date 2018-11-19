package model

import (
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

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

	todo := Todo{
		UserID: "dummy user",
		Task:   "dummy task",
		Due:    time.Now(),
	}
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnError(wantErr)
	mock.ExpectExec("CREATE TABLE todo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO todo").WithArgs("dummy user", "dummy task", AnyTime{}).WillReturnResult(sqlmock.NewResult(1, 1))
	model := TodoMySqlModel{
		db: db,
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
	// Table already exist but no row affected
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectExec("INSERT INTO todo").WithArgs("dummy user", "dummy task", AnyTime{}).WillReturnResult(sqlmock.NewResult(1, 0))
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Create(todo)
	wantErr = errors.New("No record")
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Create(%#v) == %#v, want %#v", todo, err, nil)
	}
}

func TestNewTodoMySqlModel(t *testing.T) {
	model := NewTodoMySqlModel()
	if model.db == nil {
		t.Errorf("Result NewTodoMySqlModel() == %#v", model.db)
	}

}

func TestTodoMySqlModelList(t *testing.T) {
	wantErr := errors.New("Dummy error")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	//Success
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectQuery("SELECT id, task, done, pin, due FROM todo WHERE user_id=?").WithArgs("dummy user").WillReturnRows(
		sqlmock.NewRows([]string{
			"id",
			"task",
			"done",
			"pin",
			"due",
		}).AddRow(
			1,
			"task",
			false,
			true,
			time.Now(),
		))
	model := TodoMySqlModel{
		db: db,
	}
	_, err = model.List("dummy user")
	if err != nil {
		t.Errorf("Result TodoMySqlModel.List(%q) == %v, want %v", "dummy user", err, nil)
	}

	//No Table
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectQuery("SELECT id, task, done, pin, due FROM todo WHERE user_id=?").WithArgs("dummy user").WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	_, err = model.List("dummy user")
	if err == nil {
		t.Errorf("Result TodoMySqlModel.List(%q) == %v, want %v", "dummy user", err, wantErr)
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
	_, err = model.List("dummy user")
	if err == nil {
		t.Errorf("Result TodoMySqlModel.List(%q) == %#v, want %#v", "dummy user", err, wantErr)
	}

	//Wrong col type
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectQuery("SELECT id, task, done, pin, due FROM todo WHERE user_id=?").WithArgs("dummy user").WillReturnRows(
		sqlmock.NewRows([]string{
			"id",
			"task",
			"done",
			"pin",
			"due",
		}).AddRow(
			1,
			"task",
			false,
			true,
			"wrong date",
		))
	model = TodoMySqlModel{
		db: db,
	}
	_, err = model.List("dummy user")
	wantErr = errors.New(`sql: Scan error on column index 4, name "due": unsupported Scan, storing driver.Value type string into type *time.Time`)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.List(%q) == %v, want %v", "dummy user", err, wantErr)
	}
}

func TestTodoMySqlModelPin(t *testing.T) {
	wantErr := errors.New("dummy error")
	// Success
	todo := Todo{
		ID:  1,
		Pin: true,
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo SET pin=?").WithArgs(true, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	model := TodoMySqlModel{
		db: db,
	}
	err = model.Pin(todo)
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Pin(%#v) == %#v, want %#v", todo, err, nil)
	}
	// Error when insert row
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo SET pin=?").WithArgs(true, 1).WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Pin(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Pin(%#v) == %#v, want %#v", todo, err, wantErr)
	}
	// No row affected
	wantErr = errors.New("No record")
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo SET pin=?").WithArgs(true, 1).WillReturnResult(sqlmock.NewResult(1, 0))
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Pin(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Pin(%#v) == %#v, want %#v", todo, err, wantErr)
	}
}

func TestTodoMySqlModelDone(t *testing.T) {
	wantErr := errors.New("dummy error")
	// Success
	todo := Todo{
		ID:   1,
		Done: true,
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo SET done=?").WithArgs(true, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	model := TodoMySqlModel{
		db: db,
	}
	err = model.Done(todo)
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Done(%#v) == %#v, want %#v", todo, err, nil)
	}
	// Error when insert row
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo SET done=?").WithArgs(true, 1).WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Done(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Done(%#v) == %#v, want %#v", todo, err, wantErr)
	}
	// No row affected
	wantErr = errors.New("No record")
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo SET done=?").WithArgs(true, 1).WillReturnResult(sqlmock.NewResult(1, 0))
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Done(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Done(%#v) == %#v, want %#v", todo, err, wantErr)
	}
}

func TestTodoMySqlModelRemind(t *testing.T) {
	wantErr := errors.New("Dummy error")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	//Success
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectQuery("SELECT user_id, id, task, done, pin, due FROM todo ORDER BY user_id, done, pin DESC, due").WillReturnRows(
		sqlmock.NewRows([]string{
			"user_id",
			"id",
			"task",
			"done",
			"pin",
			"due",
		}).AddRow(
			"dummy user",
			1,
			"task",
			false,
			true,
			time.Now(),
		))
	model := TodoMySqlModel{
		db: db,
	}
	_, err = model.Remind()
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Remind() == %v, want %v", err, nil)
	}

	//No Table
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectQuery("SELECT user_id, id, task, done, pin, due FROM todo ORDER BY user_id, done, pin DESC, due").WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	_, err = model.Remind()
	if err == nil {
		t.Errorf("Result TodoMySqlModel.Remind() == %v, want %v", err, wantErr)
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
	_, err = model.Remind()
	if err == nil {
		t.Errorf("Result TodoMySqlModel.Remind() == %#v, want %#v", err, wantErr)
	}
	//Wrong col type
	mock.ExpectQuery("SELECT 1 FROM todo LIMIT 1").WillReturnRows(
		sqlmock.NewRows([]string{"dummy_col"}).AddRow("1"))
	mock.ExpectQuery("SELECT user_id, id, task, done, pin, due FROM todo ORDER BY user_id, done, pin DESC, due").WillReturnRows(
		sqlmock.NewRows([]string{
			"user_id",
			"id",
			"task",
			"done",
			"pin",
			"due",
		}).AddRow(
			"dummy user",
			1,
			"task",
			false,
			true,
			"wrong date",
		))
	model = TodoMySqlModel{
		db: db,
	}
	_, err = model.Remind()
	wantErr = errors.New(`sql: Scan error on column index 5, name "due": unsupported Scan, storing driver.Value type string into type *time.Time`)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Remind() == %v, want %v", err, wantErr)
	}
}

func TestTodoMySqlModelEdit(t *testing.T) {
	wantErr := errors.New("dummy error")
	// Success
	todo := Todo{
		ID:   1,
		Task: "dummy",
		Due:  time.Now(),
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo").WithArgs(todo.Task, todo.Due, todo.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	model := TodoMySqlModel{
		db: db,
	}
	err = model.Edit(todo)
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Edit(%#v) == %#v, want %#v", todo, err, nil)
	}
	// Error when insert row
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo").WithArgs(todo.Task, todo.Due, todo.ID).WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Edit(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Edit(%#v) == %#v, want %#v", todo, err, wantErr)
	}
	// No row affected
	wantErr = errors.New("No record")
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE todo").WithArgs(todo.Task, todo.Due, todo.ID).WillReturnResult(sqlmock.NewResult(1, 0))
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Edit(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Edit(%#v) == %#v, want %#v", todo, err, wantErr)
	}
}

func TestTodoMySqlModelDelete(t *testing.T) {
	wantErr := errors.New("dummy error")
	// Success
	todo := Todo{
		ID: 1,
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM todo").WithArgs(todo.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	model := TodoMySqlModel{
		db: db,
	}
	err = model.Delete(todo)
	if err != nil {
		t.Errorf("Result TodoMySqlModel.Delete(%#v) == %#v, want %#v", todo, err, nil)
	}
	// Error when insert row
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM todo").WithArgs(todo.ID).WillReturnError(wantErr)
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Delete(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Delete(%#v) == %#v, want %#v", todo, err, wantErr)
	}
	// No row affected
	wantErr = errors.New("No record")
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM todo").WithArgs(todo.ID).WillReturnResult(sqlmock.NewResult(1, 0))
	model = TodoMySqlModel{
		db: db,
	}
	err = model.Delete(todo)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Result TodoMySqlModel.Delete(%#v) == %#v, want %#v", todo, err, wantErr)
	}
}
