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

.controller('imageList', ['$scope', 'ImageService', function($scope, ImageService){
	$scope.images = [];
	$scope.hidensfw = true;
	$scope.max = 30;

	ImageService.subscribe(function(message) {
		$scope.images.push(message);
		$scope.$apply();
	});

	$scope.get = function(message) {
		$scope.max = 30;
		$scope.images = [];
		ImageService.send(message);
	};
}]);
