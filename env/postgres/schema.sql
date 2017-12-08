--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.5

-- Started on 2017-12-01 22:26:57

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 1 (class 3079 OID 12387)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 2433 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 3 (class 3079 OID 16822)
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- TOC entry 2434 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- TOC entry 4 (class 3079 OID 16741)
-- Name: temporal_tables; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS temporal_tables WITH SCHEMA public;


--
-- TOC entry 2435 (class 0 OID 0)
-- Dependencies: 4
-- Name: EXTENSION temporal_tables; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION temporal_tables IS 'temporal tables';


--
-- TOC entry 5 (class 3079 OID 16696)
-- Name: tsm_system_rows; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS tsm_system_rows WITH SCHEMA public;


--
-- TOC entry 2436 (class 0 OID 0)
-- Dependencies: 5
-- Name: EXTENSION tsm_system_rows; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION tsm_system_rows IS 'TABLESAMPLE method which accepts number of rows as a limit';


--
-- TOC entry 2 (class 3079 OID 16965)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 2437 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- TOC entry 716 (class 1247 OID 17024)
-- Name: control_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE control_type AS ENUM (
    'dropdown',
    'checkbox',
    'radio'
);


ALTER TYPE control_type OWNER TO postgres;

--
-- TOC entry 277 (class 1255 OID 16977)
-- Name: update_array_elements(jsonb, text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION update_array_elements(arr jsonb, key text, value jsonb) RETURNS jsonb
    LANGUAGE sql
    AS $$
    select jsonb_agg(jsonb_build_object(k, case when k <> key then v else value end))
    from jsonb_array_elements(arr) e(e), 
    lateral jsonb_each(e) p(k, v)
$$;


ALTER FUNCTION public.update_array_elements(arr jsonb, key text, value jsonb) OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16981)
-- Name: image_annotation_data_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_annotation_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_annotation_data_id_seq OWNER TO monkey;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 216 (class 1259 OID 16978)
-- Name: annotation_data; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_data (
    id bigint DEFAULT nextval('image_annotation_data_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    annotation jsonb,
    annotation_type_id bigint NOT NULL
);


ALTER TABLE annotation_data OWNER TO monkey;

--
-- TOC entry 218 (class 1259 OID 16995)
-- Name: annotation_type; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_type (
    id bigint NOT NULL,
    name text
);


ALTER TABLE annotation_type OWNER TO monkey;

--
-- TOC entry 224 (class 1259 OID 17070)
-- Name: annotations_per_app_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE annotations_per_app_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE annotations_per_app_id_seq OWNER TO monkey;

--
-- TOC entry 223 (class 1259 OID 17060)
-- Name: annotations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_app (
    id bigint DEFAULT nextval('annotations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE annotations_per_app OWNER TO monkey;

--
-- TOC entry 205 (class 1259 OID 16717)
-- Name: annotations_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE annotations_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE annotations_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 204 (class 1259 OID 16711)
-- Name: annotations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_country (
    id bigint DEFAULT nextval('annotations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE annotations_per_country OWNER TO monkey;

--
-- TOC entry 220 (class 1259 OID 17044)
-- Name: donations_per_app_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE donations_per_app_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE donations_per_app_id_seq OWNER TO monkey;

--
-- TOC entry 219 (class 1259 OID 17034)
-- Name: donations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_app (
    id bigint DEFAULT nextval('donations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE donations_per_app OWNER TO monkey;

--
-- TOC entry 203 (class 1259 OID 16704)
-- Name: donations_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE donations_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE donations_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 202 (class 1259 OID 16698)
-- Name: donations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_country (
    id bigint DEFAULT nextval('donations_per_country_id_seq'::regclass) NOT NULL,
    country_code text,
    count bigint
);


ALTER TABLE donations_per_country OWNER TO monkey;

--
-- TOC entry 191 (class 1259 OID 16417)
-- Name: image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image (
    id bigint NOT NULL,
    image_provider_id bigint,
    key text,
    unlocked boolean,
    hash bigint
);


ALTER TABLE image OWNER TO monkey;

--
-- TOC entry 201 (class 1259 OID 16668)
-- Name: image_annotation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_annotation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_annotation_id_seq OWNER TO monkey;

--
-- TOC entry 200 (class 1259 OID 16663)
-- Name: image_annotation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_annotation (
    id bigint DEFAULT nextval('image_annotation_id_seq'::regclass) NOT NULL,
    image_id bigint NOT NULL,
    num_of_valid integer NOT NULL,
    num_of_invalid integer NOT NULL,
    fingerprint_of_last_modification text,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    label_id bigint NOT NULL,
    uuid uuid
);


ALTER TABLE image_annotation OWNER TO monkey;

--
-- TOC entry 209 (class 1259 OID 16785)
-- Name: image_annotation_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_annotation_history (
    id bigint NOT NULL,
    image_id bigint NOT NULL,
    annotations jsonb,
    num_of_valid integer NOT NULL,
    num_of_invalid integer NOT NULL,
    fingerprint_of_last_modification text,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    label_id bigint,
    uuid uuid
);


ALTER TABLE image_annotation_history OWNER TO monkey;

--
-- TOC entry 211 (class 1259 OID 16906)
-- Name: image_annotation_refinement_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_annotation_refinement_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_annotation_refinement_id_seq OWNER TO monkey;

--
-- TOC entry 210 (class 1259 OID 16898)
-- Name: image_annotation_refinement; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_annotation_refinement (
    id bigint DEFAULT nextval('image_annotation_refinement_id_seq'::regclass) NOT NULL,
    annotation_data_id bigint,
    label_id bigint,
    num_of_valid integer,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    fingerprint_of_last_modification text
);


ALTER TABLE image_annotation_refinement OWNER TO monkey;

--
-- TOC entry 195 (class 1259 OID 16467)
-- Name: image_classification_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_classification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_classification_id_seq OWNER TO monkey;

--
-- TOC entry 190 (class 1259 OID 16415)
-- Name: image_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_id_seq OWNER TO monkey;

--
-- TOC entry 2438 (class 0 OID 0)
-- Dependencies: 190
-- Name: image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: monkey
--

ALTER SEQUENCE image_id_seq OWNED BY image.id;


--
-- TOC entry 192 (class 1259 OID 16434)
-- Name: image_provider_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_provider_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_provider_id_seq OWNER TO monkey;

--
-- TOC entry 189 (class 1259 OID 16385)
-- Name: image_provider; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_provider (
    id bigint DEFAULT nextval('image_provider_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE image_provider OWNER TO monkey;

--
-- TOC entry 199 (class 1259 OID 16620)
-- Name: report_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE report_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE report_id_seq OWNER TO monkey;

--
-- TOC entry 198 (class 1259 OID 16617)
-- Name: image_report; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_report (
    id bigint DEFAULT nextval('report_id_seq'::regclass) NOT NULL,
    reason text,
    image_id bigint
);


ALTER TABLE image_report OWNER TO monkey;

--
-- TOC entry 197 (class 1259 OID 16487)
-- Name: image_validation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_validation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_validation_id_seq OWNER TO monkey;

--
-- TOC entry 196 (class 1259 OID 16470)
-- Name: image_validation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation (
    id bigint DEFAULT nextval('image_validation_id_seq'::regclass) NOT NULL,
    image_id bigint,
    label_id bigint,
    num_of_valid integer,
    num_of_invalid integer,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    fingerprint_of_last_modification text
);


ALTER TABLE image_validation OWNER TO monkey;

--
-- TOC entry 208 (class 1259 OID 16759)
-- Name: image_validation_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation_history (
    id bigint NOT NULL,
    image_id bigint,
    label_id bigint,
    num_of_valid integer,
    num_of_invalid integer,
    sys_period tstzrange NOT NULL,
    fingerprint_of_last_modification text
);


ALTER TABLE image_validation_history OWNER TO monkey;

--
-- TOC entry 194 (class 1259 OID 16447)
-- Name: name_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE name_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE name_id_seq OWNER TO monkey;

--
-- TOC entry 193 (class 1259 OID 16437)
-- Name: label; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label (
    id bigint DEFAULT nextval('name_id_seq'::regclass) NOT NULL,
    name text,
    parent_id bigint
);


ALTER TABLE label OWNER TO monkey;

--
-- TOC entry 226 (class 1259 OID 17084)
-- Name: label_suggestion_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE label_suggestion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE label_suggestion_id_seq OWNER TO monkey;

--
-- TOC entry 225 (class 1259 OID 17074)
-- Name: label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_suggestion (
    id bigint DEFAULT nextval('label_suggestion_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE label_suggestion OWNER TO monkey;

--
-- TOC entry 215 (class 1259 OID 16939)
-- Name: quiz_answer_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE quiz_answer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE quiz_answer_id_seq OWNER TO monkey;

--
-- TOC entry 214 (class 1259 OID 16934)
-- Name: quiz_answer; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE quiz_answer (
    id bigint DEFAULT nextval('quiz_answer_id_seq'::regclass) NOT NULL,
    quiz_question_id bigint,
    label_id bigint
);


ALTER TABLE quiz_answer OWNER TO monkey;

--
-- TOC entry 213 (class 1259 OID 16926)
-- Name: quiz_question_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE quiz_question_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE quiz_question_id_seq OWNER TO monkey;

--
-- TOC entry 212 (class 1259 OID 16923)
-- Name: quiz_question; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE quiz_question (
    id bigint DEFAULT nextval('quiz_question_id_seq'::regclass) NOT NULL,
    question text,
    refines_label_id bigint,
    recommended_control control_type,
    allow_unknown boolean,
    allow_other boolean
);


ALTER TABLE quiz_question OWNER TO monkey;

--
-- TOC entry 222 (class 1259 OID 17057)
-- Name: validations_per_app_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE validations_per_app_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE validations_per_app_id_seq OWNER TO monkey;

--
-- TOC entry 221 (class 1259 OID 17047)
-- Name: validations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_app (
    id bigint DEFAULT nextval('validations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE validations_per_app OWNER TO monkey;

--
-- TOC entry 207 (class 1259 OID 16727)
-- Name: validations_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE validations_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE validations_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 206 (class 1259 OID 16724)
-- Name: validations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_country (
    id bigint DEFAULT nextval('validations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE validations_per_country OWNER TO monkey;

--
-- TOC entry 2191 (class 2604 OID 16420)
-- Name: image id; Type: DEFAULT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image ALTER COLUMN id SET DEFAULT nextval('image_id_seq'::regclass);


--
-- TOC entry 2274 (class 2606 OID 17002)
-- Name: annotation_type annotation_type_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2276 (class 2606 OID 17033)
-- Name: annotation_type annotation_type_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_name_unique UNIQUE (name);


--
-- TOC entry 2286 (class 2606 OID 17069)
-- Name: annotations_per_app annotations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2288 (class 2606 OID 17067)
-- Name: annotations_per_app annotations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2249 (class 2606 OID 16723)
-- Name: annotations_per_country annotations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2251 (class 2606 OID 16721)
-- Name: annotations_per_country annotations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2278 (class 2606 OID 17043)
-- Name: donations_per_app donations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2280 (class 2606 OID 17041)
-- Name: donations_per_app donations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2245 (class 2606 OID 16710)
-- Name: donations_per_country donations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2247 (class 2606 OID 16708)
-- Name: donations_per_country donations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2272 (class 2606 OID 16984)
-- Name: annotation_data image_annotation_data_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2239 (class 2606 OID 16667)
-- Name: image_annotation image_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2242 (class 2606 OID 16821)
-- Name: image_annotation image_annotation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_label_uniquekey UNIQUE (image_id, label_id);


--
-- TOC entry 2259 (class 2606 OID 16910)
-- Name: image_annotation_refinement image_annotation_refinement_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2261 (class 2606 OID 17021)
-- Name: image_annotation_refinement image_annotation_refinement_label_annotation_data_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_annotation_data_unique UNIQUE (annotation_data_id, label_id);


--
-- TOC entry 2216 (class 2606 OID 16652)
-- Name: image image_hash_unique_key; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_hash_unique_key UNIQUE (hash);


--
-- TOC entry 2218 (class 2606 OID 16425)
-- Name: image image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2222 (class 2606 OID 16427)
-- Name: image image_key_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_key_unique UNIQUE (image_provider_id, key);


--
-- TOC entry 2212 (class 2606 OID 16389)
-- Name: image_provider image_provider_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_provider
    ADD CONSTRAINT image_provider_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2229 (class 2606 OID 16474)
-- Name: image_validation image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2232 (class 2606 OID 16811)
-- Name: image_validation image_validation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_label_uniquekey UNIQUE (image_id, label_id);


--
-- TOC entry 2225 (class 2606 OID 16444)
-- Name: label label_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2290 (class 2606 OID 17083)
-- Name: label_suggestion label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2292 (class 2606 OID 17081)
-- Name: label_suggestion label_suggestion_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_name_unique UNIQUE (name);


--
-- TOC entry 2268 (class 2606 OID 16938)
-- Name: quiz_answer quiz_id_pley; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_id_pley PRIMARY KEY (id);


--
-- TOC entry 2264 (class 2606 OID 16943)
-- Name: quiz_question quiz_question_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2236 (class 2606 OID 16624)
-- Name: image_report report_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT report_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2282 (class 2606 OID 17056)
-- Name: validations_per_app validations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2284 (class 2606 OID 17054)
-- Name: validations_per_app validations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2253 (class 2606 OID 16736)
-- Name: validations_per_country validations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2255 (class 2606 OID 16734)
-- Name: validations_per_country validations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2269 (class 1259 OID 17008)
-- Name: fki_annotation_data_annotation_type_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_annotation_type_fkey ON annotation_data USING btree (annotation_type_id);


--
-- TOC entry 2270 (class 1259 OID 16991)
-- Name: fki_image_annotation_data_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_data_annotation_id_fkey ON annotation_data USING btree (image_annotation_id);


--
-- TOC entry 2237 (class 1259 OID 16819)
-- Name: fki_image_annotation_label_id_key; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_label_id_key ON image_annotation USING btree (label_id);


--
-- TOC entry 2213 (class 1259 OID 16433)
-- Name: fki_image_provider_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_provider_id_fkey ON image USING btree (image_provider_id);


--
-- TOC entry 2256 (class 1259 OID 16922)
-- Name: fki_image_quiz_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_image_annotation_id_fkey ON image_annotation_refinement USING btree (annotation_data_id);


--
-- TOC entry 2257 (class 1259 OID 16916)
-- Name: fki_image_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_label_id_fkey ON image_annotation_refinement USING btree (label_id);


--
-- TOC entry 2234 (class 1259 OID 16633)
-- Name: fki_image_report_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_report_image_id_fkey ON image_report USING btree (image_id);


--
-- TOC entry 2226 (class 1259 OID 16480)
-- Name: fki_image_validation_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_image_id_fkey ON image_validation USING btree (image_id);


--
-- TOC entry 2227 (class 1259 OID 16486)
-- Name: fki_image_validation_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_label_id_fkey ON image_validation USING btree (label_id);


--
-- TOC entry 2223 (class 1259 OID 16892)
-- Name: fki_label_parent_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_parent_id_fkey ON label USING btree (parent_id);


--
-- TOC entry 2265 (class 1259 OID 16949)
-- Name: fki_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_label_id_fkey ON quiz_answer USING btree (label_id);


--
-- TOC entry 2262 (class 1259 OID 16964)
-- Name: fki_quiz_question_refines_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_question_refines_label_id_fkey ON quiz_question USING btree (refines_label_id);


--
-- TOC entry 2266 (class 1259 OID 16955)
-- Name: fki_quiz_quiz_question_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_quiz_question_id_fkey ON quiz_answer USING btree (quiz_question_id);


--
-- TOC entry 2240 (class 1259 OID 16684)
-- Name: image_annotation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_image_id_index ON image_annotation USING btree (image_id);


--
-- TOC entry 2243 (class 1259 OID 16976)
-- Name: image_annotation_uuid_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_uuid_index ON image_annotation USING btree (uuid);


--
-- TOC entry 2214 (class 1259 OID 16683)
-- Name: image_hash_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_hash_index ON image USING btree (hash);


--
-- TOC entry 2219 (class 1259 OID 16681)
-- Name: image_image_provider_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_image_provider_index ON image USING btree (image_provider_id);


--
-- TOC entry 2220 (class 1259 OID 16682)
-- Name: image_key_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_key_index ON image USING btree (key);


--
-- TOC entry 2230 (class 1259 OID 16685)
-- Name: image_validation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_image_id_index ON image_validation USING btree (image_id);


--
-- TOC entry 2233 (class 1259 OID 16686)
-- Name: image_validation_label_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_label_id_index ON image_validation USING btree (label_id);


--
-- TOC entry 2308 (class 2620 OID 16792)
-- Name: image_annotation image_annotation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_annotation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON image_annotation FOR EACH ROW EXECUTE PROCEDURE versioning('sys_period', 'image_annotation_history', 'true');


--
-- TOC entry 2307 (class 2620 OID 16784)
-- Name: image_validation image_validation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_validation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON image_validation FOR EACH ROW EXECUTE PROCEDURE versioning('sys_period', 'image_validation_history', 'true');


--
-- TOC entry 2306 (class 2606 OID 17003)
-- Name: annotation_data annotation_data_annotation_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_annotation_type_fkey FOREIGN KEY (annotation_type_id) REFERENCES annotation_type(id);


--
-- TOC entry 2305 (class 2606 OID 16986)
-- Name: annotation_data image_annotation_data_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2298 (class 2606 OID 16671)
-- Name: image_annotation image_annotation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2299 (class 2606 OID 16814)
-- Name: image_annotation image_annotation_label_id_key; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_label_id_key FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2301 (class 2606 OID 17015)
-- Name: image_annotation_refinement image_annotation_refinement_annotation_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_annotation_data_id_fkey FOREIGN KEY (annotation_data_id) REFERENCES annotation_data(id);


--
-- TOC entry 2300 (class 2606 OID 16911)
-- Name: image_annotation_refinement image_annotation_refinement_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2293 (class 2606 OID 16612)
-- Name: image image_image_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_image_provider_id_fkey FOREIGN KEY (image_provider_id) REFERENCES image_provider(id);


--
-- TOC entry 2297 (class 2606 OID 16628)
-- Name: image_report image_report_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT image_report_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2295 (class 2606 OID 16475)
-- Name: image_validation image_validation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2296 (class 2606 OID 16481)
-- Name: image_validation image_validation_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2294 (class 2606 OID 16887)
-- Name: label label_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES label(id);


--
-- TOC entry 2303 (class 2606 OID 16944)
-- Name: quiz_answer quiz_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2302 (class 2606 OID 16959)
-- Name: quiz_question quiz_question_refines_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_refines_label_id_fkey FOREIGN KEY (refines_label_id) REFERENCES label(id);


--
-- TOC entry 2304 (class 2606 OID 16950)
-- Name: quiz_answer quiz_quiz_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_quiz_question_id_fkey FOREIGN KEY (quiz_question_id) REFERENCES quiz_question(id);


--
-- TOC entry 2432 (class 0 OID 0)
-- Dependencies: 7
-- Name: public; Type: ACL; Schema: -; Owner: monkey
--

REVOKE ALL ON SCHEMA public FROM postgres;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO monkey;
GRANT ALL ON SCHEMA public TO PUBLIC;


-- Completed on 2017-12-01 22:26:58

--
-- PostgreSQL database dump complete
--

