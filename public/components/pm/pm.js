'use strict';

angular.module('chat').config(['$stateProvider', function($stateProvider) {
	$stateProvider.state('pm', {
		url: '/pm?friendId',
		templateUrl: 'components/pm/pm.html',
		controller: ['$scope', '$state', '$stateParams', 'ws', 'appState', 'page', '$interval', function($scope, $state, $stateParams, ws, appState, page, $interval) {
			appState.state = $state;
			$scope.appState = appState;
			$scope.model = {
				friend : appState.friends.friends[$stateParams.friendId]
			};

			$scope.model.friend.seq = Math.floor(Math.random()*1000);
			$scope.model.friend.messagesNotDelivered = {};

			$scope.sendhandler = function(event) {
                if ((event.type === 'keypress' && event.keyCode===13) || event.type === 'click') {
                    $scope.sendmessage();
                }
            };

            $scope.sendmessage = function() {
                if ($scope.model.inputMessage) {

                    ws.send({
                        action: 'privateMessage',
                        data: {
                        	to: $stateParams.friendId,
                            from: $scope.model.friend.myId,
                            message: sjcl.encrypt(
                            	$scope.model.friend.password,
                            	$scope.model.inputMessage
                            ),
                            seq: $scope.model.friend.seq
                        }
                    });

                    $scope.model.friend.messagesNotDelivered[($scope.model.friend.seq++).toString()] = $scope.model.friend.messages.length;

                    $scope.model.friend.messages.push({
                    	from: 'me',
                    	message: $scope.model.inputMessage,
                    	delivered: false
                    });

                    $scope.model.inputMessage = '';
                }
            };
            
            var hiddenState = {
                unreadMessages: false,
                intervalPromise: null
            };

            $scope.$on('visibilityChanged', function(event,hidden) {
                if (!hidden && hiddenState.unreadMessages) {
                    hiddenState.unreadMessages = false;
                    $interval.cancel(hiddenState.intervalPromise);
                    page.setDefaultTitle();
                }
            });

            $scope.$on('newPmNotVisible', function() {
                if (!hiddenState.unreadMessages) {
                    hiddenState.unreadMessages = true;

                    var blank = true,
                        message = '*********************';
                    hiddenState.intervalPromise = $interval(function() {
                        if (blank) {
                            blank = false;
                            page.title = message;
                        } else {
                            blank = true;
                            page.title = '';
                        }
                    }, 1000);
                }
            });
		}]
	});
}]);
