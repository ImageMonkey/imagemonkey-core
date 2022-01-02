AnnotationNavbarComponent = {
    template: "#annotation-navigationbar-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: true,
            avatar: "/img/default.png",
            username: "guest"
        }
    },
    methods: {
        save: function() {
            EventBus.$emit("saveChangesInUnifiedMode");
        },
        discard: function() {
            EventBus.$emit("discardChangesInUnifiedMode");
        },
        avatarClicked: function() {
            let username = this.$store.getters.username;
            if (username !== "")
                window.location.href = "/profile/" + username;
            else
                window.location.href = "/login";
        },
        username: function() {
            return this.username;
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

        let username = this.$store.getters.username;
        if (username !== "") {
            this.username = username;
            this.avatar = imageMonkeyApi.getAvatarUrl(username);
        } else {
            this.username = "guest";
            this.avatar = "/img/default.png";
        }
    }
};