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

  Settings.prototype.setAnnotationMode = function(value) {
    return this.set("annotationmode", value);
  }

  Settings.prototype.isLabelViewFirstTimeOpened = function() {
    return this.get("labelviewfirsttimeopened", false);
  }

  Settings.prototype.setLabelViewFirstTimeOpened = function(opened) {
    this.set("labelviewfirsttimeopened", opened);
  }

  return Settings;
}());