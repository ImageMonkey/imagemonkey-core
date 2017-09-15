--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.5

-- Started on 2017-09-13 20:55:58

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 2200 (class 1262 OID 16384)
-- Name: imagemonkey; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE imagemonkey WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8';


ALTER DATABASE imagemonkey OWNER TO postgres;

\connect imagemonkey

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
-- TOC entry 2202 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 187 (class 1259 OID 16417)
-- Name: image; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE image (
    id bigint NOT NULL,
    image_provider_id bigint,
    key text,
    unlocked boolean
);


ALTER TABLE image OWNER TO postgres;

--
-- TOC entry 192 (class 1259 OID 16467)
-- Name: image_classification_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE image_classification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_classification_id_seq OWNER TO postgres;

--
-- TOC entry 191 (class 1259 OID 16450)
-- Name: image_classification; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE image_classification (
    image_id bigint,
    label_id bigint,
    id bigint DEFAULT nextval('image_classification_id_seq'::regclass) NOT NULL
);


ALTER TABLE image_classification OWNER TO postgres;

--
-- TOC entry 186 (class 1259 OID 16415)
-- Name: image_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE image_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_id_seq OWNER TO postgres;

--
-- TOC entry 2203 (class 0 OID 0)
-- Dependencies: 186
-- Name: image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE image_id_seq OWNED BY image.id;


--
-- TOC entry 188 (class 1259 OID 16434)
-- Name: image_provider_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE image_provider_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_provider_id_seq OWNER TO postgres;

--
-- TOC entry 185 (class 1259 OID 16385)
-- Name: image_provider; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE image_provider (
    id bigint DEFAULT nextval('image_provider_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE image_provider OWNER TO postgres;

--
-- TOC entry 198 (class 1259 OID 16620)
-- Name: report_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE report_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE report_id_seq OWNER TO postgres;

--
-- TOC entry 197 (class 1259 OID 16617)
-- Name: image_report; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE image_report (
    id bigint DEFAULT nextval('report_id_seq'::regclass) NOT NULL,
    reason text,
    image_id bigint
);


ALTER TABLE image_report OWNER TO postgres;

--
-- TOC entry 194 (class 1259 OID 16487)
-- Name: image_validation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE image_validation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE image_validation_id_seq OWNER TO postgres;

--
-- TOC entry 193 (class 1259 OID 16470)
-- Name: image_validation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE image_validation (
    id bigint DEFAULT nextval('image_validation_id_seq'::regclass) NOT NULL,
    image_id bigint,
    label_id bigint,
    num_of_valid integer,
    num_of_invalid integer
);


ALTER TABLE image_validation OWNER TO postgres;

--
-- TOC entry 190 (class 1259 OID 16447)
-- Name: name_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE name_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE name_id_seq OWNER TO postgres;

--
-- TOC entry 189 (class 1259 OID 16437)
-- Name: label; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE label (
    id bigint DEFAULT nextval('name_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE label OWNER TO postgres;

--
-- TOC entry 196 (class 1259 OID 16606)
-- Name: rd_table; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE rd_table (
    rd numeric
);


ALTER TABLE rd_table OWNER TO postgres;

--
-- TOC entry 195 (class 1259 OID 16600)
-- Name: tmp_table; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE tmp_table (
    rd numeric,
    id bigint,
    key text
);


ALTER TABLE tmp_table OWNER TO postgres;

--
-- TOC entry 2046 (class 2604 OID 16420)
-- Name: image id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image ALTER COLUMN id SET DEFAULT nextval('image_id_seq'::regclass);


--
-- TOC entry 2065 (class 2606 OID 16466)
-- Name: image_classification image_classification_id_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_classification
    ADD CONSTRAINT image_classification_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2055 (class 2606 OID 16425)
-- Name: image image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2057 (class 2606 OID 16427)
-- Name: image image_key_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_key_unique UNIQUE (image_provider_id, key);


--
-- TOC entry 2052 (class 2606 OID 16389)
-- Name: image_provider image_provider_id_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_provider
    ADD CONSTRAINT image_provider_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2069 (class 2606 OID 16474)
-- Name: image_validation image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2059 (class 2606 OID 16444)
-- Name: label label_id_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2061 (class 2606 OID 16446)
-- Name: label label_name_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY label
    ADD CONSTRAINT label_name_unique UNIQUE (name);


--
-- TOC entry 2072 (class 2606 OID 16624)
-- Name: image_report report_id_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT report_id_pkey PRIMARY KEY (id);


--
-- TOC entry 2062 (class 1259 OID 16458)
-- Name: fki_image_classification_image_id_fkey; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX fki_image_classification_image_id_fkey ON image_classification USING btree (image_id);


--
-- TOC entry 2063 (class 1259 OID 16464)
-- Name: fki_image_classification_label_id_fkey; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX fki_image_classification_label_id_fkey ON image_classification USING btree (label_id);


--
-- TOC entry 2053 (class 1259 OID 16433)
-- Name: fki_image_provider_id_fkey; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX fki_image_provider_id_fkey ON image USING btree (image_provider_id);


--
-- TOC entry 2070 (class 1259 OID 16633)
-- Name: fki_image_report_image_id_fkey; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX fki_image_report_image_id_fkey ON image_report USING btree (image_id);


--
-- TOC entry 2066 (class 1259 OID 16480)
-- Name: fki_image_validation_image_id_fkey; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX fki_image_validation_image_id_fkey ON image_validation USING btree (image_id);


--
-- TOC entry 2067 (class 1259 OID 16486)
-- Name: fki_image_validation_label_id_fkey; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX fki_image_validation_label_id_fkey ON image_validation USING btree (label_id);


--
-- TOC entry 2074 (class 2606 OID 16453)
-- Name: image_classification image_classification_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_classification
    ADD CONSTRAINT image_classification_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2075 (class 2606 OID 16459)
-- Name: image_classification image_classification_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_classification
    ADD CONSTRAINT image_classification_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


--
-- TOC entry 2073 (class 2606 OID 16612)
-- Name: image image_image_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image
    ADD CONSTRAINT image_image_provider_id_fkey FOREIGN KEY (image_provider_id) REFERENCES image_provider(id);


--
-- TOC entry 2078 (class 2606 OID 16628)
-- Name: image_report image_report_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_report
    ADD CONSTRAINT image_report_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2076 (class 2606 OID 16475)
-- Name: image_validation image_validation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_image_id_fkey FOREIGN KEY (image_id) REFERENCES image(id);


--
-- TOC entry 2077 (class 2606 OID 16481)
-- Name: image_validation image_validation_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY image_validation
    ADD CONSTRAINT image_validation_label_id_fkey FOREIGN KEY (label_id) REFERENCES label(id);


-- Completed on 2017-09-13 20:55:58

--
-- PostgreSQL database dump complete
--

