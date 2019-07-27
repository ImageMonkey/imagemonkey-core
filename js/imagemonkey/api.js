var ImageMonkeyApi = (function () {
    function ImageMonkeyApi(baseUrl) {
        this.baseUrl = baseUrl;
        this.apiVersion = 'v1';
		this.token = '';
        this.availableLabels = null;
    };

	ImageMonkeyApi.prototype.setToken = function(token) {
		this.token = token;
	}

    ImageMonkeyApi.prototype.getAvailableLabels = function(useCache = false) {
        if(useCache && this.availableLabels) {
            var inst = this;
            return new Promise(function(resolve, reject) {
                resolve(inst.availableLabels);
            });
        } else {
            var inst = this;
            return new Promise(function(resolve, reject) {
                var url = inst.baseUrl + "/" + inst.apiVersion + "/label?detailed=true";
                var xhr = new XMLHttpRequest();
                xhr.responseType = "json";
                xhr.open("GET", url);
                xhr.onload = function() {
                    var jsonResponse = xhr.response;
                    resolve(jsonResponse);
                }
                xhr.onerror = reject;
                xhr.send();
            });

        }
    }

    ImageMonkeyApi.prototype.labelImage = function(imageId, data) {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = inst.baseUrl + "/" + inst.apiVersion + "/donation/" + imageId + "/labelme";
            var xhr = new XMLHttpRequest();
            xhr.open("POST", url);
            xhr.setRequestHeader("Content-Type", "application/json");
            xhr.onload = function() {
                resolve();
            }
            xhr.onerror = reject;
            xhr.send(JSON.stringify(data));
        });
    }

	ImageMonkeyApi.prototype.acceptTrendingLabel = function(labelName, labelType, labelDescription, labelPlural) {
		var inst = this;
		return new Promise(function(resolve, reject) {
			var url = inst.baseUrl + "/" + inst.apiVersion + "/trendinglabels/" + labelName + "/accept";
			var xhr = new XMLHttpRequest();
			xhr.open("POST", url);
			xhr.setRequestHeader("Content-Type", "application/json");
			xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                resolve();
            }
            xhr.onerror = reject;
            xhr.send(JSON.stringify({"label": {"type": labelType, "description": labelDescription, "plural": labelPlural}}));
		});
	}

	ImageMonkeyApi.prototype.getTrendingLabels = function() {
		var inst = this;
		return new Promise(function(resolve, reject) {
			var url = inst.baseUrl + "/" + inst.apiVersion + "/trendinglabels";
			var xhr = new XMLHttpRequest();
			xhr.responseType = "json";
			xhr.open("GET", url);
			xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
			xhr.onload = function() {
				var jsonResponse = xhr.response;
				resolve(jsonResponse);
			}
			xhr.onerror = reject;
			xhr.send();
		});
	}
    return ImageMonkeyApi;
}());
