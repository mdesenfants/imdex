var box = $('#username');
var output = $('#output');

function createSetupFunction(id, loadedImage) {
  return function() {
    var imageTarget = $("#"+id);
    if (loadedImage.width == loadedImage.height) {
      imageTarget.attr("src", loadedImage.src)
        .addClass("loaded")
        .removeClass("loading")
        .fadeIn("slow");
    } else {
      imageTarget.remove();
    }
  }
}

function getUser(user) {
  if (user !== null && user !== "") {
    var address = "/find/"+encodeURIComponent(user);
    $.get(address)
    .done(function(data) {
      var imageKeys = Object.keys(data.images);
      if(imageKeys.length > 0) {
        for (var i = 0; i < imageKeys.length; i++) {
          var source = data.images[imageKeys[i]];
          img = new Image();
          img.src = source.thumbnail;

          output.append('<a class="container" href="'+source.url+'"><img class="loading" id="'+source.id+'"/></a>');
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
      output.html("Try again later.")
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
  refreshPage();

  $(window).on('hashchange', refreshPage);

  box.change(function() {
    output.fadeOut("slow", function() {
      output.html("");
      window.location.hash = encodeURIComponent(box.val());
    }).fadeIn("slow");
  });
});
