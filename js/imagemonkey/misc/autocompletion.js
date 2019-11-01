var AutoCompletion = (function() {
    function AutoCompletion(id, entries) {
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
                    // delegate back to autocomplete, but extract the last term
                    response($.ui.autocomplete.filter(
                        entries, extractLast(request.term)));
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
                }
            });

    }
    return AutoCompletion;
}());