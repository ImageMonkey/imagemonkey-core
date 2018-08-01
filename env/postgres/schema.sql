--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.5

-- Started on 2018-08-01 22:10:37

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
-- TOC entry 2648 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 3 (class 3079 OID 1232685)
-- Name: temporal_tables; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS temporal_tables WITH SCHEMA public;


--
-- TOC entry 2649 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION temporal_tables; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION temporal_tables IS 'temporal tables';


--
-- TOC entry 2 (class 3079 OID 1232688)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 2650 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- TOC entry 566 (class 1247 OID 1232700)
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
-- TOC entry 821 (class 1247 OID 1233666)
-- Name: label_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE label_type AS ENUM (
    'normal',
    'refinement',
    'refinement_category'
);


ALTER TYPE label_type OWNER TO postgres;

--
-- TOC entry 273 (class 1255 OID 1232709)
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
-- TOC entry 187 (class 1259 OID 1232710)
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
-- TOC entry 188 (class 1259 OID 1232712)
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
-- TOC entry 189 (class 1259 OID 1232719)
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
-- TOC entry 190 (class 1259 OID 1232721)
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
-- TOC entry 191 (class 1259 OID 1232728)
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
-- TOC entry 192 (class 1259 OID 1232730)
-- Name: account_permission; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE account_permission (
    id bigint DEFAULT nextval('account_permission_id_seq'::regclass) NOT NULL,
    can_remove_label boolean,
    account_id bigint
);


ALTER TABLE account_permission OWNER TO monkey;

