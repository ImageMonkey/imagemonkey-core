--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.5

-- Started on 2018-07-08 21:22:40

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
-- TOC entry 2609 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 3 (class 3079 OID 164561)
-- Name: temporal_tables; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS temporal_tables WITH SCHEMA public;


--
-- TOC entry 2610 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION temporal_tables; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION temporal_tables IS 'temporal tables';


--
-- TOC entry 2 (class 3079 OID 164564)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 2611 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- TOC entry 562 (class 1247 OID 164576)
-- Name: control_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE control_type AS ENUM (
    'dropdown',
    'checkbox',
    'radio',
    'color tags'
);


ALTER TYPE control_type OWNER TO postgres;

--
-- TOC entry 269 (class 1255 OID 164585)
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
-- TOC entry 187 (class 1259 OID 164586)
-- Name: access_token_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE access_token_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE access_token_id_seq OWNER TO monkey;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 188 (class 1259 OID 164588)
-- Name: access_token; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE access_token (
    id bigint DEFAULT nextval('access_token_id_seq'::regclass) NOT NULL,
    user_id bigint,
    token text,
    expiration_time bigint
);


ALTER TABLE access_token OWNER TO monkey;

--
-- TOC entry 189 (class 1259 OID 164595)
-- Name: account_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE account_id_seq OWNER TO monkey;

--
-- TOC entry 190 (class 1259 OID 164597)
-- Name: account; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE account (
    id bigint DEFAULT nextval('account_id_seq'::regclass) NOT NULL,
    name text,
    hashed_password text,
    email text,
    profile_picture text,
    created bigint NOT NULL,
    is_moderator boolean NOT NULL
);


ALTER TABLE account OWNER TO monkey;

--
-- TOC entry 191 (class 1259 OID 164604)
-- Name: account_permission_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE account_permission_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE account_permission_id_seq OWNER TO monkey;

--
-- TOC entry 192 (class 1259 OID 164606)
-- Name: account_permission; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE account_permission (
    id bigint DEFAULT nextval('account_permission_id_seq'::regclass) NOT NULL,
    can_remove_label boolean,
    account_id bigint
);


ALTER TABLE account_permission OWNER TO monkey;

--
-- TOC entry 193 (class 1259 OID 164610)
-- Name: image_annotation_data_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_annotation_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_annotation_data_id_seq OWNER TO monkey;

