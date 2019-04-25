  var canvas, annotator;
  var detailedCanvas = null;
  var numOfPendingRequests = 0;
  var autoAnnotations = null;
  var labelId = null;
  var annotationInfo = new AnnotationInfo();
  var UnifiedModeStates = {uninitialized: 0, fetchedLabels: 1, fetchedAnnotations: 2, initialized: 3};
  var unifiedModePopulated = UnifiedModeStates.uninitialized;
  var annotationSettings = new AnnotationSettings();
  var colorPicker = null;
  var existingAnnotations = null;
  var browserFingerprint = null;
  var deleteObjectsPopupShown = false;
  var unifiedModeAnnotations = {};
  var pluralAnnotations = false;
  var pluralLabels = null;
  var initializeLabelsLstAftLoadDelayed = false;

  {{ if eq .annotationMode "browse" }}
  var browseModeLastSelectedAnnotatorMenuItem = null;
  {{ end }}

  function getActiveAnnotationMenuItem(ignoreAutoLoadButton) {
    var ret = "";
    $("#annotatorMenu").children().each(function (){
      if($(this).hasClass("active")){
        if(ignoreAutoLoadButton) {
          if($(this).attr("id") !== "loadAutoAnnotationsMenuItem") {
            ret = $(this).attr("id");
            return;
          }
        } else {
          ret = $(this).attr("id");
          return;
        }
      }
    });

    return ret;
  }

  function zoomIn() {
    canvas.fabric().setZoom(canvas.fabric().getZoom() * 1.1);
  }

  function zoomOut() {
    canvas.fabric().setZoom(canvas.fabric().getZoom() / 1.1);
  }

  function AnnotationInfo () {
    this.imageId = "";
    this.validationId = "";
    this.origImageWidth = 0;
    this.origImageHeight = 0;
    this.annotationId = "";
    this.imageUrl = "";
    this.imageUnlocked = false;
  }

  function handleUnannotatedImageResponse(data) {
    existingAnnotations = null;
    autoAnnotations = null;
    initializeLabelsLstAftLoadDelayed = false;

    if(data !== null) {
      annotationInfo.imageId = data.uuid;
      annotationInfo.origImageWidth = data.width;
      annotationInfo.origImageHeight = data.height;
      annotationInfo.validationId = data.validation.uuid;
      annotationInfo.imageUrl = data.url;
      annotationInfo.imageUnlocked = data.unlocked;

      if("auto_annotations" in data) {
        if(data["auto_annotations"].length !== 0)
          autoAnnotations = data["auto_annotations"];
      }

      setLabel(data.label.label, data.label.sublabel, data.label.accessor);
      changeNavHeader("default");
    }
    else {
      annotationInfo.imageId = "";
      annotationInfo.origImageWidth = 720; //width of the oops-no-annotation-left image
      annotationInfo.origImageHeight = 720; //height of the oops-no-annotation-left image
      annotationInfo.validationId = "";
      annotationInfo.imageUrl = "";
      annotationInfo.imageUnlocked = false;
    }

    showHideAutoAnnotationsLoadButton();

    if(canvas !== undefined && canvas !== null) {
      annotator.reset();
    }
    addMainCanvas();
    populateCanvas(getUrlFromImageUrl(annotationInfo.imageUrl, annotationInfo.imageUnlocked), false);
    changeControl(annotator);
    numOfPendingRequests = 0;

    if(data === null)
      changeNavHeader("noimage");

    {{ if eq .annotationMode "browse" }}
    if(browseModeLastSelectedAnnotatorMenuItem === null) {
      if(annotator) {
        annotationSettings.loadPreferedAnnotationTool(annotator);
        annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
      }
    }
    else {
      changeMenuItem(browseModeLastSelectedAnnotatorMenuItem);
      annotator.setShape(browseModeLastSelectedAnnotatorMenuItem);
    }
    {{ else }}
    if(annotator) {
      annotationSettings.loadPreferedAnnotationTool(annotator);
      annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
    }
    {{ end }}


    {{ if eq .annotationView "unified" }}
      unifiedModeAnnotations = {};
      if(annotationInfo.imageId !== "")
        populateUnifiedModeToolbox(annotationInfo.imageId);
    {{ end }}

    populateRevisionsDropdown(0, 0);
    showHideRevisionsDropdown();
  }

  function handleAnnotatedImageResponse(data) {
    annotationInfo.imageId = data.image.uuid;
    annotationInfo.origImageWidth = data.image.width;
    annotationInfo.origImageHeight = data.image.height;
    annotationInfo.annotationId = data.uuid;
    annotationInfo.imageUrl = data.image.url;
    annotationInfo.imageUnlocked = data.image.unlocked;

    initializeLabelsLstAftLoadDelayed = false;
    autoAnnotations = null;
    existingAnnotations = data["annotations"];
    showHideAutoAnnotationsLoadButton();

    setLabel(data.validation.label, data.validation.sublabel, null);

    if(canvas !== undefined && canvas !== null) {
      annotator.reset();
    }
    addMainCanvas();
    populateCanvas(getUrlFromImageUrl(data.image.url, data.image.unlocked), false);
    changeControl(annotator);
    numOfPendingRequests = 0;
    showHideControls(true);

    {{ if eq .annotationMode "browse" }}
    if(browseModeLastSelectedAnnotatorMenuItem === null) {
      if(annotator) {
        annotationSettings.loadPreferedAnnotationTool(annotator);
        annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
      }
    }
    else {
      changeMenuItem(browseModeLastSelectedAnnotatorMenuItem);
      annotator.setShape(browseModeLastSelectedAnnotatorMenuItem);
    }
    {{ else }}
    if(annotator) {
      annotationSettings.loadPreferedAnnotationTool(annotator);
      annotator.setPolygonVertexSize(new Settings().getPolygonVertexSize());
    }
    {{ end }}

    populateRevisionsDropdown(data["num_revisions"], data["revision"]);
    showHideRevisionsDropdown();

    {{ if eq .annotationView "unified" }}
      addLabelToLabelLst(data.validation.label, data.validation.sublabel, data.uuid);
      var firstItem = $('#annotationLabelsLst').children('.labelslstitem').first();
      if(firstItem && firstItem.length === 1) {
        $(firstItem[0]).addClass("grey inverted");
      }

      $("#unifiedModeLabelsLstLoadingIndicator").hide();
    {{ end }}
  }

  function changeNavHeader(mode) {
    if(mode === "default") {
      $("#labelContainer").css("margin-top", "-2em");
      $("#navHeader").css("min-height", "290px");
      $("#navHeader").show();
    } else if(mode === "noimage") {
      $("#labelContainer").css("margin-top", "0");
      $("#navHeader").css("min-height", "200px");
      $("#annotationControlsGrid").hide();
      $("#navHeader").show();
    } else {
      $("#labelContainer").css("margin-top", "0");
      $("#navHeader").css("min-height", "200px");
      $("#navHeader").show();
    }
  }



  function getUnannotatedImage(validationId) {
    var url = '';

    if(validationId === undefined)
      url = '{{ .apiBaseUrl }}/v1/annotate?add_auto_annotations=true' + ((labelId === null) ? "" : ("&label_id=" + labelId));
    else
      url = '{{ .apiBaseUrl }}/v1/annotate?validation_id=' + validationId;

    showHideControls(false);

    $.ajax({
      url: url,
      dataType: 'json',
      type: 'GET',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data) {
        handleUnannotatedImageResponse(data);
      },
      error: function (xhr, options, err) {
        handleUnannotatedImageResponse(null);
      }
    });
  }

  function populatePluralsAndLoadData() {
    getPluralLabels(function() {
      {{ if eq .annotationMode "default" }}
        {{ if eq .validationId "" }}
          getUnannotatedImage();
        {{ else }}
          getUnannotatedImage({{ .validationId }});
        {{ end }}
      {{ end }}

      {{ if eq .annotationMode "refine" }}
        {{ if ne .annotationId "" }}
          getAnnotatedImage({{ .annotationId }}, {{ .annotationRevision }});
        {{ end }}
      {{ end }}

      {{ if eq .annotationMode "browse" }}
      $("#loadingSpinner").hide();
      {{ end }}
    });
  }

  function populateUnifiedModeToolbox(imageId) {
    $("#unifiedModeLabelsLstLoadingIndicator").show();
    unifiedModePopulated = UnifiedModeStates.uninitialized;
    getAnnotationsForImage(imageId);
    getLabelsForImage(imageId, true);
  }

  function getAnnotationsForImage(imageId) {
    var url = '';
    if(annotationInfo.imageUnlocked)
      url = '{{ .apiBaseUrl }}/v1/donation/' + imageId + "/annotations";
    else
      url = '{{ .apiBaseUrl }}/v1/unverified-donation/' + imageId + "/annotations" + "?token=" + getCookie("imagemonkey");
    $.ajax({
      url: url,
      dataType: 'json',
      type: 'GET',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data) {
        {{ if eq .annotationView "unified" }}
        for(var i = 0; i < data.length; i++) {
          if(data[i].validation.sublabel !== "")
            unifiedModeAnnotations[data[i].validation.sublabel + "/" + data[i].validation.label] = {annotations: data[i].annotations, 
                                                                                                    label: data[i].validation.label, 
                                                                                                    sublabel: data[i].validation.sublabel,
                                                                                                    dirty: false
                                                                                                   };
          else {
            unifiedModeAnnotations[data[i].validation.label] = {annotations: data[i].annotations, 
                                                                label: data[i].validation.label, 
                                                                sublabel: data[i].validation.sublabel,
                                                                dirty: false
                                                               };
          }
        }
        unifiedModePopulated |= UnifiedModeStates.fetchedAnnotations;

        if(unifiedModePopulated === UnifiedModeStates.initialized) {
          if(canvas.fabric().backgroundImage && canvas.fabric().backgroundImage !== undefined) {
            initializeLabelsLstAftLoadDelayed = false;
            selectLabelInUnifiedLabelsLstAfterLoad();
          }
          else { //image is not yet loaded (which we need before we can initialize the labels list), 
                 //so we need to initialize the labels list when the image is loaded
            initializeLabelsLstAftLoadDelayed = true;
          }
        }
        {{ end }}
      },
      error: function (xhr, options, err) {
      }
    });
  }

  function saveCurrentSelectLabelInUnifiedModeList() {
    if($("#label").attr("sublabel") === "")
      key = $("#label").attr("label");
    else
      key = $("#label").attr("sublabel") + "/" + $("#label").attr("label");

    if(key in unifiedModeAnnotations) {
      var existingAnnos = unifiedModeAnnotations[key].annotations;
      if(annotator.isDirty()) {
        var annos = annotator.toJSON();
        unifiedModeAnnotations[key] = {annotations: annos, label: $("#label").attr("label"), sublabel: $("#label").attr("sublabel"), dirty: true};
      }
    } else {
      var annos = annotator.toJSON();
      if(annos.length > 0) {
        unifiedModeAnnotations[key] = {annotations: annos, label: $("#label").attr("label"), sublabel: $("#label").attr("sublabel"), dirty: true};
      }
    }
  }

  function onRefinementInRefinementsLstClicked(elem) {
    $("#removeAnnotationRefinementsDlg").attr("data-to-be-removed-id", $(elem).attr("id"));
    $("#removeAnnotationRefinementsDlg").modal("show");
  }

  function onLabelInLabelLstClicked(elem) {
    var key = "";

    //before changing label, save existing annotations
    saveCurrentSelectLabelInUnifiedModeList();

    $("#annotationLabelsLst").children().each(function(i) {
      $(this).removeClass("grey inverted");
    }); 

    setLabel($(elem).attr("data-label"), $(elem).attr("data-sublabel"), null);
    $(elem).addClass("grey inverted");

    //clear existing annotations 
    annotator.deleteAll();

    key = "";
    if($(elem).attr("data-sublabel") === "")
      key = $(elem).attr("data-label");
    else
      key = $(elem).attr("data-sublabel") + "/" + $(elem).attr("data-label");

    //show annotations for selected label
    if(key in unifiedModeAnnotations) {
      annotator.loadAnnotations(unifiedModeAnnotations[key].annotations, canvas.fabric().backgroundImage.scaleX);
    }
  }

  function addRefinementToRefinementsLst(name, uuid, icon) {
    var id = "refinementlstitem-" + uuid;
    $("#annotationPropertiesLst").append('<div class="ui segment center aligned refinementlstitem" id="' + id + '"' +
                                        ' data-uuid="' + uuid +
                                          '" onclick="onRefinementInRefinementsLstClicked(this);"' +
                                          'onmouseover="this.style.backgroundColor=\'#e6e6e6\';"' +
                                          'onmouseout="this.style.backgroundColor=\'white\';"' + 
                                          'style="overflow: auto;">' +
                                          '<span class="left-floated">' +
                                            '<p><i class="' + icon + ' icon"></i> ' + name + '</p>' + 
                                          '</span><span class="right-floated"><i class="right icon delete ui red"></i></span></div>');
  }

  function addLabelToLabelLst(label, sublabel, uuid) {
    var id = "labellstitem-" + uuid;
    var displayedLabel = ((sublabel === "") ? label : sublabel + "/" + label);
    $("#annotationLabelsLst").append('<div class="ui segment center aligned labelslstitem" id="' + id + '"' +
                                        ' data-label="' + label + '" data-uuid="' + uuid +
                                        '" data-sublabel="' + sublabel + 
                                          '" onclick="onLabelInLabelLstClicked(this);"' +
                                          'onmouseover="this.style.backgroundColor=\'#e6e6e6\';"' +
                                          'onmouseout="this.style.backgroundColor=\'white\';"' + 
                                          'style="overflow: auto;"><p>' + displayedLabel + '</p></div>');
  }

  function getLabelsForImage(imageId, onlyUnlockedLabels) {
    var url = '{{ .apiBaseUrl }}/v1/donation/' + imageId + "/labels?only_unlocked_labels=" + (onlyUnlockedLabels ? "true" : "false");
    $.ajax({
      url: url,
      dataType: 'json',
      type: 'GET',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data) {
        {{ if eq .annotationView "unified" }}
        if(data !== null) {
          for(var i = 0; i < data.length; i++) {
            addLabelToLabelLst(data[i].label, '', data[i].uuid);
            if(data[i].sublabels !== null) {
              for(var j = 0; j < data[i].sublabels.length; j++) {
                addLabelToLabelLst(data[i].label, data[i].sublabels[j].name, 
                                    data[i].sublabels[j].uuid);
              }
            }
          }
        }
        unifiedModePopulated |= UnifiedModeStates.fetchedLabels;

        if(unifiedModePopulated === UnifiedModeStates.initialized) {
          if(canvas.fabric().backgroundImage && canvas.fabric().backgroundImage !== undefined) {
            initializeLabelsLstAftLoadDelayed = false;
            selectLabelInUnifiedLabelsLstAfterLoad();
          }
          else { //image is not yet loaded (which we need before we can initialize the labels list), 
                 //so we need to initialize the labels list when the image is loaded
            initializeLabelsLstAftLoadDelayed = true;
          }
        }
        {{ end }}
      },
      error: function (xhr, options, err) {
      }
    });
  }

  function selectLabelInUnifiedLabelsLstAfterLoad() {
    var unifiedModeToolboxChildren = $('#annotationLabelsLst').children('.labelslstitem');
    var foundLabelInUnifiedModeToolbox = false;
    unifiedModeToolboxChildren.each(function(index, value) {
      if(($(this).attr("data-label") === $("#label").attr("label")) && ($(this).attr("data-sublabel") === $("#label").attr("sublabel"))) {
        foundLabelInUnifiedModeToolbox = true;
        $(this).click();
        return false;
      }  
    })

    //when label not found, select first one in list
    if(!foundLabelInUnifiedModeToolbox) {
      var firstItem = unifiedModeToolboxChildren.first();
      if(firstItem && firstItem.length === 1)
        firstItem[0].click();
    }

    $("#unifiedModeLabelsLstLoadingIndicator").hide();
  }

  function getAnnotatedImage(annotationId, annotationRevision) {
    var url = '{{ .apiBaseUrl }}/v1/annotation?annotation_id=' + annotationId;

    if(annotationRevision !== -1)
      url += '&rev=' + annotationRevision;

    showHideControls(false);

    $.ajax({
      url: url,
      dataType: 'json',
      type: 'GET',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data){
        handleAnnotatedImageResponse(data);

        //if there are already annotations, do not show blacklist or unannotatable button
        $("#blacklistButton").hide();
        $("#notAnnotableButton").hide();
      }
    });
  }

  function blacklistAnnotation(validationId) {
    showHideControls(false);
    var url = '{{ .apiBaseUrl }}/v1/validation/' + validationId + '/blacklist-annotation';
    $.ajax({
      url: url,
      type: 'POST',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data, status, xhr){
        {{ if eq .annotationMode "default" }}
        getUnannotatedImage();
        {{ else }}
        $("#loadingSpinner").hide();
        clearDetailedCanvas();
        annotator.reset();
        showBrowseAnnotationImageGrid();
        {{ end }}
      }
    });
  }

  function markAsNotAnnotatable(validationId) {
    showHideControls(false);
    var url = '{{ .apiBaseUrl }}/v1/validation/' + validationId + '/not-annotatable';
    $.ajax({
      url: url,
      type: 'POST',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data, status, xhr){
        {{ if eq .annotationMode "default" }}
        getUnannotatedImage();
        {{ else }}
        $("#loadingSpinner").hide();
        clearDetailedCanvas();
        annotator.reset();
        showBrowseAnnotationImageGrid();
        {{ end }}
      }
    });
  }

  function getUrlFromImageUrl(imageUrl, imageUnlocked){
    var url = (imageUrl === "" ? "img/oops-no-annotation-left.png" : imageUrl);

    if(imageUrl !== ""){
      if(!imageUnlocked){
        url += "?token=" + getCookie("imagemonkey");
      }

      {{ if eq .annotationMode "browse" }}
      if($("#highlightParentAnnotationsCheckbox").checkbox("is checked")) {
        var labelToAnnotate = $("#label").attr("accessor");
        if(labelToAnnotate in labelAccessorsLookupTable) {
          if(!imageUnlocked) {
            url += "&highlight=" + encodeURIComponent(labelAccessorsLookupTable[labelToAnnotate]);
          } else {
            url += "?highlight=" + encodeURIComponent(labelAccessorsLookupTable[labelToAnnotate]);
          }
        }
      }
      {{ end }}
    }

    return url;
  }

  function isTrashMenuButtonEnabled(){
    return !$("#trashMenuItem").hasClass("disabled");
  }

  function isSmartAnnotationEnabled(){
    var obj = $("smartAnnotation");
    if(obj.length) {
      return obj.checkbox("is checked");
    }
    return false;
  }

  function changeMenuItem(type){
    var id = "";
    if(type === 'Rectangle')
      id = "rectMenuItem";
    else if(type === "Circle")
      id = "circleMenuItem";
    else if(type === "Polygon")
      id = "polygonMenuItem";
    else if(type === "PanMode")
      id = "panMenuItem";
    else if(type === "BlockSelection")
      id = "blockSelectMenuItem";
    else if(type === "FreeDrawing")
      id = "freeDrawingMenuItem";
    else if(type === "ForegroundSelection")
      id = "smartAnnotationFgMenuItem";
    else if(type === "BackgroundSelection")
      id = "smartAnnotationBgMenuItem";
    else if(type === "SelectMove")
      id = "selectMoveMenutItem";

    {{ if eq .annotationMode "browse" }}
    browseModeLastSelectedAnnotatorMenuItem = type;
    {{ end }}

    $("#annotatorMenu").children().each(function (){
      if($(this).attr("id") === id)
        $(this).addClass("active");
      else
        $(this).removeClass("active");
    });

  }


  function showHideSmartAnnotationControls(show){
    if(show){
      $("#circleMenuItem").hide();
      $("#polygonMenuItem").hide();
      $("#smartAnnotationFgMenuItem").show();
      $("#smartAnnotationBgMenuItem").show();
    }
    else{
      $("#circleMenuItem").show();
      $("#polygonMenuItem").show();
      $("#smartAnnotationFgMenuItem").hide();
      $("#smartAnnotationBgMenuItem").hide();
    }

    $('#annotatorMenu .item').popup({
      inline: true,
      hoverable: true
    });
  }

  function showHideControls(show){
    if(show){
      $("#doneButton").show();
      $("#blacklistButton").show();
      $("#notAnnotableButton").show();
      $("#annotatorMenu").show();
      $("#smartAnnotation").show();
      $("#showSmartAnnotationHelpDlg").show();
      $("#annotationControlsGrid").show();
      $("#annotationControlsMainArea").show();
      $("#annotationButtons").show();
      $("#loadingSpinner").hide();

      if(annotationInfo.imageUnlocked)
        $("#imageLockedLabel").hide();
      else
        $("#imageLockedLabel").show();

      $("#annotationColumnContent").show();
      $("#annotationColumnSpacer").show();
      $("#annotationPropertiesColumnSpacer").show();
    }
    else{
      $("#doneButton").hide();
      $("#blacklistButton").hide();
      $("#notAnnotableButton").hide();
      $("#annotatorMenu").hide();
      $("#smartAnnotation").hide();
      $("#showSmartAnnotationHelpDlg").hide();
      $("#annotationControlsGrid").hide();
      $("#annotationControlsMainArea").hide();
      $("#annotationButtons").hide();
      $("#loadingSpinner").show();
      $("#imageLockedLabel").hide();
      $("#annotationColumnContent").hide();
      $("#annotationColumnSpacer").hide();
      $("#annotationPropertiesColumnSpacer").hide();
    }
  }

  function pollUntilProcessed(uuid) {
    var url = "{{ .playgroundBaseUrl }}/v1/grabcut/" + uuid;
    $.getJSON(url, function (response) {
      if(jQuery.isEmptyObject(response))
        setTimeout(pollUntilProcessed(uuid), 1000);
      else{
        detailedCanvas.clearObjects();

        if(response["result"]["points"].length > 0){
          var data = [];
          data.push(response["result"]);
          annotator.setSmartAnnotationData(data);
          detailedCanvas.drawAnnotations(data, $("#smartAnnotationCanvasWrapper").attr("scaleFactor"));
        }
        
        numOfPendingRequests -= 1;
        if(numOfPendingRequests <= 0){
          $("#smartAnnotationCanvasWrapper").dimmer("hide");
          numOfPendingRequests = 0;
        }
      }
    });
  }

  function populateDetailedCanvas(force = false){
    if((detailedCanvas !== null) && !force)
      detailedCanvas.clear();
    else
      detailedCanvas = new CanvasDrawer("smartAnnotationCanvas", 0, 0);

    var maxWidth = document.getElementById("smartAnnotationContainer").clientWidth - 50; //margin
    var scaleFactor = maxWidth/annotationInfo.origImageWidth;
    if(scaleFactor > 1.0)
      scaleFactor = 1.0;

    var w = annotationInfo.origImageWidth * scaleFactor;
    var h = annotationInfo.origImageHeight * scaleFactor;

    $("#smartAnnotationCanvasWrapper").attr("width", w);
    $("#smartAnnotationCanvasWrapper").attr("height", h);
    $("#smartAnnotationCanvasWrapper").attr("scaleFactor", scaleFactor);
    //detailedCanvas = new CanvasDrawer("smartAnnotationCanvas", w, h);
    detailedCanvas.setWidth(w);
    detailedCanvas.setHeight(h);
    detailedCanvas.setCanvasBackgroundImage(canvas.fabric().backgroundImage, null);
  }

  function clearDetailedCanvas(){
    if(detailedCanvas !== null){
      detailedCanvas.clear();
    }
  }

  function getCanvasScaleFactor(){
    var maxWidth = document.getElementById("annotationAreaContainer").clientWidth - 50; //margin
    var scaleFactor = maxWidth/annotationInfo.origImageWidth;
    if(scaleFactor > 1.0)
      scaleFactor = 1.0;
    return scaleFactor;
  }

  function populateCanvas(backgroundImageUrl, initAnnotator, force=true){
    if((canvas !== null) && !force)
      annotator.reset();
    else{
      canvas = new CanvasDrawer("annotationArea");
      canvas.fabric().selection = false;
      annotator = new Annotator(canvas.fabric(), onAnnotatorObjectSelected, onAnnotatorMouseUp, onAnnotatorObjectDeselected);
    }

    var scaleFactor = getCanvasScaleFactor();

    var w = annotationInfo.origImageWidth * scaleFactor;
    var h = annotationInfo.origImageHeight * scaleFactor;

    $("#annotationAreaContainer").attr("width", w);
    $("#annotationAreaContainer").attr("height", h);
    $("#annotationAreaContainer").attr("scaleFactor", scaleFactor);
    canvas.setWidth(w);
    canvas.setHeight(h);

    if(initAnnotator){
      canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, function() {
        annotator.initHistory();
        onCanvasBackgroundImageSet();

      });
    } 
    else{
      canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, onCanvasBackgroundImageSet);
    }
  }

  function dataURItoBlob(dataURI) {
    var byteString = atob(dataURI.split(',')[1]);
    var ab = new ArrayBuffer(byteString.length);
    var ia = new Uint8Array(ab);
    for (var i = 0; i < byteString.length; i++) {
        ia[i] = byteString.charCodeAt(i);
    }
    return new Blob([ab], {type: 'image/png'});
  }

  function grabCutMe(){
    numOfPendingRequests += 1;
    $("#smartAnnotationCanvasWrapper").dimmer("show");
    var blob = dataURItoBlob(annotator.getMask());
    var formData = new FormData()
    formData.append('image', blob);
    formData.append('uuid', annotationInfo.imageId);
    $.ajax({
      url: '{{ .playgroundBaseUrl }}/v1/grabcut',
      processData: false,
      contentType: false,
      data: formData,
      type: 'POST',
      success: function(data, status, xhr){
        pollUntilProcessed(xhr.getResponseHeader("Location"));
      }
    });
  }

  function changeControl(annotator){
    if(annotationInfo.imageId === ""){
      $("#labelContainer").hide();
      $("#doneButton").hide();
      $("#bottomLabel").hide();
      $("#isPluralContainer").hide();
      annotator.block();
    }
    else{
      $("#labelContainer").show();
      $("#doneButton").show();
      $("#bottomLabel").show();
      $("#isPluralContainer").show();
      annotator.unblock();
    }
  }

  function setLabel(label, sublabel, accessor){
    $("#label").attr("label", label);
    $("#label").attr("sublabel", sublabel);

    if(accessor !== null)
      $("#label").attr("accessor", accessor);

    if(sublabel === ""){
      $("#label").text(("Annotate all: " + label));
      $("#bottomLabel").text(("Annotate all: " + label));
      $("#isPluralButton").attr("data-tooltip", "Set in case you want to annotate multiple " + label + " objects at once");
    } 
    else {
      $("#label").text(("Annotate all: " + sublabel + "/" + label));
      $("#bottomLabel").text(("Annotate all: " + sublabel + "/" + label));
      $("#isPluralButton").attr("data-tooltip", "Set in case you want to annotate multiple " + sublabel + "/" + label + " objects at once");
    }
  }

  function onAnnotatorObjectSelected(){
    if(annotator.objectsSelected()) {
        if(annotator.isSelectMoveModeEnabled()) {
          $("#trashMenuItem").removeClass("disabled");
          $("#propertiesMenuItem").removeClass("disabled");

          var strokeColor = annotator.getStrokeColorOfSelected();
          if(strokeColor !== null)
            colorPicker.setColor(strokeColor);
        }

        {{ if eq .annotationView "unified" }}
        //when object is selected, show refinements
        var refs = annotator.getRefinementsOfSelectedItem();
        var refsUuidMapping = annotationRefinementsDlg.getRefinementsUuidMapping();
        for(var i = 0; i < refs.length; i++) {
          if(refs[i] in refsUuidMapping) {
            addRefinementToRefinementsLst(refsUuidMapping[refs[i]].name, refs[i], refsUuidMapping[refs[i]].icon);
          }
        }
        $("#addRefinementButton").removeClass("disabled");
        $("#addRefinementButtonTooltip").removeAttr("data-tooltip");
        context.attach('#annotationColumn', annotationRefinementsContextMenu.data);

        {{ end }}
    } else {
      $("#trashMenuItem").addClass("disabled");
      $("#propertiesMenuItem").addClass("disabled");
    }
  }

  function onAnnotatorObjectDeselected() {
    context.destroy('#annotationColumn');

    {{ if eq .annotationView "unified" }}
    $("#annotationPropertiesLst").empty();
    $("#addRefinementButton").addClass("disabled");
    $("#addRefinementButtonTooltip").attr("data-tooltip", "Select a annotation first")
    {{ end }}
  }

  function onAnnotatorMouseUp(){
    if(isSmartAnnotationEnabled() && !annotator.isPanModeEnabled())
      grabCutMe();
  }

  function canvasHasObjects(){
    if(canvas.fabric().getObjects().length > 0)
      return true;
    return false;
  }

  function onCanvasBackgroundImageSet(){
    if(isSmartAnnotationEnabled())
      populateDetailedCanvas();

    if(existingAnnotations !== null) {
      annotator.loadAnnotations(existingAnnotations, canvas.fabric().backgroundImage.scaleX);
      //drawAnnotations(canvas.fabric(), existingAnnotations, canvas.fabric().backgroundImage.scaleX);
      existingAnnotations = annotator.toJSON(); //export JSON after loading annotations 
                                                //due to rounding we might end up with slightly different values, so we
                                                //export them in order to make sure that we don't accidentially detect 
                                                //a rounding errors as changes.
    }

    showHideControls(true);
    $("#annotationArea").css({"border-width":"1px",
                              "border-style": "solid",
                              "border-color": "#000000"});


    {{ if eq .annotationView "unified" }}
    if(initializeLabelsLstAftLoadDelayed) {
      selectLabelInUnifiedLabelsLstAfterLoad();
      initializeLabelsLstAftLoadDelayed = false;
    }
    {{ end }}
  }

  function isLoadingIndicatorVisible(){
    return $("#smartAnnotationDimmer").is(":visible"); 
  }

  function showHideAutoAnnotationsLoadButton(){
    if(autoAnnotations && (autoAnnotations.length > 0) && (!isSmartAnnotationEnabled())){
      $("#loadAutoAnnotationsMenuItem").show();
      $("#loadAutoAnnotationsMenuItem").removeClass("disabled");
      $("#loadAutoAnnotationsMenuItem").addClass("orange");
    }
    else{
      $("#loadAutoAnnotationsMenuItem").hide();
    }
  }

  function addMainCanvas() {
    $("#annotationColumnSpacer").remove();
    $("#annotationPropertiesColumnSpacer").remove();
    $("#annotationColumnContent").remove();

    var spacer = '';
    var unifiedModePropertiesLst = '';
    var w = "sixteen";
    if(isSmartAnnotationEnabled()) {
      w = "eight";

    }
    else {
      var workspaceSize = annotationSettings.loadWorkspaceSize();
      {{ if eq .annotationView "unified" }}
        if(workspaceSize === "small"){
          w = "eight";
          spacer = '<div class="four wide column" id="annotationColumnSpacer">' +
                      '<h2 class="ui center aligned header">' + 
                        '<div class="content">' +
                          'Labels' +
                        '</div>' + 
                      '</h2>' +
                      '<div class="ui basic segment">'
                        '<div class="ui segments">' +
                          '<div class="ui raised segments" style="overflow: auto; height: 50vh;" id="annotationLabelsLst">' +
                            '<div class="ui active indeterminate loader" id="unifiedModeLabelsLstLoadingIndicator"></div>' +
                          '</div>' + 
                        '</div>' + 
                      '</div>' +
                    '</div>';
          unifiedModePropertiesLstWidth = 'four';
        }
        else if(workspaceSize === "medium"){
          w = "ten";
          spacer = '<div class="three wide column" id="annotationColumnSpacer">' +
                      '<h2 class="ui center aligned header">' + 
                        '<div class="content">' +
                          'Labels' +
                        '</div>' + 
                      '</h2>' +
                      '<div class="ui basic segment">' +
                        '<div class="ui segments">' +
                          '<div class="ui raised segments" style="overflow: auto; height: 50vh;" id="annotationLabelsLst">' +
                            '<div class="ui active indeterminate loader" id="unifiedModeLabelsLstLoadingIndicator"></div>' +
                          '</div>' + 
                        '</div>' +
                      '</div>' +
                    '</div>';
          unifiedModePropertiesLstWidth = 'three';
        }
        else if(workspaceSize === "big"){
          w = "ten";
          spacer = '<div class="three wide column" id="annotationColumnSpacer">' +
                      '<h2 class="ui center aligned header">' + 
                        '<div class="content">' +
                          'Labels' +
                        '</div>' + 
                      '</h2>' +
                      '<div class="ui basic segment">' +
                        '<div class="ui segments">' +
                          '<div class="ui raised segments" style="overflow: auto; height: 50vh;" id="annotationLabelsLst">' +
                            '<div class="ui active indeterminate loader" id="unifiedModeLabelsLstLoadingIndicator"></div>' +
                          '</div>' + 
                        '</div>' + 
                      '</div>' +
                    '</div>';
          unifiedModePropertiesLstWidth = 'three';
        }


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
                                          
                                          '<div class="ui form">' +
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

                                        

                                       //'<div class="ui bottom attached segment"><div id="addRefinementButtonTooltip" data-tooltip="Select a annotation first"><div class="ui fluid disabled button" id="addRefinementButton"><i class="plus icon"></i>Add</div></div></div>' + 
                                      '</div>' + 
                                   '</div>';

      {{ else }}
        if(workspaceSize === "small"){
          w = "eight";
          spacer = '<div class="four wide column" id="annotationColumnSpacer"></div>';
        }
        else if(workspaceSize === "medium"){
          w = "ten";
          spacer = '<div class="three wide column" id="annotationColumnSpacer"></div>';
        }
        else if(workspaceSize === "big")
          w = "sixteen";
      {{ end }}
    }

 
    var data =  spacer +
                '<div class="' + w +' wide center aligned column" id="annotationColumnContent">' +
                 '<div id="annotationAreaContainer">' +
                    '<canvas id="annotationArea" imageId=""></canvas>' +
                 '</div>' +
                '</div>' + unifiedModePropertiesLst;

    $("#annotationColumn").show();
    $("#annotationColumn").append(data);

    {{ if eq .annotationView "unified" }}
    $("#addRefinementButton").click(function(e) {
      var refs = annotator.getRefinementsOfSelectedItem();
      if(refs.indexOf($('#addRefinementDropdown').dropdown('get value')) == -1 ) {
        refs.push($('#addRefinementDropdown').dropdown('get value'));
        annotator.setRefinements(refs);
        var allRefs = annotationRefinementsDlg.getRefinementsUuidMapping();
        var refIcon = "";
        if($('#addRefinementDropdown').dropdown('get value') in allRefs)
          refIcon = allRefs[$('#addRefinementDropdown').dropdown('get value')].icon;
        addRefinementToRefinementsLst($('#addRefinementDropdown').dropdown('get text'), $('#addRefinementDropdown').dropdown('get value'), refIcon);
      }
      $('#addRefinementDropdown').dropdown('restore placeholder text');
    });
    var refs = annotationRefinementsDlg.getRefinementsUuidMapping();
    for(k in refs) {
      if(refs.hasOwnProperty(k)) {
        var entry = '<div class="item" data-value="' + k + '">';
        if(refs[k].icon !== "") {
          entry += '<i class="' + refs[k].icon + ' icon"></i>';
        }
        entry += (refs[k].name + '</div>');
        $("#addRefinementDropdownMenu").append(entry);
      }
    }
    $("#addRefinementDropdown").dropdown();
    {{ end }}

    $("#annotationArea").attr("imageId", annotationInfo.imageId);
    $("#annotationArea").attr("origImageWidth", annotationInfo.origImageWidth);
    $("#annotationArea").attr("origImageHeight", annotationInfo.origImageHeight);
    $("#annotationArea").attr("validationId", annotationInfo.validationId);
  }

  function handleUpdateAnnotationsRes(res) {
    {{ if eq .annotationMode "browse" }}
    $("#loadingSpinner").hide();
    updateAnnotationsForImage(annotationInfo.annotationId, res);
    showBrowseAnnotationImageGrid();
    {{ end }}

    {{ if eq .onlyOnce true }}
    showHideControls(false);
    $("#onlyOnceDoneMessageContainer").show();
    $("#onlyOnceDoneMessage").fadeIn("slow");
    $("#loadingSpinner").hide();
    {{ end }}
  }

  function updateAnnotations(res) {
    
    if(_.isEqual(res, existingAnnotations)) {
      showHideControls(false);
      clearDetailedCanvas();
      annotator.reset();
      handleUpdateAnnotationsRes(existingAnnotations);
      return;
    }

    var postData = {}
    postData["annotations"] = res;

    var headers = {}
    if(browserFingerprint !== null)
      headers["X-Browser-Fingerprint"] = browserFingerprint;

    headers['X-App-Identifier'] = '{{ .appIdentifier }}';

    showHideControls(false);
    clearDetailedCanvas();
    annotator.reset();

    var url = "{{ .apiBaseUrl }}/v1/annotation/" + annotationInfo.annotationId;
    $.ajax({
      url: url,
      type: 'PUT',
      data: JSON.stringify(postData),
      headers: headers,
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data){
        handleUpdateAnnotationsRes(res);
      }
    });
  }

  function onAddAnnotationsDone() {
    {{ if eq .annotationMode "default" }}
    getUnannotatedImage();
    {{ else }}
    $("#loadingSpinner").hide();
    changeNavHeader("browse");
    showBrowseAnnotationImageGrid();
    {{ end }}

    {{ if eq .onlyOnce true }}
    $("#onlyOnceDoneMessage").fadeIn("slow");
    showHideControls(false);
    $("#loadingSpinner").hide();
    {{ end }}
  }

  function addAnnotations(annotations) {
    var headers = {}
    if(browserFingerprint !== null)
      headers["X-Browser-Fingerprint"] = browserFingerprint;

    headers['X-App-Identifier'] = '{{ .appIdentifier }}';

    showHideControls(false);
    clearDetailedCanvas();
    annotator.reset();

    if(annotations.length === 0) {
      onAddAnnotationsDone();
      return;
    }

    var url = "{{ .apiBaseUrl }}/v1/donation/" + annotationInfo.imageId + "/annotate";
    $.ajax({
      url: url,
      type: 'POST',
      data: JSON.stringify(annotations),
      headers: headers,
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data){
        onAddAnnotationsDone();
      }
    });
  }

  function getPluralLabels(onDoneCallback) {
    {{ if ne .annotationMode "browse" }}
    $("#loadingSpinner").show();
    {{ end }}

    var url = "{{ .apiBaseUrl }}/v1/label/plurals";
    $.ajax({
      url: url,
      type: 'GET',
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", "Bearer " + getCookie("imagemonkey"))
      },
      success: function(data) {
        pluralLabels = data;

        onDoneCallback();
      }
    });
  }
  

  $(document).ready(function(){
      var lastActiveMenuItem = "";
      $('#warningMsg').hide();
      
      $('#smartAnnotation').checkbox({
        onChange : function() {
          var enabled = isSmartAnnotationEnabled();
          if(enabled){
            annotator.enableSmartAnnotation();

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

            populateDetailedCanvas(true);
          }
          else{
            annotator.disableSmartAnnotation();

            $("#smartAnnotationContainer").remove();
            //$("#annotationColumn").prepend('<div class="four wide column" id="spacer"></div>');
          }

          addMainCanvas();
          populateCanvas(getUrlFromImageUrl(annotationInfo.imageUrl, annotationInfo.imageUnlocked), false);

          showHideSmartAnnotationControls(enabled);
          showHideAutoAnnotationsLoadButton();
        },
        beforeChecked : function() {
          if(canvasHasObjects() > 0){
            $('#discardChangesPopup').modal('show');
            return false;
          }
        },
        beforeUnchecked : function() {
          if(canvasHasObjects() > 0){
            $('#discardChangesPopup').modal('show');
            return false;
          }
        }
      });

      
      showHideSmartAnnotationControls(false);
      

      colorPicker = new Huebee($('#colorPicker')[0], {});
      colorPicker.on('change', function(color, hue, sat, lum) {
        annotator.setStrokeColorOfSelected(color);
      });

      $("#skipAnnotationDropdown").dropdown();

      Mousetrap.bind("r", function() { 
        $("#rectMenuItem").trigger("click");
      });


      annotationRefinementsContextMenu = {
        data: [
          {header: 'Refinements'},
          {text: 'Add refinements', action: function(e, selector) {
            e.preventDefault();

            if(annotator.getIdOfSelectedItem() !== "") {
              annotationRefinementsDlg.populateRefinements(annotator.getRefinementsOfSelectedItem());
              annotationRefinementsDlg.open();
            }
          }}
        ]
      }

      context.init({preventDoubleContext: false});

      $("#addAnnotationRefinementsDlgDoneButton").click(function(e) {
        annotator.setRefinements(annotationRefinementsDlg.getSelectedRefinements().split(','));
      });
      


      $("#rectMenuItem").click(function(e) {
        if(annotator !== undefined) {
          annotator.disablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setShape("Rectangle");
          changeMenuItem("Rectangle");
        }
      });

      Mousetrap.bind("c", function() { 
        $("#circleMenuItem").trigger("click");
      });

      $("#circleMenuItem").click(function(e) {
        if(annotator !== undefined) {
          annotator.disablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setShape("Circle");
          changeMenuItem("Circle");
        }
      });

      Mousetrap.bind("p", function() {
        $("#polygonMenuItem").trigger("click");
      });

      Mousetrap.bind("s", function() { 
        $("#selectMoveMenutItem").trigger("click");
      });

      $("#polygonMenuItem").click(function(e) {
        if(annotator !== undefined) {
          annotator.disablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setShape("Polygon");
          changeMenuItem("Polygon");
        }
      });

      $("#selectMoveMenutItem").click(function(e) {
        if(annotator !== undefined) {
          annotator.disablePanMode();
          annotator.setShape("");
          annotator.enableSelectMoveMode();
          changeMenuItem("SelectMove");
        }
      });

      

       $("#freeDrawingMenuItem").click(function(e) {
        if(annotator !== undefined) {
          annotator.disablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setShape("FreeDrawing");
          annotator.setBrushColor("red");
          changeMenuItem("FreeDrawing");
        }
      });

      Mousetrap.bind("y", function() {
        if(deleteObjectsPopupShown)
          $("#deletedObjectsYesButton").trigger("click");
      });

      Mousetrap.bind("n", function() {
        if(deleteObjectsPopupShown)
          $("#deleteObjectsPopup").modal("hide");
      });

      Mousetrap.bind("del", function() {
        $("#trashMenuItem").trigger("click");
      });

      $("#trashMenuItem").click(function(e) {
        if(isTrashMenuButtonEnabled()) {
          $('#deleteObjectsPopup').modal({
            onShow: function() {
              deleteObjectsPopupShown = true;
            },
            onHidden: function() {
              deleteObjectsPopupShown = false;
            }
          }).modal('show');
        }
      });

      $("#redoMenuItem").click(function(e) {
        if(annotator !== undefined)
          annotator.redo();
      });

      $("#undoMenuItem").click(function(e) {
        if(annotator !== undefined)
          annotator.undo();
      });

      Mousetrap.bind('+', function() { 
        zoomIn(); 
      });

      $("#zoomInMenuItem").click(function(e) {
        zoomIn();
      });

      Mousetrap.bind('-', function() { 
        zoomOut(); 
      });

      $("#zoomOutMenuItem").click(function(e) {
        zoomOut();
      });

      $("#removeAnnotationRefinementsDlgYesButton").click(function(e) {
        $("#"+$("#removeAnnotationRefinementsDlg").attr("data-to-be-removed-id")).remove();
        var refs = annotator.getRefinementsOfSelectedItem();
        var idx = refs.indexOf($("#removeAnnotationRefinementsDlg").attr("data-to-be-removed-id").replace("refinementlstitem-", ""));
        if(idx > -1) refs.splice(idx, 1);
        annotator.setRefinements(refs);
      });

      $("#isPluralButton").click(function(e) {
        var pluralLabel = null;
        var currentLabel = "";
        if ($("#label").attr("sublabel") !== "")
          currentLabel = $("#label").attr("sublabel") + $("#label").attr("label");
        else
          currentLabel = $("#label").attr("label");

        if(pluralLabels && currentLabel in pluralLabels) {
          pluralLabel = pluralLabels[currentLabel];
        }

        var isPluralButton = $("#isPluralButton");
        if(isPluralButton.hasClass("basic")) {
          pluralAnnotations = true;
          isPluralButton.removeClass("basic");
          isPluralButton.css("background-color", "white");
          isPluralButton.removeClass("inverted");

          if(pluralLabel)
            $("#label").text("Annotate all: " + pluralLabel);
        } else {
          pluralAnnotations = false;
          isPluralButton.removeClass("white");
          isPluralButton.addClass("basic");
          isPluralButton.addClass("inverted");
          $("#label").text("Annotate all: " + currentLabel);
        }
      });

      $('#strokeWidthSlider').on('input', function(e) {
        var val = parseInt($(this).val());
        annotator.setStrokeWidthOfSelected(val);
      });

      Mousetrap.bind("ctrl", function(e) {
        if(!e.repeat) { //if the ctrl key is held down, the event constantly fires. we are only interested in the first event 
          lastActiveMenuItem = getActiveAnnotationMenuItem(true); //remember active menu item
          $("#panMenuItem").trigger("click");
        }
      }, "keydown");

      Mousetrap.bind("ctrl", function(e) { //ctrl key released
        $("#"+lastActiveMenuItem).trigger("click");
      }, "keyup");

      $("#panMenuItem").click(function(e) {
        if(annotator !== undefined) {
          annotator.enablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setShape("");
          changeMenuItem("PanMode");
        }
      });

      $("#blockSelectMenuItem").click(function(e) {
        annotator.disablePanMode();
        annotator.disableSelectMoveMode();
        annotator.setShape("Blocks");
        changeMenuItem("BlockSelection");
        annotator.toggleGrid();
      });

      $("#deletedObjectsYesButton").click(function(e) {
        annotator.deleteSelected();
        if(!annotator.objectsSelected())
          $("#trashMenuItem").addClass("disabled");
      });

      $("#smartAnnotationFgMenuItem").click(function(e) {
        if(annotator !== undefined) {
          changeMenuItem("ForegroundSelection");
          annotator.disablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setBrushColor("white"); //do not change color (grabcut requires this!)
          annotator.setBrushWidth(10);
          annotator.setShape("FreeDrawing");
        }
      });

      $("#smartAnnotationBgMenuItem").click(function(e) {
        if(annotator !== undefined) {
          changeMenuItem("BackgroundSelection");
          annotator.disablePanMode();
          annotator.disableSelectMoveMode();
          annotator.setBrushColor("black"); //do not change color (grabcut requires this!)
          annotator.setBrushWidth(10);
          annotator.setShape("FreeDrawing");
        }
      });

      $("#loadAutoAnnotationsMenuItem").click(function(e) {
        if((autoAnnotations !== null) && !$("#loadAutoAnnotationsMenuItem").hasClass("disabled")){
          annotator.loadAutoAnnotations(autoAnnotations, getCanvasScaleFactor());
          $("#loadAutoAnnotationsMenuItem").addClass("disabled"); //once clicked, disable it
          $("#loadAutoAnnotationsMenuItem").removeClass("orange"); //and remove highlight
        }
      });

      $("#discardChangesYesButton").click(function(e) {
        annotator.deleteAll();
        $("#smartAnnotation").checkbox("toggle");
      });

      $('#showSmartAnnotationHelpDlg').click(function(){
        $('#smartAnnotationHelpDlgGif').attr('src', 'img/smart_annotation.gif');
        $('#smartAnnotationHelpDlg').modal('setting', { detachable:false }).modal('show');
      });

      $("#settingsMenuItem").click(function(e) {
        annotationSettings.setAll();
        $('#annotationSettingsPopup').modal({
          onApprove : function() {
            if(!/^\d+$/.test($("#annotationPolygonVertexSizeInput").val())) {
              $('#annotationSettingsPopupWarningMessageBoxContent').text("The polygon vertex size needs to be a numeric value!");
              $("#annotationSettingsPopupWarningMessageBox").show(200).delay(1500).hide(200);
              return false;
            }

            annotationSettings.persistAll(); 
            $('#annotationSettingsPopup').modal('hide');
            $('#annotationSettingsRefreshBrowserPopup').modal('show');
          }
        }).modal('show');
      });

      $('#blacklistButton').click(function(e) {
        {{ if (eq .sessionInformation.LoggedIn false) }}
        //in case we aren't logged in, do nothing
        return;

        {{else}}
        {{ if (eq .sessionInformation.LoggedIn true) }}
        var blacklistAnnotationUsageDlgAlreadyShown = localStorage.getItem("blacklistAnnotationUsageDlgShown");
          if(blacklistAnnotationUsageDlgAlreadyShown === null) {
            $("#blacklistAnnotationUsageDlg").modal("show");
            localStorage.setItem("blacklistAnnotationUsageDlgShown", true);
          } else {
            blacklistAnnotation(annotationInfo.validationId);
          }
        {{ else }}
          $("#blacklistAnnotationUsageDlg").modal("show");
        {{ end }}

        {{ end }}
        
      });

      $('#blacklistAnnotationUsageDlgAcceptButton').click(function(e) {
        $("#blacklistAnnotationUsageDlg").modal("hide");
        blacklistAnnotation(annotationInfo.validationId);
      });

      $('#notAnnotableButton').click(function(e) {
        var markAsUnannotatableUsageDlgAlreadyShown = localStorage.getItem("markAsUnannotatableUsageDlgShown");
        if(markAsUnannotatableUsageDlgAlreadyShown === null) {
          $("#markAsUnannotatableUsageDlg").modal("show");
          localStorage.setItem("markAsUnannotatableUsageDlgShown", true);
        } else {
          markAsNotAnnotatable(annotationInfo.validationId);
        }
      });

      $('#markAsUnannotatableUsageDlgAcceptButton').click(function(e) {
        $("#markAsUnannotatableUsageDlg").modal("hide");
        markAsNotAnnotatable(annotationInfo.validationId);
      });

      $('#doneButton').click(function(e) {
        var res = null;

        {{ if eq .annotationView "unified" }}
        saveCurrentSelectLabelInUnifiedModeList();
        if(Object.keys(unifiedModeAnnotations).length === 0) {
          $('#warningMsgText').text('Please annotate the image first.');
          $('#warningMsg').show(200).delay(1500).hide(200);
          return;
        }
        {{ else }}
        res = annotator.toJSON((pluralAnnotations ? annotationRefinementsDlg.getPluralAnnotationRefinementUuid() : null));
        if(res.length === 0) { //at least one annotation needs to be there
          $('#warningMsgText').text('Please annotate the image first.');
          $('#warningMsg').show(200).delay(1500).hide(200);
          return;
        }
        {{ end }}

        if(isLoadingIndicatorVisible()){ //in case smart annotation is currently running
          $('#warningMsgText').text('Smart Annotation is currently in progress.');
          $('#warningMsg').show(200).delay(1500).hide(200);
          return;
        }

        e.preventDefault();
        
        if(existingAnnotations !== null) {
          updateAnnotations(annotator.toJSON());
        }
        else {
          var annotations = [];
          {{ if eq .annotationView "unified" }}
          for(var key in unifiedModeAnnotations) {
            if(unifiedModeAnnotations.hasOwnProperty(key)) {
              if(unifiedModeAnnotations[key].dirty) {
                var annotation = {};
                annotation["annotations"] = unifiedModeAnnotations[key].annotations;
                annotation["label"] = unifiedModeAnnotations[key].label;
                annotation["sublabel"] = unifiedModeAnnotations[key].sublabel;
                annotations.push(annotation);
              }
            }
          }
          unifiedModeAnnotations = {};
          {{ else }}
          var annotation = {};
          annotation["annotations"] = res;
          annotation["label"] = $('#label').attr('label');
          annotation["sublabel"] = $('#label').attr('sublabel');
          annotations.push(annotation);
          {{ end }}
          addAnnotations(annotations);
        }
      });

      changeNavHeader({{ .annotationMode }});

      populatePluralsAndLoadData();

      try {
        //can fail in case someone uses uBlock origin or Co.
        new Fingerprint2().get(function(result, components){
          browserFingerprint = result;
        });
      } catch(e) {
      }
});