angular.module('todoApp', [])
  .controller('TodoListController', function ($scope, $http) {
    var todoList = this;
    todoList.isWorking = true;
    $http.get('/user-info').then(function (response) {
      todoList.user = response.data;
      todoList.isWorking = false;
    });
    todoList.isWorking = true;
    $http.get('/list').then(function (response) {
      todoList.todos = response.data;
      todoList.isWorking = false;
    });

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
        if (todo.Pin) tasks.push(todo);
      });
      return sortByDue(tasks);
    };

    todoList.nonPinTasks = function () {
      var tasks = [];
      angular.forEach(todoList.todos, function (todo) {
        if (!todo.Pin) tasks.push(todo);
      });
      return sortByDue(tasks);
    };

    todoList.setDone = function (id, status) {
      todoList.isWorking = true;
      var data = {
        "ID": id,
        "Done": status
      };
      $http.post('/done', data).then(function (response) {
        todoList.isWorking = false;
      });
    };

    todoList.setPin = function (id, status) {
      todoList.isWorking = true;
      var data = {
        "ID": id,
        "Pin": status
      };
      $http.post('/pin', data).then(function (response) {
        todoList.isWorking = false;
      });
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
      dateString = moment(date).calendar(Date.now(), {
        sameDay: '[Today] [at] H:mm',
        nextDay: '[Tomorrow] [at] H:mm',
        nextWeek: 'dddd [at] H:mm',
        lastDay: '[Yesterday] [at] H:mm',
        lastWeek: '[Last] dddd [at] H:mm',
        sameElse: 'ddd D MMM YY [at] H:mm'
      });
      return dateString
    };

    todoList.isOverdue = function(todo) {
      if(!todo.Done && (new Date()) > new Date(todo.Due)) {
        return "(overdue)";
      }
      return "";
    }

  });