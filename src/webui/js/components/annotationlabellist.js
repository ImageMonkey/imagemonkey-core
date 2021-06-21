AnnotationLabelListComponent = {
    template: "#annotation-label-list-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            labels: [],
            labelLookupTable: [],
            currentSelectedItem: null,
            visible: true 
        }
    },
    computed: {},
    methods: {
        itemSelected: function(labelUuid) {
            this.currentSelectedItem = labelUuid;
        },
        itemColor: function(labelUuid) {
            if (this.currentSelectedItem === labelUuid)
                return "bg-red-100";
            return "bg-green-100";
        },
        onUnannotatedImageDataReceived: function(data) {
            var that = this;
            let onlyUnlockedLabels = false;
            imageMonkeyApi.getLabelsForImage(data.uuid, onlyUnlockedLabels)
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
        }
    },
    beforeDestroy: function() {
        EventBus.$off("unannotatedImageDataReceived", this.onUnannotatedImageDataReceived);
    },
    mounted: function() {
        EventBus.$on("unannotatedImageDataReceived", this.onUnannotatedImageDataReceived);
    }
};
