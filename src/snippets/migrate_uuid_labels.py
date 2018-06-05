import json
import os
import uuid

path = ".." + os.path.sep + "wordlists" + os.path.sep + "en" + os.path.sep + "labels.json"

with open(path) as json_data:
    data = json.load(json_data)

    labels = data["labels"]

    for label_name in labels:
    	#print(label_name)
    	l = labels[label_name] 
    	if "accessors" in l:
    		accessors = l["accessors"]
    		new_accessors = []
    		for accessor in accessors:
    			u = str(uuid.uuid4())
    			new_accessors.append({"name": accessor, "uuid": u})

    			if accessor == ".":
    				label_accessor_uuid_mapping[label_name] = u

    		data["labels"][label_name]["accessors"] = new_accessors

    	if "has" in labels[label_name]:
	    	has = labels[label_name]["has"]
	    	for h in has:
	    		if "accessors" in has[h]:
		    		subaccessors = has[h]["accessors"]
		    		new_sublabel_accessors = []
		    		for subaccessor in subaccessors:
		    			u1 = str(uuid.uuid4())
		    			new_sublabel_accessors.append({"name": subaccessor, "uuid": u1})

		    			if subaccessor == ".":
		    				label_accessor_uuid_mapping[label_name]

		    		data["labels"][label_name]["has"][h]["accessors"] = new_accessors

with open(path, 'w') as outfile:
	json.dump(data, outfile)




