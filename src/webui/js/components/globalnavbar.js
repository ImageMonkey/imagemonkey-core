GlobalNavbarComponent = {
    template: "#global-navbar-template",
    data() {
        return {
            visible: true,
            numOfNotifications: 0
        }
    },
    methods: {
        hide: function() {
            this.visible = false;
        },
        logout: function() {
            imageMonkeyApi.logout()
                .then(function() {
                    Cookies.expire("imagemonkey");
                    window.location.href = "/"; //redirect to home page
                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        },
        profile: function() {
            let username = parseJwt(Cookies.get("imagemonkey"))["username"];
            window.location.href = "/profile/" + username; //redirect to profile page
        }
    },
    mounted: function() {
        $(".dropdown").dropdown();

        var that = this;
        if (this.$store.getters.isModerator) {
            imageMonkeyApi.getNumOfUnprocessedImageDescriptions()
                .then(function(numOfNotifications) {
                    that.numOfNotifications = numOfNotifications;
                }).catch(function(e) {
                    console.log(e.message);
                    Sentry.captureException(e);
                });
        }
    }
};