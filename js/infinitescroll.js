
    var InfiniteScroll = function(callback, activate) {
        this.paused = false;
        this.callback = callback;
        this.currentScrollPosition = null;
        this.initialize = function() {
            if(activate)
                this.activate();
        };

        this.activate = function() {
            $(window).on(
                'scroll',
                this.handleScroll.bind(this)
            );
        };

        this.pause = function() {
            this.paused = true;
        };

        this.resume = function() {
            this.paused = false;
        };

        this.deactivate = function() {
            $(window).unbind('scroll');
        };

        this.saveScrollPosition = function() {
            this.scrollPosition = window.pageYOffset;
        };

        this.restoreScrollPosition = function() {
            window.scrollTo(0, this.scrollPosition);
        };
 
        this.handleScroll = function() {
            if(!this.paused) {
                var scrollTop = $(document).scrollTop();
                var windowHeight = $(window).height();
                var height = $(document).height() - windowHeight;
                var scrollPercentage = (scrollTop / height);

                // if the scroll is more than 80% from the top, load more content.
                if(scrollPercentage > 0.80) {
                    this.doSomething();
                }
            }
        }
 
        this.doSomething = function() {
            typeof this.callback === 'function' && this.callback();
        }
 
        this.initialize();
    }