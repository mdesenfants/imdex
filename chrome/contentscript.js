$("a.author").each(function() {
  author = $(this);
  margin = author.css("margin-right");
  $(this).after('<a target="_blank" style="margin-right: '+margin+'" href="http://localhost:3000/'+author.html()+'">(imdex)</a>');
});
