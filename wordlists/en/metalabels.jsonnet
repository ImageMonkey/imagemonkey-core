local countries = import 'countries.libsonnet';
{
	"metalabels": {
		"kitchen" : {
			"description": "",
			"uuid": "6ea6beb7-b1ce-4ee1-bc94-bd35e118ddda",
			"accessors": ["."]
		},
		"landscape": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "62394d29-e80c-4606-b73c-9511fec9b80f"
        },
		"valley": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "eb43962c-f9f2-47df-9c51-c70f8accddc3"
		},
		"coastline": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "41287ff3-969e-4307-9e7f-2152df4a4203"
		},
		"construction site": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "22aae092-5fe5-4884-858b-5abf5c7dd856"
		},
		"farm": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "ca0dedef-ae45-4400-96b0-6f479f554ff7"
		},
		"park": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "130cf35b-5dd6-4646-a8f5-8569d1fd9ce9"
		},
		"riverside": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "2a792f5e-c2a7-42de-a834-e10b30d3bfe4"
		},
		"junkyard": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "95f79ddc-4e2d-40cb-ab5a-64392f228e6c"
		},
		"roadworks": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "5e0b934d-18d3-4b0b-bdce-944d7984f958"
		},
		"seaside": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "65d51128-63aa-4905-b9c4-053a2042c8b6"
		},
		"countryside": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "24771312-432b-4f2e-94db-abe53ee83eb2"
		},
		"town centre": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "a2fd481f-f4ed-47cf-8566-7c61b11849a9"
		},
		"factory": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "1d0c69a4-425d-4129-9a73-dd0db6a4bb89"
		},
		"garage": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "bb1aea85-0327-42aa-8f75-fcb0fde83202"
		},
		"cityscape": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "87528935-30aa-4554-9f7b-36531cefc788"
		},
		"shopping mall": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "019418cf-2208-451c-ae65-b0eed74c73e5"
		},
		"warehouse": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "9c8f2f62-9c96-4695-9365-ccea91100474"
		},
		"restaurant": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "e240a103-d49e-40a8-a9e6-5188b4f2eed2"
		},
		"supermarket": {
			"description": "optional",
            "accessors": ["."],
            "uuid": "916a6412-1a51-42c7-b9c8-0cc32ac342ac"
		},
        "airport": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "c32809d5-be38-4463-a320-e36be0bc4c3c"
        },
        "nature": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "cf11f750-c3ae-47bc-b1a2-208c36da1825"
        },
        "motorsport": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "b42ac15b-171b-4025-985c-3e230beefe07"
        },
        "urban": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "02a80408-270b-4f7d-8ddb-a6405f799b44"
        },
        "land": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "c35e2dc7-541b-4c32-8a30-1d9f1999699a"
        },
        "traffic": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "dd45969a-1109-46de-8d39-07d44bed2559"
        },
        "outdoor": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "9d5f1fae-8d42-4a2c-9eec-7387aedc0f11"
        },
        "island": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "9f6009af-7e7d-4455-a928-743666133998"
        },
        "indoor": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "a0983fde-2271-4ab7-bbaf-c787edd6bcfb"
        },
        "office": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "4c2e0ae0-1bdc-4621-96d2-d12978493848"
        },
        "botany": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "006870d6-6fa8-4d9a-a676-9a3a4c8fe436"
        },
        "architecture": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "8e548c53-3136-4b9a-a2f8-5524686cb528"
        },
        "industrial": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "74876d97-08ae-4e3d-bc04-dbe260d9634c"
        },
        "military": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "0fcccb28-7180-420b-8862-9e8b049f005a"
        },
        "tropical": {
            "description": "optional",
            "accessors": ["."],
            "uuid": "44050e74-f76a-43ef-8ad2-e6a644ecee7f"
        },
		"town": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "0c86f9fb-cc51-4394-8347-b7d4f95dde32"
		},



        "germany": countries["germany"],
        "poland": countries["poland"],
        "russia": countries["russia"],
        "vietnam": countries["vietnam"],
        "india": countries["india"]
	}
}
