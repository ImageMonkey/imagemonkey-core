var ImageMonkeyApi = (function() {
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
        if (useCache && this.availableLabels) {
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
            xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                if (xhr.status >= 400)
                    reject();
                else
                    resolve();
            }
            xhr.onerror = function() {
                reject();
            }
            xhr.send(JSON.stringify(data));
        });
    }

    ImageMonkeyApi.prototype.acceptTrendingLabel = function(labelName, labelType, labelDescription, labelPlural, labelRenameTo) {
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
            xhr.onerror = function() {
                reject();
            }
            xhr.onreadystatechange = function() {
                if (xhr.status >= 400) {
                    reject();
                }
            }
            xhr.send(JSON.stringify({
                "label": {
                    "type": labelType,
                    "description": labelDescription,
                    "plural": labelPlural,
                    "rename_to": labelRenameTo
                }
            }));
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

    ImageMonkeyApi.prototype.getImageCollections = function(username) {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = inst.baseUrl + "/" + inst.apiVersion + "/user/" + username + "/imagecollections";
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

    ImageMonkeyApi.prototype.getLabelAccessors = function(detailed) {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = inst.baseUrl + "/" + inst.apiVersion + "/label/accessors";
            if (detailed)
                url += "?detailed=true";

            var xhr = new XMLHttpRequest();
            xhr.responseType = "json";
            xhr.open("GET", url);
            xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                var jsonResponse = xhr.response;
                resolve(jsonResponse);
            }
            xhr.onerror = function() {
                reject();
            }
            xhr.onreadystatechange = function() {
                if (xhr.status >= 400) {
                    reject();
                }
            }
            xhr.send();
        });
    }
    ImageMonkeyApi.prototype.getLabelSuggestions = function(includeUnlocked = true) {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = inst.baseUrl + "/" + inst.apiVersion + "/label/suggestions";
            if (!includeUnlocked)
                url += "?include_unlocked=false";

            var xhr = new XMLHttpRequest();
            xhr.responseType = "json";
            xhr.open("GET", url);
            xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                var jsonResponse = xhr.response;
                resolve(jsonResponse);
            }
            xhr.onerror = function() {
                reject();
            }
            xhr.onreadystatechange = function() {
                if (xhr.status >= 400) {
                    reject();
                }
            }
            xhr.send();
        });
    }

    ImageMonkeyApi.prototype.getPluralLabels = function() {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = inst.baseUrl + "/" + inst.apiVersion + "/label/plurals";
            var xhr = new XMLHttpRequest();
            xhr.responseType = "json";
            xhr.open("GET", url);
            xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                var jsonResponse = xhr.response;
                resolve(jsonResponse);
            }
            xhr.onerror = function() {
                reject();
            }
            xhr.onreadystatechange = function() {
                if (xhr.status >= 400) {
                    reject();
                }
            }
            xhr.send();
        });
    }


    ImageMonkeyApi.prototype.getAnnotatedImage = function(annotationId, annotationRevision) {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = inst.baseUrl + "/" + inst.apiVersion + "/annotation?annotation_id=" + annotationId;
            if (annotationRevision !== -1)
                url += '&rev=' + annotationRevision;

            var xhr = new XMLHttpRequest();
            xhr.responseType = "json";
            xhr.open("GET", url);
            xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                var jsonResponse = xhr.response;
                resolve(jsonResponse);
            }
            xhr.onerror = function() {
                reject();
            }
            xhr.onreadystatechange = function() {
                if (xhr.status >= 400) {
                    reject();
                }
            }
            xhr.send();
        });
    }

    ImageMonkeyApi.prototype.getUnannotatedImage = function(validationId, labelId) {
        var inst = this;
        return new Promise(function(resolve, reject) {
            var url = "";
            if (validationId === undefined)
                url = (inst.baseUrl + "/" + inst.apiVersion + "/annotate?add_auto_annotations=true" +
                    ((labelId === null) ? "" : ("&label_id=" + labelId)));
            else
                url = inst.baseUrl + "/" + inst.apiVersion + "/annotate?validation_id=" + validationId;

            var xhr = new XMLHttpRequest();
            xhr.responseType = "json";
            xhr.open("GET", url);
            xhr.setRequestHeader("Authorization", "Bearer " + inst.token);
            xhr.onload = function() {
                var jsonResponse = xhr.response;
                resolve(jsonResponse);
            }
            xhr.onerror = function() {
                reject();
            }
            xhr.onreadystatechange = function() {
                if (xhr.status >= 400) {
                    reject();
                }
            }
            xhr.send();
        });
    }

    return ImageMonkeyApi;
}());