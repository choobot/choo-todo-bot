package model

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID     int
	UserID string
	Task   string
	Done   bool
	Pin    bool
	Due    time.Time
}

type TodoModel interface {
	List(userID string) ([]Todo, error)
	Create(todo Todo) error
	Pin(todo Todo) error
	Done(todo Todo) error
	Remind() (map[string][]Todo, error)
	Edit(todo Todo) error
}

type TodoMySqlModel struct {
	db *sql.DB
}

func (this *TodoMySqlModel) SetTimeZone() error {
	sql := `SET time_zone = 'Asia/Bangkok'`
	_, err := this.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (this *TodoMySqlModel) List(userID string) ([]Todo, error) {
	this.SetTimeZone()
	err := this.CreateTablesIfNotExist()
	if err != nil {
		return nil, err
	}
	var todos []Todo
	rows, err := this.db.Query("SELECT id, task, done, pin, due FROM todo WHERE user_id=?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var task string
		var done bool
		var pin bool
		var due time.Time
		if err := rows.Scan(&id, &task, &done, &pin, &due); err != nil {
			return nil, err
		}
		loc, _ := time.LoadLocation("Asia/Bangkok")
		due = due.In(loc)
		todo := Todo{
			ID:     id,
			UserID: userID,
			Task:   task,
			Pin:    pin,
			Done:   done,
			Due:    due,
		}
		todos = append(todos, todo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

func NewTodoMySqlModel() TodoMySqlModel {
	db, _ := sql.Open("mysql", os.Getenv("DATA_SOURCE_NAME"))
	return TodoMySqlModel{
		db: db,
	}
}

func (this *TodoMySqlModel) CreateTablesIfNotExist() error {
	sql := "SELECT 1 FROM todo LIMIT 1"
	_, err := this.db.Query(sql)
	if err != nil {
		sql = `
		CREATE TABLE todo (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			task TEXT NOT NULL,
			done BOOL NOT NULL DEFAULT FALSE,
			pin BOOL NOT NULL DEFAULT FALSE,
			due DATETIME NOT NULL
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		_, err = this.db.Exec(sql)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *TodoMySqlModel) Create(todo Todo) error {
	this.SetTimeZone()
	err := this.CreateTablesIfNotExist()
	if err != nil {
		return err
	}
	sql := `INSERT INTO todo ( user_id, task, due ) VALUES( ?, ?, ?)`
	result, err := this.db.Exec(sql, todo.UserID, todo.Task, todo.Due)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}
func (this *TodoMySqlModel) Pin(todo Todo) error {
	sql := `UPDATE todo SET pin=? WHERE id=?`
	result, err := this.db.Exec(sql, todo.Pin, todo.ID)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}

func (this *TodoMySqlModel) Done(todo Todo) error {
	sql := `UPDATE todo SET done=? WHERE id=?`
	result, err := this.db.Exec(sql, todo.Done, todo.ID)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}

func (this *TodoMySqlModel) Remind() (map[string][]Todo, error) {
	this.SetTimeZone()
	err := this.CreateTablesIfNotExist()
	if err != nil {
		return nil, err
	}
	userTodos := map[string][]Todo{}
	rows, err := this.db.Query("SELECT user_id, id, task, done, pin, due FROM todo ORDER BY user_id, done, pin DESC, due")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var id int
		var task string
		var done bool
		var pin bool
		var due time.Time
		if err := rows.Scan(&userID, &id, &task, &done, &pin, &due); err != nil {
			return nil, err
		}
		loc, _ := time.LoadLocation("Asia/Bangkok")
		due = due.In(loc)
		todo := Todo{
			ID:     id,
			UserID: userID,
			Task:   task,
			Pin:    pin,
			Done:   done,
			Due:    due,
		}
		log.Println(due)
		//Add to map
		todos := userTodos[userID]
		todos = append(todos, todo)
		userTodos[userID] = todos
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userTodos, nil
}

func (this *TodoMySqlModel) Edit(todo Todo) error {
	log.Println(todo)
	sql := `UPDATE todo SET task=?, due=? WHERE id=?`
	result, err := this.db.Exec(sql, todo.Task, todo.Due, todo.ID)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}
