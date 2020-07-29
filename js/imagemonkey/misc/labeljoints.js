var LabelJoints = (function() {
    function LabelJoints(availableLabelJoints) {
        this.availableLabelJoints = new Map(Object.entries(availableLabelJoints));
        this.usedLabelJoints = new Map();
    }

    LabelJoints.prototype.acquireLabelJoint = function() {
        let inst = this;
        for (const [key, value] of inst.availableLabelJoints.entries()) {
            if (!inst.usedLabelJoints.has(key)) {
                inst.usedLabelJoints.set(key, null);
                return key;
            }
        }
        return null;
    }

    LabelJoints.prototype.releaseLabelJoint = function(identifier) {
        this.usedLabelJoints.delete(identifier);
    }

    LabelJoints.prototype.getLabelJointUuids = function(identifier) {
        if (!this.usedLabelJoints.has(identifier))
            return null;

        let items = this.availableLabelJoints.get(identifier);
        let uuids = [];
        for (item of items) {
            uuids.push(item.uuid);
        }
        return uuids;
    }

    return LabelJoints;
}());