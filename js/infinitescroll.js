
    var InfiniteScroll = function(callback, activate) {
        this.callback = callback;
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

        this.deactivate = function() {
            $(window).unbind('scroll');
        };
 
        this.handleScroll = function() {
            var scrollTop = $(document).scrollTop();
            var windowHeight = $(window).height();
            var height = $(document).height() - windowHeight;
            var scrollPercentage = (scrollTop / height);

            // if the scroll is more than 80% from the top, load more content.
            if(scrollPercentage > 0.80) {
                this.doSomething();
            }
        }
 
        this.doSomething = function() {
            typeof this.callback === 'function' && this.callback();
        }
 
        this.initialize();
    }