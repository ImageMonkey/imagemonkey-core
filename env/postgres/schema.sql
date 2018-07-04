--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.5

-- Started on 2018-07-04 19:21:15

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
-- TOC entry 2604 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 2 (class 3079 OID 66319)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 2605 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- TOC entry 559 (class 1247 OID 66331)
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
-- TOC entry 266 (class 1255 OID 66339)
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
-- TOC entry 186 (class 1259 OID 66340)
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
-- TOC entry 187 (class 1259 OID 66342)
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
-- TOC entry 188 (class 1259 OID 66349)
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
-- TOC entry 189 (class 1259 OID 66351)
-- Name: account; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE account (
    id bigint DEFAULT nextval('account_id_seq'::regclass) NOT NULL,
    name text,
    hashed_password text,
    email text,
    profile_picture text,
    created bigint NOT NULL,
    is_moderator boolean
);


ALTER TABLE account OWNER TO monkey;

--
-- TOC entry 190 (class 1259 OID 66358)
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
-- TOC entry 191 (class 1259 OID 66360)
-- Name: account_permission; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE account_permission (
    id bigint DEFAULT nextval('account_permission_id_seq'::regclass) NOT NULL,
    can_remove_label boolean,
    account_id bigint
);


ALTER TABLE account_permission OWNER TO monkey;

--
-- TOC entry 192 (class 1259 OID 66364)
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
-- TOC entry 193 (class 1259 OID 66366)
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
-- TOC entry 194 (class 1259 OID 66373)
-- Name: annotation_type; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotation_type (
    id bigint NOT NULL,
    name text
);


ALTER TABLE annotation_type OWNER TO monkey;

