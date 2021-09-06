AnnotationBrowseFormComponent = {
    template: "#annotation-browse-form-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            searchQuery: "",
            autoCompletion: null,
            labelAccessorsLoaded: false,
            annotatedStatisticsLoaded: false,
            searchNoOptionsSelected: true,
            searchReworkExistingAnnotationsSelected: false,
            searchHighlightAnnotationsParentSelected: false,
            numberOfShownQueryResults: '',
            availableLabels: [],
            errorMessage: "",
            visible: true
        }
    },
    computed: {},
    methods: {
        /*searchNoOptionsSelected: function() {
        	return this.searchNoOptionsSelected;
        },
        searchReworkExistingAnnotationsSelected: function() {
        	return this.searchReworkExistingAnnotationsSelected;
        },
        searchHighlightAnnotationsParentSelected: function() {
        	return this.searchHighlightAnnotationsParentSelected;
        },*/
        showInlineErrorMessage: function(errorMessage) {
            this.errorMessage = errorMessage;
            setTimeout(() => this.errorMessage = "", 5000);
        },
        search: function() {
            EventBus.$emit("showWaveLoadingIndicator");

            this.numberOfShownQueryResults = 0;
            let apiCommand = null;
            let searchOption = "";
            if (this.searchNoOptionsSelected) {
                searchOption = "no-option";
                apiCommand = imageMonkeyApi.queryUnannotatedAnnotations(this.searchQuery, true);
            } else if (this.searchReworkExistingAnnotationsSelected) {
                searchOption = "rework";
                apiCommand = imageMonkeyApi.queryAnnotated(this.searchQuery, true);
            }

            let fullUrl = new URL(window.location);
            fullUrl.searchParams.set('query', this.searchQuery);
            if (searchOption !== "no-option")
                fullUrl.searchParams.set('search_option', searchOption);
            window.history.pushState({}, null, fullUrl);

            var that = this;
            apiCommand
                .then(function(data) {
                    if (data && data.length > 0) {
                        EventBus.$emit("populateUnifiedModeImageGrid", data, searchOption);
                    } else {
                        EventBus.$emit("hideWaveLoadingIndicator");
                        that.showInlineErrorMessage("Nothing found");
                    }
                }).catch(function(e) {
                    EventBus.$emit("hideWaveLoadingIndicator");
                    that.showInlineErrorMessage("Couldn't process request - please try again later");
                    Sentry.captureException(e);
                });
        },
        randomQuery: function() {
            for (const availableLabel of this.availableLabels) {
                let randomNum = Math.floor(Math.random() * this.availableLabels.length);
                this.searchQuery = this.availableLabels[randomNum];
            }
        },
        showAnnotatedStatistics: function() {
            EventBus.$emit("showAnnotatedStatisticsPopup");
        },
        populate: function() {
            var that = this;
            let promises = [imageMonkeyApi.getLabelAccessors(true)];
            if (this.$store.getters.loggedIn) {
                promises.push(imageMonkeyApi.getImageCollections(this.$store.getters.username));
                promises.push(imageMonkeyApi.getLabelSuggestions(false));
            }

            Promise.all(promises)
                .then(function(data) {

                    let availableLabels = [];
                    for (const elem of data[0]) {
                        availableLabels.push(elem.accessor);
                    }

                    if (data.length > 1) {
                        for (const elem of data[1]) {
                            availableLabels.push("image.collection='" + elem.name + "'")
                        }
                    }

                    if (data.length > 2) {
                        for (const elem of data[2]) {
                            availableLabels.push(elem);
                        }
                    }

                    that.autoCompletion = new AutoCompletion("#annotation-query", availableLabels);
                    that.availableLabels = availableLabels;
                    that.labelAccessorsLoaded = true;
                }).catch(function(e) {
                    Sentry.captureException(e);
                });
        },
        onAnnotatedStatisticsLoaded: function() {
            this.annotatedStatisticsLoaded = true;
        },
        onAnnotatedStatisticsPopupLabelClicked: function(label) {
            this.searchQuery = label;
            this.search();
        },
        onUnifiedModeImageGridCurrentlyShownImagesUpdated: function(num) {
            EventBus.$emit("hideWaveLoadingIndicator");
            this.numberOfShownQueryResults = num;
        },
        onLoadAnnotationBrowseFormLabels: function(query = null) {
            this.populate();
            if (query !== null) {
                this.searchQuery = query;
                this.search();
            }
        }
    },
    beforeDestroy: function() {
        EventBus.$off("annotatedStatisticsLoaded", this.onAnnotatedStatisticsLoaded);
        EventBus.$off("annotatedStatisticsPopupLabelClicked", this.onAnnotatedStatisticsPopupLabelClicked);
        EventBus.$off("unifiedModeImageGridCurrentlyShownImagesUpdated", this.onUnifiedModeImageGridCurrentlyShownImagesUpdated);
        EventBus.$off("loadAnnotationBrowseFormLabels", this.onLoadAnnotationBrowseFormLabels);
    },
    mounted: function() {
        EventBus.$on("annotatedStatisticsLoaded", this.onAnnotatedStatisticsLoaded);
        EventBus.$on("annotatedStatisticsPopupLabelClicked", this.onAnnotatedStatisticsPopupLabelClicked);
        EventBus.$on("unifiedModeImageGridCurrentlyShownImagesUpdated", this.onUnifiedModeImageGridCurrentlyShownImagesUpdated);
        EventBus.$on("loadAnnotationBrowseFormLabels", this.onLoadAnnotationBrowseFormLabels);
    }
};
