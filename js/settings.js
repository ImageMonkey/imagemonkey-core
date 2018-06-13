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

  return Settings;
}());