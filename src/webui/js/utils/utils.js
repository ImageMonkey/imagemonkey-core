function buildComposedLabels(label, sublabels) {
	if(sublabels === null || sublabels === undefined)
		return [label];
	composedLabels = [label];
	for(const sublabel of sublabels) {
		composedLabels.push(sublabel + "/" + label);	
	}
	return composedLabels;
}
