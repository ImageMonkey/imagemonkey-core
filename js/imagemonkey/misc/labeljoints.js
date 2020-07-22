var LabelJoints = (function() {
    function LabelJoints(availableLabelJoints) {
        this.availableLabelJoints = new Map(Object.entries(availableLabelJoints));
        this.usedLabelJoints = new Map();
    }

    LabelJoints.prototype.acquireLabelJoint = function(mode) {
        let inst = this;
        for (const [key, value] of inst.availableLabelJoints.entries()) {
            if (!inst.usedLabelJoints.has(key)) {
                inst.usedLabelJoints.set(key, null);
                return key;
            }
        }
        return null;
    }

    LabelJoints.prototype.releaseLabelJoint = function(uuid) {
        this.usedLabelJoints.delete(uuid);
    }


    return LabelJoints;
}());