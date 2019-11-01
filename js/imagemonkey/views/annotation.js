var UnifiedModeStates = {
    uninitialized: 0,
    fetchedLabels: 1,
    fetchedAnnotations: 2,
    initialized: 3
};


var AnnotationView = (function() {
    function AnnotationView(apiBaseUrl, playgroundBaseUrl, annotationMode, annotationView, annotationId, annotationRevision, validationId, loggedIn) {
        this.annotationMode = annotationMode;
        this.annotationView = annotationView;
        this.loggedIn = loggedIn;
        this.apiBaseUrl = apiBaseUrl;
        this.playgroundBaseUrl = playgroundBaseUrl;
        this.annotationRevision = annotationRevision;
        this.validationId = validationId;
        this.annotationId = annotationId;

        this.imageMonkeyApi = new ImageMonkeyApi(this.apiBaseUrl);
        this.imageMonkeyApi.setToken(getCookie("imagemonkey"));

        this.canvas = null;
        this.annotator = null;
        this.detailedCanvas = null;
        this.numOfPendingRequests = 0;
        this.autoAnnotations = null;
        this.labelId = null;
        this.annotationInfo = new AnnotationInfo();
        this.unifiedModePopulated = UnifiedModeStates.uninitialized;
        this.annotationSettings = new AnnotationSettings();
        this.colorPicker = null;
        this.existingAnnotations = null;
        this.browserFingerprint = null;
        this.deleteObjectsPopupShown = false;
        this.unifiedModeAnnotations = {};
        this.unifiedModeLabels = {};
        this.pluralAnnotations = false;
        this.pluralLabels = null;
        this.initializeLabelsLstAftLoadDelayed = false;
        this.browseModeLastSelectedAnnotatorMenuItem = null;
        this.annotationRefinementsContextMenu = {};
        this.labelAccessorsLookupTable = {};
        this.availableLabels = [];
        this.availableLabelsLookupTable = {};
        this.labelsAutoCompletion = null;
    }

    AnnotationView.prototype.setSentryDSN = function(sentryDSN) {
        try {
            Sentry.init({
                dsn: sentryDSN,
            });
        } catch (e) {}
    }


    AnnotationView.prototype.handleUpdateAnnotationsRes = function(res) {
        if (this.annotationMode === "browse") {
            $("#loadingSpinner").hide();
            this.updateAnnotationsForImage(this.annotationInfo.annotationId, res);
            showBrowseAnnotationImageGrid();
        }

        if (this.onlyOnce) {
            showHideControls(false, this.annotationInfo.imageUnlocked);
            $("#onlyOnceDoneMessageContainer").show();
            $("#onlyOnceDoneMessage").fadeIn("slow");
            $("#loadingSpinner").hide();
        }
    }

    AnnotationView.prototype.onCanvasBackgroundImageSet = function() {
        if (isSmartAnnotationEnabled())
            this.populateDetailedCanvas();

        if (this.existingAnnotations !== null) {
            this.annotator.loadAnnotations(this.existingAnnotations, this.canvas.fabric().backgroundImage.scaleX);
            this.existingAnnotations = this.annotator.toJSON(); //export JSON after loading annotations
            //due to rounding we might end up with slightly different values, so we
            //export them in order to make sure that we don't accidentially detect
            //a rounding errors as changes.
        }

        showHideControls(true, this.annotationInfo.imageUnlocked);
        $("#annotationArea").css({
            "border-width": "1px",
            "border-style": "solid",
            "border-color": "#000000"
        });


        if (this.annotationView === "unified") {
            if (this.initializeLabelsLstAftLoadDelayed) {
                this.selectLabelInUnifiedLabelsLstAfterLoad();
                this.initializeLabelsLstAftLoadDelayed = false;
            }
        }
    }

    AnnotationView.prototype.loadAnnotatedImage = function(annotationId, annotationRevision) {
        showHideControls(false, this.annotationInfo.imageUnlocked);
        var inst = this;
        this.imageMonkeyApi.getAnnotatedImage(annotationId, annotationRevision)
            .then(function(data) {
                inst.handleAnnotatedImageResponse(data);

                //if there are already annotations, do not show blacklist or unannotatable button
                $("#blacklistButton").hide();
                $("#notAnnotableButton").hide();
            }).catch(function(e) {
                Sentry.captureException(e);
            });
    }

    AnnotationView.prototype.onAnnotatorObjectDeselected = function() {
        context.destroy('#annotationColumn');

        if (this.annotationView === "unified") {
            $("#annotationPropertiesLst").empty();
            $("#addRefinementButton").addClass("disabled");
            $("#addRefinementButtonTooltip").attr("data-tooltip", "Select a annotation first")
        }
    }

    AnnotationView.prototype.onAnnotatorObjectSelected = function() {
        if (this.annotator.objectsSelected()) {
            if (this.annotator.isSelectMoveModeEnabled()) {
                $("#trashMenuItem").removeClass("disabled");
                $("#propertiesMenuItem").removeClass("disabled");

                var strokeColor = this.annotator.getStrokeColorOfSelected();
                if (strokeColor !== null)
                    this.colorPicker.setColor(strokeColor);

                if (this.annotationView === "unified") {
                    //when object is selected, show refinements
                    var refs = this.annotator.getRefinementsOfSelectedItem();
                    var refsUuidMapping = annotationRefinementsDlg.getRefinementsUuidMapping();
                    for (var i = 0; i < refs.length; i++) {
                        if (refs[i] in refsUuidMapping) {
                            addRefinementToRefinementsLst(refsUuidMapping[refs[i]].name, refs[i], refsUuidMapping[refs[i]].icon);
                        }
                    }
                    $("#addRefinementButton").removeClass("disabled");
                    $("#addRefinementButtonTooltip").removeAttr("data-tooltip");
                    context.attach('#annotationColumn', this.annotationRefinementsContextMenu.data);
                }
            }
        } else {
            $("#trashMenuItem").addClass("disabled");
            $("#propertiesMenuItem").addClass("disabled");
        }
    }

    AnnotationView.prototype.onAnnotatorMouseUp = function() {
        if (isSmartAnnotationEnabled() && !this.annotator.isPanModeEnabled())
            this.grabCutMe();
    }

    AnnotationView.prototype.populateCanvas = function(backgroundImageUrl, initAnnotator, force = true) {
        if ((this.canvas !== null) && !force)
            this.annotator.reset();
        else {
            this.canvas = new CanvasDrawer("annotationArea");
            this.canvas.fabric().selection = false;
            this.annotator = new Annotator(this.canvas.fabric(), this.onAnnotatorObjectSelected.bind(this),
                this.onAnnotatorMouseUp.bind(this), this.onAnnotatorObjectDeselected.bind(this));
        }

        var scaleFactor = getCanvasScaleFactor(this.annotationInfo);

        var w = this.annotationInfo.origImageWidth * scaleFactor;
        var h = this.annotationInfo.origImageHeight * scaleFactor;

        $("#annotationAreaContainer").attr("width", w);
        $("#annotationAreaContainer").attr("height", h);
        $("#annotationAreaContainer").attr("scaleFactor", scaleFactor);
        this.canvas.setWidth(w);
        this.canvas.setHeight(h);

        if (initAnnotator) {
            this.canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, function() {
                this.annotator.initHistory();
                this.onCanvasBackgroundImageSet();

            });
        } else {
            this.canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, this.onCanvasBackgroundImageSet.bind(this));
        }
    }

    AnnotationView.prototype.markAsNotAnnotatable = function(validationId) {
        showHideControls(false, this.annotationInfo.imageUnlocked);
        var url = this.apiBaseUrl + '/v1/validation/' + validationId + '/not-annotatable';
        var inst = this;
        $.ajax({
            url: url,
            type: 'POST',
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
            },
            success: function(data, status, xhr) {
                if (inst.annotationMode === "default")
                    inst.loadUnannotatedImage();
                else {
                    $("#loadingSpinner").hide();
                    clearDetailedCanvas(inst.detailedCanvas);
                    inst.annotator.reset();
                    showBrowseAnnotationImageGrid();
                }
            }
        });
    }



    AnnotationView.prototype.updateAnnotations = function(res) {
        if (_.isEqual(res, this.existingAnnotations)) {
            showHideControls(false, this.annotationInfo.imageUnlocked);
            clearDetailedCanvas(this.detailedCanvas);
            this.annotator.reset();
            this.handleUpdateAnnotationsRes(this.existingAnnotations);
            return;
        }

        var postData = {}
        postData["annotations"] = res;

        var headers = {}
        if (this.browserFingerprint !== null)
            headers["X-Browser-Fingerprint"] = this.browserFingerprint;

        headers['X-App-Identifier'] = this.appIdentifier;

        showHideControls(false, this.annotationInfo.imageUnlocked);
        clearDetailedCanvas(this.detailedCanvas);
        this.annotator.reset();

        var url = this.apiBaseUrl + "/v1/annotation/" + this.annotationInfo.annotationId;
        var inst = this;
        $.ajax({
            url: url,
            type: 'PUT',
            data: JSON.stringify(postData),
            headers: headers,
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
            },
            success: function(data) {
                inst.handleUpdateAnnotationsRes(res);
            }
        });
    }


    AnnotationView.prototype.addAnnotationsUnifiedMode = function() {
        var annotations = [];
        for (var key in this.unifiedModeAnnotations) {
            if (this.unifiedModeAnnotations.hasOwnProperty(key)) {
                if (this.unifiedModeAnnotations[key].dirty) {
                    var annotation = {};
                    annotation["annotations"] = this.unifiedModeAnnotations[key].annotations;
                    annotation["label"] = this.unifiedModeAnnotations[key].label;
                    annotation["sublabel"] = this.unifiedModeAnnotations[key].sublabel;
                    annotations.push(annotation);
                }
            }
        }
        this.unifiedModeLabels = {};
        this.unifiedModeAnnotations = {};
        this.addAnnotations(annotations);
    }

    AnnotationView.prototype.addAnnotations = function(annotations) {
        var headers = {}
        if (this.browserFingerprint !== null)
            headers["X-Browser-Fingerprint"] = this.browserFingerprint;

        headers['X-App-Identifier'] = this.appIdentifier;

        showHideControls(false, this.annotationInfo.imageUnlocked);
        clearDetailedCanvas(this.detailedCanvas);
        this.annotator.reset();

        if (annotations.length === 0) {
            this.onAddAnnotationsDone();
            return;
        }

        var url = this.apiBaseUrl + "/v1/donation/" + this.annotationInfo.imageId + "/annotate";
        var inst = this;
        $.ajax({
            url: url,
            type: 'POST',
            data: JSON.stringify(annotations),
            headers: headers,
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
            },
            success: function(data) {
                inst.onAddAnnotationsDone();
            }
        });
    }

    AnnotationView.prototype.onAddAnnotationsDone = function() {
        if (this.annotationMode === "default")
            this.loadUnannotatedImage();
        else {
            $("#loadingSpinner").hide();
            changeNavHeader("browse");
            showBrowseAnnotationImageGrid();
        }

        if (this.onlyOnce) {
            $("#onlyOnceDoneMessage").fadeIn("slow");
            showHideControls(false);
            $("#loadingSpinner").hide();
        }
    }

    AnnotationView.prototype.blacklistAnnotation = function(validationId) {
        showHideControls(false, this.annotationInfo.imageUnlocked);
        var url = this.apiBaseUrl + '/v1/validation/' + validationId + '/blacklist-annotation';
        var inst = this;
        $.ajax({
            url: url,
            type: 'POST',
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
            },
            success: function(data, status, xhr) {
                if (inst.annotationMode === "default")
                    inst.loadUnannotatedImage();
                else {
                    $("#loadingSpinner").hide();
                    clearDetailedCanvas(inst.detailedCanvas);
                    inst.annotator.reset();
                    showBrowseAnnotationImageGrid();
                }
            }
        });
    }

    AnnotationView.prototype.loadUnannotatedImage = function(validationId) {
        showHideControls(false, this.annotationInfo.imageUnlocked);
        var inst = this;
        this.imageMonkeyApi.getUnannotatedImage(validationId, inst.labelId)
            .then(function(data) {
                inst.handleUnannotatedImageResponse(data);
            }).catch(function() {
                inst.handleUnannotatedImageResponse(null);
            });

    }

    AnnotationView.prototype.getAnnotationInfo = function() {
        return this.annotationInfo;
    }

    AnnotationView.prototype.populateDefaultsAndLoadData = function() {
        if (this.annotationMode === "browse")
            $("#loadingSpinner").show();

        var inst = this;
        Promise.all([this.imageMonkeyApi.getPluralLabels(), this.imageMonkeyApi.getLabelAccessors(true)])
            .then(function(data) {
                inst.pluralLabels = data[0];

                for (var i = 0; i < data[1].length; i++) {
                    inst.labelAccessorsLookupTable[data[1][i].accessor] = data[1][i].parent_accessor;
                }

                if (inst.annotationMode === "default") {
                    if (inst.validationId === "")
                        inst.loadUnannotatedImage();
                    else
                        inst.loadUnannotatedImage(inst.validationId);
                }

                if (inst.annotationMode === "refine") {
                    if (inst.annotationId !== "")
                        inst.loadAnnotatedImage(inst.annotationId, inst.annotationRevision);
                }

                if (inst.annotationMode === "browse")
                    $("#loadingSpinner").hide();

            }).catch(function(e) {
                Sentry.captureException(e);
            });
    }


    AnnotationView.prototype.saveCurrentSelectLabelInUnifiedModeList = function() {
        var key = '';
        if ($("#label").attr("sublabel") === "")
            key = $("#label").attr("label");
        else
            key = $("#label").attr("sublabel") + "/" + $("#label").attr("label");

        var annos = null;
        if (key in this.unifiedModeAnnotations) {
            if (this.annotator.isDirty()) {
                annos = this.annotator.toJSON();
                this.unifiedModeAnnotations[key] = {
                    annotations: annos,
                    label: $("#label").attr("label"),
                    sublabel: $("#label").attr("sublabel"),
                    dirty: true
                };
            }
        } else {
            annos = this.annotator.toJSON();
            if (annos.length > 0) {
                this.unifiedModeAnnotations[key] = {
                    annotations: annos,
                    label: $("#label").attr("label"),
                    sublabel: $("#label").attr("sublabel"),
                    dirty: true
                };
            }
        }
    }

    AnnotationView.prototype.onLabelInLabelLstClicked = function(elem) {
        var key = "";

        //before changing label, save existing annotations
        this.saveCurrentSelectLabelInUnifiedModeList();

        $("#annotationLabelsLst").children().each(function(i) {
            $(this).removeClass("grey inverted");
        });

        setLabel($(elem).attr("data-label"), $(elem).attr("data-sublabel"), null);
        $(elem).addClass("grey inverted");

        //clear existing annotations
        this.annotator.deleteAll();

        key = "";
        if ($(elem).attr("data-sublabel") === "")
            key = $(elem).attr("data-label");
        else
            key = $(elem).attr("data-sublabel") + "/" + $(elem).attr("data-label");

        //show annotations for selected label
        if (key in this.unifiedModeAnnotations) {
            this.annotator.loadAnnotations(this.unifiedModeAnnotations[key].annotations, this.canvas.fabric().backgroundImage.scaleX);
        }
    }

    AnnotationView.prototype.populateDetailedCanvas = function(force = false) {
        if ((this.detailedCanvas !== null) && !force)
            this.detailedCanvas.clear();
        else
            this.detailedCanvas = new CanvasDrawer("smartAnnotationCanvas", 0, 0);

        var maxWidth = document.getElementById("smartAnnotationContainer").clientWidth - 50; //margin
        var scaleFactor = maxWidth / this.annotationInfo.origImageWidth;
        if (scaleFactor > 1.0)
            scaleFactor = 1.0;

        var w = this.annotationInfo.origImageWidth * scaleFactor;
        var h = this.annotationInfo.origImageHeight * scaleFactor;

        $("#smartAnnotationCanvasWrapper").attr("width", w);
        $("#smartAnnotationCanvasWrapper").attr("height", h);
        $("#smartAnnotationCanvasWrapper").attr("scaleFactor", scaleFactor);
        //detailedCanvas = new CanvasDrawer("smartAnnotationCanvas", w, h);
        this.detailedCanvas.setWidth(w);
        this.detailedCanvas.setHeight(h);
        this.detailedCanvas.setCanvasBackgroundImage(this.canvas.fabric().backgroundImage, null);
    }


    AnnotationView.prototype.grabCutMe = function() {
        this.numOfPendingRequests += 1;
        $("#smartAnnotationCanvasWrapper").dimmer("show");
        var blob = dataURItoBlob(this.annotator.getMask());
        var formData = new FormData()
        formData.append('image', blob);
        formData.append('uuid', this.annotationInfo.imageId);
        var inst = this;
        $.ajax({
            url: inst.playgroundBaseUrl + '/v1/grabcut',
            processData: false,
            contentType: false,
            data: formData,
            type: 'POST',
            success: function(data, status, xhr) {
                inst.pollUntilProcessed(xhr.getResponseHeader("Location"));
            }
        });
    }

    AnnotationView.prototype.getLabelsForImage = function(imageId, onlyUnlockedLabels) {
        var url = this.apiBaseUrl + '/v1/donation/' + imageId + "/labels?only_unlocked_labels=" + (onlyUnlockedLabels ? "true" : "false");
        var inst = this;
        $.ajax({
            url: url,
            dataType: 'json',
            type: 'GET',
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
            },
            success: function(data) {
                if (inst.annotationView === "unified") {
                    if (data !== null) {
                        for (var i = 0; i < data.length; i++) {
                            addLabelToLabelLst(data[i].label, '', data[i].uuid, false, false, data[i].unlocked, inst.loggedIn);
                            if (data[i].sublabels !== null) {
                                for (var j = 0; j < data[i].sublabels.length; j++) {
                                    addLabelToLabelLst(data[i].label, data[i].sublabels[j].name,
                                        data[i].sublabels[j].uuid, false, false, data[i].unlocked, inst.loggedIn);
                                }
                            }
                        }
                    }
                    inst.unifiedModePopulated |= UnifiedModeStates.fetchedLabels;

                    if (inst.unifiedModePopulated === UnifiedModeStates.initialized) {
                        if (inst.canvas.fabric().backgroundImage && inst.canvas.fabric().backgroundImage !== undefined) {
                            inst.initializeLabelsLstAftLoadDelayed = false;
                            inst.selectLabelInUnifiedLabelsLstAfterLoad();
                        } else { //image is not yet loaded (which we need before we can initialize the labels list),
                            //so we need to initialize the labels list when the image is loaded
                            inst.initializeLabelsLstAftLoadDelayed = true;
                        }
                    }
                }
            },
            error: function(xhr, options, err) {}
        });
    }

    AnnotationView.prototype.getAnnotationsForImage = function(imageId) {
        var url = '';
        if (this.annotationInfo.imageUnlocked)
            url = this.apiBaseUrl + '/v1/donation/' + imageId + "/annotations";
        else
            url = this.apiBaseUrl + '/v1/unverified-donation/' + imageId + "/annotations" + "?token=" + getCookie("imagemonkey");
        var inst = this;
        $.ajax({
            url: url,
            dataType: 'json',
            type: 'GET',
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
            },
            success: function(data) {
                if (inst.annotationView === "unified") {
                    for (var i = 0; i < data.length; i++) {
                        if (data[i].validation.sublabel !== "")
                            inst.unifiedModeAnnotations[data[i].validation.sublabel + "/" + data[i].validation.label] = {
                                annotations: data[i].annotations,
                                label: data[i].validation.label,
                                sublabel: data[i].validation.sublabel,
                                dirty: false
                            };
                        else {
                            inst.unifiedModeAnnotations[data[i].validation.label] = {
                                annotations: data[i].annotations,
                                label: data[i].validation.label,
                                sublabel: data[i].validation.sublabel,
                                dirty: false
                            };
                        }
                    }
                    inst.unifiedModePopulated |= UnifiedModeStates.fetchedAnnotations;

                    if (inst.unifiedModePopulated === UnifiedModeStates.initialized) {
                        if (inst.canvas.fabric().backgroundImage && inst.canvas.fabric().backgroundImage !== undefined) {
                            inst.initializeLabelsLstAftLoadDelayed = false;
                            inst.selectLabelInUnifiedLabelsLstAfterLoad();
                        } else { //image is not yet loaded (which we need before we can initialize the labels list),
                            //so we need to initialize the labels list when the image is loaded
                            inst.initializeLabelsLstAftLoadDelayed = true;
                        }
                    }
                }
            },
            error: function(xhr, options, err) {}
        });
    }


    AnnotationView.prototype.selectLabelInUnifiedLabelsLstAfterLoad = function() {
        var unifiedModeToolboxChildren = $('#annotationLabelsLst').children('.labelslstitem');
        var foundLabelInUnifiedModeToolbox = false;
        unifiedModeToolboxChildren.each(function(index, value) {
            if (($(this).attr("data-label") === $("#label").attr("label")) && ($(this).attr("data-sublabel") === $("#label").attr("sublabel"))) {
                foundLabelInUnifiedModeToolbox = true;
                $(this).click();
                return false;
            }
        })

        //when label not found, select first one in list
        if (!foundLabelInUnifiedModeToolbox) {
            var firstItem = unifiedModeToolboxChildren.first();
            if (firstItem && firstItem.length === 1)
                firstItem[0].click();
        }

        $("#unifiedModeLabelsLstLoadingIndicator").hide();
    }


    AnnotationView.prototype.handleAnnotatedImageResponse = function(data) {
        this.annotationInfo.imageId = data.image.uuid;
        this.annotationInfo.origImageWidth = data.image.width;
        this.annotationInfo.origImageHeight = data.image.height;
        this.annotationInfo.annotationId = data.uuid;
        this.annotationInfo.imageUrl = data.image.url;
        this.annotationInfo.imageUnlocked = data.image.unlocked;

        this.initializeLabelsLstAftLoadDelayed = false;
        this.autoAnnotations = null;
        this.existingAnnotations = data["annotations"];
        showHideAutoAnnotationsLoadButton(this.autoAnnotations);

        setLabel(data.validation.label, data.validation.sublabel, null);

        if (this.canvas !== undefined && this.canvas !== null) {
            this.annotator.reset();
        }
        this.addMainCanvas();
        this.populateCanvas(getUrlFromImageUrl(data.image.url, data.image.unlocked, this.annotationMode, this.labelAccessorsLookupTable), false);
        changeControl(this.annotator, this.annotationInfo.imageId);
        this.numOfPendingRequests = 0;
        showHideControls(true, this.annotationInfo.imageUnlocked);

        if (this.annotationMode === "browse") {
            if (this.browseModeLastSelectedAnnotatorMenuItem === null) {
                if (this.annotator) {
                    this.annotationSettings.loadPreferedAnnotationTool(this.annotator);
                    this.annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
                }
            } else {
                this.changeMenuItem(this.browseModeLastSelectedAnnotatorMenuItem);
                this.annotator.setShape(this.browseModeLastSelectedAnnotatorMenuItem);
            }
        } else {
            if (this.annotator) {
                this.annotationSettings.loadPreferedAnnotationTool(this.annotator);
                this.annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
            }
        }

        populateRevisionsDropdown(data["num_revisions"], data["revision"]);
        showHideRevisionsDropdown();

        if (this.annotationView === "unified") {
            addLabelToLabelLst(data.validation.label, data.validation.sublabel, data.uuid, false, false, data.validation.unlocked, this.loggedIn);
            var firstItem = $('#annotationLabelsLst').children('.labelslstitem').first();
            if (firstItem && firstItem.length === 1) {
                $(firstItem[0]).addClass("grey inverted");
            }

            $("#unifiedModeLabelsLstLoadingIndicator").hide();
        }


        var pushedUrl = window.location.pathname + "?annotation_id=" + data.uuid;
        if (data.revision !== -1)
            pushedUrl += "&rev=" + data.revision;
        if (this.annotationView === "unified")
            pushedUrl += "&view=unified";
        history.pushState({
                current_page: pushedUrl,
                previous_page: window.location.href
            },
            "", pushedUrl
        );
    }

    AnnotationView.prototype.changeMenuItem = function(type) {
        var id = "";
        if (type === 'Rectangle')
            id = "rectMenuItem";
        else if (type === "Circle")
            id = "circleMenuItem";
        else if (type === "Polygon")
            id = "polygonMenuItem";
        else if (type === "PanMode")
            id = "panMenuItem";
        else if (type === "BlockSelection")
            id = "blockSelectMenuItem";
        else if (type === "FreeDrawing")
            id = "freeDrawingMenuItem";
        else if (type === "ForegroundSelection")
            id = "smartAnnotationFgMenuItem";
        else if (type === "BackgroundSelection")
            id = "smartAnnotationBgMenuItem";
        else if (type === "SelectMove")
            id = "selectMoveMenutItem";

        if (this.annotationMode === "browse")
            this.browseModeLastSelectedAnnotatorMenuItem = type;

        $("#annotatorMenu").children().each(function() {
            if ($(this).attr("id") === id)
                $(this).addClass("active");
            else
                $(this).removeClass("active");
        });

    }



    AnnotationView.prototype.pollUntilProcessed = function(uuid) {
        var url = this.playgroundBaseUrl + "/v1/grabcut/" + uuid;
        var inst = this;
        $.getJSON(url, function(response) {
            if (jQuery.isEmptyObject(response))
                setTimeout(inst.pollUntilProcessed(uuid), 1000);
            else {
                inst.detailedCanvas.clearObjects();

                if (response["result"]["points"].length > 0) {
                    var data = [];
                    data.push(response["result"]);
                    inst.annotator.setSmartAnnotationData(data);
                    inst.detailedCanvas.drawAnnotations(data, $("#smartAnnotationCanvasWrapper").attr("scaleFactor"));
                }

                inst.numOfPendingRequests -= 1;
                if (inst.numOfPendingRequests <= 0) {
                    $("#smartAnnotationCanvasWrapper").dimmer("hide");
                    inst.numOfPendingRequests = 0;
                }
            }
        });
    }

    AnnotationView.prototype.addMainCanvas = function() {
        $("#annotationColumnSpacer").remove();
        $("#annotationPropertiesColumnSpacer").remove();
        $("#annotationColumnContent").remove();

        var spacer = '';
        var inst = this;
        var unifiedModePropertiesLst = '';
        var w = "sixteen";
        if (isSmartAnnotationEnabled()) {
            w = "eight";

        } else {
            var unifiedModePropertiesLstWidth = 'four';
            var workspaceSize = this.annotationSettings.loadWorkspaceSize();
            if (this.annotationView === "unified") {
                var spacerWidth = "three";
                if (workspaceSize === "small") {
                    w = "eight";
                    spacerWidth = "four";
                    unifiedModePropertiesLstWidth = 'four';
                } else if (workspaceSize === "medium") {
                    w = "eight";
                    spacerWidth = "four";
                    unifiedModePropertiesLstWidth = 'four';
                } else if (workspaceSize === "big") {
                    w = "eight";
                    spacerWidth = "four";
                    unifiedModePropertiesLstWidth = 'four';
                }

                var unifiedModeLabelsLstUiElems = '';
                if (this.annotationMode !== "refine") {
                    var showUnifiedModeLabelsLstUiElems = false;
                    if (this.annotationMode === "default")
                        showUnifiedModeLabelsLstUiElems = true;
                    else {
                        if (!$("#annotationsOnlyCheckbox").checkbox("is checked"))
                            showUnifiedModeLabelsLstUiElems = true;
                    }
                    if (showUnifiedModeLabelsLstUiElems) {
                        unifiedModeLabelsLstUiElems = '<div class="ui center aligned grid">' +
                            '<div class="twelve wide centered column">' +
                            '<div class="ui form">' +
                            '<div class="fields">' +
                            '<div class="field">' +
                            '<div class="ui search">' +
                            '<div class="ui center aligned action input" id="addLabelToUnifiedModeListForm">' +

                            '<div class="ui input">' +
                            '<input placeholder="Enter label..." type="text" id="addLabelsToUnifiedModeListLabels" class="mousetrap">' +
                            '</div>' +
                            '<div class="ui button" id="addLabelToUnifiedModeListButton">Add</div>' +
                            '</div>' +
                            '</div>' +
                            '</div>' +
                            '</div>' +
                            '</div>' +
                            '</div>' +
                            '</div>';
                    }
                }


                spacer = '<div class="' + spacerWidth + ' wide column" id="annotationColumnSpacer">' +
                    '<h2 class="ui center aligned header">' +
                    '<div class="content">' +
                    'Labels' +
                    '</div>' +
                    '</h2>' +
                    '<div class="ui basic segment">' +
                    '<div class="ui segments">' +
                    '<div class="ui raised segments" style="overflow: auto; height: 50vh;" id="annotationLabelsLst">' +
                    '<div class="ui active indeterminate loader" id="unifiedModeLabelsLstLoadingIndicator"></div>' +
                    '</div>' + unifiedModeLabelsLstUiElems +

                    '</div>' +
                    '</div>' +
                    '</div>';



                unifiedModePropertiesLst = '<div class="' + unifiedModePropertiesLstWidth + ' wide column" id="annotationPropertiesColumnSpacer">' +
                    '<h2 class="ui center aligned header">' +
                    '<div class="content">' +
                    'Properties' +
                    '</div>' +
                    '</h2>' +
                    '<div class="ui basic segment">' +
                    '<div class="ui segments">' +
                    '<div class="ui raised segments" style="overflow: auto; height: 50vh;" id="annotationPropertiesLst">' +
                    '</div>' +

                    '<div class="ui center aligned grid">' +
                    '<div class="twelve wide centered column">' +
                    '<div class="ui form">' +
                    '<div class="fields">' +
                    '<div class="field">' +
                    '<div class="ui search">' +
                    '<div class="ui center aligned action input" id="addRefinementForm">' +
                    '<div class="ui small search selection dropdown" id="addRefinementDropdown">' +
                    '<div class="default text">Select Refinement</div>' +
                    '<div class="menu" id="addRefinementDropdownMenu">' +
                    '</div>' +
                    '</div>' +
                    '<div id="addRefinementButtonTooltip" data-tooltip="Select a annotation first" data-position="left center">' +
                    '<div class="ui disabled button" id="addRefinementButton">Add</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>';

            } else {
                if (workspaceSize === "small") {
                    w = "eight";
                    spacer = '<div class="four wide column" id="annotationColumnSpacer"></div>';
                } else if (workspaceSize === "medium") {
                    w = "ten";
                    spacer = '<div class="three wide column" id="annotationColumnSpacer"></div>';
                } else if (workspaceSize === "big")
                    w = "sixteen";
            }
        }


        var data = spacer +
            '<div class="' + w + ' wide center aligned column" id="annotationColumnContent">' +
            '<div id="annotationAreaContainer">' +
            '<canvas id="annotationArea" imageId=""></canvas>' +
            '</div>' +
            '</div>' + unifiedModePropertiesLst;

        $("#annotationColumn").show();
        $("#annotationColumn").append(data);

        if (this.annotationView === "unified") {
            $("#addLabelToUnifiedModeListButton").click(function(e) {
                var selectedElem = null;
                var labelName = escapeHtml($("#addLabelsToUnifiedModeListLabels").val());

                if (!inst.loggedIn) {
                    if (!(labelName in inst.availableLabelsLookupTable)) {
                        $("#warningMsgText").text("Please sign in first to add new labels!");
                        $("#warningMsg").show(200).delay(1500).hide(200);
                        return
                    }
                } else { //logged in
                    var pattern = new RegExp("^[a-zA-Z ]+$");
                    if (!pattern.test(labelName)) {
                        $("#warningMsgText").text("Invalid label name " + labelName + ". (supported characters: a-zA-Z and ' ')");
                        $("#warningMsg").show(200).delay(1500).hide(200);
                        return
                    }
                }

                if (labelName in inst.availableLabelsLookupTable)
                    selectedElem = inst.availableLabelsLookupTable[labelName];
                if (selectedElem === null) {
                    if (inst.loggedIn) {
                        var tempUuid = labelName.replace(/\s/g, ""); //remove all whitespaces
                        selectedElem = {
                            "uuid": tempUuid,
                            "label": labelName,
                            "sublabel": "",
                            "newly_created": true
                        }
                    } else {
                        $("#warningMsgText").text("Please sign in first to add new labels!");
                        $("#warningMsg").show(200).delay(1500).hide(200);
                        return
                    }
                }

                var alreadyExistsInUnifiedModeLabelsLst = false;
                var elem;
                $("#annotationLabelsLst").children('.labelslstitem').each(function(idx) {
                    if ($(this).attr("data-uuid") === selectedElem.uuid) {
                        alreadyExistsInUnifiedModeLabelsLst = true;
                        elem = $(this);
                        return false;
                    }

                    //if it's a non productive label, we need to do it a bit differently
                    if (selectedElem.uuid === labelName && $(this).attr("data-label") === selectedElem.uuid &&
                        $(this).attr("data-sublabel") === "") {

                        alreadyExistsInUnifiedModeLabelsLst = true;
                        elem = $(this);
                        return false;
                    }
                });

                if (!alreadyExistsInUnifiedModeLabelsLst) {
                    if (selectedElem.sublabel !== "") {
                        inst.unifiedModeLabels[selectedElem.uuid] = {
                            "label": selectedElem.label,
                            "sublabels": [{
                                "name": selectedElem.sublabel
                            }],
                            "annotatable": true
                        };
                    } else {
                        inst.unifiedModeLabels[selectedElem.uuid] = {
                            "label": selectedElem.label,
                            "annotatable": true
                        };
                    }
                    elem = addLabelToLabelLst(selectedElem.label, selectedElem.sublabel,
                        selectedElem.uuid, true, true, false, inst.loggedIn);
                }

                //select newly added (or already existing) label
                inst.onLabelInLabelLstClicked(elem);

                $("#addLabelsToUnifiedModeListLabels").val("");
            });

            $("#addRefinementButton").click(function(e) {
                var refs = inst.annotator.getRefinementsOfSelectedItem();
                if (refs.indexOf($('#addRefinementDropdown').dropdown('get value')) == -1) {
                    refs.push($('#addRefinementDropdown').dropdown('get value'));
                    inst.annotator.setRefinements(refs);
                    var allRefs = annotationRefinementsDlg.getRefinementsUuidMapping();
                    var refIcon = "";
                    if ($('#addRefinementDropdown').dropdown('get value') in allRefs)
                        refIcon = allRefs[$('#addRefinementDropdown').dropdown('get value')].icon;
                    addRefinementToRefinementsLst($('#addRefinementDropdown').dropdown('get text'), $('#addRefinementDropdown').dropdown('get value'), refIcon);
                }
                $('#addRefinementDropdown').dropdown('restore placeholder text');
            });
            var refs = annotationRefinementsDlg.getRefinementsUuidMapping();
            for (var k in refs) {
                if (refs.hasOwnProperty(k)) {
                    var entry = '<div class="item" data-value="' + k + '">';
                    if (refs[k].icon !== "") {
                        entry += '<i class="' + refs[k].icon + ' icon"></i>';
                    }
                    entry += (refs[k].name + '</div>');
                    $("#addRefinementDropdownMenu").append(entry);
                }
            }
            $("#addRefinementDropdown").dropdown();
        }

        $("#annotationArea").attr("imageId", this.annotationInfo.imageId);
        $("#annotationArea").attr("origImageWidth", this.annotationInfo.origImageWidth);
        $("#annotationArea").attr("origImageHeight", this.annotationInfo.origImageHeight);
        $("#annotationArea").attr("validationId", this.annotationInfo.validationId);
    }

    AnnotationView.prototype.handleUnannotatedImageResponse = function(data) {
        this.existingAnnotations = null;
        this.autoAnnotations = null;
        this.initializeLabelsLstAftLoadDelayed = false;

        if (data !== null) {
            this.annotationInfo.imageId = data.uuid;
            this.annotationInfo.origImageWidth = data.width;
            this.annotationInfo.origImageHeight = data.height;
            this.annotationInfo.validationId = data.validation.uuid;
            this.annotationInfo.imageUrl = data.url;
            this.annotationInfo.imageUnlocked = data.unlocked;

            if ("auto_annotations" in data) {
                if (data["auto_annotations"].length !== 0)
                    this.autoAnnotations = data["auto_annotations"];
            }

            setLabel(data.label.label, data.label.sublabel, data.label.accessor);
            changeNavHeader("default");
        } else {
            this.annotationInfo.imageId = "";
            this.annotationInfo.origImageWidth = 720; //width of the oops-no-annotation-left image
            this.annotationInfo.origImageHeight = 720; //height of the oops-no-annotation-left image
            this.annotationInfo.validationId = "";
            this.annotationInfo.imageUrl = "";
            this.annotationInfo.imageUnlocked = false;
        }

        showHideAutoAnnotationsLoadButton(this.autoAnnotations);

        if (this.canvas !== undefined && this.canvas !== null) {
            this.annotator.reset();
        }
        this.addMainCanvas();
        this.populateCanvas(getUrlFromImageUrl(this.annotationInfo.imageUrl, this.annotationInfo.imageUnlocked, this.annotationMode,
            this.labelAccessorsLookupTable), false);
        changeControl(this.annotator, this.annotationInfo.imageId);
        this.numOfPendingRequests = 0;

        if (data === null)
            changeNavHeader("noimage");

        if (this.annotationMode === "browse") {
            if (this.browseModeLastSelectedAnnotatorMenuItem === null) {
                if (this.annotator) {
                    this.annotationSettings.loadPreferedAnnotationTool(this.annotator);
                    this.annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
                }
            } else {
                this.changeMenuItem(this.browseModeLastSelectedAnnotatorMenuItem);
                this.annotator.setShape(this.browseModeLastSelectedAnnotatorMenuItem);
            }
        } else {
            if (this.annotator) {
                this.annotationSettings.loadPreferedAnnotationTool(this.annotator);
                this.annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
            }
        }


        if (this.annotationView === "unified") {
            this.unifiedModeLabels = {};
            this.unifiedModeAnnotations = {};
            if (this.annotationInfo.imageId !== "")
                this.populateUnifiedModeToolbox(this.annotationInfo.imageId);
        }

        populateRevisionsDropdown(0, 0);
        showHideRevisionsDropdown();
    }


    AnnotationView.prototype.populateUnifiedModeToolbox = function(imageId) {
        $("#unifiedModeLabelsLstLoadingIndicator").show();
        this.unifiedModePopulated = UnifiedModeStates.uninitialized;
        this.getAnnotationsForImage(imageId);
        this.getLabelsForImage(imageId, false);

        var labelRequests = [this.imageMonkeyApi.getAvailableLabels()];
        if (this.loggedIn)
            labelRequests.push(this.imageMonkeyApi.getLabelSuggestions(false));

        var inst = this;
        Promise.all(labelRequests)
            .then(function(data) {
                for (var key in data[0]) {
                    if (data[0].hasOwnProperty(key)) {
                        inst.availableLabels.push(key);
                        inst.availableLabelsLookupTable[key] = {
                            "uuid": data[0][key].uuid,
                            "label": key,
                            "sublabel": ""
                        };
                    }

                    if (data[0][key].has) {
                        for (var subkey in data[0][key].has) {
                            if (data[0][key].has.hasOwnProperty(subkey)) {
                                inst.availableLabels.push(subkey + "/" + key);
                                inst.availableLabels[subkey + "/" + key] = {
                                    "uuid": data[0][key].has[subkey].uuid,
                                    "label": key,
                                    "sublabel": subkey
                                };
                            }
                        }
                    }
                }
                if (data.length > 1) {
                    inst.availableLabels.push(...data[1]);
                }
                inst.labelsAutoCompletion = new AutoCompletion("#addLabelsToUnifiedModeListLabels", inst.availableLabels);
            }).catch(function(e) {
                Sentry.captureException(e);
            });
    }

    AnnotationView.prototype.exec = function() {
        var inst = this;
        var lastActiveMenuItem = "";
        $('#warningMsg').hide();

        $('#smartAnnotation').checkbox({
            onChange: function() {
                var enabled = isSmartAnnotationEnabled();
                if (enabled) {
                    inst.annotator.enableSmartAnnotation();

                    $("#spacer").remove();
                    $("#annotationColumn").show();
                    $("#annotationColumn").prepend('<div class="eight wide center aligned column" id="smartAnnotationContainer">' +
                        '<div class="" id="smartAnnotationCanvasWrapper">' +
                        '<div class="ui dimmer" id="smartAnnotationDimmer">' +
                        '<div class="ui loader">' +
                        '</div>' +
                        '</div>' +
                        '<canvas id="smartAnnotationCanvas"></canvas>' +
                        '</div>' +
                        '</div>');

                    inst.populateDetailedCanvas(true);
                } else {
                    inst.annotator.disableSmartAnnotation();

                    $("#smartAnnotationContainer").remove();
                    //$("#annotationColumn").prepend('<div class="four wide column" id="spacer"></div>');
                }

                inst.addMainCanvas();
                inst.populateCanvas(inst.getUrlFromImageUrl(inst.annotationInfo.imageUrl, inst.annotationInfo.imageUnlocked, this.annotationMode,
                    this.labelAccessorsLookupTable), false);

                showHideSmartAnnotationControls(enabled);
                showHideAutoAnnotationsLoadButton(inst.autoAnnotations);
            },
            beforeChecked: function() {
                if (canvasHasObjects(inst.canvas) > 0) {
                    $('#discardChangesPopup').modal('show');
                    return false;
                }
            },
            beforeUnchecked: function() {
                if (canvasHasObjects(inst.canvas) > 0) {
                    $('#discardChangesPopup').modal('show');
                    return false;
                }
            }
        });


        showHideSmartAnnotationControls(false);


        inst.colorPicker = new Huebee($('#colorPicker')[0], {});
        inst.colorPicker.on('change', function(color, hue, sat, lum) {
            inst.annotator.setStrokeColorOfSelected(color);
        });

        $("#skipAnnotationDropdown").dropdown();

        Mousetrap.bind("r", function() {
            $("#rectMenuItem").trigger("click");
        });


        inst.annotationRefinementsContextMenu = {
            data: [{
                    header: 'Refinements'
                },
                {
                    text: 'Add refinements',
                    action: function(e, selector) {
                        e.preventDefault();

                        if (inst.annotator.getIdOfSelectedItem() !== "") {
                            annotationRefinementsDlg.populateRefinements(inst.annotator.getRefinementsOfSelectedItem());
                            annotationRefinementsDlg.open();
                        }
                    }
                }
            ]
        }

        context.init({
            preventDoubleContext: false
        });

        $("#addAnnotationRefinementsDlgDoneButton").click(function(e) {
            var refs = annotationRefinementsDlg.getSelectedRefinements().split(',');

            if (inst.annotationView === "unified") {
                $("#annotationPropertiesLst").empty();
                var allRefs = annotationRefinementsDlg.getRefinementsUuidMapping();
                for (var i = 0; i < refs.length; i++) {
                    var refIcon = "";
                    if (refs[i] in allRefs)
                        refIcon = allRefs[refs[i]].icon;
                    addRefinementToRefinementsLst(allRefs[refs[i]].name, refs[i], refIcon);
                }
            }
            inst.annotator.setRefinements(refs);
        });



        $("#rectMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setShape("Rectangle");
                inst.changeMenuItem("Rectangle");
            }
        });

        Mousetrap.bind("c", function() {
            $("#circleMenuItem").trigger("click");
        });

        $("#circleMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setShape("Circle");
                inst.changeMenuItem("Circle");
            }
        });

        Mousetrap.bind("p", function() {
            $("#polygonMenuItem").trigger("click");
        });

        Mousetrap.bind("s", function() {
            $("#selectMoveMenutItem").trigger("click");
        });

        $("#polygonMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setShape("Polygon");
                inst.changeMenuItem("Polygon");
            }
        });

        $("#selectMoveMenutItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.disablePanMode();
                inst.annotator.setShape("");
                inst.annotator.enableSelectMoveMode();
                inst.changeMenuItem("SelectMove");
            }
        });



        $("#freeDrawingMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setShape("FreeDrawing");
                inst.annotator.setBrushColor("red");
                inst.changeMenuItem("FreeDrawing");
            }
        });

        Mousetrap.bind("y", function() {
            if (inst.deleteObjectsPopupShown)
                $("#deletedObjectsYesButton").trigger("click");
        });

        Mousetrap.bind("n", function() {
            if (inst.deleteObjectsPopupShown)
                $("#deleteObjectsPopup").modal("hide");
        });

        Mousetrap.bind("del", function() {
            $("#trashMenuItem").trigger("click");
        });

        $("#trashMenuItem").click(function(e) {
            if (isTrashMenuButtonEnabled()) {
                $('#deleteObjectsPopup').modal({
                    onShow: function() {
                        inst.deleteObjectsPopupShown = true;
                    },
                    onHidden: function() {
                        inst.deleteObjectsPopupShown = false;
                    }
                }).modal('show');
            }
        });

        $("#redoMenuItem").click(function(e) {
            if (inst.annotator !== undefined)
                inst.annotator.redo();
        });

        $("#undoMenuItem").click(function(e) {
            if (inst.annotator !== undefined)
                inst.annotator.undo();
        });

        Mousetrap.bind('+', function() {
            zoomIn(inst.canvas);
        });

        $("#zoomInMenuItem").click(function(e) {
            zoomIn(inst.canvas);
        });

        Mousetrap.bind('-', function() {
            zoomOut(inst.canvas);
        });

        $("#zoomOutMenuItem").click(function(e) {
            zoomOut(inst.canvas);
        });

        $("#removeLabelFromUnifiedModeLstDlgYesButton").click(function(e) {
            var removedElemUuid = $("#removeLabelFromUnifiedModeLstDlg").attr("data-to-be-removed-uuid");
            $("#labellstitem-" + removedElemUuid).remove();
            if (removedElemUuid in inst.unifiedModeLabels) {
                var toBeRemovedAnnotationId = "";
                if (inst.unifiedModeLabels[removedElemUuid].sublabel !== undefined && inst.unifiedModeLabels[removedElemUuid].sublabel !== "")
                    toBeRemovedAnnotationId = inst.unifiedModeLabels[removedElemUuid].sublabel + "/" + inst.unifiedModeLabels[removedElemUuid].label;
                else
                    toBeRemovedAnnotationId = inst.unifiedModeLabels[removedElemUuid].label;
                if (toBeRemovedAnnotationId in inst.unifiedModeAnnotations) {
                    delete inst.unifiedModeAnnotations[toBeRemovedAnnotationId];
                    inst.annotator.deleteAll();
                }

                delete inst.unifiedModeLabels[removedElemUuid];

                //after we've deleted the label, we need to highlight another label for annotation
                //(just pick the first one)
                var unifiedModeToolboxChildren = $('#annotationLabelsLst').children('.labelslstitem');
                var firstItem = unifiedModeToolboxChildren.first();
                if (firstItem && firstItem.length === 1)
                    firstItem[0].click();
            }
        });

        $("#removeAnnotationRefinementsDlgYesButton").click(function(e) {
            $("#" + $("#removeAnnotationRefinementsDlg").attr("data-to-be-removed-id")).remove();
            var refs = inst.annotator.getRefinementsOfSelectedItem();
            var idx = refs.indexOf($("#removeAnnotationRefinementsDlg").attr("data-to-be-removed-id").replace("refinementlstitem-", ""));
            if (idx > -1) refs.splice(idx, 1);
            inst.annotator.setRefinements(refs);
        });

        $("#isPluralButton").click(function(e) {
            var pluralLabel = null;
            var currentLabel = "";
            if ($("#label").attr("sublabel") !== "")
                currentLabel = $("#label").attr("sublabel") + $("#label").attr("label");
            else
                currentLabel = $("#label").attr("label");

            if (inst.pluralLabels && currentLabel in inst.pluralLabels) {
                pluralLabel = inst.pluralLabels[currentLabel];
            }

            var isPluralButton = $("#isPluralButton");
            if (isPluralButton.hasClass("basic")) {
                inst.pluralAnnotations = true;
                isPluralButton.removeClass("basic");
                isPluralButton.css("background-color", "white");
                isPluralButton.removeClass("inverted");

                if (pluralLabel)
                    $("#label").text("Annotate all: " + pluralLabel);
            } else {
                inst.pluralAnnotations = false;
                isPluralButton.removeClass("white");
                isPluralButton.addClass("basic");
                isPluralButton.addClass("inverted");
                $("#label").text("Annotate all: " + currentLabel);
            }
        });

        $('#strokeWidthSlider').on('input', function(e) {
            var val = parseInt($(this).val());
            inst.annotator.setStrokeWidthOfSelected(val);
        });

        Mousetrap.bind("ctrl", function(e) {
            if (!e.repeat) { //if the ctrl key is held down, the event constantly fires. we are only interested in the first event
                lastActiveMenuItem = getActiveAnnotationMenuItem(true); //remember active menu item
                $("#panMenuItem").trigger("click");
            }
        }, "keydown");

        Mousetrap.bind("ctrl", function(e) { //ctrl key released
            $("#" + lastActiveMenuItem).trigger("click");
        }, "keyup");

        $("#panMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.enablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setShape("");
                inst.changeMenuItem("PanMode");
            }
        });

        $("#blockSelectMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setShape("Blocks");
                inst.changeMenuItem("BlockSelection");
                inst.annotator.toggleGrid();
            }
        });

        $("#deletedObjectsYesButton").click(function(e) {
            inst.annotator.deleteSelected();
            if (!inst.annotator.objectsSelected())
                $("#trashMenuItem").addClass("disabled");
        });

        $("#smartAnnotationFgMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.changeMenuItem("ForegroundSelection");
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setBrushColor("white"); //do not change color (grabcut requires this!)
                inst.annotator.setBrushWidth(10);
                inst.annotator.setShape("FreeDrawing");
            }
        });

        $("#smartAnnotationBgMenuItem").click(function(e) {
            if (inst.annotator !== undefined && inst.annotator) {
                inst.changeMenuItem("BackgroundSelection");
                inst.annotator.disablePanMode();
                inst.annotator.disableSelectMoveMode();
                inst.annotator.setBrushColor("black"); //do not change color (grabcut requires this!)
                inst.annotator.setBrushWidth(10);
                inst.annotator.setShape("FreeDrawing");
            }
        });

        $("#loadAutoAnnotationsMenuItem").click(function(e) {
            if ((inst.autoAnnotations !== null) && !$("#loadAutoAnnotationsMenuItem").hasClass("disabled")) {
                inst.annotator.loadAutoAnnotations(inst.autoAnnotations, getCanvasScaleFactor());
                $("#loadAutoAnnotationsMenuItem").addClass("disabled"); //once clicked, disable it
                $("#loadAutoAnnotationsMenuItem").removeClass("orange"); //and remove highlight
            }
        });

        $("#discardChangesYesButton").click(function(e) {
            inst.annotator.deleteAll();
            $("#smartAnnotation").checkbox("toggle");
        });

        $('#showSmartAnnotationHelpDlg').click(function() {
            $('#smartAnnotationHelpDlgGif').attr('src', 'img/smart_annotation.gif');
            $('#smartAnnotationHelpDlg').modal('setting', {
                detachable: false
            }).modal('show');
        });

        $("#settingsMenuItem").click(function(e) {
            inst.annotationSettings.setAll();
            $('#annotationSettingsPopup').modal({
                onApprove: function() {
                    if (!/^\d+$/.test($("#annotationPolygonVertexSizeInput").val())) {
                        $('#annotationSettingsPopupWarningMessageBoxContent').text("The polygon vertex size needs to be a numeric value!");
                        $("#annotationSettingsPopupWarningMessageBox").show(200).delay(1500).hide(200);
                        return false;
                    }

                    inst.annotationSettings.persistAll();
                    $('#annotationSettingsPopup').modal('hide');
                    $('#annotationSettingsRefreshBrowserPopup').modal('show');
                }
            }).modal('show');
        });

        $('#blacklistButton').click(function(e) {
            if (!inst.loggedIn) {
                $("#warningMsgText").text("You need to be logged in to perform this action.");
                $("#warningMsg").show(200).delay(1500).hide(200);
                //in case we aren't logged in, do nothing
                return;
            }

            if (inst.loggedIn) {
                var blacklistAnnotationUsageDlgAlreadyShown = localStorage.getItem("blacklistAnnotationUsageDlgShown");
                if (blacklistAnnotationUsageDlgAlreadyShown === null) {
                    $("#blacklistAnnotationUsageDlg").modal("show");
                    localStorage.setItem("blacklistAnnotationUsageDlgShown", true);
                } else {
                    inst.blacklistAnnotation(inst.annotationInfo.validationId);
                }
            } else {
                $("#blacklistAnnotationUsageDlg").modal("show");
            }
        });

        $('#blacklistAnnotationUsageDlgAcceptButton').click(function(e) {
            $("#blacklistAnnotationUsageDlg").modal("hide");
            inst.blacklistAnnotation(inst.annotationInfo.validationId);
        });

        $('#notAnnotableButton').click(function(e) {
            var markAsUnannotatableUsageDlgAlreadyShown = localStorage.getItem("markAsUnannotatableUsageDlgShown");
            if (markAsUnannotatableUsageDlgAlreadyShown === null) {
                $("#markAsUnannotatableUsageDlg").modal("show");
                localStorage.setItem("markAsUnannotatableUsageDlgShown", true);
            } else {
                inst.markAsNotAnnotatable(inst.annotationInfo.validationId);
            }
        });

        $('#markAsUnannotatableUsageDlgAcceptButton').click(function(e) {
            $("#markAsUnannotatableUsageDlg").modal("hide");
            inst.markAsNotAnnotatable(inst.annotationInfo.validationId);
        });

        $('#doneButton').click(function(e) {
            var res = null;

            if (inst.annotationView === "unified") {
                inst.saveCurrentSelectLabelInUnifiedModeList();
                if (Object.keys(inst.unifiedModeLabels).length === 0 && Object.keys(inst.unifiedModeAnnotations).length === 0) {
                    $('#warningMsgText').text('Please annotate the image first.');
                    $('#warningMsg').show(200).delay(1500).hide(200);
                    return;
                }
            } else {
                res = inst.annotator.toJSON((inst.pluralAnnotations ? annotationRefinementsDlg.getPluralAnnotationRefinementUuid() : null));
                if (res.length === 0) { //at least one annotation needs to be there
                    $('#warningMsgText').text('Please annotate the image first.');
                    $('#warningMsg').show(200).delay(1500).hide(200);
                    return;
                }
            }

            if (isLoadingIndicatorVisible()) { //in case smart annotation is currently running
                $('#warningMsgText').text('Smart Annotation is currently in progress.');
                $('#warningMsg').show(200).delay(1500).hide(200);
                return;
            }

            e.preventDefault();

            if (inst.existingAnnotations !== null) {
                inst.updateAnnotations(inst.annotator.toJSON());
            } else {
                if (inst.annotationView === "unified") {
                    if (Object.keys(inst.unifiedModeLabels).length > 0) {
                        var newlyAddedLabelsUnifiedMode = [];
                        for (var key in inst.unifiedModeLabels) {
                            newlyAddedLabelsUnifiedMode.push(inst.unifiedModeLabels[key]);
                        }
                        showHideControls(false, inst.annotationInfo.imageUnlocked);
                        //add missing labels first, then add annotations
                        inst.imageMonkeyApi.labelImage(inst.annotationInfo.imageId, newlyAddedLabelsUnifiedMode)
                            .then(function() {
                                inst.addAnnotationsUnifiedMode();
                            }).catch(function(e) {
                                Sentry.captureException(e);
                            });
                    } else {
                        inst.addAnnotationsUnifiedMode();
                    }
                } else {
                    var annotations = [];
                    var annotation = {};
                    annotation["annotations"] = res;
                    annotation["label"] = $('#label').attr('label');
                    annotation["sublabel"] = $('#label').attr('sublabel');
                    annotations.push(annotation);
                    inst.addAnnotations(annotations);
                }
            }
        });

        $(window).bind('popstate', function() {
            window.location.href = window.location.href;
        });

        changeNavHeader(inst.annotationMode);

        inst.populateDefaultsAndLoadData(inst.annotationMode, inst.validationId, inst.annotationRevision);

        try {
            //can fail in case someone uses uBlock origin or Co.
            new Fingerprint2().get(function(result, components) {
                inst.browserFingerprint = result;
            });
        } catch (e) {}

    }

    return AnnotationView;
}());