--
-- TOC entry 195 (class 1259 OID 66379)
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
-- TOC entry 196 (class 1259 OID 66381)
-- Name: annotations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_app (
    id bigint DEFAULT nextval('annotations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE annotations_per_app OWNER TO monkey;

--
-- TOC entry 197 (class 1259 OID 66388)
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
-- TOC entry 198 (class 1259 OID 66390)
-- Name: annotations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE annotations_per_country (
    id bigint DEFAULT nextval('annotations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE annotations_per_country OWNER TO monkey;

--
-- TOC entry 199 (class 1259 OID 66397)
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
-- TOC entry 200 (class 1259 OID 66399)
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
-- TOC entry 201 (class 1259 OID 66406)
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
-- TOC entry 202 (class 1259 OID 66408)
-- Name: donations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_app (
    id bigint DEFAULT nextval('donations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE donations_per_app OWNER TO monkey;

--
-- TOC entry 203 (class 1259 OID 66415)
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
-- TOC entry 204 (class 1259 OID 66417)
-- Name: donations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE donations_per_country (
    id bigint DEFAULT nextval('donations_per_country_id_seq'::regclass) NOT NULL,
    country_code text,
    count bigint
);


ALTER TABLE donations_per_country OWNER TO monkey;

--
-- TOC entry 205 (class 1259 OID 66424)
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
-- TOC entry 206 (class 1259 OID 66430)
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
-- TOC entry 207 (class 1259 OID 66432)
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
    revision integer
);


ALTER TABLE image_annotation OWNER TO monkey;

--
-- TOC entry 208 (class 1259 OID 66440)
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
-- TOC entry 209 (class 1259 OID 66447)
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
-- TOC entry 210 (class 1259 OID 66449)
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
-- TOC entry 211 (class 1259 OID 66457)
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
-- TOC entry 212 (class 1259 OID 66459)
-- Name: image_annotation_revision; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_annotation_revision (
    id bigint DEFAULT nextval('image_annotation_revision_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    revision integer
);


ALTER TABLE image_annotation_revision OWNER TO monkey;

--
-- TOC entry 213 (class 1259 OID 66463)
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
-- TOC entry 214 (class 1259 OID 66465)
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
-- TOC entry 2606 (class 0 OID 0)
-- Dependencies: 214
-- Name: image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: monkey
--

ALTER SEQUENCE image_id_seq OWNED BY image.id;


--
-- TOC entry 215 (class 1259 OID 66467)
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
-- TOC entry 216 (class 1259 OID 66469)
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
-- TOC entry 217 (class 1259 OID 66476)
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
-- TOC entry 218 (class 1259 OID 66478)
-- Name: image_provider; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_provider (
    id bigint DEFAULT nextval('image_provider_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE image_provider OWNER TO monkey;

--
-- TOC entry 219 (class 1259 OID 66485)
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
-- TOC entry 220 (class 1259 OID 66487)
-- Name: image_quarantine; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_quarantine (
    id bigint DEFAULT nextval('image_quarantine_id_seq'::regclass) NOT NULL,
    image_id bigint
);


ALTER TABLE image_quarantine OWNER TO monkey;

--
-- TOC entry 221 (class 1259 OID 66491)
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
-- TOC entry 222 (class 1259 OID 66493)
-- Name: image_report; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_report (
    id bigint DEFAULT nextval('report_id_seq'::regclass) NOT NULL,
    reason text,
    image_id bigint
);


ALTER TABLE image_report OWNER TO monkey;

--
-- TOC entry 223 (class 1259 OID 66500)
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
-- TOC entry 224 (class 1259 OID 66502)
-- Name: image_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_source (
    id bigint DEFAULT nextval('image_source_id_seq'::regclass) NOT NULL,
    url text,
    image_id bigint
);


ALTER TABLE image_source OWNER TO monkey;

--
-- TOC entry 225 (class 1259 OID 66509)
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
-- TOC entry 226 (class 1259 OID 66511)
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
-- TOC entry 227 (class 1259 OID 66519)
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
-- TOC entry 228 (class 1259 OID 66525)
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
-- TOC entry 229 (class 1259 OID 66527)
-- Name: image_validation_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE image_validation_source (
    id bigint DEFAULT nextval('image_validation_source_id_seq'::regclass) NOT NULL,
    image_validation_id bigint,
    image_source_id bigint
);


ALTER TABLE image_validation_source OWNER TO monkey;

--
-- TOC entry 230 (class 1259 OID 66531)
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
-- TOC entry 231 (class 1259 OID 66533)
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
-- TOC entry 232 (class 1259 OID 66540)
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
-- TOC entry 233 (class 1259 OID 66542)
-- Name: label_accessor; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_accessor (
    id bigint DEFAULT nextval('label_accessor_id_seq'::regclass) NOT NULL,
    label_id bigint,
    accessor text
);


ALTER TABLE label_accessor OWNER TO monkey;

--
-- TOC entry 234 (class 1259 OID 66549)
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
-- TOC entry 235 (class 1259 OID 66551)
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
-- TOC entry 236 (class 1259 OID 66558)
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
-- TOC entry 237 (class 1259 OID 66560)
-- Name: label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE label_suggestion (
    id bigint DEFAULT nextval('label_suggestion_id_seq'::regclass) NOT NULL,
    name text,
    proposed_by bigint
);


ALTER TABLE label_suggestion OWNER TO monkey;

--
-- TOC entry 238 (class 1259 OID 66567)
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
-- TOC entry 239 (class 1259 OID 66569)
-- Name: quiz_answer; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE quiz_answer (
    id bigint DEFAULT nextval('quiz_answer_id_seq'::regclass) NOT NULL,
    quiz_question_id bigint,
    label_id bigint
);


ALTER TABLE quiz_answer OWNER TO monkey;

--
-- TOC entry 240 (class 1259 OID 66573)
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
-- TOC entry 241 (class 1259 OID 66575)
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
    uuid uuid
);


ALTER TABLE quiz_question OWNER TO monkey;

--
-- TOC entry 242 (class 1259 OID 66582)
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
-- TOC entry 243 (class 1259 OID 66584)
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
-- TOC entry 244 (class 1259 OID 66588)
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
-- TOC entry 245 (class 1259 OID 66590)
-- Name: user_annotation_blacklist; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_annotation_blacklist (
    id bigint DEFAULT nextval('user_annotation_blacklist_id_seq'::regclass) NOT NULL,
    account_id bigint,
    image_validation_id bigint
);


ALTER TABLE user_annotation_blacklist OWNER TO monkey;

--
-- TOC entry 246 (class 1259 OID 66594)
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
-- TOC entry 247 (class 1259 OID 66596)
-- Name: user_image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE user_image (
    id bigint DEFAULT nextval('user_image_id_seq'::regclass) NOT NULL,
    image_id bigint,
    account_id bigint
);


ALTER TABLE user_image OWNER TO monkey;

--
-- TOC entry 248 (class 1259 OID 66600)
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
-- TOC entry 249 (class 1259 OID 66602)
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
-- TOC entry 250 (class 1259 OID 66606)
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
-- TOC entry 251 (class 1259 OID 66608)
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
-- TOC entry 252 (class 1259 OID 66612)
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
-- TOC entry 253 (class 1259 OID 66614)
-- Name: validations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_app (
    id bigint DEFAULT nextval('validations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE validations_per_app OWNER TO monkey;

--
-- TOC entry 254 (class 1259 OID 66621)
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
-- TOC entry 255 (class 1259 OID 66623)
-- Name: validations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE validations_per_country (
    id bigint DEFAULT nextval('validations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE validations_per_country OWNER TO monkey;

--
-- TOC entry 2257 (class 2604 OID 66630)
-- Name: image id; Type: DEFAULT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image ALTER COLUMN id SET DEFAULT nextval('image_id_seq'::regclass);


--
-- TOC entry 2286 (class 2606 OID 66632)
-- Name: access_token access_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY access_token
    ADD CONSTRAINT access_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2295 (class 2606 OID 66634)
-- Name: account_permission account_permission_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account_permission
    ADD CONSTRAINT account_permission_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2303 (class 2606 OID 66636)
-- Name: annotation_type annotation_type_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2305 (class 2606 OID 66638)
-- Name: annotation_type annotation_type_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_type
    ADD CONSTRAINT annotation_type_name_unique UNIQUE (name);


--
-- TOC entry 2307 (class 2606 OID 66640)
-- Name: annotations_per_app annotations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2309 (class 2606 OID 66642)
-- Name: annotations_per_app annotations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_app
    ADD CONSTRAINT annotations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2311 (class 2606 OID 66644)
-- Name: annotations_per_country annotations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2313 (class 2606 OID 66646)
-- Name: annotations_per_country annotations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotations_per_country
    ADD CONSTRAINT annotations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2315 (class 2606 OID 66648)
-- Name: api_token api_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2317 (class 2606 OID 66650)
-- Name: api_token api_token_token_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_token_unique UNIQUE (token);


--
-- TOC entry 2320 (class 2606 OID 66652)
-- Name: donations_per_app donations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2322 (class 2606 OID 66654)
-- Name: donations_per_app donations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_app
    ADD CONSTRAINT donations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2324 (class 2606 OID 66656)
-- Name: donations_per_country donations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2326 (class 2606 OID 66658)
-- Name: donations_per_country donations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY donations_per_country
    ADD CONSTRAINT donations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2301 (class 2606 OID 66660)
-- Name: annotation_data image_annotation_data_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2340 (class 2606 OID 66662)
-- Name: image_annotation image_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2343 (class 2606 OID 66664)
-- Name: image_annotation image_annotation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_label_uniquekey UNIQUE (image_id, label_id, auto_generated);


--
-- TOC entry 2348 (class 2606 OID 66666)
-- Name: image_annotation_refinement image_annotation_refinement_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2350 (class 2606 OID 66668)
-- Name: image_annotation_refinement image_annotation_refinement_label_annotation_data_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_annotation_data_unique UNIQUE (annotation_data_id, label_id);


--
-- TOC entry 2353 (class 2606 OID 66670)
-- Name: image_annotation_revision image_annotation_revision_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2330 (class 2606 OID 66672)
-- Name: image image_hash_unique_key; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_hash_unique_key UNIQUE (hash);


--
-- TOC entry 2332 (class 2606 OID 66674)
-- Name: image image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2336 (class 2606 OID 66676)
-- Name: image image_key_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_key_unique UNIQUE (image_provider_id, key);


--
-- TOC entry 2357 (class 2606 OID 66678)
-- Name: image_label_suggestion image_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2359 (class 2606 OID 66680)
-- Name: image_label_suggestion image_label_suggestion_image_id_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_label_suggestion_id_unique UNIQUE (label_suggestion_id, image_id);


--
-- TOC entry 2361 (class 2606 OID 66682)
-- Name: image_provider image_provider_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_provider
    ADD CONSTRAINT image_provider_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2364 (class 2606 OID 66684)
-- Name: image_quarantine image_quarantine_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2366 (class 2606 OID 66686)
-- Name: image_quarantine image_quarantine_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_unique UNIQUE (image_id);


--
-- TOC entry 2372 (class 2606 OID 66688)
-- Name: image_source image_source_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_source
    ADD CONSTRAINT image_source_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2376 (class 2606 OID 66690)
-- Name: image_validation image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2379 (class 2606 OID 66692)
-- Name: image_validation image_validation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_label_uniquekey UNIQUE (image_id, label_id);


--
-- TOC entry 2384 (class 2606 OID 66694)
-- Name: image_validation_source image_validation_source_id; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_id PRIMARY KEY (id);


--
-- TOC entry 2394 (class 2606 OID 66696)
-- Name: label_accessor label_accessor_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2396 (class 2606 OID 66698)
-- Name: label_accessor label_accessor_label_id_accessor_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_label_id_accessor_unique UNIQUE (label_id, accessor);


--
-- TOC entry 2387 (class 2606 OID 66700)
-- Name: label label_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2389 (class 2606 OID 66702)
-- Name: label label_name_parent_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_name_parent_id_unique UNIQUE (name, parent_id);


--
-- TOC entry 2400 (class 2606 OID 66704)
-- Name: label_suggestion label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2402 (class 2606 OID 66706)
-- Name: label_suggestion label_suggestion_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_name_unique UNIQUE (name);


--
-- TOC entry 2391 (class 2606 OID 66708)
-- Name: label label_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2406 (class 2606 OID 66710)
-- Name: quiz_answer quiz_id_pley; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_id_pley PRIMARY KEY (id);


--
-- TOC entry 2409 (class 2606 OID 66712)
-- Name: quiz_question quiz_question_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2411 (class 2606 OID 66974)
-- Name: quiz_question quiz_question_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_uuid_unique UNIQUE (uuid);


--
-- TOC entry 2369 (class 2606 OID 66714)
-- Name: image_report report_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT report_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2414 (class 2606 OID 66716)
-- Name: trending_label_suggestion trending_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2416 (class 2606 OID 66718)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_unique UNIQUE (label_suggestion_id);


--
-- TOC entry 2420 (class 2606 OID 66720)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_image_validation_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_image_validation_id_unique UNIQUE (account_id, image_validation_id);


--
-- TOC entry 2430 (class 2606 OID 66722)
-- Name: user_image_annotation user_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2289 (class 2606 OID 66724)
-- Name: account user_email_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_email_unique UNIQUE (email);


--
-- TOC entry 2291 (class 2606 OID 66726)
-- Name: account user_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2422 (class 2606 OID 66728)
-- Name: user_annotation_blacklist user_image_blacklist_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_image_blacklist_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2426 (class 2606 OID 66730)
-- Name: user_image user_image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2434 (class 2606 OID 66732)
-- Name: user_image_validation user_image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2293 (class 2606 OID 66734)
-- Name: account user_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account
    ADD CONSTRAINT user_name_unique UNIQUE (name);


--
-- TOC entry 2436 (class 2606 OID 66736)
-- Name: validations_per_app validations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 2438 (class 2606 OID 66738)
-- Name: validations_per_app validations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_app
    ADD CONSTRAINT validations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2440 (class 2606 OID 66740)
-- Name: validations_per_country validations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 2442 (class 2606 OID 66742)
-- Name: validations_per_country validations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY validations_per_country
    ADD CONSTRAINT validations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2287 (class 1259 OID 66743)
-- Name: fki_access_token_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_access_token_user_id_fkey ON access_token USING btree (user_id);


--
-- TOC entry 2296 (class 1259 OID 66744)
-- Name: fki_account_permission_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_account_permission_account_id_fkey ON account_permission USING btree (account_id);


--
-- TOC entry 2297 (class 1259 OID 66745)
-- Name: fki_annotation_data_annotation_type_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_annotation_type_fkey ON annotation_data USING btree (annotation_type_id);


--
-- TOC entry 2298 (class 1259 OID 66746)
-- Name: fki_annotation_data_image_annotation_revision_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_image_annotation_revision_id_fkey ON annotation_data USING btree (image_annotation_revision_id);


--
-- TOC entry 2318 (class 1259 OID 66747)
-- Name: fki_api_token_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_api_token_account_id_fkey ON api_token USING btree (account_id);


--
-- TOC entry 2299 (class 1259 OID 66748)
-- Name: fki_image_annotation_data_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_data_annotation_id_fkey ON annotation_data USING btree (image_annotation_id);


--
-- TOC entry 2338 (class 1259 OID 66749)
-- Name: fki_image_annotation_label_id_key; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_label_id_key ON image_annotation USING btree (label_id);


--
-- TOC entry 2351 (class 1259 OID 66750)
-- Name: fki_image_annotation_revision_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_revision_image_annotation_id_fkey ON image_annotation_revision USING btree (image_annotation_id);


--
-- TOC entry 2354 (class 1259 OID 66751)
-- Name: fki_image_label_suggestion_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_image_id_fkey ON image_label_suggestion USING btree (image_id);


--
-- TOC entry 2355 (class 1259 OID 66752)
-- Name: fki_image_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_label_suggestion_id_fkey ON image_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 2327 (class 1259 OID 66753)
-- Name: fki_image_provider_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_provider_id_fkey ON image USING btree (image_provider_id);


--
-- TOC entry 2362 (class 1259 OID 66754)
-- Name: fki_image_quarantine_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quarantine_image_id_fkey ON image_quarantine USING btree (image_id);


--
-- TOC entry 2345 (class 1259 OID 66755)
-- Name: fki_image_quiz_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_image_annotation_id_fkey ON image_annotation_refinement USING btree (annotation_data_id);


--
-- TOC entry 2346 (class 1259 OID 66756)
-- Name: fki_image_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_label_id_fkey ON image_annotation_refinement USING btree (label_id);


--
-- TOC entry 2367 (class 1259 OID 66757)
-- Name: fki_image_report_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_report_image_id_fkey ON image_report USING btree (image_id);


--
-- TOC entry 2370 (class 1259 OID 66758)
-- Name: fki_image_source_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_source_image_id_fkey ON image_source USING btree (image_id);


--
-- TOC entry 2373 (class 1259 OID 66759)
-- Name: fki_image_validation_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_image_id_fkey ON image_validation USING btree (image_id);


--
-- TOC entry 2374 (class 1259 OID 66760)
-- Name: fki_image_validation_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_label_id_fkey ON image_validation USING btree (label_id);


--
-- TOC entry 2381 (class 1259 OID 66761)
-- Name: fki_image_validation_source_image_source_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_source_id_fkey ON image_validation_source USING btree (image_source_id);


--
-- TOC entry 2382 (class 1259 OID 66762)
-- Name: fki_image_validation_source_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_validation_id_fkey ON image_validation_source USING btree (image_validation_id);


--
-- TOC entry 2392 (class 1259 OID 66763)
-- Name: fki_label_accessor_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_accessor_label_id_fkey ON label_accessor USING btree (label_id);


--
-- TOC entry 2397 (class 1259 OID 66764)
-- Name: fki_label_example_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_example_label_id_fkey ON label_example USING btree (label_id);


--
-- TOC entry 2385 (class 1259 OID 66765)
-- Name: fki_label_parent_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_parent_id_fkey ON label USING btree (parent_id);


--
-- TOC entry 2398 (class 1259 OID 66766)
-- Name: fki_label_suggestion_proposed_by_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_suggestion_proposed_by_fkey ON label_suggestion USING btree (proposed_by);


--
-- TOC entry 2403 (class 1259 OID 66767)
-- Name: fki_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_label_id_fkey ON quiz_answer USING btree (label_id);


--
-- TOC entry 2407 (class 1259 OID 66768)
-- Name: fki_quiz_question_refines_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_question_refines_label_id_fkey ON quiz_question USING btree (refines_label_id);


--
-- TOC entry 2404 (class 1259 OID 66769)
-- Name: fki_quiz_quiz_question_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_quiz_question_id_fkey ON quiz_answer USING btree (quiz_question_id);


--
-- TOC entry 2412 (class 1259 OID 66770)
-- Name: fki_trending_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_trending_label_suggestion_label_suggestion_id_fkey ON trending_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 2417 (class 1259 OID 66771)
-- Name: fki_user_annotation_blacklist_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_account_id_fkey ON user_annotation_blacklist USING btree (account_id);


--
-- TOC entry 2418 (class 1259 OID 66772)
-- Name: fki_user_annotation_blacklist_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_image_validation_id_fkey ON user_annotation_blacklist USING btree (image_validation_id);


--
-- TOC entry 2423 (class 1259 OID 66773)
-- Name: fki_user_image_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_account_id_fkey ON user_image USING btree (account_id);


--
-- TOC entry 2427 (class 1259 OID 66774)
-- Name: fki_user_image_annotation_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_image_annotation_id_fkey ON user_image_annotation USING btree (image_annotation_id);


--
-- TOC entry 2428 (class 1259 OID 66775)
-- Name: fki_user_image_annotation_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_user_id_fkey ON user_image_annotation USING btree (account_id);


--
-- TOC entry 2424 (class 1259 OID 66776)
-- Name: fki_user_image_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_image_id_fkey ON user_image USING btree (image_id);


--
-- TOC entry 2431 (class 1259 OID 66777)
-- Name: fki_user_image_validation_acccount_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_acccount_id_fkey ON user_image_validation USING btree (account_id);


--
-- TOC entry 2432 (class 1259 OID 66778)
-- Name: fki_user_image_validation_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_account_id_fkey ON user_image_validation USING btree (image_validation_id);


--
-- TOC entry 2341 (class 1259 OID 66779)
-- Name: image_annotation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_image_id_index ON image_annotation USING btree (image_id);


--
-- TOC entry 2344 (class 1259 OID 66780)
-- Name: image_annotation_uuid_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_uuid_index ON image_annotation USING btree (uuid);


--
-- TOC entry 2328 (class 1259 OID 66781)
-- Name: image_hash_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_hash_index ON image USING btree (hash);


--
-- TOC entry 2333 (class 1259 OID 66782)
-- Name: image_image_provider_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_image_provider_index ON image USING btree (image_provider_id);


--
-- TOC entry 2334 (class 1259 OID 66783)
-- Name: image_key_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_key_index ON image USING btree (key);


--
-- TOC entry 2337 (class 1259 OID 66784)
-- Name: image_unlocked_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_unlocked_index ON image USING btree (unlocked);


--
-- TOC entry 2377 (class 1259 OID 66785)
-- Name: image_validation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_image_id_index ON image_validation USING btree (image_id);


--
-- TOC entry 2380 (class 1259 OID 66786)
-- Name: image_validation_label_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_label_id_index ON image_validation USING btree (label_id);


--
-- TOC entry 2443 (class 2606 OID 66787)
-- Name: access_token access_token_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY access_token
    ADD CONSTRAINT access_token_user_id_fkey FOREIGN KEY (user_id) REFERENCES account(id);


--
-- TOC entry 2444 (class 2606 OID 66792)
-- Name: account_permission account_permission_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY account_permission
    ADD CONSTRAINT account_permission_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2445 (class 2606 OID 66797)
-- Name: annotation_data annotation_data_annotation_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_annotation_type_fkey FOREIGN KEY (annotation_type_id) REFERENCES annotation_type(id);


--
-- TOC entry 2446 (class 2606 OID 66802)
-- Name: annotation_data annotation_data_image_annotation_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT annotation_data_image_annotation_revision_id_fkey FOREIGN KEY (image_annotation_revision_id) REFERENCES image_annotation_revision(id);


--
-- TOC entry 2448 (class 2606 OID 66807)
-- Name: api_token api_token_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY api_token
    ADD CONSTRAINT api_token_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2447 (class 2606 OID 66812)
-- Name: annotation_data image_annotation_data_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY annotation_data
    ADD CONSTRAINT image_annotation_data_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2450 (class 2606 OID 66817)
-- Name: image_annotation image_annotation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2451 (class 2606 OID 66822)
-- Name: image_annotation image_annotation_label_id_key; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation
    ADD CONSTRAINT image_annotation_label_id_key FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2452 (class 2606 OID 66827)
-- Name: image_annotation_refinement image_annotation_refinement_annotation_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_annotation_data_id_fkey FOREIGN KEY (annotation_data_id) REFERENCES annotation_data(id);


--
-- TOC entry 2453 (class 2606 OID 66832)
-- Name: image_annotation_refinement image_annotation_refinement_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2454 (class 2606 OID 66837)
-- Name: image_annotation_revision image_annotation_revision_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2449 (class 2606 OID 66842)
-- Name: image image_image_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_image_provider_id_fkey FOREIGN KEY (image_provider_id) REFERENCES image_provider(id);


--
-- TOC entry 2455 (class 2606 OID 66847)
-- Name: image_label_suggestion image_label_suggestion_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2456 (class 2606 OID 66852)
-- Name: image_label_suggestion image_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES label_suggestion(id);


--
-- TOC entry 2457 (class 2606 OID 66857)
-- Name: image_quarantine image_quarantine_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2458 (class 2606 OID 66862)
-- Name: image_report image_report_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT image_report_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2459 (class 2606 OID 66867)
-- Name: image_source image_source_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_source
    ADD CONSTRAINT image_source_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2460 (class 2606 OID 66872)
-- Name: image_validation image_validation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2461 (class 2606 OID 66877)
-- Name: image_validation image_validation_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2462 (class 2606 OID 66882)
-- Name: image_validation_source image_validation_source_image_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_image_source_id_fkey FOREIGN KEY (image_source_id) REFERENCES image_source(id);


--
-- TOC entry 2463 (class 2606 OID 66887)
-- Name: image_validation_source image_validation_source_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY image_validation_source
    ADD CONSTRAINT image_validation_source_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2465 (class 2606 OID 66892)
-- Name: label_accessor label_accessor_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_accessor
    ADD CONSTRAINT label_accessor_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2466 (class 2606 OID 66897)
-- Name: label_example label_example_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_example
    ADD CONSTRAINT label_example_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2464 (class 2606 OID 66902)
-- Name: label label_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES label(id);


--
-- TOC entry 2467 (class 2606 OID 66907)
-- Name: label_suggestion label_suggestion_proposed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY label_suggestion
    ADD CONSTRAINT label_suggestion_proposed_by_fkey FOREIGN KEY (proposed_by) REFERENCES account(id);


--
-- TOC entry 2468 (class 2606 OID 66912)
-- Name: quiz_answer quiz_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2470 (class 2606 OID 66917)
-- Name: quiz_question quiz_question_refines_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_question
    ADD CONSTRAINT quiz_question_refines_label_id_fkey FOREIGN KEY (refines_label_id) REFERENCES label(id);


--
-- TOC entry 2469 (class 2606 OID 66922)
-- Name: quiz_answer quiz_quiz_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY quiz_answer
    ADD CONSTRAINT quiz_quiz_question_id_fkey FOREIGN KEY (quiz_question_id) REFERENCES quiz_question(id);


--
-- TOC entry 2471 (class 2606 OID 66927)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES label_suggestion(id);


--
-- TOC entry 2472 (class 2606 OID 66932)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2473 (class 2606 OID 66937)
-- Name: user_annotation_blacklist user_annotation_blacklist_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2474 (class 2606 OID 66942)
-- Name: user_image user_image_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2476 (class 2606 OID 66947)
-- Name: user_image_annotation user_image_annotation_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_image_annotation_account_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2477 (class 2606 OID 66952)
-- Name: user_image_annotation user_image_annotation_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_annotation
    ADD CONSTRAINT user_image_annotation_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES image_annotation(id);


--
-- TOC entry 2475 (class 2606 OID 66957)
-- Name: user_image user_image_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image
    ADD CONSTRAINT user_image_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2478 (class 2606 OID 66962)
-- Name: user_image_validation user_image_validation_acccount_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_acccount_id_fkey FOREIGN KEY (account_id) REFERENCES account(id);


--
-- TOC entry 2479 (class 2606 OID 66967)
-- Name: user_image_validation user_image_validation_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY user_image_validation
    ADD CONSTRAINT user_image_validation_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES image_validation(id);


--
-- TOC entry 2603 (class 0 OID 0)
-- Dependencies: 4
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO monkey;


-- Completed on 2018-07-04 19:21:16

--
-- PostgreSQL database dump complete
--

