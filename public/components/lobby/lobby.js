'use strict';

angular.module('chat').config(['$stateProvider', function($stateProvider) {
    $stateProvider.state('lobby', {
        url: '/lobby',
        templateUrl: 'components/lobby/lobby.html',
        controller: ['$scope', '$state', 'ws', 'appState', function($scope, $state, ws, appState) {
            appState.state = $state;
            $scope.appState = appState;

            if (!localStorage) throw new Error('web storage required');
            $scope.model = {
                username : localStorage.getItem('lobbyusername')
            };

            $scope.model.setupvisible = $scope.model.username ? false : true;

            $scope.sendhandler = function(event) {
                if ((event.type === 'keypress' && event.keyCode===13) || event.type === 'click') {
                    $scope.sendmessage();
                }
            };

            $scope.sendmessage = function() {
                if ($scope.model.inputmessage) {
                    ws.send({
                        action: 'lobbymessage',
                        data: {
                            username: $scope.model.username,
                            message: $scope.model.inputmessage
                        }
                    });

                    $scope.model.inputmessage = '';
                }
            };

            $scope.sessionsetuphandler = function(event) {
                if ((event.type === 'keypress' && event.keyCode===13) || (event.type === 'click')) {
                    $scope.model.setupvisible = false;
                    localStorage.setItem('lobbyusername', $scope.model.username);
                }
            };

            $scope.setuptoggle = function() {
                $scope.model.setupvisible = !$scope.model.setupvisible;
            };

            $scope.$on('togglelobbysettingsvisibility', $scope.setuptoggle);
        }]
    });
}]);