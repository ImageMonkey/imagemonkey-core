var Introduction = (function () {
  function Introduction(loggedInUser) {
    this.loggedInUser = loggedInUser;

    this.shepherd = new Shepherd.Tour({
      defaultStepOptions: {
        showCancelLink: true,
        classes: 'shepherd-theme-dark'
      }
    });


    this.shepherd.addStep('modes', {
      text: ["You can choose between two modes - <b>default mode</b> and <b>browse mode</b>",
             "In default mode you will get a random picture to label, each time you refresh your browser window.",
             "The browse mode allows you to search for pictures that already have a specific label. e.q: Let's assume " +
             "you want to search for all images that have the label <b>road</b> but not <b>pavement</b>, then you can use the " +
             "following query in browse mode: <pre>road & ~pavement</pre>" ],
      attachTo: {element: '#modeButtons', on: 'top'},
      buttons: [
        {
          action: this.shepherd.cancel,
          classes: 'shepherd-button-secondary',
          text: 'Exit'
        }, {
          action: this.shepherd.next,
          text: 'Next'
        }
      ]
    });

    var labelAddStepElement = '';
    var labelAddStepText = [];
    if(this.loggedInUser) {
      labelAddStepElement = '#labelSuggestion';
      labelAddStepText = ["One of the biggest challenges when building a publicly available dataset is to keep a curated label list.",
                          "That's why we have a two staged approach when it comes to labeling. Each new label that you enter here " +
                          "will be in a locked state, until it gets unlocked by a moderator.",
                          "This allows us to keep a tight labels list while still being able to extend the labels list with new labels " +
                          "when new type of images will be uploaded.",
                          "In case you want to have a look, here are all the <a href=\"https://github.com/bbernhard/imagemonkey-trending-labels/issues\">trending labels</a"];
    }
    else {
      labelAddStepElement = '#labelDropdown';
      labelAddStepText = ["One of the biggest challenges when building a publicly available dataset is to keep a curated label list.",
                          "That's why we have a two staged approach when it comes to labeling. Each new label that you enter here " +
                          "will be in a locked state, until it gets unlocked by a moderator.",
                          "This allows us to keep a tight labels list while still being able to extend the labels list with new labels " +
                          "when new type of images will be uploaded.",
                          "In case you want to have a look, here are all the <a href=\"https://github.com/bbernhard/imagemonkey-trending-labels/issues\">trending labels</a"];
    }

    this.shepherd.addStep('label-add', {
      text: labelAddStepText,
      attachTo: {element: labelAddStepElement, on: 'left'},
      buttons: [
        {
          action: this.shepherd.cancel,
          classes: 'shepherd-button-secondary',
          text: 'Exit'
        }, {
          action: this.shepherd.next,
          text: 'Next'
        }
      ]
    });

    this.shepherd.addStep('help-me-button', {
      text: ["Do you want to see the introduction again?","Click this button here"],
      attachTo: {element: '#helpMeButton', on: 'right'},
      buttons: [
        {
          action: this.shepherd.cancel,
          classes: 'shepherd-button-secondary',
          text: 'Exit'
        }
      ]
    });
  }

  Introduction.prototype.start = function() {
    this.shepherd.start();
  }

  return Introduction;

}());
