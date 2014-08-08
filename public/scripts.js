var box = null;
var output = null;
var status = null;

function createSetupFunction(id, loadedImage) {
	return function() {
		var imageTarget = $("#"+id);
		if (loadedImage.width < loadedImage.height) {
			imageTarget.css("width", "100%")
		} else {
			imageTarget.css("height", "100%")
		}
		$("#r"+id).show();
		imageTarget.attr("src", loadedImage.src)
			.addClass("loaded")
			.removeClass("loading")
			.fadeIn("slow");
	}
}

function getUser(user) {
	if (user !== null && user !== "") {
		var address = "/find/"+encodeURIComponent(user);
		output.html("-")

		var interval = setInterval(function() {
			switch(output.html()) {
				case "-":
					output.html("\\");
					break;
				case "\\":
					output.html("|");
					break;
				case "|":
					output.html("/");
					break;
				case "/":
					output.html("-");
					break;
				default:
					output.html("-");
					break;
			}

			$("img.loading").each(function(){
				if ($(this).attr("src") === null || $(this).attr("src") === "" ) {
					$(this).attr("src", "loading.gif");
				}
			});
		}, 200 );

		$.get(address)
		.done(function(data) {
			clearInterval(interval);
			output.html("");

			var imageKeys = Object.keys(data.images);
			if(imageKeys.length > 0) {
				for (var i = 0; i < imageKeys.length; i++) {
					var source = data.images[imageKeys[i]];
					img = new Image();
					img.src = source.thumbnail;

					output.append('<div class="resultBox" id="r'+source.id+'">'+
					'<a target="_blank" class="imgBox" href="'+source.url+'"><img class="loading" id="'+source.id+'"/></a>'+
					'</div>');
					$(img).load(createSetupFunction(source.id, img));
				}
			}
			else
			{
				output.html("Couldn't find that one...");
			}

			$("img.loading").error(function() { $(this).remove() });
		})
		.fail(function(data) {
			clearInterval(interval);
			output.html("Try again later.");
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
