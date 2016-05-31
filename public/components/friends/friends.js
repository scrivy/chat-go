'use strict';

angular.module('chat').config(['$stateProvider', function($stateProvider) {
    $stateProvider.state('friends', {
        url: '/friends',
        templateUrl: 'components/friends/friends.html',
        controller: ['$scope', '$state', 'ws', 'appState', function($scope, $state, ws, appState) {
            appState.state = $state;
            var friends = appState.friends.friends;
            $scope.appState = appState;

            $scope.model = {};
            resetAddFriend();

            $scope.sendRequest = function() {
                if ($scope.model.addFriend.friendsId) {
                    ws.send({
                        action: 'addFriend',
                        data: {
                            from: $scope.model.addFriend.requestId,
                            to: +$scope.model.addFriend.friendsId
                        }
                    });
                }
            };

            $scope.testMessage = function() {
                ws.send({
                    action: 'testMessage',
                    data: {
                        to: +$scope.model.addFriend.friendsId,
                        message: sjcl.encrypt(
                            $scope.model.addFriend.password,
                            JSON.stringify({
                                test: 'test',
                                friendsId: $scope.model.addFriend.myId
                            })
                        )
                    }
                });
            };

            $scope.deleteFriend = function(friendId) {
                delete friends[friendId];
                saveFriends();
            };

            ws.on('addFriend', function(data) {
                $scope.model.addFriend.inputAndButtonVisible = false;
                if (data.exists) {
                    $scope.model.addFriend.passInputVisible = true;
                } else {
                    $scope.model.addFriend.password = genMD5();
                    $scope.model.addFriend.passwordVisible = true;
                }
            });

            ws.on('testMessage', function(data) {
                var decrypted = JSON.parse(sjcl.decrypt($scope.model.addFriend.password, data.message));
                if (decrypted.test === 'test') {

                    friends[decrypted.friendsId] = {
                        name: $scope.model.addFriend.friendsName,
                        myId: $scope.model.addFriend.myId,
                        password: $scope.model.addFriend.password,
                        messages: []
                    };

                    saveFriends();
                    $scope.testMessage();
                    ws.send({
                        action: 'registerIds',
                        data: [$scope.model.addFriend.myId]
                    });
                    resetAddFriend();
                }
            });

            function genMD5() {
                return CryptoJS.MD5(Math.random().toString() + Math.random().toString() + Math.random().toString() + Math.random().toString()).toString();
            }

            function resetAddFriend() {
                $scope.model.addFriend = {
                    requestId : Math.floor(Math.random() * 1000),
                    friendsName: null,
                    myId : genMD5(),
                    password : null,
                    inputAndButtonVisible : true,
                    passInputVisible : false
                };

                appState.friends.addFriendVisible = false;
            }

            function saveFriends() {
                localStorage.setItem(
                    'friends',
                    JSON.stringify(friends)
                );
            }
        }]
    });
}]);
