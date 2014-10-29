function getImageService() {
	var service = {};

	service.send = function(message) {
		if (service.ws) return;

		if ("WebSocket" in window)
		{
			var socket = new WebSocket("ws://"+window.location.host+"/find/stream");

			socket.onopen = function(event) {
				console.log("Using websockets.");
				socket.send(message)
			};

			socket.onmessage = function(event) {
				source = JSON.parse(event.data);
				service.callback(source);
			};

			socket.onerror = function(event) {
				return;
			}

			socket.onclose = function(event) {
				service.ws = null;
			}

			service.ws = socket;
		}
	}

	service.subscribe = function(callback) {
		service.callback = callback;
	}

	return service;
}


angular.module('imgwaffle', [])

.factory('ImageService', getImageService)

.config(function($locationProvider) {
	$locationProvider.html5Mode(true);
})

.controller('imageList', ['$scope', '$location', 'ImageService', function($scope, $location, ImageService){
	$scope.images = [];
	$scope.hidensfw = true;
	$scope.max = 30;
	$scope.search = $location.path().substring(1, $location.path().length);
	$scope.lastSearch = '';

	ImageService.subscribe(function(message) {
		$scope.images.push(message);
		$scope.$apply();
	});

	$scope.get = function(message) {
		if (message != $scope.lastSearch)
		{
			$scope.max = 30;
			$scope.images = [];
			$location.path('/'+message);

			ImageService.send(message);
			$scope.lastSearch = message;
		}
	};

	$scope.blurOnEnter = function($event) {
		if ($event.keyCode != 13) return;
		$timeout(function() {$event.target.blur();}, 0, false);
	}

	if ($scope.search != '') {
		$scope.get($scope.search);
	}
}]);
