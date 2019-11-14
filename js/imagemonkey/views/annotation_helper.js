function isLoadingIndicatorVisible() {
    return $("#smartAnnotationDimmer").is(":visible");
}

function showHideAutoAnnotationsLoadButton(autoAnnotations) {
    if (autoAnnotations && (autoAnnotations.length > 0) && (!isSmartAnnotationEnabled())) {
        $("#loadAutoAnnotationsMenuItem").show();
        $("#loadAutoAnnotationsMenuItem").removeClass("disabled");
        $("#loadAutoAnnotationsMenuItem").addClass("orange");
    } else {
        $("#loadAutoAnnotationsMenuItem").hide();
    }
}

function canvasHasObjects(canvas) {
    if (canvas.fabric().getObjects().length > 0)
        return true;
    return false;
}

function setLabel(label, sublabel, accessor) {
    $("#label").attr("label", label);
    $("#label").attr("sublabel", sublabel);

    if (accessor !== null)
        $("#label").attr("accessor", accessor);

    if (sublabel === "") {
        $("#label").text(("Annotate all: " + label));
        $("#bottomLabel").text(("Annotate all: " + label));
        $("#isPluralButton").attr("data-tooltip", "Set in case you want to annotate multiple " + label + " objects at once");
    } else {
        $("#label").text(("Annotate all: " + sublabel + "/" + label));
        $("#bottomLabel").text(("Annotate all: " + sublabel + "/" + label));
        $("#isPluralButton").attr("data-tooltip", "Set in case you want to annotate multiple " + sublabel + "/" + label + " objects at once");
    }
}

function changeControl(annotator, imageId) {
    if (imageId === "") {
        $("#labelContainer").hide();
        $("#doneButton").hide();
        $("#bottomLabel").hide();
        $("#isPluralContainer").hide();
        annotator.block();
    } else {
        $("#labelContainer").show();
        $("#doneButton").show();
        $("#bottomLabel").show();
        $("#isPluralContainer").show();
        annotator.unblock();
    }
}

function dataURItoBlob(dataURI) {
    var byteString = atob(dataURI.split(',')[1]);
    var ab = new ArrayBuffer(byteString.length);
    var ia = new Uint8Array(ab);
    for (var i = 0; i < byteString.length; i++) {
        ia[i] = byteString.charCodeAt(i);
    }
    return new Blob([ab], {
        type: 'image/png'
    });
}

function clearDetailedCanvas(detailedCanvas) {
    if (detailedCanvas !== null) {
        detailedCanvas.clear();
    }
}

function getCanvasScaleFactor(annotationInfo) {
    var maxWidth = document.getElementById("annotationAreaContainer").clientWidth - 50; //margin
    var scaleFactor = maxWidth / annotationInfo.origImageWidth;
    if (scaleFactor > 1.0)
        scaleFactor = 1.0;
    return scaleFactor;
}


function getUrlFromImageUrl(imageUrl, imageUnlocked, annotationMode, lookupTable) {
    var url = (imageUrl === "" ? "img/oops-no-annotation-left.png" : imageUrl);

    if (imageUrl !== "") {
        if (!imageUnlocked) {
            url += "?token=" + getCookie("imagemonkey");
        }

        if (annotationMode === "browse") {
            if ($("#highlightParentAnnotationsCheckbox").checkbox("is checked")) {
                var labelToAnnotate = $("#label").attr("accessor");
                if (labelToAnnotate in lookupTable) {
                    if (!imageUnlocked) {
                        url += "&highlight=" + encodeURIComponent(lookupTable[labelToAnnotate]);
                    } else {
                        url += "?highlight=" + encodeURIComponent(lookupTable[labelToAnnotate]);
                    }
                }
            }
        }
    }
    return url;
}

