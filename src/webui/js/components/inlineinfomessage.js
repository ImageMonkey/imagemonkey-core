InlineInfoMessageComponent = {
    template: "#inline-info-message-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            message: "",
            visible: true
        }
    },
    computed: {},
    methods: {
        show: function(message) {
            this.visible = true;
            this.message = message;
        },
        hide: function() {
            this.visible = false;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("showInlineInfoMessage", this.show);
    },
    mounted: function() {
        EventBus.$on("showInlineInfoMessage", this.show);
    }
}