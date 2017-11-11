--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.5

-- Started on 2017-10-20 16:08:16

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET search_path = public, pg_catalog;

--
-- TOC entry 2249 (class 0 OID 16711)
-- Dependencies: 201
-- Data for Name: annotations_per_country; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY annotations_per_country (id, count, country_code) FROM stdin;
1	21	--
\.


--
-- TOC entry 2257 (class 0 OID 0)
-- Dependencies: 202
-- Name: annotations_per_country_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('annotations_per_country_id_seq', 21, true);


--
-- TOC entry 2247 (class 0 OID 16698)
-- Dependencies: 199
-- Data for Name: donations_per_country; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY donations_per_country (id, country_code, count) FROM stdin;
1	GB	1
2	--	2
\.


--
-- TOC entry 2258 (class 0 OID 0)
-- Dependencies: 200
-- Name: donations_per_country_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('donations_per_country_id_seq', 4, true);


--
-- TOC entry 2234 (class 0 OID 16385)
-- Dependencies: 186
-- Data for Name: image_provider; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY image_provider (id, name) FROM stdin;
2	flickr
3	pexels
4	donation
\.


--
-- TOC entry 2236 (class 0 OID 16417)
-- Dependencies: 188
-- Data for Name: image; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY image (id, image_provider_id, key, unlocked, hash) FROM stdin;
658	4	030388b6-4856-4421-bc33-e7925dccffb0	t	2874355956933665392
659	4	0f65c287-9420-4e02-a892-aa552724c6b4	t	218159733767216447
660	4	2b49ed71-4ce4-4e23-8ae1-64f36503e4f3	t	7428467669473489569
661	4	ab694b60-c157-4bcf-9294-1970be07b078	t	-30677415914899457
662	4	92f69f32-40b1-4556-91aa-f26e28e75330	t	-2235477576582668288
663	4	29911403-a297-4a18-89bb-bb6cea79048a	t	3978656574262879015
664	4	fb3bc619-eedd-46bb-8034-403c2f1385e0	t	829087890725027358
665	4	b1400b57-019c-4285-997d-0131dd116888	t	-434890200311792
666	4	31e52c37-6313-4b3b-ae1f-9c234e70159e	t	-1012762419733072897
667	4	8ca6702b-2c06-4fcd-a1f8-d16f62c76520	t	-1745211085012350977
668	4	4a9b4efc-6ed5-4add-bb59-7633cf87d6ae	t	519672605498359304
669	4	2ada8bf1-8aee-4374-a64f-4d870b4f2c89	t	-522170291684114433
670	4	39762321-8756-448f-a3ec-06e26801eab4	t	-34472371746497528
671	4	a9573baa-2dbf-48b4-9cb9-f35f51204e04	t	-218212659659175981
672	4	65e426d4-647b-4233-80bb-7b96ed02e163	t	5658632167040799
673	4	ab0165c5-3101-4301-899e-f3ab7c18f5c0	t	6017801976224545279
674	4	4b43558e-37e1-4b38-837d-f920258ec735	t	-4647759215910191104
675	4	55249a59-4505-47ba-b983-aef1d9a38f68	t	33253767512031
676	4	aa50324f-9b36-42a7-b59f-fa230b64dff2	t	2034518654573280800
677	4	40e5bead-d876-4ebe-b5b1-5f5d53fab6e4	t	-74239853582086144
\.


