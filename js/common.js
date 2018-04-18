$(document)
    .ready(function() {

      // fix menu when passed
      $('.masthead')
        .visibility({
          once: false,
          onBottomPassed: function() {
            $('.fixed.menu').transition('fade in');
          },
          onBottomPassedReverse: function() {
            $('.fixed.menu').transition('fade out');
          }
        })
      ;

      // create sidebar and attach to menu open
      $('.ui.sidebar')
        .sidebar('attach events', '.toc.item')
      ;

    });

function escapeHtml (str) {
  return jQuery('<div/>').text(str).html();
}

function getCookie(s) {
  var cookie = Cookies.get(s);
  if(typeof cookie == "undefined"){
    return "";
  }
  return cookie;
}