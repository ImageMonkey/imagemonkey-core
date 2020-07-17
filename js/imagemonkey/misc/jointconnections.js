var JointConnection = (function() {
	function JointConnection() {
		this._jointConnectionIds = [];
		this._jointConnectionPoints = [];
		this._jointConnectionLabels = [];
		this._jointConnectionAnnotationIds = [];
	}

	JointConnection.prototype.setIds = function(ids) {
		this._jointConnectionIds = ids;
	}

	JointConnection.prototype.addId = function(id) {
		this._jointConnectionIds.push(id);
	}

	JointConnection.prototype.getIds = function() {
		return this._jointConnectionIds;
	}

	JointConnection.prototype.setPoints = function(points) {
		this._jointConnectionPoints = points;
	}

	JointConnection.prototype.addPoint = function(point) {
		this._jointConnectionPoints.push(point);
	}

	JointConnection.prototype.addLabel = function(label) {
		this._jointConnectionLabels.push(label);
	}

	JointConnection.prototype.getPoints = function() {
		return this._jointConnectionPoints;
	}

	JointConnection.prototype.getPoint = function(pos) {
		return this._jointConnectionPoints[pos];
	}

	JointConnection.prototype.setAnnotationIds = function(annotationIds) {
		this._jointConnectionAnnotationIds = annotationIds;
	}

	JointConnection.prototype.getAnnotationIds = function() {
		return this._jointConnectionAnnotationIds;
	}

	JointConnection.prototype.addAnnotationId = function(annotationId) {
		this._jointConnectionAnnotationIds.push(annotationId);
	}

	JointConnection.prototype.getLabels = function() {
		return this._jointConnectionLabels;
	}
	
	return JointConnection;
}());


var JointConnections = (function() {
	function JointConnections() {
		this._jointConnectionsEnabled = false;
		this._jointConnections = new Map();
	}

	JointConnections.prototype.enable = function() {
		this._jointConnectionsEnabled = true;
	}

	JointConnections.prototype.disable = function() {
		this._jointConnectionsEnabled = false;
	}

	JointConnections.prototype.isEnabled = function() {
		return this._jointConnectionsEnabled;
	}

	JointConnections.prototype.add = function(jointConnection, uuid) {
		this._jointConnections[uuid] = jointConnection;
	}

	JointConnections.prototype.getJointConnection = function(uuid) {
		if(this._jointConnections.has(uuid))
			return this._jointConnections[uuid];
		return null;
	}

	return JointConnections;
}());
