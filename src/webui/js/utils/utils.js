function buildComposedLabels(label, uuid, sublabels) {
    if (sublabels === null || sublabels === undefined)
        return [{
            "displayname": label,
            "uuid": uuid
        }];
    composedLabels = [{
        "displayname": label,
        "uuid": uuid
    }];
    for (const sublabel of sublabels) {
        composedLabels.push({
            "displayname": sublabel.name + "/" + label,
            "uuid": sublabel.uuid
        });
    }
    return composedLabels;
}