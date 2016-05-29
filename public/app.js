'use strict';

angular.module('chat', [
  'ngAnimate',
  'ui.router'
])

.config(['$urlRouterProvider', function($urlRouterProvider) {
    $urlRouterProvider.otherwise('/lobby');
}])

.controller('topnavibar', ['$scope', 'ws', 'navigation', 'appState',
    function($scope, ws, navigation, appState) {
        $scope.appState = appState;

        $scope.model = {
            clientcount: 0,
            newPmMessage: false
        };

        ws.on('clientcount', function(data) {
            $scope.model.clientcount = data;
        });

        $scope.$on('newPmNotVisible', function() {
            $scope.model.newPmMessage = true;
            console.log($scope.model);
        });

        $scope.$on('visibilityChanged', function(e, hidden) {
            if (!hidden) {
                $scope.model.newPmMessage = false;
            }
        });

        $scope.togglelobbysettings = function() {
            navigation.togglelobbysettingsvisibility();
        };

        $scope.toggle = function(prop) {
            appState[appState.state.current.name][prop] = !appState[appState.state.current.name][prop];
        };

        $scope.toggleDeleteFriends = function() {
            appState.friends.deleteFriendsVisible = !appState.friends.deleteFriendsVisible;
        };
    }
])

.controller('htmlCtrl', ['$scope', '$rootScope', 'page', 'visibilityApiService', function($scope, $rootScope, page, visibilityApiService) {
    $scope.model = page;
}])

.factory('page', function() {
    var defaultTitle = '------------------------';

    return self = {
        title: defaultTitle,
        hasFocus: true,
        setDefaultTitle: function() {
            self.title = defaultTitle;
        }
    };
})

.service('navigation', ['$rootScope',
    function($rootScope) {
        this.togglelobbysettingsvisibility = function() {
            $rootScope.$broadcast('togglelobbysettingsvisibility');
        };
    }
])

.factory('popSound', function() {
    var pop = new Audio('pop.ogg');

    return {
        pop: function() {
            pop.play();
        }
    };
})

.factory('visibilityApiService', ['$rootScope', 'page', function($rootScope, page) {
    document.addEventListener('visibilitychange', function() {
        $rootScope.$broadcast('visibilityChanged', document.hidden);
    });

    window.onblur = function() {
        page.hasFocus = false;
    };

    window.onfocus = function() {
        page.hasFocus = true;
        $rootScope.$broadcast('visibilityChanged', document.hidden);
    };
    return {};
}])

.factory('ws', ['$rootScope', 'appState', 'popSound', 'page',
    function($rootScope, appState, popSound, page) {

        var ws = new WebSocket('ws://' + window.location.host + '/ws'),
            actions = {

                lobbymessage: function(message) {
                    appState.lobby.messages.push(message);
                    popSound.pop();
                },

                privateMessage: function(message) {
                    var friend = appState.friends.friends[message.from];

                    if (friend) {
                        friend.messages.push({
                            from: friend.name,
                            to: 'me',
                            message: sjcl.decrypt(friend.password, message.message)
                        });

                        if (document.hidden || !page.hasFocus) {
                            $rootScope.$broadcast('newPmNotVisible');
                        }

                        send({
                            action: 'privateMessageDelivered',
                            data: {
                                to: message.from,
                                from: message.to,
                                seq: message.seq
                            }
                        });
                        popSound.pop();
                    }
                },

                privateMessageDelivered: function(message) {
                    var friend = appState.friends.friends[message.from];

                    if (friend) {
                        friend.messages[friend.messagesNotDelivered[message.seq]].delivered = true;
                        delete friend.messagesNotDelivered[message.seq];
                    }
                }

            }
        ;

        ws.onmessage = function(event) {
            var message;
            try {
                message = JSON.parse(event.data);
            } catch(e) {
                console.error(e.message);
                console.error('raw message: ' + event);
                return;
            }
            console.log(message);

            if (typeof message === 'object' && message.action && message.data) {
                if (actions[message.action]) {
                    $rootScope.$apply(function() {
                        actions[message.action](message.data);
                    });
                } else {
                    throw new Error('WebSocket error, no registered action exists for "' + message.action + '"');
                }
            } else {
                throw new Error('WebSocket error, malformed incoming message');
            }
        };

        ws.onopen = function() {
            var myIds = [];
            for (var id in appState.friends.friends) {
                myIds.push(appState.friends.friends[id].myId);
            }

            if (myIds.length) {
                send({
                    action: 'registerIds',
                    data: myIds
                });
            }
        };

        function send(message) {
            if (typeof message === 'object' && message.action && message.data) {
                ws.send(JSON.stringify(message));
            } else {
                throw new Error('WebSocket error, no action or data properties provided');
            }
        }
            
        return {
            on: function(name, cb) {
                actions[name] = cb;
            },
            send: send
        };
    }
])

.directive('autoscrolldown', function() {
    return {
        link: function(scope, element, attrs) {
            scope.$watch(function() {
                return element[0].scrollHeight;
            }, function(newval, oldval) {
                element[0].scrollTop = newval;
            });
        }
    };
})

.factory('appState', [ 
    function() {
        var friends = localStorage.getItem('friends');
        friends = friends ? JSON.parse(friends) : {};

        return {
            state: null,
            lobby: {
                messages: []
            },
            friends: {
                friends: friends,
                deleteFriendsVisible: false,
                addFriendVisible: false
            }
        };
    }
]);
