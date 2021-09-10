AnnotationNavbarComponent = {
    template: "#annotation-navigationbar-template",
    data() {
        return {
            visible: true
        }
    },
    methods: {
        save: function() {
            EventBus.$emit("saveChangesInUnifiedMode");
        }
    },
    beforeDestroy: function() {
        EventBus.$off("save", this.save);
    },
    mounted: function() {
        EventBus.$on("save", this.save);
    }
};