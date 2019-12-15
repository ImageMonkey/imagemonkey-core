var AutoCompletion = (function() {
    function split(val) {
      return val.split( / \s*/ );
    }
	
    function extractLast(term) {
      return split(term).pop();
    }
	
    function AutoCompletion(id, entries) {
        var maxResults = 10;
        $(id)
            // don't navigate away from the field on tab when selecting an item
            .on("keydown", function(event) {
                if (event.keyCode === $.ui.keyCode.TAB &&
                    $(this).autocomplete("instance").menu.active) {
                    event.preventDefault();
                }
            })
            .autocomplete({
                minLength: 3,
                delay: 500,
                source: function(request, response) {
                    //delegate back to autocomplete, but extract the last term
                    var results = $.ui.autocomplete.filter(entries, extractLast(request.term));
                    response(results.slice(0, maxResults));
                },
                search: function(e, ui) {
                    //see https://stackoverflow.com/questions/40782638/jquery-autocomplete-performance-going-down-with-each-search
                    $(this).data("ui-autocomplete").menu.bindings = $();
                },
                focus: function() {
                    // prevent value inserted on focus
                    return false;
                },
                select: function(event, ui) {
                    var terms = split(this.value);
                    // remove the current input
                    terms.pop();
                    // add the selected item
                    terms.push(ui.item.value);
                    // add placeholder to get the comma-and-space at the end
                    terms.push("");
                    this.value = terms.join("");
                    return false;
                },
                response: function(event, ui) {
                    //Check if we have more than "maxResults" results
                    if (ui.content.length > maxResults) {
                        //Remove all elements until there are only maxResults remaining (the use of slice() was not supported)
                        while (ui.content.length > maxResults) {
                            ui.content.pop();
                        }
                        //Add message
                        ui.content.push({
                            'label': 'Please narrow down your search',
                            'value': ''
                        });
                    }
                }
            }).data("ui-autocomplete")._renderItem = function(ul, item) {
                //Add the .ui-state-disabled class and don't wrap in <a> if value is empty
                if (item.value == '') {
                    return $('<li class="ui-state-disabled">' + item.label + '</li>').appendTo(ul);
                } else {
                    return $("<li>")
                        .append($("<div>").text(item.label))
                        .appendTo(ul);
                }
            }
    }
    return AutoCompletion;
}());
