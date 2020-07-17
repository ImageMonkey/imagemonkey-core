var LabelJoints = (function() {
	function LabelJoints(availableLabelJoints) {
		this.availableLabelJoints = Object.entries(availableLabelJoints);
		this.usedLabelJoints = new Map();
	}

	LabelJoints.prototype.acquireLabelJoint = function(mode) {
		let inst = this;
		for (let key in inst.availableLabelJoints.keys()) {	
			if (!inst.usedLabelJoints.has(key)) {
				inst.usedLabelJoints.set(key, null); 
				return key;
			}
		}
	}

	LabelJoints.prototype.releaseLabelJoint = function(uuid) {
		this.usedLabelJoints.delete(uuid);
	}


	return LabelJoints;
}());