--
-- TOC entry 194 (class 1259 OID 164612)
-- Name: annotation_data; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_data (
    id bigint DEFAULT nextval('image_annotation_data_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    annotation jsonb,
    annotation_type_id bigint NOT NULL,
    image_annotation_revision_id bigint
);


ALTER TABLE annotation_data OWNER TO monkey;

--
-- TOC entry 195 (class 1259 OID 164619)
-- Name: annotation_type; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_type (
    id bigint NOT NULL,
    name text
);


ALTER TABLE annotation_type OWNER TO monkey;

--
-- TOC entry 196 (class 1259 OID 164625)
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
-- TOC entry 197 (class 1259 OID 164627)
-- Name: annotations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_app (
    id bigint DEFAULT nextval('annotations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE annotations_per_app OWNER TO monkey;

--
-- TOC entry 198 (class 1259 OID 164634)
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
-- TOC entry 199 (class 1259 OID 164636)
-- Name: annotations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_country (
    id bigint DEFAULT nextval('annotations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE annotations_per_country OWNER TO monkey;

--
-- TOC entry 200 (class 1259 OID 164643)
-- Name: api_token_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE api_token_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE api_token_id_seq OWNER TO monkey;

--
-- TOC entry 201 (class 1259 OID 164645)
-- Name: api_token; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE api_token (
    id bigint DEFAULT nextval('api_token_id_seq'::regclass) NOT NULL,
    description text,
    token text,
    issued_at bigint,
    account_id bigint,
    revoked boolean,
    expires_at bigint
);


ALTER TABLE api_token OWNER TO monkey;

--
-- TOC entry 202 (class 1259 OID 164652)
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
-- TOC entry 203 (class 1259 OID 164654)
-- Name: donations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_app (
    id bigint DEFAULT nextval('donations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE donations_per_app OWNER TO monkey;

--
-- TOC entry 204 (class 1259 OID 164661)
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
-- TOC entry 205 (class 1259 OID 164663)
-- Name: donations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_country (
    id bigint DEFAULT nextval('donations_per_country_id_seq'::regclass) NOT NULL,
    country_code text,
    count bigint
);


ALTER TABLE donations_per_country OWNER TO monkey;

--
-- TOC entry 206 (class 1259 OID 164670)
-- Name: image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image (
    id bigint NOT NULL,
    image_provider_id bigint,
    key text,
    unlocked boolean,
    hash bigint,
    width integer NOT NULL,
    height integer NOT NULL
);


ALTER TABLE image OWNER TO monkey;

--
-- TOC entry 207 (class 1259 OID 164676)
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
-- TOC entry 208 (class 1259 OID 164678)
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
    uuid uuid,
    auto_generated boolean,
    revision integer NOT NULL
);


ALTER TABLE image_annotation OWNER TO monkey;

--
-- TOC entry 209 (class 1259 OID 164686)
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
-- TOC entry 210 (class 1259 OID 164693)
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
-- TOC entry 211 (class 1259 OID 164695)
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
-- TOC entry 212 (class 1259 OID 164703)
-- Name: image_annotation_revision_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_annotation_revision_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_annotation_revision_id_seq OWNER TO monkey;

--
-- TOC entry 213 (class 1259 OID 164705)
-- Name: image_annotation_revision; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_annotation_revision (
    id bigint DEFAULT nextval('image_annotation_revision_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    revision integer
);


ALTER TABLE image_annotation_revision OWNER TO monkey;

--
-- TOC entry 214 (class 1259 OID 164709)
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
-- TOC entry 215 (class 1259 OID 164711)
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
-- TOC entry 2612 (class 0 OID 0)
-- Dependencies: 215
-- Name: image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: monkey
--

ALTER SEQUENCE image_id_seq OWNED BY image.id;


--
-- TOC entry 216 (class 1259 OID 164713)
-- Name: image_label_suggestion_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_label_suggestion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_label_suggestion_id_seq OWNER TO monkey;

--
-- TOC entry 217 (class 1259 OID 164715)
-- Name: image_label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_label_suggestion (
    id bigint DEFAULT nextval('image_label_suggestion_id_seq'::regclass) NOT NULL,
    label_suggestion_id bigint,
    image_id bigint,
    fingerprint_of_last_modification text,
    annotatable boolean NOT NULL
);


ALTER TABLE image_label_suggestion OWNER TO monkey;

--
-- TOC entry 218 (class 1259 OID 164722)
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
-- TOC entry 219 (class 1259 OID 164724)
-- Name: image_provider; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_provider (
    id bigint DEFAULT nextval('image_provider_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE image_provider OWNER TO monkey;

--
-- TOC entry 220 (class 1259 OID 164731)
-- Name: image_quarantine_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_quarantine_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_quarantine_id_seq OWNER TO monkey;

--
-- TOC entry 221 (class 1259 OID 164733)
-- Name: image_quarantine; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_quarantine (
    id bigint DEFAULT nextval('image_quarantine_id_seq'::regclass) NOT NULL,
    image_id bigint
);


ALTER TABLE image_quarantine OWNER TO monkey;

--
-- TOC entry 222 (class 1259 OID 164737)
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
-- TOC entry 223 (class 1259 OID 164739)
-- Name: image_report; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_report (
    id bigint DEFAULT nextval('report_id_seq'::regclass) NOT NULL,
    reason text,
    image_id bigint
);


ALTER TABLE image_report OWNER TO monkey;

--
-- TOC entry 224 (class 1259 OID 164746)
-- Name: image_source_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_source_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_source_id_seq OWNER TO monkey;

--
-- TOC entry 225 (class 1259 OID 164748)
-- Name: image_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_source (
    id bigint DEFAULT nextval('image_source_id_seq'::regclass) NOT NULL,
    url text,
    image_id bigint
);


ALTER TABLE image_source OWNER TO monkey;

--
-- TOC entry 226 (class 1259 OID 164755)
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
-- TOC entry 227 (class 1259 OID 164757)
-- Name: image_validation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation (
    id bigint DEFAULT nextval('image_validation_id_seq'::regclass) NOT NULL,
    image_id bigint,
    label_id bigint,
    num_of_valid integer,
    num_of_invalid integer,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    fingerprint_of_last_modification text,
    uuid uuid NOT NULL,
    num_of_not_annotatable integer NOT NULL
);


ALTER TABLE image_validation OWNER TO monkey;

--
-- TOC entry 228 (class 1259 OID 164765)
-- Name: image_validation_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation_history (
    id bigint NOT NULL,
    image_id bigint,
    label_id bigint,
    num_of_valid integer,
    num_of_invalid integer,
    sys_period tstzrange NOT NULL,
    fingerprint_of_last_modification text,
    uuid uuid,
    num_of_not_annotatable integer
);


ALTER TABLE image_validation_history OWNER TO monkey;

--
-- TOC entry 229 (class 1259 OID 164771)
-- Name: image_validation_source_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE image_validation_source_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_validation_source_id_seq OWNER TO monkey;

--
-- TOC entry 230 (class 1259 OID 164773)
-- Name: image_validation_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation_source (
    id bigint DEFAULT nextval('image_validation_source_id_seq'::regclass) NOT NULL,
    image_validation_id bigint,
    image_source_id bigint
);


ALTER TABLE image_validation_source OWNER TO monkey;

--
-- TOC entry 231 (class 1259 OID 164777)
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
-- TOC entry 232 (class 1259 OID 164779)
-- Name: label; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label (
    id bigint DEFAULT nextval('name_id_seq'::regclass) NOT NULL,
    name text,
    parent_id bigint,
    uuid uuid NOT NULL
);


ALTER TABLE label OWNER TO monkey;

--
-- TOC entry 233 (class 1259 OID 164786)
-- Name: label_accessor_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE label_accessor_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE label_accessor_id_seq OWNER TO monkey;

--
-- TOC entry 234 (class 1259 OID 164788)
-- Name: label_accessor; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_accessor (
    id bigint DEFAULT nextval('label_accessor_id_seq'::regclass) NOT NULL,
    label_id bigint,
    accessor text
);


ALTER TABLE label_accessor OWNER TO monkey;

--
-- TOC entry 235 (class 1259 OID 164795)
-- Name: label_example_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE label_example_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE label_example_id_seq OWNER TO monkey;

--
-- TOC entry 236 (class 1259 OID 164797)
-- Name: label_example; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_example (
    id bigint DEFAULT nextval('label_example_id_seq'::regclass),
    attribution text,
    label_id bigint,
    filename text
);


ALTER TABLE label_example OWNER TO monkey;

--
-- TOC entry 237 (class 1259 OID 164804)
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
-- TOC entry 238 (class 1259 OID 164806)
-- Name: label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_suggestion (
    id bigint DEFAULT nextval('label_suggestion_id_seq'::regclass) NOT NULL,
    name text,
    proposed_by bigint
);


ALTER TABLE label_suggestion OWNER TO monkey;

--
-- TOC entry 239 (class 1259 OID 164813)
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
-- TOC entry 240 (class 1259 OID 164815)
-- Name: quiz_answer; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE quiz_answer (
    id bigint DEFAULT nextval('quiz_answer_id_seq'::regclass) NOT NULL,
    quiz_question_id bigint,
    label_id bigint
);


ALTER TABLE quiz_answer OWNER TO monkey;

--
-- TOC entry 241 (class 1259 OID 164819)
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
-- TOC entry 242 (class 1259 OID 164821)
-- Name: quiz_question; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE quiz_question (
    id bigint DEFAULT nextval('quiz_question_id_seq'::regclass) NOT NULL,
    question text,
    refines_label_id bigint,
    recommended_control control_type,
    allow_unknown boolean,
    allow_other boolean,
    browse_by_example boolean,
    multiselect boolean,
    uuid uuid NOT NULL
);


ALTER TABLE quiz_question OWNER TO monkey;

--
-- TOC entry 243 (class 1259 OID 164828)
-- Name: trending_label_suggestion_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE trending_label_suggestion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE trending_label_suggestion_id_seq OWNER TO monkey;

--
-- TOC entry 244 (class 1259 OID 164830)
-- Name: trending_label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE trending_label_suggestion (
    id bigint DEFAULT nextval('trending_label_suggestion_id_seq'::regclass) NOT NULL,
    label_suggestion_id bigint,
    num_of_last_sent integer,
    github_issue_id bigint NOT NULL,
    closed boolean NOT NULL
);


ALTER TABLE trending_label_suggestion OWNER TO monkey;

--
-- TOC entry 245 (class 1259 OID 164834)
-- Name: user_annotation_blacklist_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE user_annotation_blacklist_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_annotation_blacklist_id_seq OWNER TO monkey;

--
-- TOC entry 246 (class 1259 OID 164836)
-- Name: user_annotation_blacklist; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_annotation_blacklist (
    id bigint DEFAULT nextval('user_annotation_blacklist_id_seq'::regclass) NOT NULL,
    account_id bigint,
    image_validation_id bigint
);


ALTER TABLE user_annotation_blacklist OWNER TO monkey;

--
-- TOC entry 247 (class 1259 OID 164840)
-- Name: user_image_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE user_image_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_image_id_seq OWNER TO monkey;

--
-- TOC entry 248 (class 1259 OID 164842)
-- Name: user_image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_image (
    id bigint DEFAULT nextval('user_image_id_seq'::regclass) NOT NULL,
    image_id bigint,
    account_id bigint
);


ALTER TABLE user_image OWNER TO monkey;

--
-- TOC entry 249 (class 1259 OID 164846)
-- Name: user_image_annotation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE user_image_annotation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_image_annotation_id_seq OWNER TO monkey;

--
-- TOC entry 250 (class 1259 OID 164848)
-- Name: user_image_annotation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_image_annotation (
    id bigint DEFAULT nextval('user_image_annotation_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    account_id bigint,
    "timestamp" timestamp with time zone
);


ALTER TABLE user_image_annotation OWNER TO monkey;

--
-- TOC entry 251 (class 1259 OID 164852)
-- Name: user_image_validation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE user_image_validation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_image_validation_id_seq OWNER TO monkey;

--
-- TOC entry 252 (class 1259 OID 164854)
-- Name: user_image_validation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_image_validation (
    id bigint DEFAULT nextval('user_image_validation_id_seq'::regclass) NOT NULL,
    image_validation_id bigint,
    account_id bigint,
    "timestamp" timestamp with time zone
);


ALTER TABLE user_image_validation OWNER TO monkey;

--
-- TOC entry 253 (class 1259 OID 164858)
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
-- TOC entry 254 (class 1259 OID 164860)
-- Name: validations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_app (
    id bigint DEFAULT nextval('validations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE validations_per_app OWNER TO monkey;

--
-- TOC entry 255 (class 1259 OID 164867)
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
-- TOC entry 256 (class 1259 OID 164869)
-- Name: validations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_country (
    id bigint DEFAULT nextval('validations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE validations_per_country OWNER TO monkey;

--
-- TOC entry 2260 (class 2604 OID 164876)
-- Name: image id; Type: DEFAULT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image ALTER COLUMN id SET DEFAULT nextval('image_id_seq'::regclass);


--
-- TOC entry 2289 (class 2606 OID 164878)
-- Name: access_token access_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY access_token
    ADD CONSTRAINT access_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2298 (class 2606 OID 164880)
-- Name: account_permission account_permission_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account_permission
    ADD CONSTRAINT account_permission_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2306 (class 2606 OID 164882)
-- Name: annotation_type annotation_type_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2308 (class 2606 OID 164884)
-- Name: annotation_type annotation_type_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_name_unique UNIQUE (name);


--
-- TOC entry 2310 (class 2606 OID 164886)
-- Name: annotations_per_app annotations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2312 (class 2606 OID 164888)
-- Name: annotations_per_app annotations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2314 (class 2606 OID 164890)
-- Name: annotations_per_country annotations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2316 (class 2606 OID 164892)
-- Name: annotations_per_country annotations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2318 (class 2606 OID 164894)
-- Name: api_token api_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2320 (class 2606 OID 164896)
-- Name: api_token api_token_token_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_token_unique UNIQUE (token);


--
-- TOC entry 2323 (class 2606 OID 164898)
-- Name: donations_per_app donations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2325 (class 2606 OID 164900)
-- Name: donations_per_app donations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2327 (class 2606 OID 164902)
-- Name: donations_per_country donations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2329 (class 2606 OID 164904)
-- Name: donations_per_country donations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2304 (class 2606 OID 164906)
-- Name: annotation_data image_annotation_data_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2343 (class 2606 OID 164908)
-- Name: image_annotation image_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2346 (class 2606 OID 164910)
-- Name: image_annotation image_annotation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_label_uniquekey UNIQUE (image_id, label_id, auto_generated);


--
-- TOC entry 2351 (class 2606 OID 164912)
-- Name: image_annotation_refinement image_annotation_refinement_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2353 (class 2606 OID 164914)
-- Name: image_annotation_refinement image_annotation_refinement_label_annotation_data_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_annotation_data_unique UNIQUE (annotation_data_id, label_id);


--
-- TOC entry 2356 (class 2606 OID 164916)
-- Name: image_annotation_revision image_annotation_revision_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2333 (class 2606 OID 164918)
-- Name: image image_hash_unique_key; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_hash_unique_key UNIQUE (hash);


--
-- TOC entry 2335 (class 2606 OID 164920)
-- Name: image image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2339 (class 2606 OID 164922)
-- Name: image image_key_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_key_unique UNIQUE (image_provider_id, key);


--
-- TOC entry 2360 (class 2606 OID 164924)
-- Name: image_label_suggestion image_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2362 (class 2606 OID 164926)
-- Name: image_label_suggestion image_label_suggestion_image_id_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_label_suggestion_id_unique UNIQUE (label_suggestion_id, image_id);


--
-- TOC entry 2364 (class 2606 OID 164928)
-- Name: image_provider image_provider_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_provider
    ADD CONSTRAINT image_provider_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2367 (class 2606 OID 164930)
-- Name: image_quarantine image_quarantine_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2369 (class 2606 OID 164932)
-- Name: image_quarantine image_quarantine_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_unique UNIQUE (image_id);


--
-- TOC entry 2375 (class 2606 OID 164934)
-- Name: image_source image_source_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_source
    ADD CONSTRAINT image_source_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2379 (class 2606 OID 164936)
-- Name: image_validation image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2382 (class 2606 OID 164938)
-- Name: image_validation image_validation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_label_uniquekey UNIQUE (image_id, label_id);


--
-- TOC entry 2387 (class 2606 OID 164940)
-- Name: image_validation_source image_validation_source_id; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_id PRIMARY KEY (id);


--
-- TOC entry 2397 (class 2606 OID 164942)
-- Name: label_accessor label_accessor_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2399 (class 2606 OID 164944)
-- Name: label_accessor label_accessor_label_id_accessor_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_label_id_accessor_unique UNIQUE (label_id, accessor);


--
-- TOC entry 2390 (class 2606 OID 164946)
-- Name: label label_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2392 (class 2606 OID 164948)
-- Name: label label_name_parent_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_name_parent_id_unique UNIQUE (name, parent_id);


--
-- TOC entry 2403 (class 2606 OID 164950)
-- Name: label_suggestion label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2405 (class 2606 OID 164952)
-- Name: label_suggestion label_suggestion_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_name_unique UNIQUE (name);


--
-- TOC entry 2394 (class 2606 OID 164954)
-- Name: label label_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2409 (class 2606 OID 164956)
-- Name: quiz_answer quiz_id_pley; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_id_pley PRIMARY KEY (id);


--
-- TOC entry 2412 (class 2606 OID 164958)
-- Name: quiz_question quiz_question_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2414 (class 2606 OID 164960)
-- Name: quiz_question quiz_question_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2372 (class 2606 OID 164962)
-- Name: image_report report_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT report_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2417 (class 2606 OID 164964)
-- Name: trending_label_suggestion trending_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2419 (class 2606 OID 164966)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_unique UNIQUE (label_suggestion_id);


--
-- TOC entry 2423 (class 2606 OID 164968)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_image_validation_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_image_validation_id_unique UNIQUE (account_id, image_validation_id);


--
-- TOC entry 2433 (class 2606 OID 164970)
-- Name: user_image_annotation user_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2292 (class 2606 OID 164972)
-- Name: account user_email_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_email_unique UNIQUE (email);


--
-- TOC entry 2294 (class 2606 OID 164974)
-- Name: account user_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2425 (class 2606 OID 164976)
-- Name: user_annotation_blacklist user_image_blacklist_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_image_blacklist_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2429 (class 2606 OID 164978)
-- Name: user_image user_image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2437 (class 2606 OID 164980)
-- Name: user_image_validation user_image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2296 (class 2606 OID 164982)
-- Name: account user_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_name_unique UNIQUE (name);


--
-- TOC entry 2439 (class 2606 OID 164984)
-- Name: validations_per_app validations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2441 (class 2606 OID 164986)
-- Name: validations_per_app validations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2443 (class 2606 OID 164988)
-- Name: validations_per_country validations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2445 (class 2606 OID 164990)
-- Name: validations_per_country validations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2290 (class 1259 OID 164991)
-- Name: fki_access_token_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_access_token_user_id_fkey ON access_token USING btree (user_id);


--
-- TOC entry 2299 (class 1259 OID 164992)
-- Name: fki_account_permission_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_account_permission_account_id_fkey ON account_permission USING btree (account_id);


--
-- TOC entry 2300 (class 1259 OID 164993)
-- Name: fki_annotation_data_annotation_type_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_annotation_type_fkey ON annotation_data USING btree (annotation_type_id);


--
-- TOC entry 2301 (class 1259 OID 164994)
-- Name: fki_annotation_data_image_annotation_revision_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_image_annotation_revision_id_fkey ON annotation_data USING btree (image_annotation_revision_id);


--
-- TOC entry 2321 (class 1259 OID 164995)
-- Name: fki_api_token_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_api_token_account_id_fkey ON api_token USING btree (account_id);


--
-- TOC entry 2302 (class 1259 OID 164996)
-- Name: fki_image_annotation_data_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_data_annotation_id_fkey ON annotation_data USING btree (image_annotation_id);


--
-- TOC entry 2341 (class 1259 OID 164997)
-- Name: fki_image_annotation_label_id_key; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_label_id_key ON image_annotation USING btree (label_id);


--
-- TOC entry 2354 (class 1259 OID 164998)
-- Name: fki_image_annotation_revision_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_revision_image_annotation_id_fkey ON image_annotation_revision USING btree (image_annotation_id);


--
-- TOC entry 2357 (class 1259 OID 164999)
-- Name: fki_image_label_suggestion_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_image_id_fkey ON image_label_suggestion USING btree (image_id);


--
-- TOC entry 2358 (class 1259 OID 165000)
-- Name: fki_image_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_label_suggestion_id_fkey ON image_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 2330 (class 1259 OID 165001)
-- Name: fki_image_provider_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_provider_id_fkey ON image USING btree (image_provider_id);


--
-- TOC entry 2365 (class 1259 OID 165002)
-- Name: fki_image_quarantine_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quarantine_image_id_fkey ON image_quarantine USING btree (image_id);


--
-- TOC entry 2348 (class 1259 OID 165003)
-- Name: fki_image_quiz_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_image_annotation_id_fkey ON image_annotation_refinement USING btree (annotation_data_id);


--
-- TOC entry 2349 (class 1259 OID 165004)
-- Name: fki_image_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_label_id_fkey ON image_annotation_refinement USING btree (label_id);


--
-- TOC entry 2370 (class 1259 OID 165005)
-- Name: fki_image_report_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_report_image_id_fkey ON image_report USING btree (image_id);


--
-- TOC entry 2373 (class 1259 OID 165006)
-- Name: fki_image_source_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_source_image_id_fkey ON image_source USING btree (image_id);


--
-- TOC entry 2376 (class 1259 OID 165007)
-- Name: fki_image_validation_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_image_id_fkey ON image_validation USING btree (image_id);


--
-- TOC entry 2377 (class 1259 OID 165008)
-- Name: fki_image_validation_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_label_id_fkey ON image_validation USING btree (label_id);


--
-- TOC entry 2384 (class 1259 OID 165009)
-- Name: fki_image_validation_source_image_source_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_source_id_fkey ON image_validation_source USING btree (image_source_id);


--
-- TOC entry 2385 (class 1259 OID 165010)
-- Name: fki_image_validation_source_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_validation_id_fkey ON image_validation_source USING btree (image_validation_id);


--
-- TOC entry 2395 (class 1259 OID 165011)
-- Name: fki_label_accessor_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_accessor_label_id_fkey ON label_accessor USING btree (label_id);


--
-- TOC entry 2400 (class 1259 OID 165012)
-- Name: fki_label_example_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_example_label_id_fkey ON label_example USING btree (label_id);


--
-- TOC entry 2388 (class 1259 OID 165013)
-- Name: fki_label_parent_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_parent_id_fkey ON label USING btree (parent_id);


--
-- TOC entry 2401 (class 1259 OID 165014)
-- Name: fki_label_suggestion_proposed_by_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_suggestion_proposed_by_fkey ON label_suggestion USING btree (proposed_by);


--
-- TOC entry 2406 (class 1259 OID 165015)
-- Name: fki_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_label_id_fkey ON quiz_answer USING btree (label_id);


--
-- TOC entry 2410 (class 1259 OID 165016)
-- Name: fki_quiz_question_refines_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_question_refines_label_id_fkey ON quiz_question USING btree (refines_label_id);


--
-- TOC entry 2407 (class 1259 OID 165017)
-- Name: fki_quiz_quiz_question_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_quiz_question_id_fkey ON quiz_answer USING btree (quiz_question_id);


--
-- TOC entry 2415 (class 1259 OID 165018)
-- Name: fki_trending_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_trending_label_suggestion_label_suggestion_id_fkey ON trending_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 2420 (class 1259 OID 165019)
-- Name: fki_user_annotation_blacklist_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_account_id_fkey ON user_annotation_blacklist USING btree (account_id);


--
-- TOC entry 2421 (class 1259 OID 165020)
-- Name: fki_user_annotation_blacklist_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_image_validation_id_fkey ON user_annotation_blacklist USING btree (image_validation_id);


--
-- TOC entry 2426 (class 1259 OID 165021)
-- Name: fki_user_image_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_account_id_fkey ON user_image USING btree (account_id);


--
-- TOC entry 2430 (class 1259 OID 165022)
-- Name: fki_user_image_annotation_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_image_annotation_id_fkey ON user_image_annotation USING btree (image_annotation_id);


--
-- TOC entry 2431 (class 1259 OID 165023)
-- Name: fki_user_image_annotation_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_user_id_fkey ON user_image_annotation USING btree (account_id);


--
-- TOC entry 2427 (class 1259 OID 165024)
-- Name: fki_user_image_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_image_id_fkey ON user_image USING btree (image_id);


--
-- TOC entry 2434 (class 1259 OID 165025)
-- Name: fki_user_image_validation_acccount_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_acccount_id_fkey ON user_image_validation USING btree (account_id);


--
-- TOC entry 2435 (class 1259 OID 165026)
-- Name: fki_user_image_validation_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_account_id_fkey ON user_image_validation USING btree (image_validation_id);


--
-- TOC entry 2344 (class 1259 OID 165027)
-- Name: image_annotation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_image_id_index ON image_annotation USING btree (image_id);


--
-- TOC entry 2347 (class 1259 OID 165028)
-- Name: image_annotation_uuid_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_uuid_index ON image_annotation USING btree (uuid);


--
-- TOC entry 2331 (class 1259 OID 165029)
-- Name: image_hash_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_hash_index ON image USING btree (hash);


--
-- TOC entry 2336 (class 1259 OID 165030)
-- Name: image_image_provider_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_image_provider_index ON image USING btree (image_provider_id);


--
-- TOC entry 2337 (class 1259 OID 165031)
-- Name: image_key_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_key_index ON image USING btree (key);


--
-- TOC entry 2340 (class 1259 OID 165032)
-- Name: image_unlocked_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_unlocked_index ON image USING btree (unlocked);


--
-- TOC entry 2380 (class 1259 OID 165033)
-- Name: image_validation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_image_id_index ON image_validation USING btree (image_id);


--
-- TOC entry 2383 (class 1259 OID 165034)
-- Name: image_validation_label_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_label_id_index ON image_validation USING btree (label_id);


--
-- TOC entry 2483 (class 2620 OID 165221)
-- Name: image_annotation image_annotation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_annotation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON image_annotation FOR EACH ROW EXECUTE PROCEDURE versioning('sys_period', 'image_annotation_history', 'true');


--
-- TOC entry 2484 (class 2620 OID 165220)
-- Name: image_validation image_validation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_validation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON image_validation FOR EACH ROW EXECUTE PROCEDURE versioning('sys_period', 'image_validation_history', 'true');


--
-- TOC entry 2446 (class 2606 OID 165035)
-- Name: access_token access_token_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY access_token
    ADD CONSTRAINT access_token_user_id_fkey FOREIGN KEY (user_id) REFERENCES account(id);


--
-- TOC entry 2447 (class 2606 OID 165040)
-- Name: account_permission account_permission_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account_permission
    ADD CONSTRAINT account_permission_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2448 (class 2606 OID 165045)
-- Name: annotation_data annotation_data_annotation_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_annotation_type_fkey FOREIGN KEY (annotation_type_id) REFERENCES annotation_type(id);


--
-- TOC entry 2449 (class 2606 OID 165050)
-- Name: annotation_data annotation_data_image_annotation_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_image_annotation_revision_id_fkey FOREIGN KEY (image_annotation_revision_id) REFERENCES image_annotation_revision(id);


--
-- TOC entry 2451 (class 2606 OID 165055)
-- Name: api_token api_token_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2450 (class 2606 OID 165060)
-- Name: annotation_data image_annotation_data_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2453 (class 2606 OID 165065)
-- Name: image_annotation image_annotation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2454 (class 2606 OID 165070)
-- Name: image_annotation image_annotation_label_id_key; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_label_id_key FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2455 (class 2606 OID 165075)
-- Name: image_annotation_refinement image_annotation_refinement_annotation_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_annotation_data_id_fkey FOREIGN KEY (annotation_data_id) REFERENCES annotation_data(id);


--
-- TOC entry 2456 (class 2606 OID 165080)
-- Name: image_annotation_refinement image_annotation_refinement_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2457 (class 2606 OID 165085)
-- Name: image_annotation_revision image_annotation_revision_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2452 (class 2606 OID 165090)
-- Name: image image_image_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_image_provider_id_fkey FOREIGN KEY (image_provider_id) REFERENCES image_provider(id);


--
-- TOC entry 2458 (class 2606 OID 165095)
-- Name: image_label_suggestion image_label_suggestion_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2459 (class 2606 OID 165100)
-- Name: image_label_suggestion image_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES label_suggestion(id);


--
-- TOC entry 2460 (class 2606 OID 165105)
-- Name: image_quarantine image_quarantine_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2461 (class 2606 OID 165110)
-- Name: image_report image_report_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT image_report_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2462 (class 2606 OID 165115)
-- Name: image_source image_source_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_source
    ADD CONSTRAINT image_source_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2463 (class 2606 OID 165120)
-- Name: image_validation image_validation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2464 (class 2606 OID 165125)
-- Name: image_validation image_validation_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2465 (class 2606 OID 165130)
-- Name: image_validation_source image_validation_source_image_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_image_source_id_fkey FOREIGN KEY (image_source_id) REFERENCES image_source(id);


--
-- TOC entry 2466 (class 2606 OID 165135)
-- Name: image_validation_source image_validation_source_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2468 (class 2606 OID 165140)
-- Name: label_accessor label_accessor_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2469 (class 2606 OID 165145)
-- Name: label_example label_example_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_example
    ADD CONSTRAINT label_example_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2467 (class 2606 OID 165150)
-- Name: label label_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES label(id);


--
-- TOC entry 2470 (class 2606 OID 165155)
-- Name: label_suggestion label_suggestion_proposed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_proposed_by_fkey FOREIGN KEY (proposed_by) REFERENCES account(id);


--
-- TOC entry 2471 (class 2606 OID 165160)
-- Name: quiz_answer quiz_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2473 (class 2606 OID 165165)
-- Name: quiz_question quiz_question_refines_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_refines_label_id_fkey FOREIGN KEY (refines_label_id) REFERENCES label(id);


--
-- TOC entry 2472 (class 2606 OID 165170)
-- Name: quiz_answer quiz_quiz_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_quiz_question_id_fkey FOREIGN KEY (quiz_question_id) REFERENCES quiz_question(id);


--
-- TOC entry 2474 (class 2606 OID 165175)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES label_suggestion(id);


--
-- TOC entry 2475 (class 2606 OID 165180)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2476 (class 2606 OID 165185)
-- Name: user_annotation_blacklist user_annotation_blacklist_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2477 (class 2606 OID 165190)
-- Name: user_image user_image_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2479 (class 2606 OID 165195)
-- Name: user_image_annotation user_image_annotation_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_image_annotation_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2480 (class 2606 OID 165200)
-- Name: user_image_annotation user_image_annotation_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_image_annotation_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2478 (class 2606 OID 165205)
-- Name: user_image user_image_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2481 (class 2606 OID 165210)
-- Name: user_image_validation user_image_validation_acccount_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_acccount_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2482 (class 2606 OID 165215)
-- Name: user_image_validation user_image_validation_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2608 (class 0 OID 0)
-- Dependencies: 5
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO monkey;


-- Completed on 2018-07-08 21:22:41

--
-- PostgreSQL database dump complete
--

