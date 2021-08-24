AnnotationLabelListComponent = {
    template: "#annotation-label-list-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            labels: [],
            labelLookupTable: [],
            currentSelectedItem: null,
            visible: true,
            addLabelInput: null,
            addedButNotCommittedLabels: {},
            toBeRemovedLabelUuids: [],
            availableLabelsLookupTable: {},
            availableLabels: [],
            labelsAutoCompletion: null,
            imageId: null,
            annotations: {}
        }
    },
    computed: {},
    methods: {
        generateUuid: function() {
            return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
                (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16));
        },
        itemSelected: function(labelUuid) {
            this.currentSelectedItem = labelUuid;
            let annotationsForLabel = [];

            let displayName = null;
            if (labelUuid in this.labelLookupTable)
                displayName = this.labelLookupTable[labelUuid];

            if (displayName in this.annotations)
                annotationsForLabel = this.annotations[displayName];
            EventBus.$emit("labelSelected");
            EventBus.$emit("drawAnnotations", annotationsForLabel);
        },
        itemColor: function(labelUuid) {
            if (this.currentSelectedItem === labelUuid)
                return "bg-red-100";
            return "bg-green-100";
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
        },
        getLabelsForImage: function(imageId, toBeSelectedValidationId) {
            this.reset();
            this.imageId = imageId;

            var that = this;
            let onlyUnlockedLabels = false;
            imageMonkeyApi.getLabelsForImage(imageId, onlyUnlockedLabels)
                .then(function(entries) {
                    let composedLabels = []
                    let labelSelected = false;
                    for (const entry of entries) {
                        composedLabels.push(...buildComposedLabels(entry.label, entry.uuid, entry.sublabels));
                        if (toBeSelectedValidationId !== null) {
                            if ("validation" in entry) {
                                if (entry["validation"]["uuid"] === toBeSelectedValidationId) {
                                    labelSelected = true;
                                    that.itemSelected(entry["uuid"]);
                                }
                            }
                        }
                    }

                    if (!labelSelected)
                        EventBus.$emit("noLabelSelected");

                    that.labelLookupTable = {}
                    for (const composedLabel of composedLabels) {
                        that.labelLookupTable[composedLabel.uuid] = composedLabel.displayname;
                    }

                    that.labels = composedLabels;
                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        },
        getAnnotationsForImage: function(imageId, imageUnlocked) {
            var that = this;
            imageMonkeyApi.getAnnotationsForImage(imageId, imageUnlocked)
                .then(function(annotations) {
                    for (const annotation of annotations) {
                        let displayName = getDisplayName(annotation.validation.label, annotation.validation.sublabel);
                        that.annotations[displayName] = annotation.annotations;
                    }
                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        },
        onUnannotatedImageDataReceived: function(data, validationId) {
            this.getLabelsForImage(data.uuid, validationId);
            this.getAnnotationsForImage(data.uuid, data.unlocked);
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

                if (!labelExistsInLabelList(splittedLabel, this.labels)) {
                    let newLabel = {
                        "uuid": this.generateUuid()
                    };
                    this.addedButNotCommittedLabels[splittedLabel] = newLabel;
                    this.labels.push(...buildComposedLabels(splittedLabel, newLabel.uuid, []));
                    this.labelLookupTable[newLabel.uuid] = getDisplayName(splittedLabel, "");
                    this.currentSelectedItem = newLabel.uuid;
                } else {
                    EventBus.$emit("duplicateLabelAdded", splittedLabel);
                }
            }
            this.addLabelInput = null;
        },
        onConfirmRemoveLabel: function(label) {
            if (label in this.addedButNotCommittedLabels)
                delete this.addedButNotCommittedLabels[label];
            else {
                this.toBeRemovedLabelUuids.push(this.labels[label].uuid);
            }
            delete this.labelLookupTable[this.labels[label].uuid];
            removeLabelFromLabelList(label, this.labels);
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