--
-- TOC entry 2245 (class 0 OID 16663)
-- Dependencies: 197
-- Data for Name: image_annotation; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY image_annotation (id, image_id, annotations, num_of_valid, num_of_invalid) FROM stdin;
127	659	[{"top": 52, "left": 21, "width": 893, "height": 490}]	1	0
126	664	[{"top": 44, "left": 153, "width": 693, "height": 551}]	0	0
128	666	[{"top": 69, "left": 94, "width": 455, "height": 437}]	1	0
129	667	[{"top": 24, "left": 241, "width": 563, "height": 605}]	1	0
125	663	[{"top": 167, "left": 338, "width": 408, "height": 371}]	6	1
123	658	[{"top": 152, "left": 161, "width": 137, "height": 323}, {"top": 152, "left": 403, "width": 320, "height": 401}, {"top": 192, "left": 543, "width": 300, "height": 213}, {"top": 79, "left": 123, "width": 228, "height": 25}, {"top": 37, "left": 659, "width": 198, "height": 79}]	3	0
124	660	[{"top": 4, "left": 143, "width": 183, "height": 331}]	14	1
\.


--
-- TOC entry 2259 (class 0 OID 0)
-- Dependencies: 198
-- Name: image_annotation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('image_annotation_id_seq', 129, true);


--
-- TOC entry 2260 (class 0 OID 0)
-- Dependencies: 192
-- Name: image_classification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('image_classification_id_seq', 5, true);


--
-- TOC entry 2261 (class 0 OID 0)
-- Dependencies: 187
-- Name: image_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('image_id_seq', 677, true);


--
-- TOC entry 2262 (class 0 OID 0)
-- Dependencies: 189
-- Name: image_provider_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('image_provider_id_seq', 4, true);


--
-- TOC entry 2243 (class 0 OID 16617)
-- Dependencies: 195
-- Data for Name: image_report; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY image_report (id, reason, image_id) FROM stdin;
7	Other Violation: additional info	658
8	Sensitive Content: additional info	661
9	Other Violation: additional info	662
10	Other Violation: test	658
11	Copyright Violation: bla	658
12	Other Violation: 	659
13	Sensitive Content: hhhh	661
14	Sensitive Content: hhhhhhhh	661
\.


--
-- TOC entry 2238 (class 0 OID 16437)
-- Dependencies: 190
-- Data for Name: label; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY label (id, name) FROM stdin;
43	banana
44	cat
45	dog
46	tennis ball
47	pizza
48	car
49	orange
50	apple
51	TV
52	smartphone
53	cup
54	glass
55	spoon
56	egg
57	bullet
58	tree
\.


--
-- TOC entry 2241 (class 0 OID 16470)
-- Dependencies: 193
-- Data for Name: image_validation; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY image_validation (id, image_id, label_id, num_of_valid, num_of_invalid) FROM stdin;
535	673	57	0	0
537	675	45	0	0
538	676	49	0	0
539	677	56	0	0
521	659	43	11	0
520	658	46	12	0
525	663	48	10	0
528	666	50	1	0
529	667	45	1	0
522	660	44	10	26
526	664	48	6	0
523	661	44	12	29
524	662	44	14	37
527	665	44	25	24
536	674	44	0	1
530	668	51	1	0
531	669	43	1	0
532	670	54	1	0
533	671	50	1	0
534	672	58	1	0
\.


--
-- TOC entry 2263 (class 0 OID 0)
-- Dependencies: 194
-- Name: image_validation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('image_validation_id_seq', 539, true);


--
-- TOC entry 2264 (class 0 OID 0)
-- Dependencies: 191
-- Name: name_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('name_id_seq', 58, true);


--
-- TOC entry 2265 (class 0 OID 0)
-- Dependencies: 196
-- Name: report_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('report_id_seq', 14, true);


--
-- TOC entry 2251 (class 0 OID 16724)
-- Dependencies: 203
-- Data for Name: validations_per_country; Type: TABLE DATA; Schema: public; Owner: monkey
--

COPY validations_per_country (id, count, country_code) FROM stdin;
1	28	--
29	1	AT
\.


--
-- TOC entry 2266 (class 0 OID 0)
-- Dependencies: 204
-- Name: validations_per_country_id_seq; Type: SEQUENCE SET; Schema: public; Owner: monkey
--

SELECT pg_catalog.setval('validations_per_country_id_seq', 29, true);


-- Completed on 2017-10-20 16:08:16

--
-- PostgreSQL database dump complete
--

