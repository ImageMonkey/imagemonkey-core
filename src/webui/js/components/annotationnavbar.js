AnnotationNavbarComponent = {
    template: "#annotation-navigationbar-template",
    data() {
        return {
            visible: false
        }
    },
    methods: {
        save: function() {},
        onImageInImageGridClicked: function(imageId) {
            this.visible = true;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("imageInImageGridClicked", this.onImageInImageGridClicked);
    },
    mounted: function() {
        EventBus.$on("imageInImageGridClicked", this.onImageInImageGridClicked);
    }
};