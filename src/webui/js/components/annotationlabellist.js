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
            availableLabels: []
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
        },
        getLabelsForImage: function(imageId) {
            this.reset();

            var that = this;
            let onlyUnlockedLabels = false;
            imageMonkeyApi.getLabelsForImage(imageId, onlyUnlockedLabels)
                .then(function(entries) {
                    let composedLabels = []
                    for (const entry of entries) {
                        composedLabels.push(...buildComposedLabels(entry.label, entry.uuid, entry.sublabels));
                    }

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
                    //TODO
                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        },
        onUnannotatedImageDataReceived: function(data) {
            this.getLabelsForImage(data.uuid);
            this.getAnnotationsForImage(data.uuid, data.unlocked);
        },
        onAddLabel: function() {
            let splittedLabels = this.addLabelInput.split(new Settings().getLabelSeparator()).map(item => item.trim());

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
                }).catch(function(e) {
                    Sentry.captureException(e);
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
