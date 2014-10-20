$("a.author").each(function() {
  author = $(this);
  margin = author.css("margin-right");
  $(this).after('<a target="_blank" style="margin-right: '+margin+'" href="http://imgwaffle.com/'+author.html()+'">(imgwaffle)</a>');
});
