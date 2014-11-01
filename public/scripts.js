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

.controller('image', ['$scope', function($scope) {
	$scope.showMenu = false;
	$scope.animated = false;
	$scope.activeImage = $scope.image.thumbnail;

	$scope.setNewImage = function() {
		$scope.animated = !$scope.animated;
		$scope.activeImage = $scope.animated ? $scope.image.animated : $scope.image.thumbnail;
	};
}])

.controller('imageList', ['$scope', '$location', 'ImageService', function($scope, $location, ImageService){
	var emSize = parseFloat(getComputedStyle(document.getElementsByTagName("body")[0], null)["font-size"]);
	var boxSize = (15 * emSize) + (emSize * 1.4);
	var boxesPerRow = Math.floor($(window).width() / boxSize) - 1;

	$scope.images = [];
	$scope.hidensfw = true;
	$scope.max = boxesPerRow * 3;
	$scope.lastSearch = '';
	$scope.searching = false;

	ImageService.subscribe(function(message) {
		$scope.searching = false;
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
			$scope.searching = true;
		}
	};

	$scope.getSearch = function() {
		return $location.path().substring(1, $location.path().length);
	};

	$scope.blurOnEnter = function($event) {
		if ($event.keyCode != 13) return;
		$timeout(function() {$event.target.blur();}, 0, false);
	}

	$scope.$watch(function() { return $location.path(); }, function(url) {
		$scope.search = $scope.getSearch();
		$scope.get($scope.search);
	});

	$scope.search = $scope.getSearch();
	if ($scope.search != '') {
		$scope.get($scope.search);
	}

	$(window).scroll(function() {
		if ($scope.max < $scope.images.length && $(window).scrollTop() + $(window).height() == $(document).height()) {
			$scope.max += boxesPerRow * 2;
			$scope.$apply();
		}
	});
}]);
