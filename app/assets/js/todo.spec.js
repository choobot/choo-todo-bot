describe('todoApp', function () {

    beforeEach(module('todoApp'));

    var $controller, $rootScope, $httpBackend;

    beforeEach(inject(function ($injector, _$controller_, _$rootScope_) {
        // The injector unwraps the underscores (_) from around the parameter names when matching
        $controller = _$controller_;
        $rootScope = _$rootScope_;
        $httpBackend = $injector.get('$httpBackend')
        var reqHandler = $httpBackend.when('GET', '/list')
            .respond([
                {
                    "Id": 1,
                    "Task": "Task 1",
                    "Done": true,
                    "Pin": false,
                    "Due": "2018-11-08T12:27:00+07:00"
                },
                {
                    "Id": 2,
                    "Task": "Task 2",
                    "Done": false,
                    "Pin": false,
                    "Due": "2018-11-12T12:27:00+07:00"
                },
                {
                    "Id": 3,
                    "Task": "Task 3",
                    "Done": true,
                    "Pin": false,
                    "Due": "2018-11-11T12:27:00+07:00"
                },
                {
                    "Id": 4,
                    "Task": "Task 4",
                    "Done": false,
                    "Pin": true,
                    "Due": "2018-11-10T12:27:00+07:00"
                },
                {
                    "Id": 5,
                    "Task": "Task 5",
                    "Done": false,
                    "Pin": true,
                    "Due": "2018-11-09T12:27:00+07:00"
                }
            ]);
        $httpBackend.when('POST', '/pin')
            .respond();
        $httpBackend.when('POST', '/done')
            .respond();
        $httpBackend.when('POST', '/edit')
            .respond();
        $httpBackend.when('POST', '/delete')
            .respond();
        $httpBackend.when('GET', '/user-info')
            .respond({
                "oauthPicture": "oauthPicture",
                "oauthName": "oauthName"
            });
    }));

    afterEach(function () {
        $httpBackend.verifyNoOutstandingExpectation();
        $httpBackend.verifyNoOutstandingRequest();
    });

    describe('TodoListController', function () {
        it('shoud be created and loading a todo list', function () {
            var todoList = $controller('TodoListController', { $scope: $rootScope });
            $httpBackend.flush();
            expect(todoList).toBeDefined();
        });

        it('shoud get /user-info', function () {
            $httpBackend.expectGET('/user-info');
            var todoList = $controller('TodoListController', { $scope: $rootScope });
            $httpBackend.flush();
        });

        it('shoud get /list', function () {
            $httpBackend.expectGET('/list');
            var todoList = $controller('TodoListController', { $scope: $rootScope });
            $httpBackend.flush();
        });

        it('shoud load data to todos', function () {
            var todoList = $controller('TodoListController', { $scope: $rootScope });
            $httpBackend.flush();
            var tasks = todoList.todos.length;
            expect(tasks).toEqual(5);
        });

        describe('setPin(id)', function () {
            it('shoud post to /pin', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                $httpBackend.expectPOST('/pin');
                todoList.setPin('1');
                $httpBackend.flush();
            });
        });

        describe('setDone(id)', function () {
            it('shoud post to /done', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                $httpBackend.expectPOST('/done');
                todoList.setDone('1');
                $httpBackend.flush();
            });
        });

        describe('remaining()', function () {
            it('shoud return number of remaining tasks', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var tasks = todoList.remaining();
                expect(tasks).toEqual(3);
            });
        });

        describe('pinTasks()', function () {
            it('shoud return pin tasks', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var tasks = todoList.pinTasks();
                expect(tasks.length).toEqual(2);
            });
            it('shoud return sorted tasks by due date/time', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var tasks = todoList.pinTasks();
                expect(tasks).toEqual([
                    {
                        "Id": 5,
                        "Task": "Task 5",
                        "Done": false,
                        "Pin": true,
                        "Due": "2018-11-09T12:27:00+07:00"
                    },
                    {
                        "Id": 4,
                        "Task": "Task 4",
                        "Done": false,
                        "Pin": true,
                        "Due": "2018-11-10T12:27:00+07:00"
                    }

                ]);
            });
        });

        describe('nonPinTasks()', function () {
            it('shoud return non pin tasks', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var tasks = todoList.nonPinTasks();
                expect(tasks.length).toEqual(3);
            });
            it('shoud return sorted tasks by due date/time', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var tasks = todoList.nonPinTasks();
                expect(tasks).toEqual([
                    {
                        "Id": 1,
                        "Task": "Task 1",
                        "Done": true,
                        "Pin": false,
                        "Due": "2018-11-08T12:27:00+07:00"
                    },
                    {
                        "Id": 3,
                        "Task": "Task 3",
                        "Done": true,
                        "Pin": false,
                        "Due": "2018-11-11T12:27:00+07:00"
                    },
                    {
                        "Id": 2,
                        "Task": "Task 2",
                        "Done": false,
                        "Pin": false,
                        "Due": "2018-11-12T12:27:00+07:00"
                    }
                ]);
            });
        });

        describe('isOverdue()', function () {
            it('shoud return (overdue) if remaining task due date is in the past', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var todo = {
                    Done: false,
                    Due: "1980-11-18T12:00:00+07:00"
                };
                var result = todoList.isOverdue(todo);
                expect(result).toEqual("(overdue)");
            });
            it('shoud return blank if remaining task due date is in the future', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var todo = {
                    Done: false,
                    Due: "2200-11-18T12:00:00+07:00"
                };
                var result = todoList.isOverdue(todo);
                expect(result).toEqual("");
            });
            it('shoud return blank if completed task', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var todo = {
                    Done: true,
                    Due: "1980-11-18T12:00:00+07:00"
                };
                var result = todoList.isOverdue(todo);
                expect(result).toEqual("");
            });

        });

        describe('toEdit()', function () {
            it('shoud copy todo to editing todo', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var todo = {
                    ID: 1,
                    Task: "dummy",
                    Due: "1980-11-18T13:00:00+07:00"
                };
                todoList.toEdit(todo);
                expect(todoList.editTodo).toEqual(todo);
            });
            it('shoud copy todo due to editing due with a proper format for HTML5 input', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var todo = {
                    ID: 1,
                    Task: "dummy",
                    Due: "1980-11-18T13:00:00+00:00"
                };
                todoList.toEdit(todo);
                expect(todoList.editDue).toEqual("1980-11-18T13:00");
            });
        });

        describe('toDelete()', function () {
            it('shoud copy todo to deleting todo', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                var todo = {
                    ID: 1,
                    Task: "dummy",
                    Due: "1980-11-18T12:00:00+07:00"
                };
                todoList.toDelete(todo);
                expect(todoList.deleteTodo).toEqual(todo);
            });
        });

        describe('delete(id)', function () {
            it('shoud post to /delete', function () {
                var todoList = $controller('TodoListController', { $scope: $rootScope });
                $httpBackend.flush();
                $httpBackend.expectPOST('/delete');
                var todo = {
                    ID: 1,
                    Task: "dummy",
                    Due: "1980-11-18T12:00:00+07:00"
                };
                todoList.delete(todo);
                $httpBackend.flush();
            });
        });
    });
});