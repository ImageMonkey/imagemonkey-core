AnnotationBrowseFormContainerComponent = {
    template: "#annotation-browse-form-container-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: true
        }
    },
    methods: {
        hide: function() {
            this.visible = false;
        },
        onShowAnnotationBrowseMode: function() {
            this.visible = true;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("showAnnotationBrowseMode", this.onShowAnnotationBrowseMode);
    },
    mounted: function() {
        EventBus.$on("showAnnotationBrowseMode", this.onShowAnnotationBrowseMode);
    }
}