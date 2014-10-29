createLinks = function() {
  author = $(this);
  margin = author.css("margin-right");
  $(this).after('<a class="waffle" target="_blank" style="margin-right: '+margin+'" href="http://imgwaffle.com/'+author.html()+'">(imgwaffle)</a>');
}

$("a.author").each(createLinks);

$(window).on('hashchange', function() {
  $(".waffle").remove();
  $("a.author").each(createLinks)
});