--
-- TOC entry 193 (class 1259 OID 1232734)
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
-- TOC entry 194 (class 1259 OID 1232736)
-- Name: annotation_data; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_data (
    id bigint DEFAULT nextval('image_annotation_data_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    annotation jsonb,
    annotation_type_id bigint NOT NULL,
    image_annotation_revision_id bigint,
    uuid uuid NOT NULL
);


ALTER TABLE annotation_data OWNER TO monkey;

--
-- TOC entry 195 (class 1259 OID 1232743)
-- Name: annotation_type; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_type (
    id bigint NOT NULL,
    name text
);


ALTER TABLE annotation_type OWNER TO monkey;

--
-- TOC entry 196 (class 1259 OID 1232749)
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
-- TOC entry 197 (class 1259 OID 1232751)
-- Name: annotations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_app (
    id bigint DEFAULT nextval('annotations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE annotations_per_app OWNER TO monkey;

--
-- TOC entry 198 (class 1259 OID 1232758)
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
-- TOC entry 199 (class 1259 OID 1232760)
-- Name: annotations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_country (
    id bigint DEFAULT nextval('annotations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE annotations_per_country OWNER TO monkey;

--
-- TOC entry 200 (class 1259 OID 1232767)
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
-- TOC entry 201 (class 1259 OID 1232769)
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
-- TOC entry 202 (class 1259 OID 1232776)
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
-- TOC entry 203 (class 1259 OID 1232778)
-- Name: donations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_app (
    id bigint DEFAULT nextval('donations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE donations_per_app OWNER TO monkey;

--
-- TOC entry 204 (class 1259 OID 1232785)
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
-- TOC entry 205 (class 1259 OID 1232787)
-- Name: donations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_country (
    id bigint DEFAULT nextval('donations_per_country_id_seq'::regclass) NOT NULL,
    country_code text,
    count bigint
);


ALTER TABLE donations_per_country OWNER TO monkey;

--
-- TOC entry 206 (class 1259 OID 1232794)
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
-- TOC entry 207 (class 1259 OID 1232800)
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
-- TOC entry 208 (class 1259 OID 1232802)
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
-- TOC entry 209 (class 1259 OID 1232810)
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
-- TOC entry 210 (class 1259 OID 1232817)
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
-- TOC entry 211 (class 1259 OID 1232819)
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
-- TOC entry 212 (class 1259 OID 1232827)
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
-- TOC entry 213 (class 1259 OID 1232829)
-- Name: image_annotation_revision; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_annotation_revision (
    id bigint DEFAULT nextval('image_annotation_revision_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    revision integer
);


ALTER TABLE image_annotation_revision OWNER TO monkey;

--
-- TOC entry 214 (class 1259 OID 1232833)
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
-- TOC entry 215 (class 1259 OID 1232835)
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
-- TOC entry 2651 (class 0 OID 0)
-- Dependencies: 215
-- Name: image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: monkey
--

ALTER SEQUENCE image_id_seq OWNED BY image.id;


--
-- TOC entry 216 (class 1259 OID 1232837)
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
-- TOC entry 217 (class 1259 OID 1232839)
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
-- TOC entry 218 (class 1259 OID 1232846)
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
-- TOC entry 219 (class 1259 OID 1232848)
-- Name: image_provider; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_provider (
    id bigint DEFAULT nextval('image_provider_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE image_provider OWNER TO monkey;

--
-- TOC entry 220 (class 1259 OID 1232855)
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
-- TOC entry 221 (class 1259 OID 1232857)
-- Name: image_quarantine; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_quarantine (
    id bigint DEFAULT nextval('image_quarantine_id_seq'::regclass) NOT NULL,
    image_id bigint
);


ALTER TABLE image_quarantine OWNER TO monkey;

--
-- TOC entry 222 (class 1259 OID 1232861)
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
-- TOC entry 223 (class 1259 OID 1232863)
-- Name: image_report; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_report (
    id bigint DEFAULT nextval('report_id_seq'::regclass) NOT NULL,
    reason text,
    image_id bigint
);


ALTER TABLE image_report OWNER TO monkey;

--
-- TOC entry 224 (class 1259 OID 1232870)
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
-- TOC entry 225 (class 1259 OID 1232872)
-- Name: image_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_source (
    id bigint DEFAULT nextval('image_source_id_seq'::regclass) NOT NULL,
    url text,
    image_id bigint
);


ALTER TABLE image_source OWNER TO monkey;

--
-- TOC entry 226 (class 1259 OID 1232879)
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
-- TOC entry 227 (class 1259 OID 1232881)
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
-- TOC entry 228 (class 1259 OID 1232889)
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
-- TOC entry 229 (class 1259 OID 1232895)
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
-- TOC entry 230 (class 1259 OID 1232897)
-- Name: image_validation_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation_source (
    id bigint DEFAULT nextval('image_validation_source_id_seq'::regclass) NOT NULL,
    image_validation_id bigint,
    image_source_id bigint
);


ALTER TABLE image_validation_source OWNER TO monkey;

--
-- TOC entry 231 (class 1259 OID 1232901)
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
-- TOC entry 232 (class 1259 OID 1232903)
-- Name: label; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label (
    id bigint DEFAULT nextval('name_id_seq'::regclass) NOT NULL,
    name text,
    parent_id bigint,
    uuid uuid NOT NULL,
    label_type label_type
);


ALTER TABLE label OWNER TO monkey;

--
-- TOC entry 233 (class 1259 OID 1232910)
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
-- TOC entry 234 (class 1259 OID 1232912)
-- Name: label_accessor; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_accessor (
    id bigint DEFAULT nextval('label_accessor_id_seq'::regclass) NOT NULL,
    label_id bigint,
    accessor text
);


ALTER TABLE label_accessor OWNER TO monkey;

--
-- TOC entry 235 (class 1259 OID 1232919)
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
-- TOC entry 236 (class 1259 OID 1232921)
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
-- TOC entry 258 (class 1259 OID 1233456)
-- Name: label_refinement_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE label_refinement_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE label_refinement_id_seq OWNER TO monkey;

--
-- TOC entry 257 (class 1259 OID 1233446)
-- Name: label_refinement; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_refinement (
    id bigint DEFAULT nextval('label_refinement_id_seq'::regclass) NOT NULL,
    label_id bigint,
    refinement_type_id bigint
);


ALTER TABLE label_refinement OWNER TO monkey;

--
-- TOC entry 237 (class 1259 OID 1232928)
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
-- TOC entry 238 (class 1259 OID 1232930)
-- Name: label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_suggestion (
    id bigint DEFAULT nextval('label_suggestion_id_seq'::regclass) NOT NULL,
    name text,
    proposed_by bigint
);


ALTER TABLE label_suggestion OWNER TO monkey;

--
-- TOC entry 239 (class 1259 OID 1232937)
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
-- TOC entry 240 (class 1259 OID 1232939)
-- Name: quiz_answer; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE quiz_answer (
    id bigint DEFAULT nextval('quiz_answer_id_seq'::regclass) NOT NULL,
    quiz_question_id bigint,
    label_id bigint
);


ALTER TABLE quiz_answer OWNER TO monkey;

--
-- TOC entry 241 (class 1259 OID 1232943)
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
-- TOC entry 242 (class 1259 OID 1232945)
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
-- TOC entry 260 (class 1259 OID 1233475)
-- Name: refinement_type_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE refinement_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE refinement_type_id_seq OWNER TO monkey;

--
-- TOC entry 259 (class 1259 OID 1233465)
-- Name: refinement_type; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE refinement_type (
    id bigint DEFAULT nextval('refinement_type_id_seq'::regclass) NOT NULL,
    name text,
    uuid uuid
);


ALTER TABLE refinement_type OWNER TO monkey;

--
-- TOC entry 243 (class 1259 OID 1232952)
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
-- TOC entry 244 (class 1259 OID 1232954)
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
-- TOC entry 245 (class 1259 OID 1232958)
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
-- TOC entry 246 (class 1259 OID 1232960)
-- Name: user_annotation_blacklist; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_annotation_blacklist (
    id bigint DEFAULT nextval('user_annotation_blacklist_id_seq'::regclass) NOT NULL,
    account_id bigint,
    image_validation_id bigint
);


ALTER TABLE user_annotation_blacklist OWNER TO monkey;

--
-- TOC entry 247 (class 1259 OID 1232964)
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
-- TOC entry 248 (class 1259 OID 1232966)
-- Name: user_image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_image (
    id bigint DEFAULT nextval('user_image_id_seq'::regclass) NOT NULL,
    image_id bigint,
    account_id bigint
);


ALTER TABLE user_image OWNER TO monkey;

--
-- TOC entry 249 (class 1259 OID 1232970)
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
-- TOC entry 250 (class 1259 OID 1232972)
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
-- TOC entry 251 (class 1259 OID 1232976)
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
-- TOC entry 252 (class 1259 OID 1232978)
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
-- TOC entry 253 (class 1259 OID 1232982)
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
-- TOC entry 254 (class 1259 OID 1232984)
-- Name: validations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_app (
    id bigint DEFAULT nextval('validations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE validations_per_app OWNER TO monkey;

--
-- TOC entry 255 (class 1259 OID 1232991)
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
-- TOC entry 256 (class 1259 OID 1232993)
-- Name: validations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_country (
    id bigint DEFAULT nextval('validations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE validations_per_country OWNER TO monkey;

--
-- TOC entry 2277 (class 2604 OID 1233000)
-- Name: image id; Type: DEFAULT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image ALTER COLUMN id SET DEFAULT nextval('image_id_seq'::regclass);


--
-- TOC entry 2308 (class 2606 OID 1233002)
-- Name: access_token access_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY access_token
    ADD CONSTRAINT access_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2317 (class 2606 OID 1233004)
-- Name: account_permission account_permission_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account_permission
    ADD CONSTRAINT account_permission_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2320 (class 2606 OID 1233759)
-- Name: annotation_data annotation_data_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2327 (class 2606 OID 1233006)
-- Name: annotation_type annotation_type_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2329 (class 2606 OID 1233008)
-- Name: annotation_type annotation_type_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_name_unique UNIQUE (name);


--
-- TOC entry 2331 (class 2606 OID 1233010)
-- Name: annotations_per_app annotations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2333 (class 2606 OID 1233012)
-- Name: annotations_per_app annotations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2335 (class 2606 OID 1233014)
-- Name: annotations_per_country annotations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2337 (class 2606 OID 1233016)
-- Name: annotations_per_country annotations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2339 (class 2606 OID 1233018)
-- Name: api_token api_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2341 (class 2606 OID 1233020)
-- Name: api_token api_token_token_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_token_unique UNIQUE (token);


--
-- TOC entry 2344 (class 2606 OID 1233022)
-- Name: donations_per_app donations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2346 (class 2606 OID 1233024)
-- Name: donations_per_app donations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2348 (class 2606 OID 1233026)
-- Name: donations_per_country donations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2350 (class 2606 OID 1233028)
-- Name: donations_per_country donations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2325 (class 2606 OID 1233030)
-- Name: annotation_data image_annotation_data_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2364 (class 2606 OID 1233032)
-- Name: image_annotation image_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2367 (class 2606 OID 1233034)
-- Name: image_annotation image_annotation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_label_uniquekey UNIQUE (image_id, label_id, auto_generated);


--
-- TOC entry 2372 (class 2606 OID 1233036)
-- Name: image_annotation_refinement image_annotation_refinement_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2374 (class 2606 OID 1233038)
-- Name: image_annotation_refinement image_annotation_refinement_label_annotation_data_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_annotation_data_unique UNIQUE (annotation_data_id, label_id);


--
-- TOC entry 2377 (class 2606 OID 1233040)
-- Name: image_annotation_revision image_annotation_revision_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2354 (class 2606 OID 1233042)
-- Name: image image_hash_unique_key; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_hash_unique_key UNIQUE (hash);


--
-- TOC entry 2356 (class 2606 OID 1233044)
-- Name: image image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2360 (class 2606 OID 1233046)
-- Name: image image_key_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_key_unique UNIQUE (image_provider_id, key);


--
-- TOC entry 2381 (class 2606 OID 1233048)
-- Name: image_label_suggestion image_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2383 (class 2606 OID 1233050)
-- Name: image_label_suggestion image_label_suggestion_image_id_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_label_suggestion_id_unique UNIQUE (label_suggestion_id, image_id);


--
-- TOC entry 2385 (class 2606 OID 1233052)
-- Name: image_provider image_provider_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_provider
    ADD CONSTRAINT image_provider_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2388 (class 2606 OID 1233054)
-- Name: image_quarantine image_quarantine_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2390 (class 2606 OID 1233056)
-- Name: image_quarantine image_quarantine_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_unique UNIQUE (image_id);


--
-- TOC entry 2396 (class 2606 OID 1233058)
-- Name: image_source image_source_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_source
    ADD CONSTRAINT image_source_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2400 (class 2606 OID 1233060)
-- Name: image_validation image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2403 (class 2606 OID 1233062)
-- Name: image_validation image_validation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_label_uniquekey UNIQUE (image_id, label_id);


--
-- TOC entry 2408 (class 2606 OID 1233064)
-- Name: image_validation_source image_validation_source_id; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_id PRIMARY KEY (id);


--
-- TOC entry 2418 (class 2606 OID 1233066)
-- Name: label_accessor label_accessor_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2420 (class 2606 OID 1233068)
-- Name: label_accessor label_accessor_label_id_accessor_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_label_id_accessor_unique UNIQUE (label_id, accessor);


--
-- TOC entry 2411 (class 2606 OID 1233070)
-- Name: label label_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2413 (class 2606 OID 1233072)
-- Name: label label_name_parent_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_name_parent_id_unique UNIQUE (name, parent_id);


--
-- TOC entry 2474 (class 2606 OID 1233450)
-- Name: label_refinement label_refinement_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_refinement
    ADD CONSTRAINT label_refinement_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2476 (class 2606 OID 1233489)
-- Name: label_refinement label_refinement_label_id_refinement_type_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_refinement
    ADD CONSTRAINT label_refinement_label_id_refinement_type_id_unique UNIQUE (refinement_type_id, label_id);


--
-- TOC entry 2424 (class 2606 OID 1233074)
-- Name: label_suggestion label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2426 (class 2606 OID 1233076)
-- Name: label_suggestion label_suggestion_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_name_unique UNIQUE (name);


--
-- TOC entry 2415 (class 2606 OID 1233078)
-- Name: label label_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2430 (class 2606 OID 1233761)
-- Name: quiz_answer quiz_answer_label_id_quiz_question_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_answer_label_id_quiz_question_unique UNIQUE (quiz_question_id, label_id);


--
-- TOC entry 2432 (class 2606 OID 1233080)
-- Name: quiz_answer quiz_id_pley; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_id_pley PRIMARY KEY (id);


--
-- TOC entry 2435 (class 2606 OID 1233082)
-- Name: quiz_question quiz_question_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2437 (class 2606 OID 1233084)
-- Name: quiz_question quiz_question_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2478 (class 2606 OID 1233469)
-- Name: refinement_type refinement_type_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY refinement_type
    ADD CONSTRAINT refinement_type_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2480 (class 2606 OID 1233474)
-- Name: refinement_type refinement_type_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY refinement_type
    ADD CONSTRAINT refinement_type_name_unique UNIQUE (name);


--
-- TOC entry 2482 (class 2606 OID 1233487)
-- Name: refinement_type refinement_type_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY refinement_type
    ADD CONSTRAINT refinement_type_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2393 (class 2606 OID 1233086)
-- Name: image_report report_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT report_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2440 (class 2606 OID 1233088)
-- Name: trending_label_suggestion trending_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2442 (class 2606 OID 1233090)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_unique UNIQUE (label_suggestion_id);


--
-- TOC entry 2446 (class 2606 OID 1233092)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_image_validation_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_image_validation_id_unique UNIQUE (account_id, image_validation_id);


--
-- TOC entry 2458 (class 2606 OID 1233094)
-- Name: user_image_annotation user_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2311 (class 2606 OID 1233096)
-- Name: account user_email_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_email_unique UNIQUE (email);


--
-- TOC entry 2313 (class 2606 OID 1233098)
-- Name: account user_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2448 (class 2606 OID 1233100)
-- Name: user_annotation_blacklist user_image_blacklist_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_image_blacklist_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2453 (class 2606 OID 1233102)
-- Name: user_image user_image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2462 (class 2606 OID 1233104)
-- Name: user_image_validation user_image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2315 (class 2606 OID 1233106)
-- Name: account user_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_name_unique UNIQUE (name);


--
-- TOC entry 2464 (class 2606 OID 1233108)
-- Name: validations_per_app validations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2466 (class 2606 OID 1233110)
-- Name: validations_per_app validations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2468 (class 2606 OID 1233112)
-- Name: validations_per_country validations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2470 (class 2606 OID 1233114)
-- Name: validations_per_country validations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2309 (class 1259 OID 1233115)
-- Name: fki_access_token_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_access_token_user_id_fkey ON access_token USING btree (user_id);


--
-- TOC entry 2318 (class 1259 OID 1233116)
-- Name: fki_account_permission_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_account_permission_account_id_fkey ON account_permission USING btree (account_id);


--
-- TOC entry 2321 (class 1259 OID 1233117)
-- Name: fki_annotation_data_annotation_type_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_annotation_type_fkey ON annotation_data USING btree (annotation_type_id);


--
-- TOC entry 2322 (class 1259 OID 1233118)
-- Name: fki_annotation_data_image_annotation_revision_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_image_annotation_revision_id_fkey ON annotation_data USING btree (image_annotation_revision_id);


--
-- TOC entry 2342 (class 1259 OID 1233119)
-- Name: fki_api_token_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_api_token_account_id_fkey ON api_token USING btree (account_id);


--
-- TOC entry 2323 (class 1259 OID 1233120)
-- Name: fki_image_annotation_data_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_data_annotation_id_fkey ON annotation_data USING btree (image_annotation_id);


--
-- TOC entry 2362 (class 1259 OID 1233121)
-- Name: fki_image_annotation_label_id_key; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_label_id_key ON image_annotation USING btree (label_id);


--
-- TOC entry 2375 (class 1259 OID 1233122)
-- Name: fki_image_annotation_revision_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_revision_image_annotation_id_fkey ON image_annotation_revision USING btree (image_annotation_id);


--
-- TOC entry 2378 (class 1259 OID 1233123)
-- Name: fki_image_label_suggestion_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_image_id_fkey ON image_label_suggestion USING btree (image_id);


--
-- TOC entry 2379 (class 1259 OID 1233124)
-- Name: fki_image_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_label_suggestion_id_fkey ON image_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 2351 (class 1259 OID 1233125)
-- Name: fki_image_provider_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_provider_id_fkey ON image USING btree (image_provider_id);


--
-- TOC entry 2386 (class 1259 OID 1233126)
-- Name: fki_image_quarantine_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quarantine_image_id_fkey ON image_quarantine USING btree (image_id);


--
-- TOC entry 2369 (class 1259 OID 1233127)
-- Name: fki_image_quiz_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_image_annotation_id_fkey ON image_annotation_refinement USING btree (annotation_data_id);


--
-- TOC entry 2370 (class 1259 OID 1233128)
-- Name: fki_image_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_label_id_fkey ON image_annotation_refinement USING btree (label_id);


--
-- TOC entry 2391 (class 1259 OID 1233129)
-- Name: fki_image_report_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_report_image_id_fkey ON image_report USING btree (image_id);


--
-- TOC entry 2394 (class 1259 OID 1233130)
-- Name: fki_image_source_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_source_image_id_fkey ON image_source USING btree (image_id);


--
-- TOC entry 2397 (class 1259 OID 1233131)
-- Name: fki_image_validation_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_image_id_fkey ON image_validation USING btree (image_id);


--
-- TOC entry 2398 (class 1259 OID 1233132)
-- Name: fki_image_validation_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_label_id_fkey ON image_validation USING btree (label_id);


--
-- TOC entry 2405 (class 1259 OID 1233133)
-- Name: fki_image_validation_source_image_source_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_source_id_fkey ON image_validation_source USING btree (image_source_id);


--
-- TOC entry 2406 (class 1259 OID 1233134)
-- Name: fki_image_validation_source_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_validation_id_fkey ON image_validation_source USING btree (image_validation_id);


--
-- TOC entry 2416 (class 1259 OID 1233135)
-- Name: fki_label_accessor_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_accessor_label_id_fkey ON label_accessor USING btree (label_id);


--
-- TOC entry 2421 (class 1259 OID 1233136)
-- Name: fki_label_example_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_example_label_id_fkey ON label_example USING btree (label_id);


--
-- TOC entry 2409 (class 1259 OID 1233137)
-- Name: fki_label_parent_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_parent_id_fkey ON label USING btree (parent_id);


--
-- TOC entry 2471 (class 1259 OID 1233464)
-- Name: fki_label_refinement_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_refinement_label_id_fkey ON label_refinement USING btree (label_id);


--
-- TOC entry 2472 (class 1259 OID 1233483)
-- Name: fki_label_refinement_refinement_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_refinement_refinement_id_fkey ON label_refinement USING btree (refinement_type_id);


--
-- TOC entry 2422 (class 1259 OID 1233138)
-- Name: fki_label_suggestion_proposed_by_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_suggestion_proposed_by_fkey ON label_suggestion USING btree (proposed_by);


--
-- TOC entry 2427 (class 1259 OID 1233139)
-- Name: fki_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_label_id_fkey ON quiz_answer USING btree (label_id);


--
-- TOC entry 2433 (class 1259 OID 1233140)
-- Name: fki_quiz_question_refines_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_question_refines_label_id_fkey ON quiz_question USING btree (refines_label_id);


--
-- TOC entry 2428 (class 1259 OID 1233141)
-- Name: fki_quiz_quiz_question_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_quiz_question_id_fkey ON quiz_answer USING btree (quiz_question_id);


--
-- TOC entry 2438 (class 1259 OID 1233142)
-- Name: fki_trending_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_trending_label_suggestion_label_suggestion_id_fkey ON trending_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 2443 (class 1259 OID 1233143)
-- Name: fki_user_annotation_blacklist_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_account_id_fkey ON user_annotation_blacklist USING btree (account_id);


--
-- TOC entry 2444 (class 1259 OID 1233144)
-- Name: fki_user_annotation_blacklist_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_image_validation_id_fkey ON user_annotation_blacklist USING btree (image_validation_id);


--
-- TOC entry 2449 (class 1259 OID 1233145)
-- Name: fki_user_image_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_account_id_fkey ON user_image USING btree (account_id);


--
-- TOC entry 2455 (class 1259 OID 1233146)
-- Name: fki_user_image_annotation_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_image_annotation_id_fkey ON user_image_annotation USING btree (image_annotation_id);


--
-- TOC entry 2456 (class 1259 OID 1233147)
-- Name: fki_user_image_annotation_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_user_id_fkey ON user_image_annotation USING btree (account_id);


--
-- TOC entry 2450 (class 1259 OID 1233148)
-- Name: fki_user_image_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_image_id_fkey ON user_image USING btree (image_id);


--
-- TOC entry 2459 (class 1259 OID 1233149)
-- Name: fki_user_image_validation_acccount_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_acccount_id_fkey ON user_image_validation USING btree (account_id);


--
-- TOC entry 2460 (class 1259 OID 1233150)
-- Name: fki_user_image_validation_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_account_id_fkey ON user_image_validation USING btree (image_validation_id);


--
-- TOC entry 2365 (class 1259 OID 1233151)
-- Name: image_annotation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_image_id_index ON image_annotation USING btree (image_id);


--
-- TOC entry 2368 (class 1259 OID 1233152)
-- Name: image_annotation_uuid_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_uuid_index ON image_annotation USING btree (uuid);


--
-- TOC entry 2352 (class 1259 OID 1233153)
-- Name: image_hash_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_hash_index ON image USING btree (hash);


--
-- TOC entry 2357 (class 1259 OID 1233154)
-- Name: image_image_provider_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_image_provider_index ON image USING btree (image_provider_id);


--
-- TOC entry 2358 (class 1259 OID 1233155)
-- Name: image_key_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_key_index ON image USING btree (key);


--
-- TOC entry 2361 (class 1259 OID 1233156)
-- Name: image_unlocked_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_unlocked_index ON image USING btree (unlocked);


--
-- TOC entry 2401 (class 1259 OID 1233157)
-- Name: image_validation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_image_id_index ON image_validation USING btree (image_id);


--
-- TOC entry 2404 (class 1259 OID 1233158)
-- Name: image_validation_label_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_label_id_index ON image_validation USING btree (label_id);


--
-- TOC entry 2451 (class 1259 OID 1233159)
-- Name: user_image_account_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX user_image_account_id_index ON user_image USING btree (account_id);


--
-- TOC entry 2454 (class 1259 OID 1233160)
-- Name: user_image_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX user_image_image_id_index ON user_image USING btree (image_id);


--
-- TOC entry 2522 (class 2620 OID 1233161)
-- Name: image_annotation image_annotation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_annotation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON image_annotation FOR EACH ROW EXECUTE PROCEDURE versioning('sys_period', 'image_annotation_history', 'true');


--
-- TOC entry 2523 (class 2620 OID 1233162)
-- Name: image_validation image_validation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_validation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON image_validation FOR EACH ROW EXECUTE PROCEDURE versioning('sys_period', 'image_validation_history', 'true');


--
-- TOC entry 2483 (class 2606 OID 1233163)
-- Name: access_token access_token_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY access_token
    ADD CONSTRAINT access_token_user_id_fkey FOREIGN KEY (user_id) REFERENCES account(id);


--
-- TOC entry 2484 (class 2606 OID 1233168)
-- Name: account_permission account_permission_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account_permission
    ADD CONSTRAINT account_permission_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2485 (class 2606 OID 1233173)
-- Name: annotation_data annotation_data_annotation_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_annotation_type_fkey FOREIGN KEY (annotation_type_id) REFERENCES annotation_type(id);


--
-- TOC entry 2486 (class 2606 OID 1233178)
-- Name: annotation_data annotation_data_image_annotation_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_image_annotation_revision_id_fkey FOREIGN KEY (image_annotation_revision_id) REFERENCES image_annotation_revision(id);


--
-- TOC entry 2488 (class 2606 OID 1233183)
-- Name: api_token api_token_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2487 (class 2606 OID 1233188)
-- Name: annotation_data image_annotation_data_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2490 (class 2606 OID 1233193)
-- Name: image_annotation image_annotation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2491 (class 2606 OID 1233198)
-- Name: image_annotation image_annotation_label_id_key; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_label_id_key FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2492 (class 2606 OID 1233203)
-- Name: image_annotation_refinement image_annotation_refinement_annotation_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_annotation_data_id_fkey FOREIGN KEY (annotation_data_id) REFERENCES annotation_data(id);


--
-- TOC entry 2493 (class 2606 OID 1233208)
-- Name: image_annotation_refinement image_annotation_refinement_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2494 (class 2606 OID 1233213)
-- Name: image_annotation_revision image_annotation_revision_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2489 (class 2606 OID 1233218)
-- Name: image image_image_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_image_provider_id_fkey FOREIGN KEY (image_provider_id) REFERENCES image_provider(id);


--
-- TOC entry 2495 (class 2606 OID 1233223)
-- Name: image_label_suggestion image_label_suggestion_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2496 (class 2606 OID 1233228)
-- Name: image_label_suggestion image_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES label_suggestion(id);


--
-- TOC entry 2497 (class 2606 OID 1233233)
-- Name: image_quarantine image_quarantine_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2498 (class 2606 OID 1233238)
-- Name: image_report image_report_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT image_report_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2499 (class 2606 OID 1233243)
-- Name: image_source image_source_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_source
    ADD CONSTRAINT image_source_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2500 (class 2606 OID 1233248)
-- Name: image_validation image_validation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2501 (class 2606 OID 1233253)
-- Name: image_validation image_validation_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2502 (class 2606 OID 1233258)
-- Name: image_validation_source image_validation_source_image_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_image_source_id_fkey FOREIGN KEY (image_source_id) REFERENCES image_source(id);


--
-- TOC entry 2503 (class 2606 OID 1233263)
-- Name: image_validation_source image_validation_source_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2505 (class 2606 OID 1233268)
-- Name: label_accessor label_accessor_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2506 (class 2606 OID 1233273)
-- Name: label_example label_example_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_example
    ADD CONSTRAINT label_example_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2504 (class 2606 OID 1233278)
-- Name: label label_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES label(id);


--
-- TOC entry 2520 (class 2606 OID 1233459)
-- Name: label_refinement label_refinement_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_refinement
    ADD CONSTRAINT label_refinement_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2521 (class 2606 OID 1233478)
-- Name: label_refinement label_refinement_refinement_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_refinement
    ADD CONSTRAINT label_refinement_refinement_type_id_fkey FOREIGN KEY (refinement_type_id) REFERENCES refinement_type(id);


--
-- TOC entry 2507 (class 2606 OID 1233283)
-- Name: label_suggestion label_suggestion_proposed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_proposed_by_fkey FOREIGN KEY (proposed_by) REFERENCES account(id);


--
-- TOC entry 2508 (class 2606 OID 1233288)
-- Name: quiz_answer quiz_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2510 (class 2606 OID 1233293)
-- Name: quiz_question quiz_question_refines_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_refines_label_id_fkey FOREIGN KEY (refines_label_id) REFERENCES label(id);


--
-- TOC entry 2509 (class 2606 OID 1233298)
-- Name: quiz_answer quiz_quiz_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_quiz_question_id_fkey FOREIGN KEY (quiz_question_id) REFERENCES quiz_question(id);


--
-- TOC entry 2511 (class 2606 OID 1233303)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES label_suggestion(id);


--
-- TOC entry 2512 (class 2606 OID 1233308)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2513 (class 2606 OID 1233313)
-- Name: user_annotation_blacklist user_annotation_blacklist_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2514 (class 2606 OID 1233318)
-- Name: user_image user_image_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2516 (class 2606 OID 1233323)
-- Name: user_image_annotation user_image_annotation_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_image_annotation_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2517 (class 2606 OID 1233328)
-- Name: user_image_annotation user_image_annotation_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_image_annotation_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2515 (class 2606 OID 1233333)
-- Name: user_image user_image_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2518 (class 2606 OID 1233338)
-- Name: user_image_validation user_image_validation_acccount_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_acccount_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2519 (class 2606 OID 1233343)
-- Name: user_image_validation user_image_validation_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2647 (class 0 OID 0)
-- Dependencies: 5
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO monkey;


-- Completed on 2018-08-01 22:10:38

--
-- PostgreSQL database dump complete
--

