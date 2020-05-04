var Settings = (function () {
  function Settings() {
    var inst=this;
  }

  Settings.prototype.get = function(key, defaultValue) {
    var value = localStorage.getItem(key);
    if(value === null)
      return defaultValue;
    return value;
  }

  Settings.prototype.set = function(key, value) {
    localStorage.setItem(key, value);
  }

  Settings.prototype.getAddLabelHotkey = function() {
    return this.get("addlabelhotkey", "shift+enter");
  }

  Settings.prototype.setAddLabelHotkey = function(value) {
    return this.set("addlabelhotkey", value);
  }

  Settings.prototype.getLabelSeparator = function() {
    return this.get("labelseparator", ",");
  }

  Settings.prototype.setLabelSeparator = function(value) {
    return this.set("labelseparator", value);
  }

  //annotation mode is set in AnnotationSettings()
  Settings.prototype.getAnnotationMode = function() {
    return this.get("annotationmode", "default");
  }

  Settings.prototype.getPolygonVertexSize = function() {
    return this.get("polygonvertexsize", 5);
  }

  Settings.prototype.getDefaultImageDescriptionLanguage = function() {
    return this.get("defaultimagedescriptionlanguage", "en");
  }

  Settings.prototype.setAnnotationMode = function(value) {
    return this.set("annotationmode", value);
  }

  Settings.prototype.setPolygonVertexSize = function(value) {
    return this.set("polygonvertexsize", value);
  }

  Settings.prototype.isLabelViewFirstTimeOpened = function() {
    return this.get("labelviewfirsttimeopened", false);
  }

  Settings.prototype.setLabelViewFirstTimeOpened = function(opened) {
    this.set("labelviewfirsttimeopened", opened);
  }

  Settings.prototype.setDefaultImageDescriptionLanguage = function(language) {
    this.set("defaultimagedescriptionlanguage", language);
  }

  Settings.prototype.vimBindingsEnabled = function() {
	var val = this.get("vimbindingsenabled", false);
	if(typeof(val) == "string")
		return (val == "true" ? true : false);
	return val;
  }

  Settings.prototype.setVimBindingsEnabled = function(enabled) {
    this.set("vimbindingsenabled", enabled);
  }

  return Settings;
}());


var AnnotationSettings = (function() {
    function AnnotationSettings() {
        var inst = this;
    }

    AnnotationSettings.prototype.getPreferedAnnotationTool = function() {
        var radioButtonId = $("#preferedAnnotationToolCheckboxes :radio:checked").attr('id');
        if (radioButtonId === "preferedRectangleAnnotationToolCheckboxInput") {
            return "Rectangle";
        }
        if (radioButtonId === "preferedCircleAnnotationToolCheckboxInput") {
            return "Circle";
        }
        if (radioButtonId === "preferedPolygonAnnotationToolCheckboxInput") {
            return "Polygon";
        }
        return "Rectangle";
    }

    AnnotationSettings.prototype.getWorkspaceSize = function() {
        var radioButtonId = $("#annotationWorkspaceSizeCheckboxes :radio:checked").attr('id');
        if (radioButtonId === "annotationWorkspaceSizeSmallCheckboxInput") {
            return "small";
        }
        if (radioButtonId === "annotationWorkspaceSizeMediumCheckboxInput") {
            return "medium";
        }
        if (radioButtonId === "annotationWorkspaceSizeBigCheckboxInput") {
            return "big";
        }
        return "small";
    }

    AnnotationSettings.prototype.getAnnotationMode = function() {
        var radioButtonId = $("#annotationModeCheckboxes :radio:checked").attr('id');
        if (radioButtonId === "annotationDefaultModeCheckboxInput") {
            return "default";
        }
        if (radioButtonId === "annotationBrowseModeCheckboxInput") {
            return "browse";
        }
        return "default";
    }

    AnnotationSettings.prototype.getPolygonVertexSize = function() {
        var polygonVertexSize = $("#annotationPolygonVertexSizeInput").val();
        return polygonVertexSize;
    }

    AnnotationSettings.prototype.persistAll = function() {
        var settings = new Settings();

        var preferedAnnotationTool = this.getPreferedAnnotationTool();
        localStorage.setItem('preferedAnnotationTool', preferedAnnotationTool); //store in local storage
        var workspaceSize = this.getWorkspaceSize();
        localStorage.setItem('annotationWorkspaceSize', workspaceSize);
        var annotationMode = this.getAnnotationMode();
        settings.setAnnotationMode(annotationMode);
        var polygonVertexSize = this.getPolygonVertexSize();
        settings.setPolygonVertexSize(polygonVertexSize);
    }

    AnnotationSettings.prototype.setAll = function() {
        this.setPreferedAnnotationTool();
        this.setWorkspaceSize();
        this.setAnnotationMode();
        this.setPolygonVertexSize();
    }

    AnnotationSettings.prototype.setWorkspaceSize = function() {
        var workspaceSize = localStorage.getItem("annotationWorkspaceSize");
        if (workspaceSize === "small") {
            $("#annotationWorkspaceSizeMediumCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeBigCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeSmallCheckbox").checkbox("set checked");
        } else if (workspaceSize === "medium") {
            $("#annotationWorkspaceSizeSmallCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeBigCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeMediumCheckbox").checkbox("check");
        } else if (workspaceSize === "big") {
            $("#annotationWorkspaceSizeSmallCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeMediumCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeBigCheckbox").checkbox("set checked");
        }
    }

    AnnotationSettings.prototype.setPreferedAnnotationTool = function() {
        var preferedAnnotationTool = localStorage.getItem("preferedAnnotationTool");
        if (preferedAnnotationTool === "Rectangle") {
            $("#preferedCircleAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedRectangleAnnotationToolCheckbox").checkbox("set checked");
        } else if (preferedAnnotationTool === "Circle") {
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedRectangleAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedCircleAnnotationToolCheckbox").checkbox("check");
        } else if (preferedAnnotationTool === "Polygon") {
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedCircleAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set checked");
        }
    }

    AnnotationSettings.prototype.setAnnotationMode = function() {
        var settings = new Settings();
        annotationMode = settings.getAnnotationMode();
        if (annotationMode === "default") {
            $("#annotationBrowseModeCheckbox").checkbox("set unchecked");
            $("#annotationDefaultModeCheckbox").checkbox("set checked");
        } else if (annotationMode === "browse") {
            $("#annotationDefaultModeCheckbox").checkbox("set unchecked");
            $("#annotationBrowseModeCheckbox").checkbox("check");
        }
    }

    AnnotationSettings.prototype.setPolygonVertexSize = function() {
        var settings = new Settings();
        $("#annotationPolygonVertexSizeInput").val(settings.getPolygonVertexSize());
    }

    AnnotationSettings.prototype.loadPreferedAnnotationTool = function(annotationView, annotator) {
        var preferedAnnotationTool = localStorage.getItem("preferedAnnotationTool");
        if ((preferedAnnotationTool === "Rectangle") || (preferedAnnotationTool === "Circle") || (preferedAnnotationTool === "Polygon")) {
            annotationView.changeMenuItem(preferedAnnotationTool);
            annotator.setShape(preferedAnnotationTool);
        }
    }

    AnnotationSettings.prototype.loadWorkspaceSize = function() {
        var workspaceSize = localStorage.getItem("annotationWorkspaceSize");
        if ((workspaceSize === "small") || (workspaceSize === "medium") || (workspaceSize === "big")) {
            return workspaceSize;
        }
        return "small";
    }

    return AnnotationSettings;
}());
