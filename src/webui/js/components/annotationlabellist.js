AnnotationLabelListComponent = {
    template: "#annotation-label-list-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            labels: [],
            labelLookupTable: [],
            currentSelectedItem: null,
            previousSelectedItem: null,
            visible: true,
            addLabelInput: null,
            addedButNotCommittedLabels: {},
            toBeRemovedLabelUuids: [],
            availableLabelsLookupTable: {},
            availableLabels: [],
            labelsAutoCompletion: null,
            imageId: null,
            annotations: {},
            notCommittedAnnotations: {}
        }
    },
    computed: {},
    methods: {
        generateUuid: function() {
            return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
                (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16));
        },
        getCurrentSelectedLabelUuid: function() {
            return this.currentSelectedItem;
        },
        getPreviousSelectedLabelUuid: function() {
            return this.previousSelectedItem;
        },
        getAnnotationsForLabelUuid: function(labelUuid) {
            let annotationsForLabel = [];
            let displayName = null;
            if (labelUuid in this.labelLookupTable)
                displayName = this.labelLookupTable[labelUuid];

            if (displayName in this.annotations)
                annotationsForLabel = this.annotations[displayName];
            else if (labelUuid in this.notCommittedAnnotations)
                annotationsForLabel = this.notCommittedAnnotations[labelUuid];
            return annotationsForLabel;
        },
        itemSelected: function(labelUuid, emitEvent = true) {
            this.previousSelectedItem = this.currentSelectedItem;
            this.currentSelectedItem = labelUuid;
            if (emitEvent)
                EventBus.$emit("labelSelected", this.currentSelectedItem, this.previousSelectedItem);
        },
        entryCanBeRemoved: function(labelName) {
            if (labelName in this.addedButNotCommittedLabels)
                return true;
            return false;
        },
        itemColor: function(labelUuid) {
            if (this.currentSelectedItem === labelUuid)
                return "bg-gray-200";
            return "bg-white";
        },
        removeLabel: function(label) {
            EventBus.$emit("removeLabel", label);
        },
        reset: function() {
            this.toBeRemovedLabelUuids = [];
            this.addedButNotCommittedLabels = {};
            this.labels = [];
            this.addLabelInput = null;
            this.labelLookupTable = [];
            this.imageId = null;
            this.annotations = {};
            this.notCommittedAnnotations = {};
            this.currentSelectedItem = null;
            this.previousSelectedItem = null;
        },
        getLabelsAndAnnotationsForImage: function(imageId, toBeSelectedValidationId, imageUnlocked) {
            this.reset();
            this.imageId = imageId;

            var that = this;
            let onlyUnlockedLabels = false;

            let promises = [imageMonkeyApi.getLabelsForImage(imageId, onlyUnlockedLabels),
                imageMonkeyApi.getAnnotationsForImage(imageId, imageUnlocked)
            ]
            Promise.all(promises)
                .then(function(data) {
                    //labels
                    let composedLabels = []
                    let selectedLabelUuid = null;
                    if (data[0] !== null) {
                        for (const entry of data[0]) {
                            composedLabels.push(...buildComposedLabels(entry.label, entry.uuid, entry.sublabels));
                            if (toBeSelectedValidationId !== null) {
                                if ("validation" in entry) {
                                    if (entry["validation"]["uuid"] === toBeSelectedValidationId) {
                                        selectedLabelUuid = entry["uuid"];
                                    }
                                }
                            }
                        }
                    }

                    that.labelLookupTable = {}
                    for (const composedLabel of composedLabels) {
                        that.labelLookupTable[composedLabel.uuid] = composedLabel.displayname;
                    }

                    that.labels = composedLabels;

                    //annotations
                    for (const annotation of data[1]) {
                        let displayName = getDisplayName(annotation.validation.label, annotation.validation.sublabel);
                        that.annotations[displayName] = annotation.annotations;
                    }

                    if (selectedLabelUuid !== null)
                        that.itemSelected(selectedLabelUuid, false);

                    $("#add-labels-input").focus();
                    EventBus.$emit("imageSpecificLabelsAndAnnotationsLoaded");

                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        },
        onUnannotatedImageDataReceived: function(imageId, validationId, imageUnlocked) {
            this.getLabelsAndAnnotationsForImage(imageId, validationId, imageUnlocked);
        },
        onAddLabel: function() {
            let labelString = $("#add-labels-input").val();
            let splittedLabels = labelString.split(new Settings().getLabelSeparator()).map(item => item.trim());

            for (const splittedLabel of splittedLabels) {
                if (!this.$store.getters.loggedIn) {
                    if (!(splittedLabel in this.availableLabelsLookupTable)) {
                        EventBus.$emit("unauthenticatedAccess");
                        return
                    }
                }

                //do not add empty label
                if (splittedLabel === "") {
                    EventBus.$emit("emptyLabelAdded");
                    continue;
                }

                if (!labelExistsInLabelList(splittedLabel, this.labels)) {
                    let newLabel = {
                        "uuid": this.generateUuid()
                    };
                    this.addedButNotCommittedLabels[splittedLabel] = newLabel;
                    this.labels.push(...buildComposedLabels(splittedLabel, newLabel.uuid, []));
                    this.labelLookupTable[newLabel.uuid] = getDisplayName(splittedLabel, "");
                    this.itemSelected(newLabel.uuid);
                } else {
                    EventBus.$emit("duplicateLabelAdded", splittedLabel);
                }
            }
            this.addLabelInput = null;
        },
        onConfirmRemoveLabel: function(label) {
            let labelUuid = getLabelUuidForLabelInLabelList(label, this.labels);
            if (label in this.addedButNotCommittedLabels)
                delete this.addedButNotCommittedLabels[label];
            else {
                this.toBeRemovedLabelUuids.push(labelUuid);
            }

            delete this.labelLookupTable[labelUuid];
            removeLabelFromLabelList(label, this.labels);

            //if removed item was selected, select a different one
            if (labelUuid === this.getCurrentSelectedLabelUuid()) {
                let nextLabelUuid = (this.labels.length > 0) ? this.labels[0].uuid : null;
                this.itemSelected(nextLabelUuid);
                if (nextLabelUuid === null)
                    EventBus.$emit("noLabelSelected");
            }
        },
        getAvailableLabelsAndLabelSuggestions: function() {
            let labelRequests = [imageMonkeyApi.getAvailableLabels()];
            if (this.$store.getters.loggedIn)
                labelRequests.push(imageMonkeyApi.getLabelSuggestions(false));

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
                                    inst.availableLabelsLookupTable[subkey + "/" + key] = {
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
                    inst.labelsAutoCompletion = new AutoCompletion("#add-labels-input", inst.availableLabels);
                    EventBus.$emit("labelsAndLabelSuggestionsLoaded");
                }).catch(function(e) {
                    Sentry.captureException(e);
                });

        },
        persistNewlyAddedLabels: function() {
            let addedLabels = [];
            for (const label in this.addedButNotCommittedLabels) {
                addedLabels.push({
                    label: label,
                    annotatable: true
                });
            }
            if (addedLabels.length > 0)
                return imageMonkeyApi.labelImage(this.imageId, addedLabels);
            return new Promise((resolve) => {
                resolve(null);
            });
        },
        persistAnnotations: function() {
            var annotations = [];
            for (var key in this.notCommittedAnnotations) {
                if (this.notCommittedAnnotations.hasOwnProperty(key)) {
                    let displayLabel = this.labelLookupTable[key];
                    let label = displayLabel;
                    let sublabel = "";
                    if (displayLabel in this.availableLabelsLookupTable) {
                        label = this.availableLabelsLookupTable[displayLabel].label;
                        sublabel = this.availableLabelsLookupTable[displayLabel].sublabel;
                    }
                    var annotation = {};
                    annotation["annotations"] = this.notCommittedAnnotations[key];
                    annotation["label"] = label;
                    annotation["sublabel"] = sublabel;
                    annotations.push(annotation);
                }
            }
            if (annotations.length > 0) {
                return imageMonkeyApi.addAnnotations(this.imageId, annotations);
            }
            return new Promise((resolve) => {
                resolve(null);
            });
        },
        updateAnnotations: function(annotations, labelUuid) {
            this.notCommittedAnnotations[labelUuid] = annotations;
        },
        ctrlSPressed: function() {
            EventBus.$emit("ctrl+sPressed");
        },
        ctrlDPressed: function() {
            EventBus.$emit("ctrl+dPressed");
        }
    },
    beforeDestroy: function() {
        EventBus.$off("unannotatedImageDataReceived", this.onUnannotatedImageDataReceived);
        EventBus.$off("confirmRemoveLabel", this.onConfirmRemoveLabel);
    },
    mounted: function() {
        EventBus.$on("unannotatedImageDataReceived", this.onUnannotatedImageDataReceived);
        EventBus.$on("confirmRemoveLabel", this.onConfirmRemoveLabel);
        this.getAvailableLabelsAndLabelSuggestions();
    }

};