"use strict";
angular.module('todoApp', [])
  .controller('TodoListController', function ($scope, $http) {
    var todoList = this;
    todoList.editTodo = {};
    todoList.deleteTodo = {};
    todoList.editDue = "";
    showWorking();
    $http.get('/user-info')
      .then(function (response) {
        todoList.user = response.data;
        hideWorking();
      })
      .catch(hideWorking);
    showWorking();
    $http.get('/list')
      .then(function (response) {
        todoList.todos = response.data;
        hideWorking();
      })
      .catch(hideWorking);

    todoList.remaining = function () {
      var count = 0;
      angular.forEach(todoList.todos, function (todo) {
        count += todo.Done ? 0 : 1;
      });

      return count;
    };

    todoList.pinTasks = function () {
      var tasks = [];
      angular.forEach(todoList.todos, function (todo) {
        if (todo.Pin) {
          tasks.push(todo);
        }
      });
      return sortByDue(tasks);
    };

    todoList.nonPinTasks = function () {
      var tasks = [];
      angular.forEach(todoList.todos, function (todo) {
        if (!todo.Pin) {
          tasks.push(todo);
        }
      });
      return sortByDue(tasks);
    };

    todoList.setDone = function (id, status) {
      showWorking();
      var data = {
        "ID": id,
        "Done": status
      };
      $http.post('/done', data)
        .then(hideWorking)
        .catch(hideWorking);
    };

    todoList.setPin = function (id, status) {
      showWorking();
      var data = {
        "ID": id,
        "Pin": status
      };
      $http.post('/pin', data)
        .then(hideWorking)
        .catch(hideWorking);
    };

    function sortByDue(tasks) {
      return tasks.sort(function (a, b) {
        if (a.Due < b.Due) {
          return -1;
        } else if (a.Due > b.Due) {
          return 1;
        }
        return 0;
      });
    }

    todoList.formatDate = function (date) {
      var dateString = moment(date).calendar(Date.now(), {
        sameDay: '[Today] [at] H:mm',
        nextDay: '[Tomorrow] [at] H:mm',
        nextWeek: 'dddd [at] H:mm',
        lastDay: '[Yesterday] [at] H:mm',
        lastWeek: '[Last] dddd [at] H:mm',
        sameElse: 'ddd D MMM YY [at] H:mm'
      });
      return dateString;
    };

    todoList.isOverdue = function (todo) {
      if (!todo.Done && (new Date()) > new Date(todo.Due)) {
        return "(overdue)";
      }
      return "";
    }

    function formatDateInput(date) {
      return moment(date).format('YYYY-MM-DD[T]HH:mm');
    };

    todoList.toEdit = function (todo) {
      todoList.editTodo = todo;
      todoList.editDue = formatDateInput(todo.Due);
    };

    todoList.edit = function () {
      showWorking();
      todoList.editTodo.Task = $("#task-input").val();
      todoList.editDue = $("#due-input").val();
      todoList.editTodo.Due = new Date(todoList.editDue);
      var data = {
        "ID": todoList.editTodo.ID,
        "Task": todoList.editTodo.Task,
        "Due": todoList.editTodo.Due
      };
      $http.post('/edit', data)
        .then(function () {
          todoList.updateEditTodoLocal(todoList.editTodo);
          hideWorking();
        })
        .catch(hideWorking);
    };

    todoList.updateEditTodoLocal = function(editTodo) {
      for (var i = 0; i < todoList.todos.length; i++) {
        if (todoList.todos[i].ID == editTodo.ID) {
          console.log(editTodo);
          console.log(todoList.todos[i]);
          todoList.todos[i] = editTodo;
          return;
        }
      }
    }

    function showWorking() {
      todoList.isWorking = true;
    }

    function hideWorking() {
      todoList.isWorking = false;
    }

    todoList.toDelete = function (todo) {
      todoList.deleteTodo = todo;
    };

    todoList.delete = function () {
      showWorking();
      var data = {
        "ID": todoList.deleteTodo.ID
      };
      $http.post('/delete', data)
        .then(function () {
          todoList.updateDeleteTodoLocal(todoList.deleteTodo);
          hideWorking();
        })
        .catch(hideWorking);
    };

    todoList.updateDeleteTodoLocal = function(deleteTodo) {
      for (var i = 0; i < todoList.todos.length; i++) {
        if (todoList.todos[i].ID == deleteTodo.ID) {
          todoList.todos.splice(i, 1);
          return;
        }
      }
    }

  });