function showHideControls(show, imageUnlocked) {
    if (show) {
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

        if (imageUnlocked)
            $("#imageLockedLabel").hide();
        else
            $("#imageLockedLabel").show();

        $("#annotationColumnContent").show();
        $("#annotationColumnSpacer").show();
        $("#annotationPropertiesColumnSpacer").show();
    } else {
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

function showHideSmartAnnotationControls(show) {
    if (show) {
        $("#circleMenuItem").hide();
        $("#polygonMenuItem").hide();
        $("#smartAnnotationFgMenuItem").show();
        $("#smartAnnotationBgMenuItem").show();
    } else {
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

function isTrashMenuButtonEnabled() {
    return !$("#trashMenuItem").hasClass("disabled");
}

function isSmartAnnotationEnabled() {
    var obj = $("#smartAnnotation");
    if (obj.length) {
        return obj.checkbox("is checked");
    }
    return false;
}

function onLabelInLabelLstRemoveClicked(elem) {
    $("#removeLabelFromUnifiedModeLstDlg").attr("data-to-be-removed-uuid", $(elem).parent().attr("data-uuid"));
    $("#removeLabelFromUnifiedModeLstDlg").modal("show");
}

function addLabelToLabelLst(label, sublabel, uuid, allowRemove = false, newlyCreated = false, isUnlocked = false, loggedIn = false) {
    var id = "labellstitem-" + uuid;
    var displayedLabel = ((sublabel === "") ? label : sublabel + "/" + label);
    if (allowRemove) {
        displayedLabel = ('<span class="left-floated">' + displayedLabel +
            '</span><span class="right-floated" onclick="onLabelInLabelLstRemoveClicked(this);">' +
            '<i class="right icon delete ui red"></i></span>');
    } else {
        displayedLabel = '<p>' + displayedLabel + '</p>';
    }

    var disabledStr = " ";
    var onClickCallback = "annotationView.onLabelInLabelLstClicked(this);";
    var tooltip = "";
    if (!isUnlocked && !loggedIn) {
        disabledStr = " disabled ";
        onClickCallback = "";
        tooltip = ' data-content="Please login to annotate this label"';
    }

    var elem = $('<div class="ui' + disabledStr + 'segment center aligned labelslstitem" id="' + id + '"' + tooltip +
        ' data-label="' + label + '" data-uuid="' + uuid +
        '" data-sublabel="' + sublabel +
        '" data-newly-created="' + newlyCreated +
        '" onclick="' + onClickCallback + '"' +
        ' onmouseover="this.style.backgroundColor=\'#e6e6e6\';"' +
        ' onmouseout="this.style.backgroundColor=\'white\';"' +
        ' style="overflow: auto;">' + displayedLabel + '</div>');

    if (tooltip !== "") {
        $(elem)
            .popup({
                inline: true,
                hoverable: true,
                position: 'bottom center',
                delay: {
                    show: 300,
                    hide: 300
                }
            });
    }

    $("#annotationLabelsLst").append(elem);

    return elem;
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

function onRefinementInRefinementsLstClicked(elem) {
    $("#removeAnnotationRefinementsDlg").attr("data-to-be-removed-id", $(elem).attr("id"));
    $("#removeAnnotationRefinementsDlg").modal("show");
}

function changeNavHeader(mode) {
    if (mode === "default") {
        $("#labelContainer").css("margin-top", "-2em");
        $("#navHeader").css("min-height", "290px");
        $("#navHeader").show();
    } else if (mode === "noimage") {
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

function getActiveAnnotationMenuItem(ignoreAutoLoadButton) {
    var ret = "";
    $("#annotatorMenu").children().each(function() {
        if ($(this).hasClass("active")) {
            if (ignoreAutoLoadButton) {
                if ($(this).attr("id") !== "loadAutoAnnotationsMenuItem") {
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

function zoomIn(canvas) {
    canvas.fabric().setZoom(canvas.fabric().getZoom() * 1.1);
}

function zoomOut(canvas) {
    canvas.fabric().setZoom(canvas.fabric().getZoom() / 1.1);
}

function AnnotationInfo() {
    this.imageId = "";
    this.validationId = "";
    this.origImageWidth = 0;
    this.origImageHeight = 0;
    this.annotationId = "";
    this.imageUrl = "";
    this.imageUnlocked = false;
}
