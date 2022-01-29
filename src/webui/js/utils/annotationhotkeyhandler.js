var AnnotationHotkeyHandler = (function() {
    function AnnotationHotkeyHandler() {
	}

	AnnotationHotkeyHandler.prototype.drawRectangle = function(action) {
		Mousetrap.bind("r", action);
	}

	AnnotationHotkeyHandler.prototype.drawCircle = function(action) {
		Mousetrap.bind("c", action);
	}

	AnnotationHotkeyHandler.prototype.drawPolygon = function(action) {
		Mousetrap.bind("p", action);
	}

	AnnotationHotkeyHandler.prototype.selectMove = function(action) {
		Mousetrap.bind("s", action);
	}

	AnnotationHotkeyHandler.prototype.deleteAnnotation = function(action) {
		Mousetrap.bind("del", action);
	}

	AnnotationHotkeyHandler.prototype.zoomOut = function(action) {
		Mousetrap.bind("-", action);
	}

	AnnotationHotkeyHandler.prototype.zoomIn = function(action) {
		Mousetrap.bind("+", action);
	}

	return AnnotationHotkeyHandler;
}());
