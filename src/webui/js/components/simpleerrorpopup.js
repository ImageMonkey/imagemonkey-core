SimpleErrorPopupComponent = {
    template: "#simple-error-popup-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: false,
            msg: null,
            closeable: true,
        }
    },
    methods: {
        show: function(msg, closeable = true) {
            this.msg = msg
            this.visible = true;
            this.closeable = closeable;
        },
        hide: function() {
            this.visible = false;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("showErrorPopup", this.show);
    },
    mounted: function() {
        EventBus.$on("showErrorPopup", this.show);
    }
};
