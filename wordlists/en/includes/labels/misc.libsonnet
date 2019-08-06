{
	"labels": {
		"pizza": {
			"description": "optional",
			"uuid": "be8270fd-2c5c-47ff-b938-0555e5201a18",
			"accessors": ["."],
			"isa": ["food"],
			"plural": "pizzas"
		},
		"orange": {
			"description": "optional",
			"uuid": "5ca7ccad-3b8c-4c9a-ac27-44cddc96d4fa",
			"accessors": ["."],
			"isa": ["food"],
			"plural": "oranges"
		},
		"apple": {
			"description": "optional",
			"uuid": "f81cf567-4798-4e4d-95f9-b430cf04ee55",
			"accessors": ["."],
			"isa": ["food"],
			"plural": "apples",
			"quiz": [
				{
					"question": "The apple's color is...",
					"uuid": "a4bcc81f-bebb-4c03-9950-eb92decccfba",
					"accessors": [".color"],
					"answers": [
						{
							"name": "red",
							"uuid": "7fbcde03-3778-4cd2-bab0-bbaedeb68f35"
						},
						{
							"name": "green",
							"uuid": "f89d9385-516f-41f4-b83e-74481b479050"
						},
						{ 
							"name": "yellow",
							"uuid": "37dc1749-8f34-4640-8fc2-abe5d2fa2b62"
						}
					],
					"allow_unknown": true,
					"allow_other": true,
					"browse_by_example": false,
					"multiselect": true,
					"control_type": "radio"
				}
			]
		},
		"banana": {
			"description": "optional",
			"uuid": "1c1dfe7c-0978-4c59-8728-7857a3867296",
			"accessors": ["."],
			"isa": ["food"],
			"plural": "bananas"
		},
		"car": {
			"description": "optional",
			"uuid": "67f02864-cb77-48aa-811c-7633a7d7d564",
			"accessors": ["."],
			"isa": ["vehicle"],
			"plural": "cars",
			"has": {
				"wheel": {
					"description": "optional",
					"uuid": "dd74f30a-7908-4120-8a8f-b12082e1b04c",
					"accessors": [".has"]
				},
				"headlight": {
					"description": "optional",
					"uuid": "e8eb2b56-921f-4b6e-985a-f10331c22fe7",
					"accessors": [".has"]
				}
			},
			"quiz": [
				{
					"question": "It's a...",
					"uuid": "e6d01e97-d53b-4da5-9c71-5f3fe36a10b9",
					"accessors": [".brand"],
					"answers": [
						{
							"name": "Seat",
							"uuid": "03f08e3e-442d-4fae-8cb8-edb945e5931a"
						}, 
						{
							"name": "Renault",
							"uuid": "cd4950c6-2da8-4dfd-8dc0-53a834e63faf"
						},
						{
							"name": "Peugot",
							"uuid": "fd2924a3-89f8-44a9-95cc-e8a745407106"
						},
						{
							"name": "BMW",
							"uuid": "8553c39d-408e-46b5-9543-66594a9d8cab"
						},
						{
							"name": "Ford",
							"uuid": "a7cf8f7d-bd75-4fa5-b3d6-bdb4bf7b0cc0"
						},
						{
							"name": "Opel",
							"uuid": "665a8c44-dab5-4cf7-af56-b3a6ccd3d62b"
						},
						{
							"name": "Alfa Romeo",
							"uuid": "6a76d6bd-c54d-4c74-9243-991d870f6d92"
						},
						{
							"name": "Chevrolet",
							"uuid": "04235cbd-9ee6-4808-91ff-084069b2fffa"
						},
						{
							"name": "Porsche",
							"uuid": "db0f5870-dba2-4f81-8fd6-4ad43a25c5fb"
						},
						{
							"name": "Honda",
							"uuid": "47e55f6c-d7d4-4bf0-8d17-c16938513bb7"
						},
						{
							"name": "Subaru",
							"uuid": "0a1f762b-72f0-4825-aa11-5953b8b9f866"
						},
						{
							"name": "Mazda",
							"uuid": "bd998d0c-caa0-437d-a83f-4d84d9e224b6"
						}, 
						{
							"name": "Mitsubishi",
							"uuid": "96b93795-987f-4a8d-905a-9944c54a2736"
						},
						{
							"name": "Lexus",
							"uuid": "9b03de96-6d8b-47b3-ac78-07d599a9a467"
						}, 
						{
							"name": "Toyota",
							"uuid": "3c89b447-65e3-4510-9f64-277cfa795e35"
						},
						{
							"name": "Volkswagen",
							"uuid": "677de3a6-6fc7-4731-a0ea-99bf546be8c4"
						},
						{
							"name": "Suzuki",
							"uuid": "69584d5d-8a66-4c55-995d-390303e92d0e"
						},
						{
							"name": "Mercedes-Benz",
							"uuid": "a0f76a59-73bf-4be2-a16b-b0dea8b455a4"
						},
						{
							"name": "Saab",
							"uuid": "0021b298-37cf-478c-b3b7-f0df57353a8d"
						},
						{
							"name": "Audi",
							"uuid": "b1f81da5-d758-444d-adc6-9215c38797f3"
						},
						{
							"name": "Kia",
							"uuid": "d3c2c5a0-3918-4914-a437-9232850ec29c"
						},
						{
							"name": "Land Rover",
							"uuid": "8b5fbe19-dfe1-498b-bd03-42e01965a716"
						},
						{
							"name": "Doge",
							"uuid": "c14da9b8-4f57-4b27-8db1-4f02cf78688d"
						},
						{
							"name": "Chrysler",
							"uuid": "6997418f-adc9-4812-a153-4443f5706150"
						}, 
						{
							"name": "Hummer",
							"uuid": "f75126dd-ae6e-43c9-8f4a-b690a67eb278"
						}, 
						{
							"name": "Hyundai",
							"uuid": "8c99f918-caca-469e-b628-299499fa7985"
						},
						{
							"name": "Jaguar",
							"uuid": "a97e90d3-5dee-44bf-9d03-797b493e3328"
						},
						{
							"name": "Jeep",
							"uuid": "f8c7154a-5379-40e3-89b9-227c39359dad"
						},
						{
							"name": "Nissan",
							"uuid": "b236ba54-dd16-492f-ba65-72a17f1ffce5"
						},
						{
							"name": "Volvo",
							"uuid": "e0762258-2703-4c7a-8fd0-6f135ac0399e"
						},
						{
							"name": "Daewoo",
							"uuid": "64d7ffee-380f-47b3-b7ce-111c44abc2c8"
						},
						{
							"name": "Fiat",
							"uuid": "b501f40a-ec22-4439-ace9-e2116e150274"
						}, 
						{
							"name": "MINI",
							"uuid": "ac0c6383-3163-45c2-98ad-765be1c9690c"
						},
						{
							"name": "Smart",
							"uuid": "7ca3a10f-5ffd-4a1d-8b34-ee2a3e419094"
						}
					],
					"allow_other": true,
					"allow_unknown": true,
					"multiselect": false,
					"control_type": "dropdown"

				},
				{
					"question": "The color of the car is...",
					"uuid": "f39743c8-f1ce-4750-b7ab-f57d15ac7270",
					"accessors": [".color"],
					"answers": [
						{
							"name": "white",
							"uuid": "cb704076-838a-4509-be88-d3401dab98d0"
						},
						{
							"name": "silver",
							"uuid": "f0f874b2-da85-4d36-978d-1b5d7e1ac416"
						},
						{
							"name": "black",
							"uuid": "5825406c-14f6-47ac-9c05-c5654f3305a8"
						},
						{
							"name": "grey",
							"uuid": "d8d7afe2-d363-4227-96d7-c39d3cb0683e"
						},
						{ 
							"name": "blue",
							"uuid": "c87757a6-0f0c-4710-97e3-eb9b9fea33b0"
						},
						{
							"name": "red",
							"uuid": "5509fc39-a4a9-40c6-bba4-0496b51c1972"
						},
						{
							"name": "brown",
							"uuid": "e1ceff72-427a-4994-bfb6-25d9805e6be0"
						},
						{
							"name": "green",
							"uuid": "92e3a70a-7438-4c1c-9b23-f5a51cfa9333"
						}
					],
					"allow_other": true,
					"allow_unknown": true,
					"multiselect": false,
					"control_type": "color tags"
				}
			]
		},
		"bicycle": {
			"description": "optional",
			"uuid": "c8cfc6a0-1a20-4e89-b879-d7378b882939",
			"accessors": ["."],
			"isa": ["vehicle"],
			"plural": "bicycles"
		},
		"building": {
			"description": "optional",
			"uuid": "3619dc01-f1e2-4791-9ddd-56550c2a6b7d",
			"accessors": ["."],
			"isa": [],
			"plural": "buildings"
		},
		"TV": {
			"description": "optional",
			"uuid": "3699ac56-1356-4695-9195-fbcf18471736",
			"accessors": ["."],
			"isa": ["electronics"],
			"plural": "TVs"
		},
		"smartphone": {
			"description": "optional",
			"uuid": "729f8759-ed57-4a45-984e-73b0f3f96e14",
			"accessors": ["."],
			"isa": ["electronics"],
			"plural": "smartphones"
		},
		"cup": {
			"description": "optional",
			"uuid": "654c2298-aa63-4d7d-b154-c8b521c26532",
			"accessors": ["."],
			"isa": [],
			"plural": "cups",
			"has": {
				"handle": {
					"description": "optional",
					"accessors": [".has"],
					"uuid": "8087f4cd-4102-47fc-bb90-0824b0b6897e"
				}
			}
		},
		"glass": {
			"description": "optional",
			"uuid": "f08885f5-e0fa-467f-addf-459013170516",
			"accessors": ["."],
			"isa": [],
			"plural": "glasses"
		},
		"spoon": {
			"description": "optional",
			"uuid": "fa12ab77-acef-40ba-90e1-aba5a8e2c245",
			"accessors": ["."],
			"isa": [],
			"plural": "spoons"
		},
		"egg": {
			"description": "optional",
			"uuid": "fe374440-d36d-4330-9145-ef92507f2c9d",
			"accessors": ["."],
			"isa": ["food"],
			"plural": "eggs"
		},
		"tennis ball": {
			"description": "optional",
			"uuid": "d040c820-cd47-4159-87e2-a0fe69fa678c",
			"accessors": ["."],
			"isa": ["sports"],
			"plural": "tennis balls"
		},
		"bullet": {
			"description": "optional",
			"uuid": "b4f05e83-60c9-4ad6-9d01-222f8d16d4ed",
			"accessors": ["."],
			"isa": [],
			"plural": "bullets"
		},
		"tree": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "de9c51d5-b633-4a92-be3f-2e09a7ed5dc4",
			"isa": [],
			"plural": "trees"
		},
		"person": {
			"description": "optional",
			"uuid": "64766828-a943-433f-8800-1901cebf959d",
			"accessors": ["."],
			"plural": "persons",
			"has": {
				"head": {
					"description": "optional",
					"uuid": "a9b9c7c8-9340-4b91-a573-d1f115b6d137",
					"accessors": [".has"]
				},
				"hand": {
					"description": "optional",
					"uuid": "9cf2d1ab-69a8-48eb-9adf-fd3977c56722",
					"accessors": [".has"]
				},
				"foot": {
					"description": "optional",
					"uuid": "9c62425d-874b-4b50-93fc-3f4de7ef1bd5",
					"accessors": [".has"]
				},
				"face": {
					"description": "optional",
					"uuid": "8184620a-b6bb-4a3e-9c69-9a75a18b7734",
					"accessors": [".has"]
				}
			}
		},
		"grass": {
			"description": "optional",
			"accessors": ["."],
			"plural": "grasses",
			"uuid": "677600b9-9e2e-48c2-8c4a-efbb72a0753f"
		},
		"road": {
			"description": "optional",
			"accessors": ["."],
			"plural": "roads",
			"uuid": "26ba089d-d11b-46ff-8f40-e292ba0e7624"
		},
		"bollard": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bollards",
			"uuid": "58899bce-7214-4930-8beb-cb06faa189e5"
		},
		"pavement": {
			"description": "optional",
			"accessors": ["."],
			"plural": "pavements",
			"uuid": "f81d3628-5f98-41e9-9967-1a13f88160c7"
		},
		"foliage": {
			"description": "optional",
			"accessors": ["."],
			"plural": "foliages",
			"uuid": "54f0c7a6-6c20-4e96-bf4a-28bebbc6580e"
		},
		"bush": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bushes",
			"uuid": "6e045775-6798-4b70-aaa8-b76b79ede454"
		},
		"chair": {
			"description": "optional",
			"accessors": ["."],
			"plural": "chairs",
			"uuid": "ace0d018-f335-4947-8604-69323f375b1b"
		},
		"plate": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "ba766c25-a0db-4458-b792-a9ec4f1777ac"
		},
		"wall": {
			"description": "optional",
			"accessors": ["."],
			"plural": "walls",
			"uuid": "9f6048e9-f867-4003-bb09-4087f087bea3"
		},
		"street light": {
			"description": "optional",
			"accessors": ["."],
			"plural": "street lights",
			"uuid": "430b0e22-bd5f-40a3-be13-15fbaf7eff99"
		},
		"sea": {
			"description": "optional",
			"accessors": ["."],
			"plural": "seas",
			"uuid": "157b4dc1-59de-4c49-830c-2c9ddb66dfee"
		},
		"river": {
			"description": "optional",
			"accessors": ["."],
			"plural": "rivers",
			"uuid": "d25e8353-2b45-434c-b8e7-8fe98ab2925b"
		},
		"lake": {
			"description": "optional",
			"accessors": ["."],
			"plural": "lakes",
			"uuid": "ddf7feaa-8c5f-4cab-84f2-6f1804998650"
		},
		"mushroom": {
			"description": "optional",
			"accessors": ["."],
			"plural": "mushrooms",
			"uuid": "f7781c7d-b963-45ad-b09e-677a2c626827"
		},
		"book": {
			"description": "optional",
			"accessors": ["."],
			"plural": "books",
			"uuid": "735c2bda-dd22-4240-a373-135cbc53a05b"
		},
		"bag": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bags",
			"uuid": "3bb8f4ee-7eb5-4a61-b1cc-0c50aac8df72"
		},
		"bridge": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bridges",
			"uuid": "50416c82-401a-469a-b2c4-85c64e4eee3a"
		},
		"onion": {
			"description": "optional",
			"accessors": ["."],
			"plural": "onions",
			"uuid": "caeabb57-4f03-4ff2-bcfb-fba30bf70de6"
		},
		"sun": {
			"description": "optional",
			"accessors": ["."],
			"plural": "suns",
			"uuid": "c6818c11-6c0a-4f51-904b-4c3e07bbbc2d"
		},
		"window": {
			"description": "optional",
			"accessors": ["."],
			"plural": "windows",
			"uuid": "ff212ebc-c56c-47d3-843e-3e4d8693c3e2"
		},
		"sunglasses": {
			"description": "optional",
			"accessors": ["."],
			"plural": "sunglasses",
			"uuid": "e15d84c5-1afa-439e-b089-ba1ce8b32d0f"
		},
		"house": {
			"description": "optional",
			"accessors": ["."],
			"uuid": "22bbce9a-9f50-4f38-ba11-6f2963d19809",
			"plural": "houses",
			"has": {
				"roof": {
					"description": "optional",
					"uuid": "b21f0558-956e-4739-9dbe-247e93ac42c0",
					"accessors": [".has"]
				}
			}
		},
		"van": {
			"description": "optional",
			"accessors": ["."],
			"plural": "vans",
			"uuid": "1849f4b6-f6a9-4ee1-8422-9662d4daccf9"
		},
		"sky": {
			"description": "optional",
			"accessors": ["."],
			"plural": "skies",
			"uuid": "459c447b-2919-475b-b6db-b5dfc3d3d676"
		},
		"truck": {
			"description": "optional",
			"accessors": ["."],
			"plural": "trucks",
			"uuid": "60c08783-9869-4d56-804f-14e14152b02f",
			"has": {
				"cabin": {
					"description": "optional",
					"uuid": "ff4187df-b9e3-4c4c-b342-07bbf2e38310",
					"accessors": [".has"]
				}
			}
		},
		"motorbike": {
			"description": "optional",
			"accessors": ["."],
			"plural": "motorbikes",
			"uuid": "3ec690a8-0f00-4bb1-851b-51739bffc95d"
		},
		"sofa": {
			"description": "optional",
			"accessors": ["."],
			"plural": "sofas",
			"uuid": "3d186e64-38e5-404c-9452-30ba5173db45"
		},
		"traffic light": {
			"description": "optional",
			"accessors": ["."],
			"plural": "traffic lights",
			"uuid": "1cf4fbe3-f381-4d10-af4d-5ac5199c94b8"
		},
		"cabbage": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cabbages",
			"uuid": "ae65d146-6428-453b-9f80-ddfadcd49e5e"
		},
		"bus": {
			"description": "optional",
			"accessors": ["."],
			"plural": "busses",
			"uuid": "8d1c1663-1569-4c8f-a72b-5def6695a942"
		},
		"excavator": {
			"description": "optional",
			"accessors": ["."],
			"plural": "excavators",
			"uuid": "e2a19fe2-9b15-49cd-83d2-3110d2b9d1ea"
		},
		"backhoe loader": {
			"description": "optional",
			"accessors": ["."],
			"plural": "backhoe loaders",
			"uuid": "8cc8ae96-6a8e-4221-aa10-9ef2059ac678"
		},
		"bulldozer": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bulldozers",
			"uuid": "39a47c82-a0f2-4b9c-8f22-00028b88004f"
		},
		"tractor": {
			"description": "optional",
			"accessors": ["."],
			"plural": "tractors",
			"uuid": "4ecf2dcf-cf95-46f5-b2fa-5f7ccdb3a4cd"
		},
		"fork lift": {
			"description": "optional",
			"accessors": ["."],
			"plural": "fork lifts",
			"uuid": "4d0d2782-ce7d-489c-8038-60dcace23ce7"
		},
		"crane": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cranes",
			"uuid": "e9a3970e-6627-47e0-84f3-29a9c5e9161a"
		},
		"helicopter": {
			"description": "optional",
			"accessors": ["."],
			"plural": "helicopters",
			"uuid": "e37f1c27-f188-4423-9ffc-f975b9890932"
		},
		"boat": {
			"description": "optional",
			"accessors": ["."],
			"plural": "boats",
			"uuid": "86bf68c3-31af-41fa-8e35-29fbe6ecda27"
		},
		"ship": {
			"description": "optional",
			"accessors": ["."],
			"plural": "ships",
			"uuid": "574aeaa2-7136-4e9c-a748-9b35446779de"
		},
		"table": {
			"description": "optional",
			"accessors": ["."],
			"plural": "tables",
			"uuid": "61db30b9-9903-4255-b878-f23679401c90"
		},
		"desk": {
			"description": "optional",
			"accessors": ["."],
			"plural": "desks",
			"uuid": "bba7aae9-3892-4367-9859-dbfdd05b9ab5"
		},
		"jet airliner": {
			"description": "optional",
			"accessors": ["."],
			"plural": "jet airliners",
			"uuid": "e6870b75-9413-4e9f-8059-1e746732007e"
		},
		"tree trunk": {
			"description": "optional",
			"accessors": ["."],
			"plural": "tree trunks",
			"uuid": "2af2b428-f5dc-4998-af0f-6a4ed5ade8cc"
		},
		"pillar": {
			"description": "optional",
			"accessors": ["."],
			"plural": "pillars",
			"uuid": "a90f30ae-50bc-478d-9fec-c8f8f62f16b9"
		},
		"vegetation": {
			"description": "optional",
			"accessors": ["."],
			"plural": "vegetations",
			"uuid": "6d321ba0-a0ab-48de-ae42-feaf75035708"
		},
		"woman": {
			"description": "optional",
			"accessors": ["."],
			"plural": "women",
			"uuid": "eecd25ba-8ae1-43e5-b3a5-e43f3e7cdd67"
		},
		"man": {
			"description": "optional",
			"accessors": ["."],
			"plural": "men",
			"uuid": "57d71782-fcb6-49bc-aac4-3122de81b89b"
		},
		"shelf": {
			"description": "optional",
			"accessors": ["."],
			"plural": "shelfs",
			"uuid": "2722232d-2e5a-408b-bff7-f4c5e1b280b5"
		},
		"bench": {
			"description": "optional",
			"accessors": ["."],
			"plural": "benches",
			"uuid": "773fcfc8-5a91-4a3e-900d-ba7f03828b93"
		},
		"park bench": {
			"description": "optional",
			"accessors": ["."],
			"plural": "park benches",
			"uuid": "70a063a4-fed4-49fb-988f-cd40276cfb76"
		},
		"stool": {
			"description": "optional",
			"accessors": ["."],
			"plural": "stools",
			"uuid": "6e720002-2a85-42b9-ac91-a4792000c2da"
		},
		"cupboard": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cupboards",
			"uuid": "9ee04fe2-4012-4a37-a97b-235af0519099"
		},
		"shop front": {
			"description": "optional",
			"accessors": ["."],
			"plural": "shop fronts",
			"uuid": "1f887b33-d1ab-4559-ac30-f595831cd1d7"
		},
		"door": {
			"description": "optional",
			"accessors": ["."],
			"plural": "doors",
			"uuid": "0291fabe-39d8-4ec9-b539-76bd77fb5c96"
		},
		"traffic cone": {
			"description": "optional",
			"accessors": ["."],
			"plural": "traffic cones",
			"uuid": "d753a069-2b21-490a-987b-ed6d055eea43"
		},
		"church": {
			"description": "optional",
			"accessors": ["."],
			"plural": "churches",
			"uuid": "ac3c81eb-7e0a-40e6-acc9-e1c58ec02c36"
		},
		"laptop": {
			"description": "optional",
			"accessors": ["."],
			"plural": "laptops",
			"uuid": "f605385b-86e1-4074-a123-5c0057bf911b"
		},
		"tomato": {
			"description": "optional",
			"accessors": ["."],
			"plural": "tomatoes",
			"uuid": "5b8a204d-8dee-4422-9afe-fd7567a5b633"
		},
		"steps": {
			"description": "optional",
			"accessors": ["."],
			"plural": "steps",
			"uuid": "269e038e-7395-4bf1-ab98-1eb57c90374c"
		},
		"path": {
			"description": "optional",
			"accessors": ["."],
			"plural": "paths",
			"uuid": "689421e5-5779-4ba0-83c3-1ad145397d5d"
		},
		"gate": {
			"description": "optional",
			"accessors": ["."],
			"plural": "gates",
			"uuid": "de63bd0c-bed2-4af2-9ae4-5bc24e035362"
		},
		"hat": {
			"description": "optional",
			"accessors": ["."],
			"plural": "hats",
			"uuid": "13127a19-78b7-4f36-b67a-cc4ffaf4ac13"
		},
		"litter": {
			"description": "optional",
			"accessors": ["."],
			"plural": "litters",
			"uuid": "95b3f6e8-d434-4dff-9b00-abe114a4b923"
		},
		"puddle": {
			"description": "optional",
			"accessors": ["."],
			"plural": "puddles",
			"uuid": "9599ced7-c4de-44db-a626-38c6dd43ef9d"
		},
		"mountain": {
			"description": "optional",
			"accessors": ["."],
			"plural": "mountains",
			"uuid": "2c4306a7-242f-49d5-a817-774c43ff985e"
		},
		"rucksack": {
			"description": "optional",
			"accessors": ["."],
			"plural": "rucksacks",
			"uuid": "f702402b-a6ea-4c57-bc1a-bb173e17715f"
		},
		"flag": {
			"description": "optional",
			"accessors": ["."],
			"plural": "flags",
			"uuid": "651b2ef9-3c5c-4f73-900a-4f344de4b644"
		},
		"statue": {
			"description": "optional",
			"accessors": ["."],
			"plural": "statues",
			"uuid": "f641d6e9-a887-4d8c-ae9d-4d9e27fdc279"
		},
		"spectacles": {
			"description": "optional",
			"accessors": ["."],
			"plural": "spectacles",
			"uuid": "1a139fb1-9c1a-4f2d-8b3a-5dab56bd8e43"
		},
		"floor": {
			"description": "optional",
			"accessors": ["."],
			"plural": "floors",
			"uuid": "cab8e973-3c19-426b-8239-ccc5a899368b"
		},
		"basket": {
			"description": "optional",
			"accessors": ["."],
			"plural": "baskets",
			"uuid": "030f8011-453a-42d6-985f-ce7cdd17a6db"
		},
		"napkin": {
			"description": "optional",
			"accessors": ["."],
			"plural": "napkins",
			"uuid": "f5dda70a-8316-4833-83be-f94b62c0e381"
		},
		"plant": {
			"description": "optional",
			"accessors": ["."],
			"plural": "plants",
			"uuid": "c98e3948-040b-43d8-bd3a-97536de10a9e"
		},
		"pen": {
			"description": "optional",
			"accessors": ["."],
			"plural": "pens",
			"uuid": "cabd5ba9-fd89-4310-be88-957a82cc481f"
		},
		"camera": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cameras",
			"uuid": "51b99c35-8c36-4862-83f2-5b361df36aef"
		},
		"snow": {
			"description": "optional",
			"accessors": ["."],
			"plural": "snow",
			"uuid": "f958ea3d-5b45-4109-a606-9b6342712f00"
		},
		"flower": {
			"description": "optional",
			"accessors": ["."],
			"plural": "flowers",
			"uuid": "7528260c-bc75-48d5-9f9a-30c8c3dcdf8f"
		},
		"umbrella": {
			"description": "optional",
			"accessors": ["."],
			"plural": "umbrellas",
			"uuid": "09d6f999-0e55-4bd7-a9bf-b658389d7a3d"
		},
		"bed": {
			"description": "optional",
			"accessors": ["."],
			"plural": "beds",
			"uuid": "40735064-ecac-47c3-8020-3adafcad7a99"
		},
		"chain": {
			"description": "optional",
			"accessors": ["."],
			"plural": "chains",
			"uuid": "5a427d30-b685-46e1-bbdb-2f19f47f56cd"
		},
		"strawberry": {
			"description": "optional",
			"accessors": ["."],
			"plural": "strawberries",
			"uuid": "661797ef-31f9-4969-9c22-0b5704603969"
		},
		"soil": {
			"description": "optional",
			"accessors": ["."],
			"plural": "soil",
			"uuid": "d3ca96c8-13e3-44a3-b14e-72032e881139"
		},
		"bottle": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bottles",
			"uuid": "cbe76e44-759e-4edd-a5a3-87ef98e398a3"
		},
		"field": {
			"description": "optional",
			"accessors": ["."],
			"plural": "fields",
			"uuid": "19f062dc-1f03-46fe-b672-87931596358b"
		},
		"forest": {
			"description": "optional",
			"accessors": ["."],
			"plural": "forests",
			"uuid": "476c449a-096c-4ed5-9b19-c4b19cf5b434"
		},
		"beach": {
			"description": "optional",
			"accessors": ["."],
			"plural": "beaches",
			"uuid": "1e37be04-df57-47c7-9f3e-fd7cbb01f980"
		},
		"carpet": {
			"description": "optional",
			"accessors": ["."],
			"plural": "carpets",
			"uuid": "a475be1e-963c-412e-bf4b-fd57ba069e0c"
		},
		"rock": {
			"description": "optional",
			"accessors": ["."],
			"plural": "rocks",
			"uuid": "c7cb83a3-1317-48dd-bd66-c93c2941e30a"
		},
		"chimney": {
			"description": "optional",
			"accessors": ["."],
			"plural": "chimneys",
			"uuid": "4051de80-f782-47be-a806-878ccdabd2a2"
		},
		"fork": {
			"description": "optional",
			"accessors": ["."],
			"plural": "forks",
			"uuid": "9b087567-b49b-47da-afbb-a6f5e9f0cf0f"
		},
		"cactus": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cactuses",
			"uuid": "2de14381-5204-431e-996a-261a411c50e8"
		},
		"knife": {
			"description": "optional",
			"accessors": ["."],
			"plural": "knifes",
			"uuid": "0db1c5b5-2feb-41fd-bbd0-33ff156f36c0"
		},
		"guitar": {
			"description": "optional",
			"accessors": ["."],
			"plural": "guitars",
			"uuid": "00e8851f-715e-4c03-8597-f2b2545a3ca2"
		},
		"football": {
			"description": "optional",
			"accessors": ["."],
			"plural": "footballs",
			"uuid": "9d0c112d-34eb-4d04-ac5c-0995af42dbc6"
		},
		"sand": {
			"description": "optional",
			"accessors": ["."],
			"plural": "sands",
			"uuid": "fd5e9951-4e65-471e-9ade-b08d85eedd98"
		},
		"temple": {
			"description": "optional",
			"accessors": ["."],
			"plural": "temples",
			"uuid": "7ccdfd4d-e13a-4518-8d10-e3743fe97d80"
		},
		"carrot": {
			"description": "optional",
			"accessors": ["."],
			"plural": "carrots",
			"uuid": "9eb38bb1-7468-4933-a277-aa89d18f5d4f"
		},
		"fence": {
			"description": "optional",
			"accessors": ["."],
			"plural": "fences",
			"uuid": "f45f733f-a497-473c-9254-b1d52f32ab4c"
		},
		"wrist watch": {
			"description": "optional",
			"accessors": ["."],
			"plural": "wrist watches",
			"uuid": "2ee8481d-4319-4249-b85e-788859e628a9"
		},
		"glove": {
			"description": "optional",
			"accessors": ["."],
			"plural": "gloves",
			"uuid": "ad31d449-4f82-45d4-a907-02d453dd3737"
		},
		"coat": {
			"description": "optional",
			"accessors": ["."],
			"plural": "coats",
			"uuid": "f321bfbd-897c-4edc-9529-dd329e5d82be"
		},
		"pineapple": {
			"description": "optional",
			"accessors": ["."],
			"plural": "pineapples",
			"uuid": "1ab94463-3344-4615-a5a4-661e26d2d3b9"
		},
		"bowl": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bowls",
			"uuid": "168690ff-7da0-4b7b-bcdb-d740dc355f15"
		},
		"sculpture": {
			"description": "optional",
			"accessors": ["."],
			"plural": "sculptures",
			"uuid": "879940b2-247b-4079-9349-3b834b58a8c2"
		},
		"ladder": {
			"description": "optional",
			"accessors": ["."],
			"plural": "ladders",
			"uuid": "88b3789c-4acf-411b-bb2e-6c487532b7b4"
		},
		"train": {
			"description": "optional",
			"accessors": ["."],
			"plural": "trains",
			"uuid": "dcf65e07-a71f-4226-8efd-ec6846960ba2"
		},
		"swimming pool": {
			"description": "optional",
			"accessors": ["."],
			"plural": "swimming pools",
			"uuid": "d4a2bec5-5e2a-4e30-818e-4f975d5c5c78"
		},
		"castle": {
			"description": "optional",
			"accessors": ["."],
			"plural": "castles",
			"uuid": "2897e51d-218c-4202-b3ce-fd82ab9ea434"
		},
		"harbour": {
			"description": "optional",
			"accessors": ["."],
			"plural": "harbours",
			"uuid": "0e6c12b1-8fe9-4022-985c-6b9ea3cf0bca"
		},
		"tower": {
			"description": "optional",
			"accessors": ["."],
			"plural": "towers",
			"uuid": "996093dc-ac0c-4458-b7f4-0d4e12edc42d"
		},
		"pallet": {
			"description": "optional",
			"accessors": ["."],
			"plural": "pallets",
			"uuid": "d7db16e2-ed10-49f1-b1c1-458993c7ec21"
		},
		"barrier": {
			"description": "optional",
			"accessors": ["."],
			"plural": "barriers",
			"uuid": "a9640f8d-9c1c-457c-91de-62149b1b3951"
		},
		"helmet": {
			"description": "optional",
			"accessors": ["."],
			"plural": "helmets",
			"uuid": "14abedd0-85b9-4b83-a103-647bcfd1003f"
		},
		"wind turbine": {
			"description": "optional",
			"accessors": ["."],
			"plural": "wind turbines",
			"uuid": "b647f5f9-7bf4-4aa5-a9cf-74238b77e11b"
		},
		"typewriter": {
			"description": "optional",
			"accessors": ["."],
			"plural": "typewriters",
			"uuid": "42e1f70e-15ad-4f5c-bd6c-a92d1344d423"
		},
		"clock": {
			"description": "optional",
			"accessors": ["."],
			"plural": "clocks",
			"uuid": "282e4dd7-d519-49ea-858b-f100309ce0bf"
		},
		"skyscraper": {
			"description": "optional",
			"accessors": ["."],
			"plural": "skyscrapers",
			"uuid": "30431063-77e0-47cd-9241-2fc4c7e06b76"
		},
		"meat": {
			"description": "optional",
			"accessors": ["."],
			"plural": "meat",
			"uuid": "9edffd21-99b8-4ee1-89f0-1bf4ad1ff92f"
		},
		"cake": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cakes",
			"uuid": "6ac249f5-c741-4280-8397-d4783b4fcca0"
		},
		"rope": {
			"description": "optional",
			"accessors": ["."],
			"plural": "ropes",
			"uuid": "db9accb6-e4e0-4898-bb68-bb3df7807f63"
		},
		"towel": {
			"description": "optional",
			"accessors": ["."],
			"plural": "towels",
			"uuid": "14523e0c-9910-491e-985e-8d8a60b5993a"
		},
		"bread": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bread",
			"uuid": "dd2a451a-042f-4994-b9d5-fa7ddbdce5a1"
		},
		"bus shelter": {
			"description": "optional",
			"accessors": ["."],
			"plural": "bus shelters",
			"uuid": "fbdb1202-3d1d-47af-bec4-06cdc2e3189f"
		},
		"jeans": {
			"description": "optional",
			"accessors": ["."],
			"plural": "jeans",
			"uuid": "aa947d16-4f65-4338-b814-507a2d5acac9"
		},
        "telephone box": {
            "description": "optional",
            "accessors": ["."],
            "plural": "telephone boxes",
            "uuid": "6439a508-960c-4672-b50d-5faf3f895f17"
        },
        "ball": {
            "description": "optional",
            "accessors": ["."],
            "plural": "balls",
            "uuid": "04c61035-a112-4834-94be-1a9bd0ef1a9f"
        },
        "cathedral": {
            "description": "optional",
            "accessors": ["."],
            "plural": "cathedrals",
            "uuid": "85c8e7f4-c6ab-4e1c-b845-9798fc82236d"
        },
        "canal": {
            "description": "optional",
            "accessors": ["."],
            "plural": "canal",
            "uuid": "1a506fa1-3e6b-4100-bf48-8889bca2b93e"
        },
        "padlock": {
            "description": "optional",
            "accessors": ["."],
            "plural": "padlocks",
            "uuid": "02d71ded-3d01-421b-8ef5-4c1300300f1c"
        },
        "barrel": {
            "description": "optional",
            "accessors": ["."],
            "plural": "barrels",
            "uuid": "07bdca85-e003-442d-9464-a82d62388aae"
        },
        "salad": {
            "description": "optional",
            "accessors": ["."],
            "plural": "salads",
            "uuid": "30a1e602-d453-4f99-a1c5-9b3529da09c5"
        },
        "tent": {
            "description": "optional",
            "accessors": ["."],
            "plural": "tents",
            "uuid": "fda43d96-1c96-45f1-be67-34973ad62059"
        },
        "curtain": {
            "description": "optional",
            "accessors": ["."],
            "plural": "curtains",
            "uuid": "90c9a7b5-d79d-4ed5-8cb7-3f41b6a150e4"
        },
        "washing machine": {
            "description": "optional",
            "accessors": ["."],
            "plural": "washing machines",
            "uuid": "e1c0679a-f6bf-428c-85de-01853570bef6"
        },
        "tie": {
            "description": "optional",
            "accessors": ["."],
            "plural": "ties",
            "uuid": "59d7f359-4da0-4338-abd0-1e9ae0543abb"
        },
        "raspberry": {
            "description": "optional",
            "accessors": ["."],
            "plural": "raspberries",
            "uuid": "11a6e3be-c590-466c-b4dc-4c47e8d5c49d"
        },
        "parking space": {
            "description": "optional",
            "accessors": ["."],
            "plural": "parking space",
            "uuid": "01c5957f-fc79-41c6-85ac-417a07667c2e"
        },
        "skateboard": {
            "description": "optional",
            "accessors": ["."],
            "plural": "skateboards",
            "uuid": "4bdc78c2-e921-45bb-9e0b-1cb9cf879599"
        },
        "computer monitor": {
            "description": "optional",
            "accessors": ["."],
            "plural": "computer monitors",
            "uuid": "d6a993dc-09df-4569-9416-29ce350a97e6"
        },
        "picture frame": {
            "description": "optional",
            "accessors": ["."],
            "plural": "picture frames",
            "uuid": "199a70a0-16d4-4eb5-a5e7-9527193477a4"
        },
        "blanket": {
            "description": "optional",
            "accessors": ["."],
            "plural": "blankets",
            "uuid": "445d397f-5145-45ab-abe7-48980163e5ea"
        },
        "cushion": {
            "description": "optional",
            "accessors": ["."],
            "plural": "cushions",
            "uuid": "5bea1931-d8e3-47c6-be24-8ab016694a74"
        },
        "blueberry": {
            "description": "optional",
            "accessors": ["."],
            "plural": "blueberries",
            "uuid": "5c3eae85-f3f6-4572-a789-f37b9b2cc8b7"
        },
        "jar": {
            "description": "optional",
            "accessors": ["."],
            "plural": "jar",
            "uuid": "0bcbc385-b00c-475e-b8f7-5d248da9a296"
        },
		"cucumber": {
			"description": "optional",
			"accessors": ["."],
			"plural": "cucumber",
			"uuid": "d38f5335-91d8-40a2-a9f4-a2ddc8970132" 
		},
		"bucket": {
			"description": "optional",
			"accessors": ["."],
			"plural": "buckets",
			"uuid": "10d32e6d-ba38-4336-a563-9c2a3ce982ee"
		},
		"escalator": {
			"description": "optional",
			"accessors": ["."],
			"plural": "escalator",
			"uuid": "7be9984a-ef3a-49f8-a68e-1c3382831d1a"
		},
		"smoke": {
			"description": "optional",
			"accessors": ["."],
			"plural": "smoke",
			"uuid": "1f0ebe47-b86e-4a1e-b892-e2dd60434fa3"
		},
		"ceiling": {
			"description": "optional",
			"accessors": ["."],
			"plural": "ceilings",
			"uuid": "bc3cf624-eb9b-4c81-a2d2-693fe27ef00e"
		},
		"balcony": {
			"description": "optional",
			"accessors": ["."],
			"plural": "balcony",
			"uuid": "fc937332-6647-4729-aa62-f4edb237639b"
		},
		"tool": {
			"description": "optional",
			"accessors": ["."],
			"plural": "tools",
			"uuid": "21ac38ea-4c47-414d-9fc5-dd3f65d316f9"
		},
		"suit": {
			"description": "optional",
			"accessors": ["."],
			"plural": "suits",
			"uuid": "521d4178-b144-49d2-8e5b-016c33e69c7a"
		},
		"dress": {
			"description": "optional",
			"accessors": ["."],
			"plural": "dresses",
			"uuid": "9b572278-e0b7-47ac-b721-641fe6a9a386"
		},
		"shoe": {
			"description": "optional",
			"accessors": ["."],
			"plural": "shoes",
			"uuid": "14c7d814-00f2-47f2-b4c5-f3f8872994bc"
		}
	}
}
