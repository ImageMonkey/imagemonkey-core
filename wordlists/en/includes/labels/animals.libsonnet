{
	"labels": {
		"deer": {
			"description": "optional",
			"accessors": ["."],
			"plural": "deer",
			"uuid": "51dc0cce-a8e2-4368-ae52-253f5e7f7e16",
			"has": {
				"head": {
					"description": "optional",
					"uuid": "284fbd7f-bdfc-4426-bc19-f1d2a0606f84",
					"accessors": [".has"]
				}
			}
		},
		"lizard": {
			"description": "optional",
			"accessors": ["."],
			"plural": "lizards",
			"uuid": "7b60239c-2716-4572-8458-cfbd0fae5912"
		},
		"owl": {
			"description": "optional",
			"accessors": ["."],
			"plural": "owls",
			"uuid": "a4882938-8566-4bcb-b5a8-b01263b1622c"
		},
		"squirrel": {
			"description": "optional",
			"accessors": ["."],
			"plural": "squirrels",
			"uuid": "e948b99c-06e5-4d08-b7d1-deec3fc77b18",
			"has": {
				"head": {
					"description": "optional",
					"uuid": "63eec330-3950-41cd-b303-9d887c744126",
					"accessors": [".has"]
				}
			}
		},
		"snail": {
			"description": "optional",
			"accessors": ["."],
			"plural": "snails",
			"uuid": "20643981-dceb-4c3e-b3fd-8cb329650b65"
		},
		"moth": {
			"description": "optional",
			"accessors": ["."],
			"plural": "moths",
			"uuid": "cf97eff5-93bf-4d26-b8a2-4c5266c3641e"
		},
		"bird": {
			"description": "optional",
			"accessors": ["."],
			"plural": "birds",
			"uuid": "e5e2bb31-6950-4a18-b98f-f9a45e3c1eb7",
			"has": {
				"head": {
					"description": "optional",
					"accessors": [".has"],
					"uuid": "bc7f4dc3-0037-4517-893e-91e31725432d"
				},
				"wing": {
					"description": "optional",
					"accessors": [".has"],
					"uuid": "b7be845d-6192-4ef9-bbc3-be3b2cf0fe0b"
				},
				"foot": {
					"description": "optional",
					"accessors": [".has"],
					"uuid": "8828a901-b273-4e91-8805-54d4f6d516fb"
				}
			}
		},
		"dog": {
			"description": "optional",
			"uuid": "bac46d8e-0655-46d5-a393-36c69a18ee2c",
			"accessors": ["."],
			"isa": ["animal"],
			"plural": "dogs",
			"has": {
				"eye": {
					"description": "optional",
					"uuid": "f14d4609-b018-4641-b9c1-bf93fe220050",
					"accessors": [".has"]
				},
				"ear": {
					"description": "optional",
					"uuid": "889c1585-5c6c-4e3e-a757-514849d81521",
					"accessors": [".has"]
				},
				"mouth": {
					"description": "optional",
					"uuid": "d4304606-7d1f-4803-b7b4-7d37dcc30714",
					"accessors": [".has"]
				},
				"tail": {
					"description": "optional",
					"uuid": "3fb21e22-05cc-4700-acf9-806896e2bbb9",
					"accessors": [".has"]
				},
				"paw": {
					"description": "optional",
					"uuid": "a7725a15-95f0-4f3c-ab8b-8655662bdd46",
					"accessors": [".has"]
				},
				"nose": {
					"description": "optional",
					"uuid": "d8fa312b-38d3-4e50-9a6d-980c6c111af9",
					"accessors": [".has"]
				},
				"head": {
					"description": "optional",
					"uuid": "edb62ea7-d952-4953-96cc-97c31be71933",
					"accessors": [".has"]
				}
			},

			"quiz": [{
					"question": "What's the size of the dog",
					"uuid": "1e52541f-9a47-463d-b0cb-e32e1d3441f3",
					"accessors": [".size"],
					"answers": [{
							"name": "big",
							"uuid": "23c84822-5468-409c-aca8-22df259333fa"
						},
						{
							"name": "medium",
							"uuid": "b23519ff-8b2f-45c6-a22a-4dcabaee4649"
						},
						{
							"name": "small",
							"uuid": "f4a78d49-63e9-4ccd-951f-33b5231a83b6"
						}
					],
					"allow_unknown": true,
					"allow_other": true,
					"browse_by_example": false,
					"multiselect": false,
					"control_type": "radio"
				},
				{
					"question": "It's a...",
					"uuid": "66cf612f-dd80-4732-b7de-3f3d2055588d",
					"accessors": [".type"],
					"answers": [{
							"name": "adult",
							"uuid": "06fd1e49-b580-4ca6-aaff-34ca39062995"
						},
						{
							"name": "puppy",
							"uuid": "47dae924-d1c3-4ded-8479-ea965d157948"
						}
					],
					"allow_unknown": true,
					"allow_other": true,
					"browse_by_example": false,
					"multiselect": false,
					"control_type": "radio"
				},
				{
					"question": "What am I?",
					"uuid": "ceb2275b-4948-4d20-8cad-b0677259193e",
					"accessors": [".breed"],
					"answers": [{
							"name": "Labrador Retriever",
							"uuid": "d33ac818-ca09-45b0-b3b0-e16a4e422147",
							"examples": [{
								"filename": "LabradorRetriever.jpg"
							}]
						},
						{
							"name": "English Cocker Spaniel",
							"uuid": "8a725d6a-9ce8-4759-89ae-218a266e8dfd",
							"examples": [{
								"filename": "EnglishCockerSpaniel.jpg"
							}]
						},
						{
							"name": "English Springer Spaniel",
							"uuid": "ab7e8185-017b-44d8-9b7e-cda53d2fd0e7",
							"examples": [{
								"filename": "EnglishSpringerSpaniel.jpg"
							}]
						},
						{
							"name": "German Shepherd",
							"uuid": "430fd4e2-7bc7-4a97-b317-6d98e89d9fcf",
							"examples": [{
								"filename": "GermanShepherd.jpg"
							}]
						},
						{
							"name": "Staffordshire Bull Terrier",
							"uuid": "267736de-9e7f-4cce-b073-00e6458bdf65",
							"examples": [{
								"filename": "StaffordshireBullTerrier.jpg"
							}]
						},
						{
							"name": "Golden Retriever",
							"uuid": "20444b53-8cd0-4b60-b9a3-897c033229b3",
							"examples": [{
								"filename": "GoldenRetriever.jpg"
							}]
						},
						{
							"name": "Boxer",
							"uuid": "7ef05e8a-6cc5-43d5-a729-642b46368bcc",
							"examples": [{
								"filename": "Boxer.jpg"
							}]
						},
						{
							"name": "Beagle",
							"uuid": "8d699ef9-292b-4d71-8b87-711f7d14ac57",
							"examples": [{
								"filename": "Beagle.jpg"
							}]
						},
						{
							"name": "Dachshund",
							"uuid": "859defaa-ddc4-4540-b4fc-03d7383ca79c",
							"examples": [{
								"filename": "Dachshund.jpg"
							}]
						},
						{
							"name": "Poodle",
							"uuid": "435d46b6-d6cf-428e-8f33-d910a65f77f6",
							"examples": [{
								"filename": "Poodle.jpg"
							}]
						},
						{
							"name": "Rottweiler",
							"uuid": "97ac9993-5765-4804-9a88-9605c6474858",
							"examples": [{
								"filename": "Rottweiler.jpg"
							}]
						},
						{
							"name": "Siberian Husky",
							"uuid": "dc12147e-c6d7-480b-953c-b5c7843b3a15",
							"examples": [{
								"filename": "SiberianHusky.jpg"
							}]
						},
						{
							"name": "Bulldog",
							"uuid": "6ae55ed6-0c91-4bb1-b824-f2a83454d200",
							"examples": [{
								"filename": "Bulldog.jpg"
							}]
						},
						{
							"name": "Mops",
							"uuid": "ac775710-d593-493b-ad01-2ff52035ad4b",
							"examples": [{
								"filename": "Mops.jpg"
							}]
						},
						{
							"name": "Dalmatiner",
							"uuid": "48fbd2d5-4b25-4af6-801d-f45bbd3337b1",
							"examples": [{
								"filename": "Dalmatiner.jpg"
							}]
						},
						{
							"name": "Great Dane",
							"uuid": "390fe1ab-b6a2-459e-8cb7-f340d53f8933",
							"examples": [{
								"filename": "GreatDane.jpg"
							}]
						}




					],
					"allow_unknown": true,
					"allow_other": true,
					"browse_by_example": true,
					"multiselect": false,
					"control_type": "dropdown"
				}

			]
		},
		"cat": {
			"description": "optional",
			"uuid": "05c02fc5-9095-41e4-acb1-68303654ebb1",
			"accessors": ["."],
			"isa": ["animal"],
			"plural": "cats",
			"has": {
				"eye": {
					"description": "optional",
					"uuid": "d6a550b6-4384-43dc-a2e9-4abb6ae80462",
					"accessors": [".has"]
				},
				"ear": {
					"description": "optional",
					"uuid": "e6f48b23-4553-421b-be6d-0600905078c8",
					"accessors": [".has"]
				},
				"mouth": {
					"description": "optional",
					"uuid": "3bf40bb8-f671-46d5-9f6b-efe385010508",
					"accessors": [".has"]
				},
				"tail": {
					"description": "optional",
					"uuid": "44fdc5c6-15b9-4a3e-9f4d-c4e3a9d1d55a",
					"accessors": [".has"]
				},
				"paw": {
					"description": "optional",
					"uuid": "de5b8963-b32f-4050-a5ce-069ec27f67b2",
					"accessors": [".has"]
				},
				"nose": {
					"description": "optional",
					"uuid": "09c167cd-186b-4b93-a972-ddf1f9763c8f",
					"accessors": [".has"]
				},
				"head": {
					"description": "optional",
					"uuid": "33ed8c33-28fc-4920-8b43-1ffb89e5bef1",
					"accessors": [".has"]
				}
			}
		},
		"ant": {
			"description": "optional",
			"accessors": ["."],
			"plural": "ants",
			"uuid": "0245c27a-b45a-4ab0-bc1c-3397651b994e"
		},
		"elephant": {
			"description": "optional",
			"accessors": ["."],
			"plural": "elephants",
			"uuid": "76b1eb48-70b3-43d4-8890-5a967649a00d"
		},
		"fish": {
			"description": "optional",
			"accessors": ["."],
			"plural": "fishes",
			"uuid": "ab9bf24e-acc0-4857-85f8-907447c97127"
		},
		"horse": {
			"description": "optional",
			"accessors": ["."],
			"plural": "horses",
			"uuid": "6cdadd8b-6dd1-4ebb-bf7e-ae388ec9cdf2",
			"has": {
				"head": {
					"description": "optional",
					"uuid": "3cebcaae-dbbe-4357-9b31-f39b5cd4f8ee",
					"accessors": [".has"]
				}
			}
		},
		"sheep": {
			"description": "optional",
			"accessors": ["."],
			"plural": "sheeps",
			"uuid": "297932ac-5f16-4164-8997-a9d3343b2648"
		},
		"cow": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cows",
			"uuid": "04f9c653-b660-4471-8541-aaf0f33958a0",
			"has": {
				"head": {
					"description": "optional",
					"uuid": "23355225-a68e-47c3-aed1-c2840711afc3",
					"accessors": [".has"]
				}
			}
		},
		"butterfly": {
			"description": "optional",
			"accessors": ["."],
			"plural": "butterflies",
			"uuid": "9cb91307-ab62-489a-b774-0dea18e7f9ab"
		},
		"rabbit": {
			"description": "optional",
			"accessors": ["."],
			"plural": "rabbits",
			"uuid": "a86d4227-50c3-4ded-ace6-02d7321af718"
		},
		"bee": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bees",
			"uuid": "26233e92-cc48-4e80-bc47-5f8c84b4ab47"
		},
		"beetle": {
			"description": "optional",
			"accessors": ["."],
			"plural": "beatles",
			"uuid": "2a19e98a-3b0b-4a7d-85c7-31a49dfe3cf5"
		},
		"snake": {
			"description": "optional",
			"accessors": ["."],
			"plural": "snakes",
			"uuid": "8e650c03-85a0-445d-9706-7a69a94b43c9"
		},
		"frog": {
			"description": "optional",
			"accessors": ["."],
			"plural": "frogs",
			"uuid": "fdd05a68-ff42-47f8-8905-00611b3a17bf"
		},
		"jellyfish": {
			"description": "optional",
			"accessors": ["."],
			"plural": "jellyfish",
			"uuid": "16cc13ae-f7e8-4347-b362-9c70a61f74d8"
		},
		"spider": {
			"description": "optional",
			"accessors": ["."],
			"plural": "spiders",
			"uuid": "f7a12c25-bc80-4f58-883b-5c237f26d893"
		},
		"giraffe": {
			"description": "optional",
			"accessors": ["."],
			"plural": "giraffes",
			"uuid": "8b8d649e-a867-46ba-bee8-6ef007dadc45"
		},
		"parrot": {
			"description": "optional",
			"accessors": ["."],
			"plural": "parrots",
			"uuid": "b976951f-df87-4bfc-be87-9c3757bed400"
		},
		"wasp": {
			"description": "optional",
			"accessors": ["."],
			"plural": "wasps",
			"uuid": "a0ea8930-071c-4680-80ce-8ec9cbbb4acd"
		},
		"ladybird": {
			"description": "optional",
			"accessors": ["."],
			"plural": "ladybirds",
			"uuid": "3890426a-a97b-45df-8ce9-8395caeb8315"
		},
		"chicken": {
			"description": "optional",
			"accessors": ["."],
			"plural": "chickens",
			"uuid": "ddbc2e8b-f8a8-498e-9a5b-fb95bbd33a2e"
		},
		"tiger": {
			"description": "optional",
			"accessors": ["."],
			"plural": "tigers",
			"uuid": "2d6e8268-bf45-43eb-9f9c-c2dfa3ac5ba2"
		}
	}
}
