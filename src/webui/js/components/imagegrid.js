ImageGridComponent = {
    template: "#image-grid-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            infiniteScroll: null,
            imageGridData: null,
            numOfLastFetchedImg: 0,
            defaultBatchSize: 50,
            numberOfCurrentlyShownResults: 0,
            numberOfQueryResults: 0,
            currentlyVisibleImageGridData: null,
            visible: true
        }
    },
    computed: {},
    methods: {
        clear: function() {
            this.imageGridData = null;
            this.numOfLastFetchedImg = 0;
            this.currentlyVisibleImageGridData = 0;
            this.numberOfCurrentlyShownResults = 0;
            this.numberOfQueryResults = 0;
            this.infiniteScroll.deactivate();
        },
        show: function() {
            this.visible = true;
        },
        hide: function() {
            this.visible = false;
        },
        imageStyle(greyedOut) {
            if (greyedOut)
                return "grey-out";
            return "";
        },
        imageClicked: function(imageId, validationId) {
            EventBus.$emit("imageInImageGridClicked", imageId, validationId);
        },
        populate: function(data, options) {
            this.clear();

            this.imageGridData = [];
            this.currentlyVisibleImageGridData = [];

            let sizes = [];
            for (const elem of data) {
                sizes.push({
                    "width": elem["image"]["width"],
                    "height": elem["image"]["height"]
                });
            }

            this.numberOfQueryResults = data.length;

            let justifiedLayout = require('justified-layout');
            let justifiedLayoutGeometry = justifiedLayout(sizes, {
                "fullWidthBreakoutRowCadence": false,
                "containerWidth": document.getElementById(this.$el.id).clientWidth,
                "boxSpacing": {
                    "horizontal": 10,
                    "vertical": 100
                }
            });

            for (var i = 0; i < justifiedLayoutGeometry.boxes.length; i++) {
                let imageUnlocked = data[i]["image"]["unlocked"];
                let imageUrl = data[i]["image"]["url"];
                if (!imageUnlocked)
                    imageUrl += "?token=" + getCookie("imagemonkey");
                imageUrl += (((imageUnlocked === true) ? '?' : '&') + "width=" + Math.round(justifiedLayoutGeometry.boxes[i].width, 0) +
                    "&height=" + Math.round(justifiedLayoutGeometry.boxes[i].height, 0));

                let tooltipText = '';
                if (options === "rework") {
                    if (data[i].validation.sublabel !== "")
                        tooltipText = data[i].validation.sublabel + "/" + data[i].validation.label;
                    else
                        tooltipText = data[i].validation.label;
                } else
                    tooltipText = data[i].label.accessor;

                this.imageGridData.push({
                    top: justifiedLayoutGeometry.boxes[i].top,
                    left: justifiedLayoutGeometry.boxes[i].left,
                    width: justifiedLayoutGeometry.boxes[i].width,
                    height: justifiedLayoutGeometry.boxes[i].height,
                    imageUuid: data[i].image.uuid,
                    validationId: data[i].uuid,
                    imageUrl: imageUrl,
                    tooltipText: tooltipText,
                    greyedOut: false
                });
            }

            this.loadNextImagesInImageGrid();
            this.infiniteScroll.activate();

        },
        loadNextImagesInImageGrid: function() {
            let from = this.numOfLastFetchedImg;
            let n = this.defaultBatchSize;
            if ((this.numOfLastFetchedImg + this.defaultBatchSize) > this.imageGridData.length) {
                n = this.imageGridData.length - this.numOfLastFetchedImg;
            }

            if (n === 0)
                return;

            let currentDateTime = new Date().getTime();
            for (var i = from; i < (from + n); i++) {
                document.getElementById(this.$el.id).style.height = ((this.imageGridData[(from + n - 1)].top +
                    this.imageGridData[(from + n - 1)].height) + "px");
                this.currentlyVisibleImageGridData.push(this.imageGridData[i]);
            }

            this.numOfLastFetchedImg += n;
            this.numberOfCurrentlyShownResults = this.numOfLastFetchedImg;

            let numberOfShownQueryResults = this.numberOfCurrentlyShownResults + "/" + this.numberOfQueryResults + " results shown";
            EventBus.$emit("unifiedModeImageGridCurrentlyShownImagesUpdated", numberOfShownQueryResults);


        },
        onGreyOutImageInImageGrid: function(imageId) {
            let idx = this.imageGridData.findIndex((obj => obj.imageUuid == imageId));
            if (idx !== -1)
                this.imageGridData[idx].greyedOut = true;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("populateUnifiedModeImageGrid", this.populate);
        EventBus.$off("greyOutImageInImageGrid", this.onGreyOutImageInImageGrid);
    },
    mounted: function() {
        this.infiniteScroll = new InfiniteScroll(this.loadNextImagesInImageGrid, false);

        EventBus.$on("populateUnifiedModeImageGrid", this.populate);
        EventBus.$on("greyOutImageInImageGrid", this.onGreyOutImageInImageGrid);
    }
};