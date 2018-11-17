angular.module('todoApp', [])
  .controller('TodoListController', function ($scope, $http) {
    var todoList = this;
    var workingCount = 0;
    addWorking();
    $http.get('/user-info').then(function (response) {
      todoList.user = response.data;
      doneWorking();
    });
    addWorking();
    $http.get('/list').then(function (response) {
      todoList.todos = response.data;
      doneWorking();
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
      addWorking();
      var data = {
        "ID": id,
        "Done": status
      };
      $http.post('/done', data).then(function (response) {
        doneWorking();
      });
    };

    todoList.setPin = function (id, status) {
      addWorking();
      var data = {
        "ID": id,
        "Pin": status
      };
      $http.post('/pin', data).then(function (response) {
        doneWorking();
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

    function addWorking() {
      workingCount++;
      document.getElementById("working").style.display = "inline";
    }

    function doneWorking() {
      workingCount--;
      if (workingCount <= 0) {
        document.getElementById("working").style.display = "none";
      }
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