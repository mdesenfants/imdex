var box = null;
var output = null;
var progress = null;
var status = null;
var checkbox = null;
var hideNSFWstatus = true;

function createSetupFunction(id, loadedImage, nsfw) {
	return function() {
		var imageTarget = $("#"+id);
		var responseTarget = $("#r"+id)
		if (loadedImage.width < loadedImage.height) {
			imageTarget.css("width", "100%")
		} else {
			imageTarget.css("height", "100%")
		}

		responseTarget.removeClass("loading");

		if (imageTarget.hasClass("nsfw")) {
			imageTarget.attr("actualSrc", loadedImage.src)
		}

		if(hideNSFWstatus && imageTarget.hasClass("nsfw"))
		{
			imageTarget.attr("src", "nsfw.gif");
		} else {
			imageTarget.attr("src", loadedImage.src);
		}

		imageTarget.css("margin-top", -Math.abs(imageTarget.parent().height()-imageTarget.height())/2).hide().css("visibility", "visible");

		imageTarget.fadeIn("slow");
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

function insertImage(source) {
	var img = new Image();
	img.src = source.thumbnail;

	output.append('<div class="resultBox" id="r'+source.id+'"><a target="_blank" class="imgBox" href="'+source.url+'"><img id="'+source.id+'" '+(source.nsfw ? 'class="nsfw"':'')+'"/></a>'+
	'</div>');
	$(img).load(createSetupFunction(source.id, img));
	$(img).error(function() { $("#r"+source.id).remove()})
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
				insertImage(source);
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

					insertImage(source);
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

function getCookie(key) {
	var cookie = ";"+document.cookie;
	var values = cookie.split(";");
	for (var i = 0; i < values.length; i++) {
		var pair = values[i].split("=");
		if (pair[0].trim() == key.trim()) {
			return pair[1].trim();
		}
	}

	return ""
}

function showNSFW() {
	hideNSFWstatus = false;
	var date = new Date();
	date.setTime(date.getTime() + 30*24*60*60*1000)
	document.cookie="hidensfw=false; expires="+date.toGMTString();
	$(".nsfw").each(function() {
		$(this).attr("src", $(this).attr("actualsrc"))
			.css("margin-top", ($(this).parent().height()-$(this).height())/2);
	});
}

function hideNSFW() {
	hideNSFWstatus = true;
	document.cookie="hidensfw=true; expires=Thu, 01 Jan 1970 00:00:00 GMT";
	$(".nsfw").each(function() {
		$(this).attr("src", "nsfw.gif")
			.css("margin-top", ($(this).parent().height()-$(this).height())/2);
	});
}

$(function () {
	box = $('#username');
	output = $('#output');
	progress = $('#progress');
	checkbox = $('#nsfw');


	hideNSFWstatus = getCookie("hidensfw") == "";

	checkbox[0].checked = hideNSFWstatus;

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

	checkbox.change(function() {
		if ($(this).is(":checked")) {
			hideNSFW();
		} else {
			showNSFW();
		}
	});
});
