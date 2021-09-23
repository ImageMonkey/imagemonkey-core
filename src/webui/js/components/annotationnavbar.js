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
        },
        discard: function() {
            EventBus.$emit("discardChangesInUnifiedMode");
        }
    },
    beforeDestroy: function() {
        EventBus.$off("save", this.save);

        Mousetrap.unbind("ctrl+s");
        Mousetrap.unbind("ctrl+d");
    },
    mounted: function() {
        EventBus.$on("save", this.save);

        let inst = this;
        Mousetrap.bind("ctrl+s", function(e) {
            if (inst.visible) {
                e.preventDefault();
                inst.save();
            }
        });

        Mousetrap.bind("ctrl+d", function(e) {
            if (inst.visible) {
                e.preventDefault();
                inst.discard();
            }
        });
    }
};