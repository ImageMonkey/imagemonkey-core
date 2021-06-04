AnnotationLabelListComponent = {
    template: "#annotation-label-list-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            labels: [],
            labelLookupTable: [],
            currentSelectedItem: null
        }
    },
    computed: {
    },
    methods: {
        itemSelected: function(labelUuid) {
            this.currentSelectedItem = labelUuid;
        },
        itemColor: function(labelUuid) {
            if (this.currentSelectedItem === labelUuid)
                return "bg-red-100";
            return "bg-green-100";
        }
    },
    mounted: function() {
        var that = this;
        EventBus.$on("unannotatedImageDataReceived", (data) => {
            let onlyUnlockedLabels = false;
            imageMonkeyApi.getLabelsForImage(data.uuid, onlyUnlockedLabels)
                .then(function(entries) {
                    let composedLabels = []
                    for (const entry of entries) {
                        composedLabels.push(...buildComposedLabels(entry.label, entry.uuid, entry.sublabels));
                    }

                    this.labelLookupTable = {}
                    for (const composedLabel of composedLabels) {
                        this.labelLookupTable[composedLabel.uuid] = composedLabel.displayname;
                    }

                    that.labels = composedLabels;
                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        });
    }
};
