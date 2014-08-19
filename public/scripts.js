var box = null;
var output = null;
var progress = null;
var status = null;

function createSetupFunction(id, loadedImage) {
	return function() {
		var imageTarget = $("#"+id);
		if (loadedImage.width < loadedImage.height) {
			imageTarget.css("width", "100%")
		} else {
			imageTarget.css("height", "100%")
		}

		$("#r"+id)
			.removeClass("loading")
			.show();

		imageTarget.attr("src", loadedImage.src)
			.css("margin-top", (imageTarget.parent().height()-imageTarget.height())/2)
			.hide()
			.css("visibility", "visible")
			.fadeIn("slow");
	}
}

function errorResult(interval) {
	if (interval) clearInterval(interval);
	progress.html("Try again later.");
}

function notFoundResult(interval) {
	if (interval) clearInterval(interval);
	progress.html("Couldn't find that one...");
}

function getUser(user) {
	if (user !== null && user !== "") {
		output.html('');
		progress.html("-").show();

		var interval = setInterval(function() {
			switch(progress.html()) {
				case "-":
					progress.html("\\");
					break;
				case "\\":
					progress.html("|");
					break;
				case "|":
					progress.html("/");
					break;
				case "/":
					progress.html("-");
					break;
				default:
					progress.html("-");
					break;
			}

			$("img.loading").each(function(){
				if ($(this).attr("src") === null || $(this).attr("src") === "" ) {
					$(this).attr("src", "loading.gif");
				}
			});
		}, 200 );

		if ("WebSocket" in window)
		{
			var socket = new WebSocket("ws://"+window.location.host+"/find/stream");
			socket.onopen = function(event) {
				console.log("Using websockets.");
				socket.send(user);
			};

			socket.onmessage = function(event) {
				clearInterval(interval);
				progress.hide();

				source = JSON.parse(event.data);

				img = new Image();
				img.src = source.thumbnail;

				output.append('<div class="resultBox" id="r'+source.id+'">'+
				'<a target="_blank" class="imgBox" href="'+source.url+'"><img id="'+source.id+'"/></a>'+
				'</div>');
				$(img).load(createSetupFunction(source.id, img));
				$(img).error(function() { $("#r"+source.id).remove()})
			};

			socket.onerror = function(event) {
				errorResult(interval);
			}

			socket.onclose = function(event) {
				if (output.children().length === 0) {
					notFoundResult(interval);
				}
			}

			return
		}

		var address = "/find/"+encodeURIComponent(user);

		console.log("Using http.")

		$.get(address)
		.done(function(data) {
			clearInterval(interval);
			progress.hide();

			var imageKeys = Object.keys(data.images);
			if(imageKeys.length > 0) {
				for (var i = 0; i < imageKeys.length; i++) {
					var source = data.images[imageKeys[i]];
					img = new Image();
					img.src = source.thumbnail;

					output.append('<div class="resultBox" id="r'+source.id+'">'+
					'<a target="_blank" class="imgBox" href="'+source.url+'"><img id="'+source.id+'"/></a>'+
					'</div>');
					$(img).load(createSetupFunction(source.id, img));
					$(img).error(function() { $("#r"+source.id).remove()})
				}
			}
			else
			{
				notFoundResult(null);
			}
		})
		.fail(function(data) {
			errorResult(null);
		});
	}
}

function refreshPage() {
	var curUser = decodeURIComponent(window.location.hash.slice(1));
	if (curUser && curUser !== "") {
		getUser(curUser);
		box.val(curUser)
	}
}

$(function () {
	box = $('#username');
	output = $('#output');
	progress = $('#progress')

	refreshPage();

	$(window).on('hashchange', refreshPage);

	box.change(function() {
		window.location.hash = encodeURIComponent(box.val());
	});

	box.keypress(function(e) {
		if(e.which == 13) {
			window.location.hash = encodeURIComponent(box.val());
		}
	});
});
