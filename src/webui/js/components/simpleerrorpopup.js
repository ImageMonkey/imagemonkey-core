SimpleErrorPopupComponent = {
    template: "#simple-error-popup-template",
    delimiters: ['${', '}$'],
	data() {
        return {
            visible: false,
			msg: null
        }
    },
    methods: {
        show: function(msg) {
			this.msg = msg
            this.visible = true;
        },
		hide: function() {
			this.visible = false;
		}
    },
    mounted: function() {
    }
};
