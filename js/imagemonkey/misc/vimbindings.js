const LabelBrowseModeVimBindingsMode = {
    Browse: 'summer',
	Label: 'label'
}

var LabelBrowseModeVimBindings = (function() {
	function LabelBrowseModeVimBindings(mode) {
		this.imageIds = [];
		this.enableListeners(); 
		this.lastSelectedItemIdx = null;
		this.mode = mode;
	}

	LabelBrowseModeVimBindings.prototype.setMode = function(mode) {
		this.mode = mode;	
	}

	LabelBrowseModeVimBindings.prototype.reset = function() { 
		this.disableListeners();
		this.imageIds = [];
	}

	LabelBrowseModeVimBindings.prototype.addImageId = function(imageId) { 
		this.imageIds.push(imageId);
		if(this.mode == LabelBrowseModeVimBindingsMode.Browse) {
			if(this.lastSelectedItemIdx === null && this.imageIds.length > 0) {
				this.lastSelectedItemIdx = 0;
				$("#"+this.imageIds[this.lastSelectedItemIdx]).addClass("item-selected");
			}	
		} 
	}

	LabelBrowseModeVimBindings.prototype._next = function(nextItemIdx) {
		if(nextItemIdx < this.imageIds.length && nextItemIdx >= 0) {	
			$("#"+this.imageIds[nextItemIdx]).addClass("item-selected");

			if(this.imageIds.length > 1)
				$("#"+this.imageIds[this.lastSelectedItemIdx]).removeClass("item-selected");

			this.lastSelectedItemIdx = nextItemIdx;
		}	
	}
	
	LabelBrowseModeVimBindings.prototype.enableListeners = function() { 
		var inst = this;
		Mousetrap.bind("i", function() {
			if(inst.mode == LabelBrowseModeVimBindingsMode.Label) {
				$("#labelSuggestion").focus();
			}
		});

		Mousetrap.bind("right", function() {
        	if(inst.mode == LabelBrowseModeVimBindingsMode.Browse) {
				if(inst.imageIds.length > 0) {
					inst._next(inst.lastSelectedItemIdx+1);	
				}
			}
		});

		Mousetrap.bind("left", function() {
			if(inst.mode == LabelBrowseModeVimBindingsMode.Browse) {
				if(inst.imageIds.length > 0) {
					inst._next(inst.lastSelectedItemIdx-1);	
				}
			}
		});

		Mousetrap.bind("enter", function() {
			if(inst.mode == LabelBrowseModeVimBindingsMode.Browse) {
				if(inst.lastSelectedItemIdx !== null) {
					$("#vimModeStatusBar").show();	
					$("#"+inst.imageIds[inst.lastSelectedItemIdx]).click();
				}
			} else if(inst.mode == LabelBrowseModeVimBindingsMode.Label) {
				if($("#vimModeStatusBarInput").is(":focus")) {
					var cmd = $("#vimModeStatusBarInput").val();
					$("#vimModeStatusBarInput").val("");
					$("#labelSuggestion").val("");
					switch(cmd) {
						case ":wq":
							$("#doneButton").click();
							break;
						default:
							$("#vimModeStatusBarInput").val("Invalid command: " +cmd);
							break;
					}	
				}
			}
		});

		Mousetrap.bind("escape", function() {
			if(inst.mode === LabelBrowseModeVimBindingsMode.Label) {
				if($("#labelSuggestion").is(":focus"))
					$("#labelSuggestion").blur();
				if($("#vimModeStatusBarInput").is(":focus")) {
					$("#vimModeStatusBarInput").val("");
					$("#vimModeStatusBarInput").blur();
				}
			}
		});

		Mousetrap.bind(":", function() {
			if(inst.mode === LabelBrowseModeVimBindingsMode.Label && !$("#labelSuggestion").is(":focus")) {
				$("#vimModeStatusBarInput").val(":");
				$("#vimModeStatusBarInput").focus();
			}
		});
	}

	LabelBrowseModeVimBindings.prototype.disableListeners = function() { 
		Mousetrap.unbind("i");
		Mousetrap.unbind("left");
		Mousetrap.unbind("right");
		Mousetrap.unbind("escape");
		Mousetrap.unbind("enter");
		Mousetrap.unbind("escape");
	}

	return LabelBrowseModeVimBindings;
}());
