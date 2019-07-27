--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.12
-- Dumped by pg_dump version 9.6.12

-- Started on 2019-07-27 16:46:59 CEST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 1 (class 3079 OID 12429)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 4300 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 2 (class 3079 OID 4440345)
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- TOC entry 4301 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';


--
-- TOC entry 4 (class 3079 OID 4440331)
-- Name: temporal_tables; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS temporal_tables WITH SCHEMA public;


--
-- TOC entry 4302 (class 0 OID 0)
-- Dependencies: 4
-- Name: EXTENSION temporal_tables; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION temporal_tables IS 'temporal tables';


--
-- TOC entry 3 (class 3079 OID 4440334)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 4303 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 1977 (class 1247 OID 4441846)
-- Name: agg_areaweightedstats; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.agg_areaweightedstats AS (
	count integer,
	distinctcount integer,
	geom public.geometry,
	totalarea double precision,
	meanarea double precision,
	totalperimeter double precision,
	meanperimeter double precision,
	weightedsum double precision,
	weightedmean double precision,
	maxareavalue double precision,
	minareavalue double precision,
	maxcombinedareavalue double precision,
	mincombinedareavalue double precision,
	sum double precision,
	mean double precision,
	max double precision,
	min double precision
);


ALTER TYPE public.agg_areaweightedstats OWNER TO postgres;

--
-- TOC entry 1980 (class 1247 OID 4441849)
-- Name: agg_areaweightedstatsstate; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.agg_areaweightedstatsstate AS (
	count integer,
	distinctvalues double precision[],
	unionedgeom public.geometry,
	totalarea double precision,
	totalperimeter double precision,
	weightedsum double precision,
	maxareavalue double precision[],
	minareavalue double precision[],
	combinedweightedareas double precision[],
	sum double precision,
	max double precision,
	min double precision
);


ALTER TYPE public.agg_areaweightedstatsstate OWNER TO postgres;

--
-- TOC entry 1983 (class 1247 OID 4441851)
-- Name: control_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.control_type AS ENUM (
    'dropdown',
    'checkbox',
    'radio',
    'color tags'
);


ALTER TYPE public.control_type OWNER TO postgres;

--
-- TOC entry 1986 (class 1247 OID 4441861)
-- Name: geomvaltxt; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.geomvaltxt AS (
	geom public.geometry,
	val double precision,
	txt text
);


ALTER TYPE public.geomvaltxt OWNER TO postgres;

--
-- TOC entry 1989 (class 1247 OID 4441863)
-- Name: label_bot_label_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.label_bot_label_type AS ENUM (
    'meta',
    'normal'
);


ALTER TYPE public.label_bot_label_type OWNER TO postgres;

--
-- TOC entry 1992 (class 1247 OID 4441868)
-- Name: label_bot_state_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.label_bot_state_type AS ENUM (
    'pending',
    'building',
    'build-failed',
    'build-success',
    'waiting for moderator approval',
    'accepted',
    'retry',
    'build-canceled',
    'already exists'
);


ALTER TYPE public.label_bot_state_type OWNER TO postgres;

--
-- TOC entry 1995 (class 1247 OID 4441888)
-- Name: label_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.label_type AS ENUM (
    'normal',
    'refinement',
    'refinement_category',
    'meta'
);


ALTER TYPE public.label_type OWNER TO postgres;

--
-- TOC entry 1998 (class 1247 OID 4441898)
-- Name: state_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.state_type AS ENUM (
    'unknown',
    'locked',
    'unlocked'
);


ALTER TYPE public.state_type OWNER TO postgres;

--
-- TOC entry 1500 (class 1255 OID 4441905)
-- Name: _st_areaweightedsummarystats_finalfn(public.agg_areaweightedstatsstate); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_areaweightedsummarystats_finalfn(aws public.agg_areaweightedstatsstate) RETURNS public.agg_areaweightedstats
    LANGUAGE plpgsql
    AS $_$
    DECLARE
        a RECORD;
        maxarea double precision = 0.0;
        minarea double precision = (($1).combinedweightedareas)[1];
        imax int := 1;
        imin int := 1;
        ret agg_areaweightedstats;
    BEGIN
        -- Search for the max and the min areas in the array of all distinct values
        FOR a IN SELECT n, (($1).combinedweightedareas)[n] warea
                 FROM generate_series(1, array_length(($1).combinedweightedareas, 1)) n LOOP
            IF a.warea > maxarea THEN
                imax := a.n;
                maxarea = a.warea;
            END IF;
            IF a.warea < minarea THEN
                imin := a.n;
                minarea = a.warea;
            END IF;
        END LOOP;

        ret := (($1).count,
                array_length(($1).distinctvalues, 1),
                ($1).unionedgeom,
                ($1).totalarea,
                ($1).totalarea / ($1).count,
                ($1).totalperimeter,
                ($1).totalperimeter / ($1).count,
                ($1).weightedsum,
                ($1).weightedsum / ($1).totalarea,
                (($1).maxareavalue)[2],
                (($1).minareavalue)[2],
                (($1).distinctvalues)[imax],
                (($1).distinctvalues)[imin],
                ($1).sum,
                ($1).sum / ($1).count,
                ($1).max,
                ($1).min
               )::agg_areaweightedstats;
        RETURN ret;
    END;
$_$;


ALTER FUNCTION public._st_areaweightedsummarystats_finalfn(aws public.agg_areaweightedstatsstate) OWNER TO postgres;

--
-- TOC entry 1501 (class 1255 OID 4441906)
-- Name: _st_areaweightedsummarystats_statefn(public.agg_areaweightedstatsstate, public.geometry); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_areaweightedsummarystats_statefn(aws public.agg_areaweightedstatsstate, geom public.geometry) RETURNS public.agg_areaweightedstatsstate
    LANGUAGE sql
    AS $_$
    SELECT _ST_AreaWeightedSummaryStats_StateFN($1, ($2, 1)::geomval);
$_$;


ALTER FUNCTION public._st_areaweightedsummarystats_statefn(aws public.agg_areaweightedstatsstate, geom public.geometry) OWNER TO postgres;

--
-- TOC entry 1502 (class 1255 OID 4441907)
-- Name: _st_areaweightedsummarystats_statefn(public.agg_areaweightedstatsstate, public.geomval); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_areaweightedsummarystats_statefn(aws public.agg_areaweightedstatsstate, gv public.geomval) RETURNS public.agg_areaweightedstatsstate
    LANGUAGE plpgsql
    AS $_$
    DECLARE
        i int;
        ret agg_areaweightedstatsstate;
        newcombinedweightedareas double precision[] := ($1).combinedweightedareas;
        newgeom geometry := ($2).geom;
        geomtype text := GeometryType(($2).geom);
    BEGIN
        -- If the geometry is a GEOMETRYCOLLECTION extract the polygon part
        IF geomtype = 'GEOMETRYCOLLECTION' THEN
            newgeom := ST_CollectionExtract(newgeom, 3);
        END IF;
        -- Skip anything that is not a polygon
        IF newgeom IS NULL OR ST_IsEmpty(newgeom) OR geomtype = 'POINT' OR geomtype = 'LINESTRING' OR geomtype = 'MULTIPOINT' OR geomtype = 'MULTILINESTRING' THEN
            ret := aws;
        -- At the first iteration the state parameter is always NULL
        ELSEIF $1 IS NULL THEN
            ret := (1,                                 -- count
                    ARRAY[($2).val],                   -- distinctvalues
                    newgeom,                           -- unionedgeom
                    ST_Area(newgeom),                  -- totalarea
                    ST_Perimeter(newgeom),             -- totalperimeter
                    ($2).val * ST_Area(newgeom),       -- weightedsum
                    ARRAY[ST_Area(newgeom), ($2).val], -- maxareavalue
                    ARRAY[ST_Area(newgeom), ($2).val], -- minareavalue
                    ARRAY[ST_Area(newgeom)],           -- combinedweightedareas
                    ($2).val,                          -- sum
                    ($2).val,                          -- max
                    ($2).val                           -- min
                   )::agg_areaweightedstatsstate;
        ELSE
            -- Search for the new value in the array of distinct values
            SELECT n
            FROM generate_series(1, array_length(($1).distinctvalues, 1)) n
            WHERE (($1).distinctvalues)[n] = ($2).val
            INTO i;

            -- If the value already exists, increment the corresponding area with the new area
            IF NOT i IS NULL THEN
                newcombinedweightedareas[i] := newcombinedweightedareas[i] + ST_Area(newgeom);
            END IF;
            ret := (($1).count + 1,                                     -- count
                    CASE WHEN i IS NULL                                 -- distinctvalues
                         THEN array_append(($1).distinctvalues, ($2).val)
                         ELSE ($1).distinctvalues
                    END,
                    ST_Union(($1).unionedgeom, newgeom),                -- unionedgeom
                    ($1).totalarea + ST_Area(newgeom),                  -- totalarea
                    ($1).totalperimeter + ST_Perimeter(newgeom),        -- totalperimeter
                    ($1).weightedsum + ($2).val * ST_Area(newgeom),     -- weightedsum
                    CASE WHEN ST_Area(newgeom) > (($1).maxareavalue)[1] -- maxareavalue
                         THEN ARRAY[ST_Area(newgeom), ($2).val]
                         ELSE ($1).maxareavalue
                    END,
                    CASE WHEN ST_Area(newgeom) < (($1).minareavalue)[1] -- minareavalue
                         THEN ARRAY[ST_Area(newgeom), ($2).val]
                         ELSE ($1).minareavalue
                    END,
                    CASE WHEN i IS NULL                                 -- combinedweightedareas
                         THEN array_append(($1).combinedweightedareas, ST_Area(newgeom))
                         ELSE newcombinedweightedareas
                    END,
                    ($1).sum + ($2).val,                                -- sum
                    greatest(($1).max, ($2).val),                       -- max
                    least(($1).min, ($2).val)                           -- min
                   )::agg_areaweightedstatsstate;
        END IF;
        RETURN ret;
    END;
$_$;


ALTER FUNCTION public._st_areaweightedsummarystats_statefn(aws public.agg_areaweightedstatsstate, gv public.geomval) OWNER TO postgres;

--
-- TOC entry 1503 (class 1255 OID 4441908)
-- Name: _st_areaweightedsummarystats_statefn(public.agg_areaweightedstatsstate, public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_areaweightedsummarystats_statefn(aws public.agg_areaweightedstatsstate, geom public.geometry, val double precision) RETURNS public.agg_areaweightedstatsstate
    LANGUAGE sql
    AS $_$
   SELECT _ST_AreaWeightedSummaryStats_StateFN($1, ($2, $3)::geomval);
$_$;


ALTER FUNCTION public._st_areaweightedsummarystats_statefn(aws public.agg_areaweightedstatsstate, geom public.geometry, val double precision) OWNER TO postgres;

--
-- TOC entry 1504 (class 1255 OID 4441909)
-- Name: _st_bufferedunion_finalfn(public.geomval); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_bufferedunion_finalfn(gv public.geomval) RETURNS public.geometry
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT ST_Buffer(($1).geom, -($1).val, 'endcap=square join=mitre')
$_$;


ALTER FUNCTION public._st_bufferedunion_finalfn(gv public.geomval) OWNER TO postgres;

--
-- TOC entry 1506 (class 1255 OID 4441910)
-- Name: _st_bufferedunion_statefn(public.geomval, public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_bufferedunion_statefn(gv public.geomval, geom public.geometry, bufsize double precision DEFAULT 0.0) RETURNS public.geomval
    LANGUAGE sql IMMUTABLE
    AS $_$
    SELECT CASE WHEN $1 IS NULL AND $2 IS NULL THEN
                    NULL
                WHEN $1 IS NULL THEN
                    (ST_Buffer($2, CASE WHEN $3 IS NULL THEN 0.0 ELSE $3 END, 'endcap=square join=mitre'),
                     CASE WHEN $3 IS NULL THEN 0.0 ELSE $3 END
                    )::geomval
                WHEN $2 IS NULL THEN
                    $1
                ELSE (ST_Union(($1).geom,
                           ST_Buffer($2, CASE WHEN $3 IS NULL THEN 0.0 ELSE $3 END, 'endcap=square join=mitre')
                          ),
                  ($1).val
                 )::geomval
       END;
$_$;


ALTER FUNCTION public._st_bufferedunion_statefn(gv public.geomval, geom public.geometry, bufsize double precision) OWNER TO postgres;

--
-- TOC entry 1507 (class 1255 OID 4441911)
-- Name: _st_differenceagg_statefn(public.geometry, public.geometry, public.geometry); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_differenceagg_statefn(geom1 public.geometry, geom2 public.geometry, geom3 public.geometry) RETURNS public.geometry
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
       newgeom geometry;
       differ geometry;
    BEGIN
        -- First pass: geom1 is NULL
        IF geom1 IS NULL AND NOT ST_IsEmpty(geom2) THEN 
            IF geom3 IS NULL OR ST_Area(geom3) = 0 THEN
                newgeom = geom2;
            ELSE
                newgeom = CASE
                              WHEN ST_Area(ST_Intersection(geom2, geom3)) = 0 OR ST_IsEmpty(ST_Intersection(geom2, geom3)) THEN geom2
                              ELSE ST_Difference(geom2, geom3)
                           END;
            END IF;
        ELSIF NOT ST_IsEmpty(geom1) AND ST_Area(geom3) > 0 THEN
            BEGIN
                differ = ST_Difference(geom1, geom3);
            EXCEPTION
            WHEN OTHERS THEN
                BEGIN
                    RAISE NOTICE 'ST_DifferenceAgg(): Had to buffer geometries by 0.000001 to compute the difference...';
                    differ = ST_Difference(ST_Buffer(geom1, 0.000001), ST_Buffer(geom3, 0.000001));
                EXCEPTION
                WHEN OTHERS THEN
                    BEGIN
                        RAISE NOTICE 'ST_DifferenceAgg(): Had to buffer geometries by 0.00001 to compute the difference...';
                        differ = ST_Difference(ST_Buffer(geom1, 0.00001), ST_Buffer(geom3, 0.00001));
                    EXCEPTION
                    WHEN OTHERS THEN
                        differ = geom1;
                    END;
                END;
            END;
            newgeom = CASE
                          WHEN ST_Area(ST_Intersection(geom1, geom3)) = 0 OR ST_IsEmpty(ST_Intersection(geom1, geom3)) THEN geom1
                          ELSE differ
                      END;
        ELSE
            newgeom = geom1;
        END IF;

        IF NOT ST_IsEmpty(newgeom) THEN
            newgeom = ST_CollectionExtract(newgeom, 3);
        END IF;

        IF newgeom IS NULL THEN
            newgeom = ST_GeomFromText('MULTIPOLYGON EMPTY', ST_SRID(geom2));
        END IF;

        RETURN newgeom;
    END;
$$;


ALTER FUNCTION public._st_differenceagg_statefn(geom1 public.geometry, geom2 public.geometry, geom3 public.geometry) OWNER TO postgres;

--
-- TOC entry 1508 (class 1255 OID 4441912)
-- Name: _st_removeoverlaps_finalfn(public.geomvaltxt[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_removeoverlaps_finalfn(gvtarray public.geomvaltxt[]) RETURNS public.geometry[]
    LANGUAGE sql
    AS $$
    WITH gvt AS (
         SELECT unnest(gvtarray) gvt
    ), geoms AS (
         SELECT ST_RemoveOverlaps(array_agg(((gvt).geom, (gvt).val)::geomval), max((gvt).txt)) geom
         FROM gvt
    )
    SELECT array_agg(geom) FROM geoms;
$$;


ALTER FUNCTION public._st_removeoverlaps_finalfn(gvtarray public.geomvaltxt[]) OWNER TO postgres;

--
-- TOC entry 1509 (class 1255 OID 4441913)
-- Name: _st_removeoverlaps_statefn(public.geomvaltxt[], public.geometry); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry) RETURNS public.geomvaltxt[]
    LANGUAGE sql
    AS $_$
    SELECT _ST_RemoveOverlaps_StateFN($1, geom, NULL, 'NO_MERGE');
$_$;


ALTER FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry) OWNER TO postgres;

--
-- TOC entry 1510 (class 1255 OID 4441914)
-- Name: _st_removeoverlaps_statefn(public.geomvaltxt[], public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry, val double precision) RETURNS public.geomvaltxt[]
    LANGUAGE sql
    AS $_$
    SELECT _ST_RemoveOverlaps_StateFN($1, $2, $3, 'LARGEST_VALUE');
$_$;


ALTER FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry, val double precision) OWNER TO postgres;

--
-- TOC entry 1511 (class 1255 OID 4441915)
-- Name: _st_removeoverlaps_statefn(public.geomvaltxt[], public.geometry, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry, mergemethod text) RETURNS public.geomvaltxt[]
    LANGUAGE sql
    AS $_$
    SELECT _ST_RemoveOverlaps_StateFN($1, $2, ST_Area($2), $3);
$_$;


ALTER FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry, mergemethod text) OWNER TO postgres;

--
-- TOC entry 1512 (class 1255 OID 4441916)
-- Name: _st_removeoverlaps_statefn(public.geomvaltxt[], public.geometry, double precision, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry, val double precision, mergemethod text) RETURNS public.geomvaltxt[]
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
        newgvtarray geomvaltxt[];
    BEGIN
        IF gvtarray IS NULL THEN
            RETURN array_append(newgvtarray, (geom, val, mergemethod)::geomvaltxt);
        END IF;
    RETURN array_append(gvtarray, (geom, val, mergemethod)::geomvaltxt);
    END;
$$;


ALTER FUNCTION public._st_removeoverlaps_statefn(gvtarray public.geomvaltxt[], geom public.geometry, val double precision, mergemethod text) OWNER TO postgres;

--
-- TOC entry 1513 (class 1255 OID 4441917)
-- Name: _st_splitagg_statefn(public.geometry[], public.geometry, public.geometry); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_splitagg_statefn(geomarray public.geometry[], geom1 public.geometry, geom2 public.geometry) RETURNS public.geometry[]
    LANGUAGE sql
    AS $_$
    SELECT _ST_SplitAgg_StateFN($1, $2, $3, 0.0);
$_$;


ALTER FUNCTION public._st_splitagg_statefn(geomarray public.geometry[], geom1 public.geometry, geom2 public.geometry) OWNER TO postgres;

--
-- TOC entry 1514 (class 1255 OID 4441918)
-- Name: _st_splitagg_statefn(public.geometry[], public.geometry, public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._st_splitagg_statefn(geomarray public.geometry[], geom1 public.geometry, geom2 public.geometry, tolerance double precision) RETURNS public.geometry[]
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
        newgeomarray geometry[];
        geom3 geometry;
        newgeom geometry;
        geomunion geometry;
    BEGIN
        -- First pass: geomarray is NULL
       IF geomarray IS NULL THEN
            geomarray = array_append(newgeomarray, geom1);
        END IF;

        IF NOT geom2 IS NULL THEN
            -- 2) Each geometry in the array - geom2
            FOREACH geom3 IN ARRAY geomarray LOOP
                newgeom = ST_Difference(geom3, geom2);
                IF tolerance > 0 THEN
                    newgeom = ST_TrimMulti(newgeom, tolerance);
                END IF;
                IF NOT newgeom IS NULL AND NOT ST_IsEmpty(newgeom) THEN
                    newgeomarray = array_append(newgeomarray, newgeom);
                END IF;
            END LOOP;

        -- 3) gv1 intersecting each geometry in the array
            FOREACH geom3 IN ARRAY geomarray LOOP
                newgeom = ST_Intersection(geom3, geom2);
                IF tolerance > 0 THEN
                    newgeom = ST_TrimMulti(newgeom, tolerance);
                END IF;
                IF NOT newgeom IS NULL AND NOT ST_IsEmpty(newgeom) THEN
                    newgeomarray = array_append(newgeomarray, newgeom);
                END IF;
            END LOOP;
        ELSE
            newgeomarray = geomarray;
        END IF;
        RETURN newgeomarray;
    END;
$$;


ALTER FUNCTION public._st_splitagg_statefn(geomarray public.geometry[], geom1 public.geometry, geom2 public.geometry, tolerance double precision) OWNER TO postgres;

--
-- TOC entry 1515 (class 1255 OID 4441919)
-- Name: ellipse(double precision, double precision, double precision, double precision, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ellipse(x double precision, y double precision, rx double precision, ry double precision, rotation double precision) RETURNS public.geometry
    LANGUAGE sql
    AS $$
   SELECT ST_Translate( ST_Rotate( ST_Scale( ST_Buffer(ST_MakePoint(0,0)::geometry, 0.5)::geometry, rx, ry), rotation), x, y)       
$$;


ALTER FUNCTION public.ellipse(x double precision, y double precision, rx double precision, ry double precision, rotation double precision) OWNER TO postgres;

--
-- TOC entry 1516 (class 1255 OID 4441920)
-- Name: sp_get_image_annotation_coverage(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.sp_get_image_annotation_coverage(imageid text DEFAULT NULL::text) RETURNS TABLE(image_id bigint, area integer, annotated_percentage integer)
    LANGUAGE plpgsql
    AS $_$
DECLARE
  imageid ALIAS FOR $1;
  sql VARCHAR;
  sql_where_cond VARCHAR;
BEGIN
	
	sql_where_cond := '';
	IF imageid IS NOT NULL THEN
		sql_where_cond := 'AND i.key = ' || quote_literal(imageid);
	END IF;

	sql := 'WITH all_annotations AS (
			    SELECT an.image_id as image_id, d.id as annotation_data_id, d.annotation as annotation, t.name as annotation_type
			    FROM image_annotation an 
			    JOIN annotation_data d ON d.image_annotation_id = an.id
			    JOIN annotation_type t ON t.id = d.annotation_type_id
			    JOIN image i ON i.id = an.image_id
			    WHERE i.unlocked = true AND an.auto_generated = false ' 
			|| sql_where_cond || 
			'), 
			ellipse_annotations AS (
			    SELECT a.image_id, a.annotation_data_id as id, 
			    Ellipse( (a.annotation->''left'')::text::float, 
			             (a.annotation->''top'')::text::float, 
			             2* (a.annotation->''rx'')::text::float, 
			             2* (a.annotation->''ry'')::text::float, 
			             CASE 
			                WHEN a.annotation->''angle'' is null THEN 0 
			                ELSE (a.annotation->''angle'')::text::float
			             END
			           ) as geom
			    FROM all_annotations a 
			    WHERE annotation_type = ''ellipse''
			),
			polygon_annotations AS (
			  -- ST_MakePolygon might return a polygon with intersecting points. In order to fix that, one needs to call ST_MakeValid on the resulting polygon.
			  --Unfortunately, this is _really_ slow (especially, if a lot of polygons are affected). In order to circumvent that, we create a ConvexHull around the
			  --polygon. This works way faster and should also be precise enough for our purpose.
				SELECT q.image_id, q.annotation_data_id as id, ST_ConvexHull(ST_MakePolygon(ST_GeomFromText(''LINESTRING('' || 
			                                                                  string_agg((((q.annotation->''x'')::text) || '' '' || ((q.annotation->''y'')::text)), '','') 
			                                                                  || '','' || (array_agg((q.annotation->''x'')::text))[1] || '' '' || (array_agg((q.annotation->''y'')::text))[1] 
			                                                                  || '')''))) as geom
			    FROM
			    (
			        SELECT a.image_id, a.annotation_data_id, jsonb_array_elements(a.annotation->''points'') as  annotation
			        FROM all_annotations a 
			        WHERE a.annotation_type = ''polygon'' AND jsonb_array_length(a.annotation->''points'') > 2
			    ) q
			    GROUP BY q.image_id, q.annotation_data_id
			),
			rectangle_annotations AS (
			    SELECT a.image_id, a.annotation_data_id as id, ST_MakePolygon(ST_MakeLine(
			       ARRAY[
			             ST_MakePoint((a.annotation->''left'')::text::integer, (a.annotation->''top'')::text::integer), 
			             ST_MakePoint((a.annotation->''left'')::text::float + (a.annotation->''width'')::text::float, (a.annotation->''top'')::text::float),
			             ST_MakePoint((a.annotation->''left'')::text::float + (a.annotation->''width'')::text::float, 
			                                                    (a.annotation->''top'')::text::float + (a.annotation->''height'')::text::float),
			             ST_MakePoint((a.annotation->''left'')::text::float, (a.annotation->''top'')::text::float + (a.annotation->''height'')::text::float),
			             ST_MakePoint((a.annotation->''left'')::text::float, (a.annotation->''top'')::text::float)
			            ])) as geom
			    FROM all_annotations a 
			    WHERE a.annotation_type = ''rect''
			    --GROUP BY a.annotation_data_id, a.annotation
			),
			all_annotation_areas AS (
			    SELECT id, image_id, geom from polygon_annotations
			    UNION 
			    SELECT id, image_id, geom from rectangle_annotations
			    UNION
			    SELECT id, image_id, geom from ellipse_annotations
			)
			SELECT i.id as image_id, round(sum(q.area))::integer as area, round(((SUM(q.area)/(i.width * i.height)) * 100))::integer as annotated_percentage
			    FROM
			    (                                                                                   
			        SELECT a.id, a.image_id, ST_Area(ST_DifferenceAgg(a.geom, b.geom)) as area
			        FROM all_annotation_areas a
			        LEFT JOIN all_annotation_areas b
			        ON (ST_Contains(a.geom, b.geom) OR -- Select all the containing, contained and overlapping polygons
			            ST_Contains(b.geom, a.geom) OR
			            ST_Overlaps(a.geom, b.geom)) AND
			            (ST_Area(a.geom) < ST_Area(b.geom) OR -- Make sure bigger polygons are removed from smaller ones
			            (ST_Area(a.geom) = ST_Area(b.geom) AND -- If areas are equal, arbitrarily remove one from the other but in a determined order so it''s not done twice.
			              a.id < b.id)) AND (a.image_id = b.image_id)
			        GROUP BY a.id, a.image_id
			        HAVING ST_Area(ST_DifferenceAgg(a.geom, b.geom)) > 0 AND NOT ST_IsEmpty(ST_DifferenceAgg(a.geom, b.geom))
			    ) q
			    JOIN image i ON q.image_id = i.id
			    GROUP BY i.id';

	RETURN QUERY EXECUTE sql;

END;
$_$;


ALTER FUNCTION public.sp_get_image_annotation_coverage(imageid text) OWNER TO postgres;

--
-- TOC entry 1517 (class 1255 OID 4441922)
-- Name: st_adduniqueid(name, name, boolean, boolean); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_adduniqueid(tablename name, columnname name, replacecolumn boolean DEFAULT false, indexit boolean DEFAULT true) RETURNS boolean
    LANGUAGE sql
    AS $_$
    SELECT ST_AddUniqueID('public', $1, $2, $3, $4)
$_$;


ALTER FUNCTION public.st_adduniqueid(tablename name, columnname name, replacecolumn boolean, indexit boolean) OWNER TO postgres;

--
-- TOC entry 1519 (class 1255 OID 4441923)
-- Name: st_adduniqueid(name, name, name, boolean, boolean); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_adduniqueid(schemaname name, tablename name, columnname name, replacecolumn boolean DEFAULT false, indexit boolean DEFAULT true) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
    DECLARE
        seqname text;
        fqtn text;
    BEGIN
        IF replacecolumn IS NULL THEN
            replacecolumn = false;
        END IF;
        IF indexit IS NULL THEN
            indexit = true;
        END IF;
         -- Determine the complete name of the table
        fqtn := '';
        IF length(schemaname) > 0 THEN
            fqtn := quote_ident(schemaname) || '.';
        END IF;
        fqtn := fqtn || quote_ident(tablename);

        -- Check if the requested column name already exists
        IF ST_ColumnExists(schemaname, tablename, columnname) THEN
            IF replacecolumn THEN
                EXECUTE 'ALTER TABLE ' || fqtn || ' DROP COLUMN ' || columnname;
            ELSE
                RAISE NOTICE 'Column already exist. Set the ''replacecolumn'' argument to ''true'' if you want to replace the column.';
                RETURN false;
            END IF;
        END IF;

        -- Create a new sequence
        seqname = schemaname || '_' || tablename || '_seq';
        EXECUTE 'DROP SEQUENCE IF EXISTS ' || quote_ident(seqname);
        EXECUTE 'CREATE SEQUENCE ' || quote_ident(seqname);

        -- Add the new column and update it with nextval('sequence')
        EXECUTE 'ALTER TABLE ' || fqtn || ' ADD COLUMN ' || columnname || ' INTEGER';
        EXECUTE 'UPDATE ' || fqtn || ' SET ' || columnname || ' = nextval(''' || seqname || ''')';

        IF indexit THEN
            EXECUTE 'CREATE INDEX ' || tablename || '_' || columnname || '_idx ON ' || fqtn || ' USING btree(' || columnname || ');';
        END IF;

        RETURN true;
    END;
$$;


ALTER FUNCTION public.st_adduniqueid(schemaname name, tablename name, columnname name, replacecolumn boolean, indexit boolean) OWNER TO postgres;

--
-- TOC entry 1520 (class 1255 OID 4441924)
-- Name: st_bufferedsmooth(public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_bufferedsmooth(geom public.geometry, bufsize double precision DEFAULT 0) RETURNS public.geometry
    LANGUAGE sql IMMUTABLE
    AS $_$
    SELECT ST_Buffer(ST_Buffer($1, $2), -$2)
$_$;


ALTER FUNCTION public.st_bufferedsmooth(geom public.geometry, bufsize double precision) OWNER TO postgres;

--
-- TOC entry 1521 (class 1255 OID 4441925)
-- Name: st_columnexists(name, name); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_columnexists(tablename name, columnname name) RETURNS boolean
    LANGUAGE sql STRICT
    AS $_$
    SELECT ST_ColumnExists('public', $1, $2)
$_$;


ALTER FUNCTION public.st_columnexists(tablename name, columnname name) OWNER TO postgres;

--
-- TOC entry 1522 (class 1255 OID 4441926)
-- Name: st_columnexists(name, name, name); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_columnexists(schemaname name, tablename name, columnname name) RETURNS boolean
    LANGUAGE plpgsql STRICT
    AS $$
    DECLARE
    BEGIN
        PERFORM 1 FROM information_schema.COLUMNS
        WHERE lower(table_schema) = lower(schemaname) AND lower(table_name) = lower(tablename) AND lower(column_name) = lower(columnname);
        RETURN FOUND;
    END;
$$;


ALTER FUNCTION public.st_columnexists(schemaname name, tablename name, columnname name) OWNER TO postgres;

--
-- TOC entry 1523 (class 1255 OID 4441927)
-- Name: st_columnisunique(name, name); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_columnisunique(tablename name, columnname name) RETURNS boolean
    LANGUAGE sql STRICT
    AS $_$
    SELECT ST_ColumnIsUnique('public', $1, $2)
$_$;


ALTER FUNCTION public.st_columnisunique(tablename name, columnname name) OWNER TO postgres;

--
-- TOC entry 1524 (class 1255 OID 4441928)
-- Name: st_columnisunique(name, name, name); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_columnisunique(schemaname name, tablename name, columnname name) RETURNS boolean
    LANGUAGE plpgsql STRICT
    AS $$
    DECLARE
        newschemaname text;
        fqtn text;
        query text;
        isunique boolean;
    BEGIN
        newschemaname := '';
        IF length(schemaname) > 0 THEN
            newschemaname := schemaname;
        ELSE
            newschemaname := 'public';
        END IF;
        fqtn := quote_ident(newschemaname) || '.' || quote_ident(tablename);

        IF NOT ST_ColumnExists(newschemaname, tablename, columnname) THEN
            RAISE NOTICE 'ST_ColumnIsUnique(): Column ''%'' does not exist... Returning NULL', columnname;
            RETURN NULL;
        END IF;

        query = 'SELECT FALSE FROM ' || fqtn || ' GROUP BY ' || columnname || ' HAVING count(' || columnname || ') > 1 LIMIT 1';
        EXECUTE QUERY query INTO isunique;
        IF isunique IS NULL THEN
              isunique = TRUE;
        END IF;
        RETURN isunique;
    END;
$$;


ALTER FUNCTION public.st_columnisunique(schemaname name, tablename name, columnname name) OWNER TO postgres;

--
-- TOC entry 1525 (class 1255 OID 4441929)
-- Name: st_createindexraster(public.raster, text, integer, boolean, boolean, boolean, boolean, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_createindexraster(rast public.raster, pixeltype text DEFAULT '32BUI'::text, startvalue integer DEFAULT 0, incwithx boolean DEFAULT true, incwithy boolean DEFAULT true, rowsfirst boolean DEFAULT true, rowscanorder boolean DEFAULT true, colinc integer DEFAULT NULL::integer, rowinc integer DEFAULT NULL::integer) RETURNS public.raster
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
        newraster raster := ST_AddBand(ST_MakeEmptyRaster(rast), pixeltype);
        x int;
        y int;
        w int := ST_Width(newraster);
        h int := ST_Height(newraster);
        rowincx int := Coalesce(rowinc, w);
        colincx int := Coalesce(colinc, h);
        rowincy int := Coalesce(rowinc, 1);
        colincy int := Coalesce(colinc, 1);
        xdir int := CASE WHEN Coalesce(incwithx, true) THEN 1 ELSE w END;
        ydir int := CASE WHEN Coalesce(incwithy, true) THEN 1 ELSE h END;
        xdflag int := Coalesce(incwithx::int, 1);
        ydflag int := Coalesce(incwithy::int, 1);
        rsflag int := Coalesce(rowscanorder::int, 1);
        newstartvalue int := Coalesce(startvalue, 0);
        newrowsfirst boolean := Coalesce(rowsfirst, true);
    BEGIN
        IF newrowsfirst THEN
            IF colincx <= (h - 1) * rowincy THEN
                RAISE EXCEPTION 'Column increment (now %) must be greater than the number of index on one column (now % pixel x % = %)...', colincx, h - 1, rowincy, (h - 1) * rowincy;
            END IF;
            --RAISE NOTICE 'abs([rast.x] - %) * % + abs([rast.y] - (% ^ ((abs([rast.x] - % + 1) % 2) | % # ))::int) * % + %', xdir::text, colincx::text, h::text, xdir::text, rsflag::text, ydflag::text, rowincy::text, newstartvalue::text;
            newraster = ST_SetBandNodataValue(
                          ST_MapAlgebra(newraster,
                                        pixeltype,
                                        'abs([rast.x] - ' || xdir::text || ') * ' || colincx::text ||
                                        ' + abs([rast.y] - (' || h::text || ' ^ ((abs([rast.x] - ' ||
                                        xdir::text || ' + 1) % 2) | ' || rsflag::text || ' # ' ||
                                        ydflag::text || '))::int) * ' || rowincy::text || ' + ' || newstartvalue::text),
                          ST_BandNodataValue(newraster)
                        );
        ELSE
            IF rowincx <= (w - 1) * colincy THEN
                RAISE EXCEPTION 'Row increment (now %) must be greater than the number of index on one row (now % pixel x % = %)...', rowincx, w - 1, colincy, (w - 1) * colincy;
            END IF;
            newraster = ST_SetBandNodataValue(
                          ST_MapAlgebra(newraster,
                                        pixeltype,
                                        'abs([rast.x] - (' || w::text || ' ^ ((abs([rast.y] - ' ||
                                        ydir::text || ' + 1) % 2) | ' || rsflag::text || ' # ' ||
                                        xdflag::text || '))::int) * ' || colincy::text || ' + abs([rast.y] - ' ||
                                        ydir::text || ') * ' || rowincx::text || ' + ' || newstartvalue::text),
                          ST_BandNodataValue(newraster)
                        );
        END IF;
        RETURN newraster;
    END;
$$;


ALTER FUNCTION public.st_createindexraster(rast public.raster, pixeltype text, startvalue integer, incwithx boolean, incwithy boolean, rowsfirst boolean, rowscanorder boolean, colinc integer, rowinc integer) OWNER TO postgres;

--
-- TOC entry 1526 (class 1255 OID 4441930)
-- Name: st_deleteband(public.raster, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_deleteband(rast public.raster, band integer) RETURNS public.raster
    LANGUAGE plpgsql
    AS $$
    DECLARE
        numband int := ST_NumBands(rast);
        bandarray int[];
    BEGIN
        IF rast IS NULL THEN
            RETURN NULL;
        END IF;
        IF band IS NULL OR band < 1 OR band > numband THEN
            RETURN rast;
        END IF;
        IF band = 1 AND numband = 1 THEN
            RETURN ST_MakeEmptyRaster(rast);
        END IF;

        -- Construct the array of band to extract skipping the band to delete
        SELECT array_agg(i) INTO bandarray
        FROM generate_series(1, numband) i
        WHERE i != band;

        RETURN ST_Band(rast, bandarray);
    END;
$$;


ALTER FUNCTION public.st_deleteband(rast public.raster, band integer) OWNER TO postgres;

--
-- TOC entry 1527 (class 1255 OID 4441931)
-- Name: st_extractpixelcentroidvalue4ma(double precision[], integer[], text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_extractpixelcentroidvalue4ma(pixel double precision[], pos integer[], VARIADIC args text[]) RETURNS double precision
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
        pixelgeom text;
        result float4;
        query text;
    BEGIN
        -- args[1] = raster width
        -- args[2] = raster height
        -- args[3] = raster upperleft x
        -- args[4] = raster upperleft y
        -- args[5] = raster scale x
        -- args[6] = raster scale y
        -- args[7] = raster skew x
        -- args[8] = raster skew y
        -- args[9] = raster SRID
        -- args[10] = geometry or raster table schema name
        -- args[11] = geometry or raster table name
        -- args[12] = geometry or raster table geometry or raster column name
        -- args[13] = geometry table value column name
        -- args[14] = method

        -- Reconstruct the pixel centroid
        pixelgeom = ST_AsText(
                      ST_Centroid(
                        ST_PixelAsPolygon(
                          ST_MakeEmptyRaster(args[1]::integer,  -- raster width
                                             args[2]::integer,  -- raster height
                                             args[3]::float,    -- raster upperleft x
                                             args[4]::float,    -- raster upperleft y
                                             args[5]::float,    -- raster scale x
                                             args[6]::float,    -- raster scale y
                                             args[7]::float,    -- raster skew x
                                             args[8]::float,    -- raster skew y
                                             args[9]::integer   -- raster SRID
                                            ),
                                          pos[0][1]::integer, -- x coordinate of the current pixel
                                          pos[0][2]::integer  -- y coordinate of the current pixel
                                         )));

        -- Query the appropriate value
        IF args[14] = 'COUNT_OF_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT count(' || quote_ident(args[13]) ||
                    ') FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'MEAN_OF_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT avg(' || quote_ident(args[13]) ||
                    ') FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';
        ----------------------------------------------------------------
        -- Methods for the ST_GlobalRasterUnion() function
        ----------------------------------------------------------------
        ELSEIF args[14] = 'COUNT_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT count(ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'FIRST_RASTER_VALUE_AT_PIXEL_CENTROID' THEN
            query = 'SELECT ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || '))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ') LIMIT 1';

        ELSEIF args[14] = 'MIN_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT min(ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'MAX_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT max(ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'SUM_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT sum(ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'MEAN_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT avg(ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'STDDEVP_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT stddev_pop(ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')))
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'RANGE_OF_RASTER_VALUES_AT_PIXEL_CENTROID' THEN
            query = 'SELECT max(val) - min(val)
                     FROM (SELECT ST_Value(' || quote_ident(args[12]) || ', ST_GeomFromText(' || quote_literal(pixelgeom) ||
                    ', ' || args[9] || ')) val
                    FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                    quote_ident(args[12]) || ')) foo';

        ELSE
            query = 'SELECT NULL';
        END IF;
--RAISE NOTICE 'query = %', query;
        EXECUTE query INTO result;
        RETURN result;
    END;
$$;


ALTER FUNCTION public.st_extractpixelcentroidvalue4ma(pixel double precision[], pos integer[], VARIADIC args text[]) OWNER TO postgres;

--
-- TOC entry 1389 (class 1255 OID 4441932)
-- Name: st_extractpixelvalue4ma(double precision[], integer[], text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_extractpixelvalue4ma(pixel double precision[], pos integer[], VARIADIC args text[]) RETURNS double precision
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
        pixelgeom text;
        result float4;
        query text;
    BEGIN
        -- args[1] = raster width
        -- args[2] = raster height
        -- args[3] = raster upperleft x
        -- args[4] = raster upperleft y
        -- args[5] = raster scale x
        -- args[6] = raster scale y
        -- args[7] = raster skew x
        -- args[8] = raster skew y
        -- args[9] = raster SRID
        -- args[10] = geometry table schema name
        -- args[11] = geometry table name
        -- args[12] = geometry table geometry column name
        -- args[13] = geometry table value column name
        -- args[14] = method

--RAISE NOTICE 'val = %', pixel[1][1][1];
--RAISE NOTICE 'y = %, x = %', pos[0][1], pos[0][2];
        -- Reconstruct the pixel square
    pixelgeom = ST_AsText(
                  ST_PixelAsPolygon(
                    ST_MakeEmptyRaster(args[1]::integer, -- raster width
                                       args[2]::integer, -- raster height
                                       args[3]::float,   -- raster upperleft x
                                       args[4]::float,   -- raster upperleft y
                                       args[5]::float,   -- raster scale x
                                       args[6]::float,   -- raster scale y
                                       args[7]::float,   -- raster skew x
                                       args[8]::float,   -- raster skew y
                                       args[9]::integer  -- raster SRID
                                      ),
                                    pos[0][1]::integer, -- x coordinate of the current pixel
                                    pos[0][2]::integer  -- y coordinate of the current pixel
                                   ));
        -- Query the appropriate value
        IF args[14] = 'COUNT_OF_POLYGONS' THEN -- Number of polygons intersecting the pixel
            query = 'SELECT count(*) FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE (ST_GeometryType(' || quote_ident(args[12]) || ') = ''ST_Polygon'' OR
                             ST_GeometryType(' || quote_ident(args[12]) || ') = ''ST_MultiPolygon'') AND
                            ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                            || quote_ident(args[12]) || ') AND
                            ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                            quote_ident(args[12]) || ')) > 0.0000000001';

        ELSEIF args[14] = 'COUNT_OF_LINESTRINGS' THEN -- Number of linestring intersecting the pixel
            query = 'SELECT count(*) FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE (ST_GeometryType(' || quote_ident(args[12]) || ') = ''ST_LineString'' OR
                             ST_GeometryType(' || quote_ident(args[12]) || ') = ''ST_MultiLineString'') AND
                             ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                             || quote_ident(args[12]) || ') AND
                             ST_Length(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), ' ||
                             quote_ident(args[12]) || ')) > 0.0000000001';

        ELSEIF args[14] = 'COUNT_OF_POINTS' THEN -- Number of points intersecting the pixel
            query = 'SELECT count(*) FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE (ST_GeometryType(' || quote_ident(args[12]) || ') = ''ST_Point'' OR
                             ST_GeometryType(' || quote_ident(args[12]) || ') = ''ST_MultiPoint'') AND
                             ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                             || quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'COUNT_OF_GEOMETRIES' THEN -- Number of geometries intersecting the pixel
            query = 'SELECT count(*) FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) || ')';

        ELSEIF args[14] = 'VALUE_OF_BIGGEST' THEN -- Value of the geometry covering the biggest area in the pixel
            query = 'SELECT ' || quote_ident(args[13]) ||
                    ' val FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ') ORDER BY ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '),
                                                        ' || quote_ident(args[12]) ||
                    ')) DESC, val DESC LIMIT 1';

        ELSEIF args[14] = 'VALUE_OF_MERGED_BIGGEST' THEN -- Value of the combined geometry covering the biggest area in the pixel
            query = 'SELECT val FROM (SELECT ' || quote_ident(args[13]) || ' val,
                                            sum(ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom)
                                            || ', '|| args[9] || '), ' || quote_ident(args[12]) ||
                    '))) sumarea FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ') GROUP BY val) foo ORDER BY sumarea DESC, val DESC LIMIT 1';

        ELSEIF args[14] = 'MIN_AREA' THEN -- Area of the geometry covering the smallest area in the pixel
            query = 'SELECT area FROM (SELECT ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '
                                                      || args[9] || '), ' || quote_ident(args[12]) ||
                    ')) area FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ')) foo WHERE area > 0.0000000001 ORDER BY area LIMIT 1';

        ELSEIF args[14] = 'VALUE_OF_MERGED_SMALLEST' THEN -- Value of the combined geometry covering the biggest area in the pixel
            query = 'SELECT val FROM (SELECT ' || quote_ident(args[13]) || ' val,
                                             sum(ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '
                                             || args[9] || '), ' || quote_ident(args[12]) ||
                    '))) sumarea FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ') AND ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                                                                     || quote_ident(args[12]) || ')) > 0.0000000001
                      GROUP BY val) foo ORDER BY sumarea ASC, val DESC LIMIT 1';

        ELSEIF args[14] = 'SUM_OF_AREAS' THEN -- Sum of areas intersecting with the pixel (no matter the value)
            query = 'SELECT sum(ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                                                                          || quote_ident(args[12]) ||
                    '))) sumarea FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ')';

        ELSEIF args[14] = 'SUM_OF_LENGTHS' THEN -- Sum of lengths intersecting with the pixel (no matter the value)
            query = 'SELECT sum(ST_Length(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                                                                          || quote_ident(args[12]) ||
                    '))) sumarea FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ')';

        ELSEIF args[14] = 'PROPORTION_OF_COVERED_AREA' THEN -- Proportion of the pixel covered by polygons (no matter the value)
            query = 'SELECT ST_Area(ST_Union(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                                                                               || quote_ident(args[12]) ||
                    ')))/ST_Area(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || ')) sumarea
                     FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                    ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                    || quote_ident(args[12]) ||
                    ')';

        ELSEIF args[14] = 'AREA_WEIGHTED_MEAN_OF_VALUES' THEN -- Mean of every geometry weighted by the area they cover
            query = 'SELECT CASE
                              WHEN sum(area) = 0 THEN 0
                              ELSE sum(area * val) /
                                   greatest(sum(area),
                                            ST_Area(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '))
                                           )
                            END
                     FROM (SELECT ' || quote_ident(args[13]) || ' val,
                                 ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                                                         || quote_ident(args[12]) || ')) area
                           FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                         ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                         || quote_ident(args[12]) ||
                    ')) foo';

        ELSEIF args[14] = 'AREA_WEIGHTED_MEAN_OF_VALUES_2' THEN -- Mean of every geometry weighted by the area they cover
            query = 'SELECT CASE
                              WHEN sum(area) = 0 THEN 0
                              ELSE sum(area * val) / sum(area)
                            END
                     FROM (SELECT ' || quote_ident(args[13]) || ' val,
                                 ST_Area(ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                                                         || quote_ident(args[12]) || ')) area
                           FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                         ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                         || quote_ident(args[12]) ||
                    ')) foo';
        ----------------------------------------------------------------
        -- Methods for the ST_GlobalRasterUnion() function
        ----------------------------------------------------------------
        ELSEIF args[14] = 'AREA_WEIGHTED_SUM_OF_RASTER_VALUES' THEN -- Sum of every pixel value weighted by the area they cover
            query = 'SELECT sum(ST_Area((gv).geom) * (gv).val)
                     FROM (SELECT ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', ' ||
                                                                   args[9] || '), ' || quote_ident(args[12]) || ') gv
                           FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                         ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                         || quote_ident(args[12]) ||
                    ')) foo';

        ELSEIF args[14] = 'SUM_OF_AREA_PROPORTIONAL_RASTER_VALUES' THEN -- Sum of the proportion of pixel values intersecting with the pixel
            query = 'SELECT sum(ST_Area((gv).geom) * (gv).val / geomarea)
                     FROM (SELECT ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', ' ||
                                                                  args[9] || '), ' || quote_ident(args[12]) || ') gv, abs(ST_ScaleX(' || quote_ident(args[12]) || ') * ST_ScaleY(' || quote_ident(args[12]) || ')) geomarea
                           FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                         ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                         || quote_ident(args[12]) ||
                    ')) foo1';

        ELSEIF args[14] = 'AREA_WEIGHTED_MEAN_OF_RASTER_VALUES' THEN -- Mean of every pixel value weighted by the maximum area they cover
            query = 'SELECT CASE
                              WHEN sum(area) = 0 THEN NULL
                              ELSE sum(area * val) /
                                   greatest(sum(area),
                                            ST_Area(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '))
                                           )
                            END
                     FROM (SELECT ST_Area((gv).geom) area, (gv).val val
                           FROM (SELECT ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', ' ||
                                                                        args[9] || '), ' || quote_ident(args[12]) || ') gv
                                 FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                               ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                         || quote_ident(args[12]) ||
                    ')) foo1) foo2';

        ELSEIF args[14] = 'AREA_WEIGHTED_MEAN_OF_RASTER_VALUES_2' THEN -- Mean of every pixel value weighted by the area they cover
            query = 'SELECT CASE
                              WHEN sum(area) = 0 THEN NULL
                              ELSE sum(area * val) / sum(area)
                            END
                     FROM (SELECT ST_Area((gv).geom) area, (gv).val val
                           FROM (SELECT ST_Intersection(ST_GeomFromText(' || quote_literal(pixelgeom) || ', ' ||
                                                                        args[9] || '), ' || quote_ident(args[12]) || ') gv
                                 FROM ' || quote_ident(args[10]) || '.' || quote_ident(args[11]) ||
                               ' WHERE ST_Intersects(ST_GeomFromText(' || quote_literal(pixelgeom) || ', '|| args[9] || '), '
                         || quote_ident(args[12]) ||
                    ')) foo1) foo2';

        ELSE
            query = 'SELECT NULL';
        END IF;
--RAISE NOTICE 'query = %', query;
        EXECUTE query INTO result;
        RETURN result;
    END;
$$;


ALTER FUNCTION public.st_extractpixelvalue4ma(pixel double precision[], pos integer[], VARIADIC args text[]) OWNER TO postgres;

--
-- TOC entry 1459 (class 1255 OID 4441934)
-- Name: st_extracttoraster(public.raster, name, name, name, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_extracttoraster(rast public.raster, schemaname name, tablename name, geomcolumnname name, method text DEFAULT 'MEAN_OF_VALUES_AT_PIXEL_CENTROID'::text) RETURNS public.raster
    LANGUAGE sql
    AS $_$
    SELECT ST_ExtractToRaster($1, 1, $2, $3, $4, NULL, $5)
$_$;


ALTER FUNCTION public.st_extracttoraster(rast public.raster, schemaname name, tablename name, geomcolumnname name, method text) OWNER TO postgres;

--
-- TOC entry 1479 (class 1255 OID 4441935)
-- Name: st_extracttoraster(public.raster, integer, name, name, name, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_extracttoraster(rast public.raster, band integer, schemaname name, tablename name, geomcolumnname name, method text DEFAULT 'MEAN_OF_VALUES_AT_PIXEL_CENTROID'::text) RETURNS public.raster
    LANGUAGE sql
    AS $_$
    SELECT ST_ExtractToRaster($1, $2, $3, $4, $5, NULL, $6)
$_$;


ALTER FUNCTION public.st_extracttoraster(rast public.raster, band integer, schemaname name, tablename name, geomcolumnname name, method text) OWNER TO postgres;

--
-- TOC entry 1488 (class 1255 OID 4441936)
-- Name: st_extracttoraster(public.raster, name, name, name, name, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_extracttoraster(rast public.raster, schemaname name, tablename name, geomcolumnname name, valuecolumnname name, method text DEFAULT 'MEAN_OF_VALUES_AT_PIXEL_CENTROID'::text) RETURNS public.raster
    LANGUAGE sql
    AS $_$
    SELECT ST_ExtractToRaster($1, 1, $2, $3, $4, $5, $6)
$_$;


ALTER FUNCTION public.st_extracttoraster(rast public.raster, schemaname name, tablename name, geomcolumnname name, valuecolumnname name, method text) OWNER TO postgres;

--
-- TOC entry 1528 (class 1255 OID 4441937)
-- Name: st_extracttoraster(public.raster, integer, name, name, name, name, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_extracttoraster(rast public.raster, band integer, schemaname name, tablename name, geomrastcolumnname name, valuecolumnname name, method text DEFAULT 'MEAN_OF_VALUES_AT_PIXEL_CENTROID'::text) RETURNS public.raster
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
    DECLARE
        query text;
        newrast raster;
        fct2call text;
        newvaluecolumnname text;
        intcount int;
    BEGIN
        -- Determine the name of the right callback function
        IF right(method, 5) = 'TROID' THEN
            fct2call = 'ST_ExtractPixelCentroidValue4ma';
        ELSE
            fct2call = 'ST_ExtractPixelValue4ma';
        END IF;

        IF valuecolumnname IS NULL THEN
            newvaluecolumnname = 'null';
        ELSE
            newvaluecolumnname = quote_literal(valuecolumnname);
        END IF;

        query = 'SELECT count(*) FROM "' || schemaname || '"."' || tablename || '" WHERE ST_Intersects($1, ' || geomrastcolumnname || ')';

        EXECUTE query INTO intcount USING rast;
        IF intcount = 0 THEN
            -- if the method should return 0 when there is no geometry involved, return a raster containing only zeros
            IF left(method, 6) = 'COUNT_' OR
               method = 'SUM_OF_AREAS' OR
               method = 'SUM_OF_LENGTHS' OR
               method = 'PROPORTION_OF_COVERED_AREA' THEN
                RETURN ST_AddBand(ST_DeleteBand(rast, band), ST_AddBand(ST_MakeEmptyRaster(rast), ST_BandPixelType(rast, band), 0, ST_BandNodataValue(rast, band)), 1, band);
            ELSE
                RETURN ST_AddBand(ST_DeleteBand(rast, band), ST_AddBand(ST_MakeEmptyRaster(rast), ST_BandPixelType(rast, band), ST_BandNodataValue(rast, band), ST_BandNodataValue(rast, band)), 1, band);
            END IF;
        END IF;

        query = 'SELECT ST_MapAlgebra($1,
                                      $2,
                                      ''' || fct2call || '(double precision[], integer[], text[])''::regprocedure,
                                      ST_BandPixelType($1, $2),
                                      NULL,
                                      NULL,
                                      NULL,
                                      NULL,
                                      ST_Width($1)::text,
                                      ST_Height($1)::text,
                                      ST_UpperLeftX($1)::text,
                                      ST_UpperLeftY($1)::text,
                                      ST_ScaleX($1)::text,
                                      ST_ScaleY($1)::text,
                                      ST_SkewX($1)::text,
                                      ST_SkewY($1)::text,
                                      ST_SRID($1)::text,' ||
                                      quote_literal(schemaname) || ', ' ||
                                      quote_literal(tablename) || ', ' ||
                                      quote_literal(geomrastcolumnname) || ', ' ||
                                      newvaluecolumnname || ', ' ||
                                      quote_literal(upper(method)) || '
                                     ) rast';
--RAISE NOTICE 'query = %', query;
        EXECUTE query INTO newrast USING rast, band;
        RETURN ST_AddBand(ST_DeleteBand(rast, band), newrast, 1, band);
    END
$_$;


ALTER FUNCTION public.st_extracttoraster(rast public.raster, band integer, schemaname name, tablename name, geomrastcolumnname name, valuecolumnname name, method text) OWNER TO postgres;

--
-- TOC entry 1529 (class 1255 OID 4441938)
-- Name: st_geotablesummary(name, name, name, name, integer, text[], text[], text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_geotablesummary(schemaname name, tablename name, geomcolumnname name DEFAULT 'geom'::name, uidcolumn name DEFAULT NULL::name, nbinterval integer DEFAULT 10, dosummary text[] DEFAULT NULL::text[], skipsummary text[] DEFAULT NULL::text[], whereclause text DEFAULT NULL::text) RETURNS TABLE(summary text, idsandtypes text, countsandareas double precision, query text, geom public.geometry)
    LANGUAGE plpgsql
    AS $$
    DECLARE
        fqtn text;
        query text;
        newschemaname name;
        summary text;
        vertex_summary record;
        area_summary record;
        findnewuidcolumn boolean := FALSE;
        newuidcolumn text;
        newuidcolumntype text;
        createidx boolean := FALSE;
        uidcolumncnt int := 0;
        whereclausewithwhere text := '';
        sval text[] = ARRAY['S1', 'IDDUP', 'S2', 'GDUP', 'GEODUP', 'S3', 'OVL', 'S4', 'GAPS', 'S5', 'TYPES', 'GTYPES', 'GEOTYPES', 'S6', 'VERTX', 'S7', 'VHISTO', 'S8', 'AREAS', 'AREA', 'S9', 'AHISTO', 'S10', 'SACOUNT', 'ALL'];
        dos1 text[] = ARRAY['S1', 'IDDUP', 'ALL'];
        dos2 text[] = ARRAY['S2', 'GDUP', 'GEODUP', 'ALL'];
        dos3 text[] = ARRAY['S3', 'OVL', 'ALL'];
        dos4 text[] = ARRAY['S4', 'GAPS', 'ALL'];
        dos5 text[] = ARRAY['S5', 'TYPES', 'GTYPES', 'GEOTYPES', 'ALL'];
        dos6 text[] = ARRAY['S6', 'VERTX', 'NPOINTS', 'ALL'];
        dos7 text[] = ARRAY['S7', 'VHISTO', 'ALL'];
        dos8 text[] = ARRAY['S8', 'AREAS', 'AREA', 'ALL'];
        dos9 text[] = ARRAY['S9', 'AHISTO', 'ALL'];
        dos10 text[] = ARRAY['S10', 'SACOUNT', 'ALL'];
        provided_uid_isunique boolean = FALSE;
        colnamearr text[];
        colnamearrlength int := 0;
        colnameidx int := 0;
        sum7nbinterval int;
        sum9nbinterval int;
        minarea double precision := 0;
        maxarea double precision := 0;
        minnp int := 0;
        maxnp int := 0;
        bydefault text;
    BEGIN
        IF geomcolumnname IS NULL THEN
            geomcolumnname = 'geom';
        END IF;
        IF nbinterval IS NULL THEN
            nbinterval = 10;
        END IF;
        IF whereclause IS NULL OR whereclause = '' THEN
            whereclause = '';
        ELSE
            whereclausewithwhere = ' WHERE ' || whereclause || ' ';
            whereclause = ' AND (' || whereclause || ')';
        END IF;
        newschemaname := '';
        IF length(schemaname) > 0 THEN
            newschemaname := schemaname;
        ELSE
            newschemaname := 'public';
        END IF;
        fqtn := quote_ident(newschemaname) || '.' || quote_ident(tablename);

        -- Validate the dosummary parameter
        IF (NOT dosummary IS NULL) THEN
            FOR i IN array_lower(dosummary, 1)..array_upper(dosummary, 1) LOOP
               dosummary[i] := upper(dosummary[i]);
            END LOOP;
            FOREACH summary IN ARRAY dosummary LOOP
                IF (NOT summary = ANY (sval)) THEN
                    RAISE EXCEPTION 'Invalid value ''%'' for the ''dosummary'' parameter...', summary;
                    RETURN;
                    EXIT;
                END IF;
            END LOOP;
        END IF;
        IF (NOT skipsummary IS NULL) THEN
            FOR i IN array_lower(skipsummary, 1)..array_upper(skipsummary, 1) LOOP
               skipsummary[i] := upper(skipsummary[i]);
            END LOOP;
            FOREACH summary IN ARRAY skipsummary LOOP
                IF (NOT summary = ANY (sval)) THEN
                    RAISE EXCEPTION 'Invalid value ''%'' for the ''skipsummary'' parameter...', summary;
                    RETURN;
                    EXIT;
                END IF;
            END LOOP;
        END IF;

        newuidcolumn = lower(uidcolumn);
        IF newuidcolumn IS NULL THEN
            newuidcolumn = 'id';
        END IF;

        -----------------------------------------------
        -- Display the number of rows selected
        query = 'SELECT  ''NUMBER OF ROWS SELECTED''::text summary, ''''::text idsandtypes, count(*)::double precision countsandareas, ''query''::text, NULL::geometry geom  FROM ' || fqtn || whereclausewithwhere;
        RETURN QUERY EXECUTE query;
        -----------------------------------------------
        -- Summary #1: Check for duplicate IDs (IDDUP)
        IF (dosummary IS NULL OR dosummary && dos1) AND (skipsummary IS NULL OR NOT (skipsummary && dos1)) THEN
            query = E'SELECT 1::text summary,\n'
                 || E'       ' || newuidcolumn || E'::text idsandtypes,\n'
                 || E'       count(*)::double precision countsandareas,\n'
                 || E'       ''SELECT * FROM ' || fqtn || ' WHERE ' || newuidcolumn || ' = '' || ' || newuidcolumn || E' || '';''::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM ' || fqtn || E'\n'
                 || ltrim(whereclausewithwhere) || CASE WHEN whereclausewithwhere = '' THEN '' ELSE E'\n' END
                 || E'GROUP BY ' || newuidcolumn || E'\n'
                 || E'HAVING count(*) > 1\n'
                 || E'ORDER BY countsandareas DESC;';

            RETURN QUERY SELECT 'SUMMARY 1 - DUPLICATE IDs (IDDUP or S1)'::text, ('DUPLICATE IDs (' || newuidcolumn::text || ')')::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 1 - Duplicate IDs (IDDUP or S1)...';

            IF ST_ColumnExists(newschemaname, tablename, newuidcolumn) THEN
                EXECUTE 'SELECT pg_typeof(' || newuidcolumn || ') FROM ' || fqtn || ' LIMIT 1' INTO newuidcolumntype;
                IF newuidcolumntype != 'geometry' AND newuidcolumntype != 'raster' THEN
                    RETURN QUERY EXECUTE query;
                    IF NOT FOUND THEN
                        RETURN QUERY SELECT '1'::text, 'No duplicate IDs...'::text, NULL::double precision, NULL::text, NULL::geometry;
                        provided_uid_isunique = TRUE;
                    END IF;
                ELSE
                    RETURN QUERY SELECT '1'::text, '''' || newuidcolumn::text || ''' is not of type numeric or text... Skipping Summary 1'::text, NULL::double precision, NULL::text, NULL::geometry;
                END IF;
            ELSE
                RETURN QUERY SELECT '1'::text, '''' || newuidcolumn::text || ''' does not exists... Skipping Summary 1'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            RETURN QUERY SELECT 'SUMMARY 1 - DUPLICATE IDs (IDDUP or S1)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 1 - Skipping Duplicate IDs (IDDUP or S1)...';
        END IF;

        -----------------------------------------------
        -- Add a unique id column if it does not exists or if the one provided is not unique
        IF (dosummary IS NULL OR dosummary && dos2 OR dosummary && dos3 OR dosummary && dos4) AND (skipsummary IS NULL OR NOT (skipsummary && dos2 AND skipsummary && dos3 AND skipsummary && dos4)) THEN

            RAISE NOTICE 'Searching for the first column containing unique values...';

            -- Construct the list of available column names (integer only)
            query = 'SELECT array_agg(column_name::text) FROM information_schema.columns WHERE table_schema = ''' || newschemaname || ''' AND table_name = ''' || tablename || ''' AND data_type = ''integer'';';
            EXECUTE query INTO colnamearr;
            colnamearrlength = array_length(colnamearr, 1);

            RAISE NOTICE '  Checking ''%''...', newuidcolumn;

            -- Search for a unique id. Search first for 'id', if no uidcolumn name is provided, or for the provided name, then the list of available column names
            WHILE (ST_ColumnExists(newschemaname, tablename, newuidcolumn) OR (newuidcolumn = 'id' AND uidcolumn IS NULL)) AND
                  NOT provided_uid_isunique AND
                  (ST_ColumnIsUnique(newschemaname, tablename, newuidcolumn) IS NULL OR NOT ST_ColumnIsUnique(newschemaname, tablename, newuidcolumn)) LOOP
                IF uidcolumn IS NULL AND colnameidx < colnamearrlength THEN
                    colnameidx = colnameidx + 1;
                    RAISE NOTICE '  ''%'' is not unique. Checking ''%''...', newuidcolumn, colnamearr[colnameidx]::text;
                    newuidcolumn = colnamearr[colnameidx];
                ELSE
                    IF upper(left(newuidcolumn, 2)) != 'ID' AND upper(newuidcolumn) != 'ID' THEN
                        RAISE NOTICE '  ''%'' is not unique. Creating ''id''...', newuidcolumn;
                        newuidcolumn = 'id';
                        uidcolumn = newuidcolumn;
                    ELSE
                        uidcolumncnt = uidcolumncnt + 1;
                        RAISE NOTICE '  ''%'' is not unique. Checking ''%''...', newuidcolumn, newuidcolumn || '_' || uidcolumncnt::text;
                        newuidcolumn = newuidcolumn || '_' || uidcolumncnt::text;
                    END IF;
                END IF;
            END LOOP;

            IF NOT ST_ColumnExists(newschemaname, tablename, newuidcolumn) THEN
                RAISE NOTICE '  Adding new unique column ''%''...', newuidcolumn;

                --EXECUTE 'DROP SEQUENCE IF EXISTS ' || quote_ident(newschemaname || '_' || tablename || '_seq');
                --EXECUTE 'CREATE SEQUENCE ' || quote_ident(newschemaname || '_' || tablename || '_seq');

                -- Add the new column and update it with nextval('sequence')
                --EXECUTE 'ALTER TABLE ' || fqtn || ' ADD COLUMN ' || newuidcolumn || ' INTEGER';
                --EXECUTE 'UPDATE ' || fqtn || ' SET ' || newuidcolumn || ' = nextval(''' || newschemaname || '_' || tablename || '_seq' || ''')';

                --EXECUTE 'CREATE INDEX ON ' || fqtn || ' USING btree(' || newuidcolumn || ');';

                query = 'SELECT ST_AddUniqueID(''' || newschemaname || ''', ''' || tablename || ''', ''' || newuidcolumn || ''', NULL, true);';
                EXECUTE query;
            ELSE
               RAISE NOTICE '  Column ''%'' exists and is unique...', newuidcolumn;
            END IF;

            -- Create a temporary unique index
            IF NOT ST_HasBasicIndex(newschemaname, tablename, newuidcolumn) THEN
                RAISE NOTICE '  Creating % index on ''%''...', (CASE WHEN whereclausewithwhere = '' THEN 'an' ELSE 'a partial' END), newuidcolumn;
                EXECUTE 'CREATE INDEX ON ' || fqtn || ' USING btree (' || newuidcolumn || ')' || whereclausewithwhere || ';';
            END IF;
        END IF;

        -----------------------------------------------
        -- Summary #2: Check for duplicate geometries (GDUP, GEODUP)
        IF (dosummary IS NULL OR dosummary && dos2) AND (skipsummary IS NULL OR NOT (skipsummary && dos2)) THEN
                query = E'SELECT 2::text summary,\n'
                     || E'       id idsandtypes,\n'
                     || E'       cnt::double precision countsandareas,\n'
                     || E'       (''SELECT * FROM ' || fqtn || ' WHERE ' || newuidcolumn || E' = ANY(ARRAY['' || id || '']);'')::text query,\n'
                     || E'       geom\n'
                     || E'FROM (SELECT string_agg(' || newuidcolumn || '::text, '', ''::text ORDER BY ' || newuidcolumn || E') id,\n'
                     || E'             count(*) cnt,\n'
                     || E'             ST_AsEWKB(' ||              geomcolumnname || E')::geometry geom\n'
                     || E'      FROM ' || fqtn || E'\n'
                     || E'    ' || ltrim(whereclausewithwhere) || CASE WHEN whereclausewithwhere = '' THEN '' ELSE E'\n' END
                     || E'      GROUP BY ST_AsEWKB(' || geomcolumnname || E')) foo\n'
                     || E'WHERE cnt > 1\n'
                     || E'ORDER BY cnt DESC;';

                RETURN QUERY SELECT 'SUMMARY 2 - DUPLICATE GEOMETRIES (GDUP, GEODUP or S2)'::text, ('DUPLICATE GEOMETRIES IDS (' || newuidcolumn || ')')::text, NULL::double precision, query, NULL::geometry;
                RAISE NOTICE 'Summary 2 - Duplicate geometries (GDUP, GEODUP or S2)...';

                IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                    RETURN QUERY EXECUTE query;
                    IF NOT FOUND THEN
                        RETURN QUERY SELECT '2'::text, 'No duplicate geometries...'::text, NULL::double precision, NULL::text, NULL::geometry;
                    END IF;
                ELSE
                    RETURN QUERY SELECT '2'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 2'::text, NULL::double precision, NULL::text, NULL::geometry;
                END IF;
            ELSE
            RETURN QUERY SELECT 'SUMMARY 2 - DUPLICATE GEOMETRIES (GDUP, GEODUP or S2)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 2 - Skipping Duplicate geometries (GDUP, GEODUP or S2)...';
        END IF;

        -----------------------------------------------
        -- Summary #3: Check for overlaps (OVL) - Skipped by default
        IF (dosummary && dos3) AND (skipsummary IS NULL OR NOT (skipsummary && dos3)) THEN
            query = E'SELECT 3::text summary,\n'
                 || E'       a.' || newuidcolumn || '::text || '', '' || b.' || newuidcolumn || E'::text idsandtypes,\n'
                 || E'       ST_Area(ST_Intersection(a.' || geomcolumnname || ', b.' || geomcolumnname || E')) countsandareas,\n'
                 || E'       ''SELECT * FROM ' || fqtn || ' WHERE ' || newuidcolumn || ' = ANY(ARRAY['' || a.' || newuidcolumn || ' || '', '' || b.' || newuidcolumn || E' || '']);''::text query,\n'
                 || E'       ST_CollectionExtract(ST_Intersection(a.' || geomcolumnname || ', b.' || geomcolumnname || E'), 3) geom\n'
                 || E'FROM (SELECT * FROM ' || fqtn || whereclausewithwhere || E') a,\n'
                 || E'     ' || fqtn || E' b\n'
                 || E'WHERE a.' || newuidcolumn || ' < b.' || newuidcolumn || E' AND\n'
                 || E'      (ST_Overlaps(a.' || geomcolumnname || ', b.' || geomcolumnname || E') OR\n'
                 || E'       ST_Contains(a.' || geomcolumnname || ', b.' || geomcolumnname || E') OR\n'
                 || E'       ST_Contains(b.' || geomcolumnname || ', a.' || geomcolumnname || E')) AND\n'
                 || E'       ST_Area(ST_Intersection(a.' || geomcolumnname || ', b.' || geomcolumnname || E')) > 0\n'
                 || E'ORDER BY ST_Area(ST_Intersection(a.' || geomcolumnname || ', b.' || geomcolumnname || ')) DESC;';

            RETURN QUERY SELECT 'SUMMARY 3 - OVERLAPPING GEOMETRIES (OVL or S3)'::text, ('OVERLAPPING GEOMETRIES IDS (' || newuidcolumn || ')')::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 3 - Overlapping geometries (OVL or S3)...';

            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                -- Create a temporary unique index
                IF NOT ST_HasBasicIndex(newschemaname, tablename, geomcolumnname) THEN
                    RAISE NOTICE '            Creating % spatial index on ''%''...', (CASE WHEN whereclausewithwhere = '' THEN 'a' ELSE 'a partial' END), geomcolumnname;
                    EXECUTE 'CREATE INDEX ON ' || fqtn || ' USING gist (' || geomcolumnname || ')' || whereclausewithwhere || ';';
                END IF;

                RAISE NOTICE '            Computing overlaps...';
                BEGIN
                    RETURN QUERY EXECUTE query;
                    IF NOT FOUND THEN
                        RETURN QUERY SELECT '3'::text, 'No overlapping geometries...'::text, NULL::double precision, NULL::text, NULL::geometry;
                    END IF;
                EXCEPTION
                WHEN OTHERS THEN
                    RETURN QUERY SELECT '3'::text, 'ERROR: Consider fixing invalid geometries and convert ST_GeometryCollection before testing for overlaps...'::text, NULL::double precision, NULL::text, NULL::geometry;
                END;
            ELSE
                RETURN QUERY SELECT '3'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 3'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            bydefault = '';
            IF dosummary IS NULL AND (skipsummary IS NULL OR NOT (skipsummary && dos3)) THEN
               bydefault = ' BY DEFAULT';
            END IF;

            RETURN QUERY SELECT 'SUMMARY 3 - OVERLAPPING GEOMETRIES (OVL or S3)'::text, ('SKIPPED' || bydefault)::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 3 - Skipping Overlapping geometries (OVL or S3)...';
        END IF;

        -----------------------------------------------
        -- Summary #4: Check for gaps (GAPS) - Skipped by default
        IF (dosummary && dos4) AND (skipsummary IS NULL OR NOT (skipsummary && dos4)) THEN
            query = E'SELECT 4::text summary,\n'
                 || E'       (ROW_NUMBER() OVER (PARTITION BY true ORDER BY ST_Area(' || geomcolumnname || E') DESC))::text idsandtypes,\n'
                 || E'       ST_Area(' || geomcolumnname || E') countsandareas,\n'
                 || E'       ''SELECT * FROM ' || fqtn || E' WHERE ' || newuidcolumn || E' = ANY(ARRAY['' || (SELECT string_agg(a.' || newuidcolumn || E'::text, '', '') FROM ' || fqtn || E' a WHERE ST_Intersects(ST_Buffer(foo.' || geomcolumnname || E', 0.000001), a.' || geomcolumnname || E')) || '']);''::text query,\n'
                 || E'       ' || geomcolumnname || E' geom\n'
                 || E'FROM (SELECT ST_Buffer(ST_SetSRID(ST_Extent(' || geomcolumnname || E')::geometry, min(ST_SRID(' || geomcolumnname || E'))), 0.01) buffer,\n'
                 || E'             (ST_Dump(ST_Difference(ST_Buffer(ST_SetSRID(ST_Extent(' || geomcolumnname || E')::geometry, min(ST_SRID(' || geomcolumnname || E'))), 0.01), ST_Union(' || geomcolumnname || E')))).*\n'
                 || E'      FROM ' || fqtn || whereclausewithwhere || E') foo\n'
                 || E'WHERE NOT ST_Intersects(geom, ST_ExteriorRing(buffer)) AND ST_Area(geom) > 0\n'
                 || E'ORDER BY countsandareas DESC;';

            RETURN QUERY SELECT 'SUMMARY 4 - GAPS (GAPS or S4)'::text, ('GAPS IDS (generated on the fly)')::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 4 - Gaps (GAPS or S4)...';

            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                RAISE NOTICE '            Computing gaps...';
                BEGIN
                    RETURN QUERY EXECUTE query;
                    IF NOT FOUND THEN
                        RETURN QUERY SELECT '4'::text, 'No gaps...'::text, NULL::double precision, NULL::text, NULL::geometry;
                    END IF;
                EXCEPTION
                WHEN OTHERS THEN
                    RETURN QUERY SELECT '4'::text, 'ERROR: Consider fixing invalid geometries and convert ST_GeometryCollection before testing for gaps...'::text, NULL::double precision, NULL::text, NULL::geometry;
                END;
            ELSE
                RETURN QUERY SELECT '4'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 4'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            bydefault = '';
            IF dosummary IS NULL AND (skipsummary IS NULL OR NOT (skipsummary && dos4)) THEN
               bydefault = ' BY DEFAULT';
            END IF;

            RETURN QUERY SELECT 'SUMMARY 4 - GAPS (GAPS or S4)'::text, ('SKIPPED' || bydefault)::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 4 - Skipping Gaps (GAPS or S4)...';
        END IF;

        -----------------------------------------------
        -- Summary #5: Check for number of NULL, INVALID, EMPTY, POINTS, LINESTRING, POLYGON, MULTIPOINT, MULTILINESTRING, MULTIPOLYGON, GEOMETRYCOLLECTION (TYPES)
        IF (dosummary IS NULL OR dosummary && dos5) AND (skipsummary IS NULL OR NOT (skipsummary && dos5)) THEN
            query = E'SELECT 5::text summary,\n'
                 || E'       CASE WHEN ST_GeometryType(' || geomcolumnname || E') IS NULL THEN ''NULL''\n'
                 || E'            WHEN ST_IsEmpty(' || geomcolumnname || ') THEN ''EMPTY '' || ST_GeometryType(' || geomcolumnname || E')\n'
                 || E'            WHEN NOT ST_IsValid(' || geomcolumnname || ') THEN ''INVALID '' || ST_GeometryType(' || geomcolumnname || E')\n'
                 || E'            ELSE ST_GeometryType(' || geomcolumnname || E')\n'
                 || E'       END idsandtypes,\n'
                 || E'       count(*)::double precision countsandareas,\n'
                 || E'       CASE WHEN ST_GeometryType(' || geomcolumnname || E') IS NULL\n'
                 || E'                 THEN ''SELECT * FROM ' || fqtn || ' WHERE ' || geomcolumnname || ' IS NULL' || whereclause || E';''\n'
                 || E'            WHEN ST_IsEmpty(' || geomcolumnname || E')\n'
                 || E'                 THEN ''SELECT * FROM ' || fqtn || ' WHERE ST_IsEmpty(' || geomcolumnname || ') AND ST_GeometryType(' || geomcolumnname || ') = '''''' || ST_GeometryType(' || geomcolumnname || ') || ''''''' || whereclause || E';''\n'
                 || E'            WHEN NOT ST_IsValid(' || geomcolumnname || E')\n'
                 || E'                 THEN ''SELECT * FROM ' || fqtn || ' WHERE NOT ST_IsValid(' || geomcolumnname || ') AND ST_GeometryType(' || geomcolumnname || ') = '''''' || ST_GeometryType(' || geomcolumnname || ') || ''''''' || whereclause || E';''\n'
                 || E'            ELSE ''SELECT * FROM ' || fqtn || ' WHERE ST_IsValid(' || geomcolumnname || ') AND NOT ST_IsEmpty(' || geomcolumnname || ') AND ST_GeometryType(' || geomcolumnname || ') = '''''' || ST_GeometryType(' || geomcolumnname || ') || ''''''' || whereclause || E';''\n'
                 || E'       END::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM ' || fqtn || E'\n'
                 || ltrim(whereclausewithwhere) || CASE WHEN whereclausewithwhere = '' THEN '' ELSE E'\n' END
                 || E'GROUP BY ST_IsValid(' || geomcolumnname || '), ST_IsEmpty(' || geomcolumnname || '), ST_GeometryType(' || geomcolumnname || E')\n'
                 || E'ORDER BY ST_GeometryType(' || geomcolumnname || ') DESC, NOT ST_IsValid(' || geomcolumnname || '), ST_IsEmpty(' || geomcolumnname || ');';

            RETURN QUERY SELECT 'SUMMARY 5 - GEOMETRY TYPES (TYPES or S5)'::text, 'TYPES'::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 5 - Geometry types (TYPES or S5)...';
            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                RETURN QUERY EXECUTE query;
                IF NOT FOUND THEN
                    RETURN QUERY SELECT '5'::text, 'No row selected...'::text, NULL::double precision, NULL::text, NULL::geometry;
                END IF;
            ELSE
                RETURN QUERY SELECT '5'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 5'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            RETURN QUERY SELECT 'SUMMARY 5 - GEOMETRY TYPES (TYPES or S5)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 5 - Skipping Geometry types (TYPES or S5)...';
        END IF;

        -----------------------------------------------
        -- Create an index on ST_NPoints(geom) if necessary so further queries are executed faster
        IF (dosummary IS NULL OR dosummary && dos6 OR dosummary && dos7) AND (skipsummary IS NULL OR NOT (skipsummary && dos6 AND skipsummary && dos7)) AND
           ST_ColumnExists(newschemaname, tablename, geomcolumnname) AND
           NOT ST_HasBasicIndex(newschemaname, tablename, NULL, 'st_npoints'::text) THEN
            RAISE NOTICE 'Creating % index on ''ST_NPoints(%)''...', (CASE WHEN whereclausewithwhere = '' THEN 'an' ELSE 'a partial' END), geomcolumnname;
            query = 'CREATE INDEX ' || left(tablename || '_' || geomcolumnname, 48) || '_st_npoints_idx ON ' || fqtn || ' USING btree (ST_NPoints(' || geomcolumnname || '))' || whereclausewithwhere || ';';
            EXECUTE query;
        END IF;

        -----------------------------------------------
        -- Summary #6: Check for polygon complexity - min number of vertexes, max number of vertexes, mean number of vertexes (VERTX).
        IF (dosummary IS NULL OR dosummary && dos6) AND (skipsummary IS NULL OR NOT (skipsummary && dos6)) THEN
            query = E'WITH points AS (SELECT ST_NPoints(' || geomcolumnname || ') nv FROM ' || fqtn || whereclausewithwhere || E'),\n'
                 || E'     agg    AS (SELECT min(nv) min, max(nv) max, avg(nv) avg FROM points)\n'
                 || E'SELECT 6::text summary,\n'
                 || E'       ''MIN number of vertexes''::text idsandtypes,\n'
                 || E'       min::double precision countsandareas,\n'
                 || E'       (''SELECT * FROM ' || fqtn || ' WHERE ST_NPoints(' || geomcolumnname || ') = '' || min::text || ''' || whereclause || E';'')::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM agg\n'
                 || E'UNION ALL\n'
                 || E'SELECT 6::text summary,\n'
                 || E'       ''MAX number of vertexes''::text idsandtypes,\n'
                 || E'       max::double precision countsandareas,\n'
                 || E'       (''SELECT * FROM ' || fqtn || ' WHERE ST_NPoints(' || geomcolumnname || ') = '' || max::text || ''' || whereclause || E';'')::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM agg\n'
                 || E'UNION ALL\n'
                 || E'SELECT 6::text summary,\n'
                 || E'       ''MEAN number of vertexes''::text idsandtypes,\n'
                 || E'       avg::double precision countsandareas,\n'
                 || E'       (''No usefull query'')::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM agg;';

            RETURN QUERY SELECT 'SUMMARY 6 - VERTEX STATISTICS (VERTX or S6)'::text, 'STATISTIC'::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 6 - Vertex statistics (VERTX or S6)...';
            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                RETURN QUERY EXECUTE query;
            ELSE
                RETURN QUERY SELECT '6'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 6'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            RETURN QUERY SELECT 'SUMMARY 6 - VERTEX STATISTICS (VERTX or S6)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 6 - Skipping Vertex statistics (VERTX or S6)...';
        END IF;

        -----------------------------------------------
        -- Summary #7: Build an histogram of the number of vertexes (VHISTO).
        IF (dosummary IS NULL OR dosummary && dos7) AND (skipsummary IS NULL OR NOT (skipsummary && dos7)) THEN
            RAISE NOTICE 'Summary 7 - Histogram of the number of vertexes (VHISTO or S7)...';

            sum7nbinterval = nbinterval;
            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN

                -- Precompute the min and max number of vertexes so we can set the number of interval to 1 if they are equal
                query = 'SELECT min(ST_NPoints(' || geomcolumnname || ')), max(ST_NPoints(' || geomcolumnname || ')) FROM ' || fqtn || whereclausewithwhere;
                EXECUTE QUERY query INTO minnp, maxnp;

                IF minnp IS NULL AND maxnp IS NULL THEN
                    query = E'WITH npoints AS (SELECT ST_NPoints(' || geomcolumnname || ') np FROM ' || fqtn || whereclausewithwhere || E'),\n'
                         || E'     histo   AS (SELECT count(*) cnt FROM npoints)\n'
                         || E'SELECT 7::text summary,\n'
                         || E'       ''NULL''::text idsandtypes,\n'
                         || E'       cnt::double precision countsandareas,\n'
                         || E'       ''SELECT *, ST_NPoints(' || geomcolumnname || ') nbpoints FROM ' || fqtn || ' WHERE ' || geomcolumnname || ' IS NULL' || whereclause || E';''::text query,\n'
                         || E'       NULL::geometry geom\n'
                         || E'FROM histo;';
                ELSE
                    IF maxnp - minnp = 0 THEN
                        RAISE NOTICE 'Summary 7: maximum number of points - minimum number of points = 0. Will create only 1 interval instead of %...', sum7nbinterval;
                        sum7nbinterval = 1;
                    ELSEIF maxnp - minnp + 1 < sum7nbinterval THEN
                        RAISE NOTICE 'Summary 7: maximum number of points - minimum number of points < %. Will create only % interval instead of %...', sum7nbinterval, maxnp - minnp + 1, sum7nbinterval;
                        sum7nbinterval = maxnp - minnp + 1;
                    END IF;

                    -- Compute the histogram
                    query = E'WITH npoints AS (SELECT ST_NPoints(' || geomcolumnname || ') np FROM ' || fqtn || whereclausewithwhere || E'),\n'
                         || E'     bins    AS (SELECT np, CASE WHEN np IS NULL THEN -1 ELSE least(floor((np - ' || minnp || ')*' || sum7nbinterval || '::numeric/(' || (CASE WHEN maxnp - minnp = 0 THEN maxnp + 0.000000001 ELSE maxnp END) - minnp || ')), ' || sum7nbinterval || ' - 1) END bin, ' || (maxnp - minnp) || '/' || sum7nbinterval || E'.0 binrange FROM npoints),\n'
                         || E'     histo  AS (SELECT bin, count(*) cnt FROM bins GROUP BY bin)\n'
                         || E'SELECT 7::text summary,\n'
                         || E'       CASE WHEN serie = -1 THEN ''NULL''::text ELSE ''['' || round(' || minnp || ' + serie * binrange)::text || '' - '' || (CASE WHEN serie = ' || sum7nbinterval || ' - 1 THEN round(' || maxnp || ')::text || '']'' ELSE round(' || minnp || E' + (serie + 1) * binrange)::text || ''['' END) END idsandtypes,\n'
                         || E'       coalesce(cnt, 0)::double precision countsandareas,\n'
                         || E'      (''SELECT *, ST_NPoints(' || geomcolumnname || ') nbpoints FROM ' || fqtn || ' WHERE ST_NPoints(' || geomcolumnname || ')'' || (CASE WHEN serie = -1 THEN '' IS NULL'' || ''' || whereclause || ''' ELSE ('' >= '' || round(' || minnp || ' + serie * binrange)::text || '' AND ST_NPoints(' || geomcolumnname || ') <'' || (CASE WHEN serie = ' || sum7nbinterval || ' - 1 THEN ''= '' || ' || maxnp || '::float8::text ELSE '' '' || round(' || minnp || ' + (serie + 1) * binrange)::text END) || ''' || whereclause || ''' || '' ORDER BY ST_NPoints(' || geomcolumnname || E') DESC'') END) || '';'')::text query,\n'
                         || E'       NULL::geometry geom\n'
                         || E'FROM generate_series(-1, ' || sum7nbinterval || E' - 1) serie\n'
                         || E'     LEFT OUTER JOIN histo ON (serie = histo.bin),\n'
                         || E'    (SELECT * FROM bins LIMIT 1) foo\n'
                         || E'ORDER BY serie;';
                END IF;
                RETURN QUERY SELECT 'SUMMARY 7 - HISTOGRAM OF THE NUMBER OF VERTEXES (VHISTO or S7)'::text, 'NUMBER OF VERTEXES INTERVALS'::text, NULL::double precision, query, NULL::geometry;
                RETURN QUERY EXECUTE query;
            ELSE
                RETURN QUERY SELECT 'SUMMARY 7 - HISTOGRAM OF THE NUMBER OF VERTEXES (VHISTO or S7)'::text, 'NUMBER OF VERTEXES INTERVALS'::text, NULL::double precision, ''::text, NULL::geometry;
                RETURN QUERY SELECT '7'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 7'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            RETURN QUERY SELECT 'SUMMARY 7 - HISTOGRAM OF THE NUMBER OF VERTEXES (VHISTO or S7)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 7 - Skipping Histogram of the number of vertexes (VHISTO or S7)...';
        END IF;

        -----------------------------------------------
        -- Create an index on ST_Area(geom) if necessary so further queries are executed faster
        IF (dosummary IS NULL OR dosummary && dos8 OR dosummary && dos9 OR dosummary && dos10) AND (skipsummary IS NULL OR NOT (skipsummary && dos8 AND skipsummary && dos9 AND skipsummary && dos10)) AND
           ST_ColumnExists(newschemaname, tablename, geomcolumnname) AND
           NOT ST_HasBasicIndex(newschemaname, tablename, NULL, 'st_area'::text) THEN
            RAISE NOTICE 'Creating % index on ''ST_Area(%)''...', (CASE WHEN whereclausewithwhere = '' THEN 'an' ELSE 'a partial' END), geomcolumnname;
            query = 'CREATE INDEX ' || left(tablename || '_' || geomcolumnname, 51) || '_st_area_idx ON ' || fqtn || ' USING btree (ST_Area(' || geomcolumnname || '))' || whereclausewithwhere || ';';
            EXECUTE query;
        END IF;

        -----------------------------------------------
        -- Summary #8: Check for polygon areas - min area, max area, mean area (AREAS)
        IF (dosummary IS NULL OR dosummary && dos8) AND (skipsummary IS NULL OR NOT (skipsummary && dos8)) THEN
            query = E'WITH areas AS (SELECT ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || whereclausewithwhere || E'),\n'
                 || E'     agg    AS (SELECT min(area) min, max(area) max, avg(area) avg FROM areas)\n'
                 || E'SELECT 8::text summary,\n'
                 || E'       ''MIN area''::text idsandtypes,\n'
                 || E'       min::double precision countsandareas,\n'
                 || E'       (''SELECT * FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') < '' || min::text || '' + 0.000000001' || whereclause || E';'')::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM agg\n'
                 || E'UNION ALL\n'
                 || E'SELECT 8::text summary,\n'
                 || E'       ''MAX area''::text idsandtypes,\n'
                 || E'       max::double precision countsandareas,\n'
                 || E'       (''SELECT * FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') > '' || max::text || '' - 0.000000001 AND ST_Area(' || geomcolumnname || ') < '' || max::text || '' + 0.000000001' || whereclause || E';'')::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM agg\n'
                 || E'UNION ALL\n'
                 || E'SELECT 8::text summary,\n'
                 || E'       ''MEAN area''::text idsandtypes,\n'
                 || E'       avg::double precision countsandareas,\n'
                 || E'       (''No usefull query'')::text query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM agg';

            RETURN QUERY SELECT 'SUMMARY 8 - GEOMETRY AREA STATISTICS (AREAS, AREA or S8)'::text, 'STATISTIC'::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 8 - Geometry area statistics (AREAS, AREA or S8)...';
            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                RETURN QUERY EXECUTE query;
            ELSE
                RETURN QUERY SELECT '8'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 8'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            RETURN QUERY SELECT 'SUMMARY 8 - GEOMETRY AREA STATISTICS (AREAS, AREA or S8)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 8 - Skipping Geometry area statistics (AREAS, AREA or S8)...';
        END IF;

        -----------------------------------------------
        -- Summary #9: Build an histogram of the areas (AHISTO)
        IF (dosummary IS NULL OR dosummary && dos9) AND (skipsummary IS NULL OR NOT (skipsummary && dos9)) THEN
            RAISE NOTICE 'Summary 9 - Histogram of areas (AHISTO or S9)...';

            sum9nbinterval = nbinterval;
            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN

                -- Precompute the min and max values so we can set the number of interval to 1 if they are equal
                query = 'SELECT min(ST_Area(' || geomcolumnname || ')), max(ST_Area(' || geomcolumnname || ')) FROM ' || fqtn || whereclausewithwhere;
                EXECUTE QUERY query INTO minarea, maxarea;
                IF maxarea IS NULL AND minarea IS NULL THEN
                    query = E'WITH values AS (SELECT ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || whereclausewithwhere || E'),\n'
                         || E'    histo  AS (SELECT count(*) cnt FROM values)\n'
                         || E'SELECT 9::text summary,\n'
                         || E'      ''NULL''::text idsandtypes,\n'
                         || E'      cnt::double precision countsandareas,\n'
                         || E'      ''SELECT *, ST_Area(' || geomcolumnname || ') FROM ' || fqtn || ' WHERE ' || geomcolumnname || ' IS NULL' || whereclause || E';''::text query,\n'
                         || E'      NULL::geometry\n'
                         || E'FROM histo;';

                    RETURN QUERY SELECT 'SUMMARY 9 - HISTOGRAM OF AREAS (AHISTO or S9)'::text, 'AREAS INTERVALS'::text, NULL::double precision, query, NULL::geometry;
                    RETURN QUERY EXECUTE query;
                ELSE
                    IF maxarea - minarea = 0 THEN
                        RAISE NOTICE 'maximum area - minimum area = 0. Will create only 1 interval instead of %...', nbinterval;
                        sum9nbinterval = 1;
                    END IF;

                    -- We make sure double precision values are converted to text using the maximum number of digits before
                    SET extra_float_digits = 3;

                    -- Compute the histogram
                    query = E'WITH areas AS (SELECT ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || whereclausewithwhere || E'),\n'
                         || E'    bins AS (SELECT area, CASE WHEN area IS NULL THEN -1 ELSE least(floor((area - ' || minarea || ')*' || sum9nbinterval || '::numeric/(' || (CASE WHEN maxarea - minarea = 0 THEN maxarea + 0.000000001 ELSE maxarea END) - minarea || ')), ' || sum9nbinterval || ' - 1) END bin, ' || (maxarea - minarea) || '/' || sum9nbinterval || E'.0 binrange FROM areas),\n'
                         || E'    histo AS (SELECT bin, count(*) cnt FROM bins GROUP BY bin)\n'
                         || E'SELECT 9::text summary,\n'
                         || E'      CASE WHEN serie = -1 THEN ''NULL''::text ELSE ''['' || (' || minarea || ' + serie * binrange)::float8::text || '' - '' || (CASE WHEN serie = ' || sum9nbinterval || ' - 1 THEN ' || maxarea || '::float8::text || '']'' ELSE (' || minarea || E' + (serie + 1) * binrange)::float8::text || ''['' END) END idsandtypes,\n'
                         || E'      coalesce(cnt, 0)::double precision countsandareas,\n'
                         || E'      (''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ')'' || (CASE WHEN serie = -1 THEN '' IS NULL'' || ''' || whereclause || ''' ELSE ('' >= '' || (' || minarea || ' + serie * binrange)::float8::text || '' AND ST_Area(' || geomcolumnname || ') <'' || (CASE WHEN serie = ' || sum9nbinterval || ' - 1 THEN ''= '' || ' || maxarea || '::float8::text ELSE '' '' || (' || minarea || ' + (serie + 1) * binrange)::float8::text END) || ''' || whereclause || ''' || '' ORDER BY ST_Area(' || geomcolumnname || E') DESC'') END) || '';'')::text query,\n'
                         || E'      NULL::geometry geom\n'
                         || E'FROM generate_series(-1, ' || sum9nbinterval || E' - 1) serie\n'
                         || E'    LEFT OUTER JOIN histo ON (serie = histo.bin),\n'
                         || E'    (SELECT * FROM bins LIMIT 1) foo\n'
                         || E'ORDER BY serie;';

                    RETURN QUERY SELECT 'SUMMARY 9 - HISTOGRAM OF AREAS (AHISTO or S9)'::text, 'AREAS INTERVALS'::text, NULL::double precision, E'SET extra_float_digits = 3;\n' || query, NULL::geometry;
                    RETURN QUERY EXECUTE query;
                    RESET extra_float_digits;
                END IF;
            ELSE
                RETURN QUERY SELECT 'SUMMARY 9 - HISTOGRAM OF AREAS (AHISTO or S9)'::text, 'AREAS INTERVALS'::text, NULL::double precision, ''::text, NULL::geometry;
                RETURN QUERY SELECT '9'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 9'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            RETURN QUERY SELECT 'SUMMARY 9 - HISTOGRAM OF AREAS (AHISTO or S9)'::text, 'SKIPPED'::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 9 - Skipping Histogram of areas (AHISTO or S9)...';
        END IF;

        -----------------------------------------------
        -- Summary #10: Build a list of the small areas (SACOUNT) < 0.1 units. Skipped by default
        IF (dosummary && dos10) AND (skipsummary IS NULL OR NOT (skipsummary && dos10)) THEN
            query = E'WITH areas AS (SELECT ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE (ST_Area(' || geomcolumnname || ') IS NULL OR ST_Area(' || geomcolumnname || ') < 0.1) ' || whereclause || E'),\n'
                 || E'     bins  AS (SELECT area,\n'
                 || E'                      CASE WHEN area IS NULL THEN -1\n'
                 || E'                           WHEN area = 0.0 THEN 0\n'
                 || E'                           WHEN area < 0.0000001 THEN 1\n'
                 || E'                           WHEN area < 0.000001 THEN 2\n'
                 || E'                           WHEN area < 0.00001 THEN 3\n'
                 || E'                           WHEN area < 0.0001 THEN 4\n'
                 || E'                           WHEN area < 0.001 THEN 5\n'
                 || E'                           WHEN area < 0.01 THEN 6\n'
                 || E'                           WHEN area < 0.1 THEN 7\n'
                 || E'                      END bin\n'
                 || E'               FROM areas),\n'
                 || E'    histo AS (SELECT bin, count(*) cnt FROM bins GROUP BY bin)\n'
                 || E'SELECT 10::text summary,\n'
                 || E'       CASE WHEN serie = -1 THEN ''NULL''\n'
                 || E'            WHEN serie = 0 THEN ''[0]''\n'
                 || E'            WHEN serie = 1 THEN '']0 - 0.0000001[''\n'
                 || E'            WHEN serie = 2 THEN ''[0.0000001 - 0.000001[''\n'
                 || E'            WHEN serie = 3 THEN ''[0.000001 - 0.00001[''\n'
                 || E'            WHEN serie = 4 THEN ''[0.00001 - 0.0001[''\n'
                 || E'            WHEN serie = 5 THEN ''[0.0001 - 0.001[''\n'
                 || E'            WHEN serie = 6 THEN ''[0.001 - 0.01[''\n'
                 || E'            WHEN serie = 7 THEN ''[0.01 - 0.1[''\n'
                 || E'       END idsandtypes,\n'
                 || E'       coalesce(cnt, 0)::double precision countsandareas,\n'
                 || E'       CASE WHEN serie = -1 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') IS NULL' || whereclause || E';''::text\n'
                 || E'            WHEN serie = 0 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') = 0' || whereclause || E';''::text\n'
                 || E'            WHEN serie = 1 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') > 0 AND ST_Area(' || geomcolumnname || ') < 0.0000001' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'            WHEN serie = 2 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') >= 0.0000001 AND ST_Area(' || geomcolumnname || ') < 0.000001' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'            WHEN serie = 3 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') >= 0.000001 AND ST_Area(' || geomcolumnname || ') < 0.00001' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'            WHEN serie = 4 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') >= 0.00001 AND ST_Area(' || geomcolumnname || ') < 0.0001' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'            WHEN serie = 5 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') >= 0.0001 AND ST_Area(' || geomcolumnname || ') < 0.001' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'            WHEN serie = 6 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') >= 0.001 AND ST_Area(' || geomcolumnname || ') < 0.01' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'            WHEN serie = 7 THEN ''SELECT *, ST_Area(' || geomcolumnname || ') area FROM ' || fqtn || ' WHERE ST_Area(' || geomcolumnname || ') >= 0.01 AND ST_Area(' || geomcolumnname || ') < 0.1' || whereclause || ' ORDER BY ST_Area(' || geomcolumnname || E') DESC;''::text\n'
                 || E'       END query,\n'
                 || E'       NULL::geometry geom\n'
                 || E'FROM generate_series(-1, 7) serie\n'
                 || E'     LEFT OUTER JOIN histo ON (serie = histo.bin),\n'
                 || E'     (SELECT * FROM bins LIMIT 1) foo\n'
                 || E'ORDER BY serie;';

            RETURN QUERY SELECT 'SUMMARY 10 - COUNT OF SMALL AREAS (SACOUNT or S10)'::text, 'SMALL AREAS INTERVALS'::text, NULL::double precision, query, NULL::geometry;
            RAISE NOTICE 'Summary 10 - Count of small areas (SACOUNT or S10)...';

            IF ST_ColumnExists(newschemaname, tablename, geomcolumnname) THEN
                RETURN QUERY EXECUTE query;
                IF NOT FOUND THEN
                    RETURN QUERY SELECT '10'::text, 'No geometry smaller than 0.1...'::text, NULL::double precision, NULL::text, NULL::geometry;
                END IF;
            ELSE
                RETURN QUERY SELECT '10'::text, '''' || geomcolumnname::text || ''' does not exists... Skipping Summary 10'::text, NULL::double precision, NULL::text, NULL::geometry;
            END IF;
        ELSE
            bydefault = '';
            IF dosummary IS NULL AND (skipsummary IS NULL OR NOT (skipsummary && dos10)) THEN
               bydefault = ' BY DEFAULT';
            END IF;
            RETURN QUERY SELECT 'SUMMARY 10 - COUNT OF AREAS (SACOUNT or S10)'::text, ('SKIPPED' || bydefault)::text, NULL::double precision, NULL::text, NULL::geometry;
            RAISE NOTICE 'Summary 10 - Skipping Count of small areas (SACOUNT or S10)...';
        END IF;

        RETURN;
    END;
$$;


ALTER FUNCTION public.st_geotablesummary(schemaname name, tablename name, geomcolumnname name, uidcolumn name, nbinterval integer, dosummary text[], skipsummary text[], whereclause text) OWNER TO postgres;

--
-- TOC entry 1530 (class 1255 OID 4441940)
-- Name: st_geotablesummary(name, name, name, name, integer, text, text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_geotablesummary(schemaname name, tablename name, geomcolumnname name, uidcolumn name, nbinterval integer, dosummary text DEFAULT NULL::text, skipsummary text DEFAULT NULL::text, whereclause text DEFAULT NULL::text) RETURNS TABLE(summary text, idsandtypes text, countsandareas double precision, query text, geom public.geometry)
    LANGUAGE sql
    AS $_$
    SELECT ST_GeoTableSummary($1, $2, $3, $4, $5, regexp_split_to_array($6, E'\\s*\,\\s'), regexp_split_to_array($7, E'\\s*\,\\s'), $8)
$_$;


ALTER FUNCTION public.st_geotablesummary(schemaname name, tablename name, geomcolumnname name, uidcolumn name, nbinterval integer, dosummary text, skipsummary text, whereclause text) OWNER TO postgres;

--
-- TOC entry 1531 (class 1255 OID 4441941)
-- Name: st_globalrasterunion(name, name, name, text, text, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_globalrasterunion(schemaname name, tablename name, rastercolumnname name, method text DEFAULT 'FIRST_RASTER_VALUE_AT_PIXEL_CENTROID'::text, pixeltype text DEFAULT NULL::text, nodataval double precision DEFAULT NULL::double precision) RETURNS public.raster
    LANGUAGE plpgsql IMMUTABLE
    AS $$
    DECLARE
        query text;
        newrast raster;
        fct2call text;
        pixeltypetxt text;
        nodatavaltxt text;
    BEGIN
        IF right(method, 5) = 'TROID' THEN
            fct2call = 'ST_ExtractPixelCentroidValue4ma';
        ELSE
            fct2call = 'ST_ExtractPixelValue4ma';
        END IF;
        IF pixeltype IS NULL THEN
            pixeltypetxt = 'ST_BandPixelType(' || quote_ident(rastercolumnname) || ')';
        ELSE
            pixeltypetxt = '''' || pixeltype || '''::text';
        END IF;
        IF nodataval IS NULL THEN
            nodatavaltxt = 'ST_BandNodataValue(' || quote_ident(rastercolumnname) || ')';
        ELSE
            nodatavaltxt = nodataval;
        END IF;
        query = 'SELECT ST_MapAlgebra(rast,
                                      1,
                                      ''' || fct2call || '(double precision[], integer[], text[])''::regprocedure,
                                      ST_BandPixelType(rast, 1),
                                      NULL,
                                      NULL,
                                      NULL,
                                      NULL,
                                      ST_Width(rast)::text,
                                      ST_Height(rast)::text,
                                      ST_UpperLeftX(rast)::text,
                                      ST_UpperLeftY(rast)::text,
                                      ST_ScaleX(rast)::text,
                                      ST_ScaleY(rast)::text,
                                      ST_SkewX(rast)::text,
                                      ST_SkewY(rast)::text,
                                      ST_SRID(rast)::text,' ||
                                      quote_literal(schemaname) || ', ' ||
                                      quote_literal(tablename) || ', ' ||
                                      quote_literal(rastercolumnname) || ',
                                      NULL' || ', ' ||
                                      quote_literal(upper(method)) || '
                                     ) rast
                 FROM (SELECT ST_AsRaster(ST_Union(rast::geometry),
                                          min(scalex),
                                          min(scaley),
                                          min(gridx),
                                          min(gridy),
                                          max(pixeltype),
                                          0,
                                          min(nodataval)
                                         ) rast
                       FROM (SELECT ' || quote_ident(rastercolumnname) || ' rast,
                                    ST_ScaleX(' || quote_ident(rastercolumnname) || ') scalex,
                                    ST_ScaleY(' || quote_ident(rastercolumnname) || ') scaley,
                                    ST_UpperLeftX(' || quote_ident(rastercolumnname) || ') gridx,
                                    ST_UpperLeftY(' || quote_ident(rastercolumnname) || ') gridy,
                                    ' || pixeltypetxt || ' pixeltype,
                                    ' || nodatavaltxt || ' nodataval
                             FROM ' || quote_ident(schemaname) || '.' || quote_ident(tablename) || '
                            ) foo1
                      ) foo2';
        EXECUTE query INTO newrast;
        RETURN newrast;
    END;
$$;


ALTER FUNCTION public.st_globalrasterunion(schemaname name, tablename name, rastercolumnname name, method text, pixeltype text, nodataval double precision) OWNER TO postgres;

--
-- TOC entry 1532 (class 1255 OID 4441942)
-- Name: st_hasbasicindex(name, name); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_hasbasicindex(tablename name, columnname name) RETURNS boolean
    LANGUAGE sql
    AS $_$
    SELECT ST_HasBasicIndex('public', $1, $2, NULL)
$_$;


ALTER FUNCTION public.st_hasbasicindex(tablename name, columnname name) OWNER TO postgres;

--
-- TOC entry 1533 (class 1255 OID 4441943)
-- Name: st_hasbasicindex(name, name, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_hasbasicindex(tablename name, columnname name, idxstring text) RETURNS boolean
    LANGUAGE sql
    AS $_$
    SELECT ST_HasBasicIndex('public', $1, $2, $3)
$_$;


ALTER FUNCTION public.st_hasbasicindex(tablename name, columnname name, idxstring text) OWNER TO postgres;

--
-- TOC entry 1534 (class 1255 OID 4441944)
-- Name: st_hasbasicindex(name, name, name, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_hasbasicindex(schemaname name, tablename name, columnname name, idxstring text) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
    DECLARE
        query text;
        coltype text;
        hasindex boolean := FALSE;
    BEGIN
        IF schemaname IS NULL OR schemaname = '' OR tablename IS NULL OR tablename = '' THEN
            RETURN NULL;
        END IF;
        -- Check if schemaname is not actually a table name and idxstring actually a column name.
        -- That's the only way to support a three parameters variant taking a schemaname, a tablename and a columnname
        IF ST_ColumnExists(tablename, columnname, idxstring) THEN
            schemaname = tablename;
            tablename = columnname;
            columnname = idxstring;
            idxstring = NULL;
        END IF;
        IF (columnname IS NULL OR columnname = '') AND (idxstring IS NULL OR idxstring = '') THEN
            RETURN NULL;
        END IF;
        IF NOT columnname IS NULL AND columnname != '' AND ST_ColumnExists(schemaname, tablename, columnname) THEN
            -- Determine the type of the column
            query := 'SELECT typname
                      FROM pg_namespace
                          LEFT JOIN pg_class ON (pg_namespace.oid = pg_class.relnamespace)
                          LEFT JOIN pg_attribute ON (pg_attribute.attrelid = pg_class.oid)
                          LEFT JOIN pg_type ON (pg_type.oid = pg_attribute.atttypid)
                      WHERE lower(nspname) = lower(''' || schemaname || ''') AND lower(relname) = lower(''' || tablename || ''') AND lower(attname) = lower(''' || columnname || ''');';
            EXECUTE QUERY query INTO coltype;
        END IF;

        IF coltype IS NULL AND (idxstring IS NULL OR idxstring = '') THEN
            RETURN NULL;
        ELSIF coltype = 'raster' THEN
            -- When column type is RASTER we ignore the column name and
            -- only check if the type of the index is gist since it is a functional
            -- index and it would be hard to check on which column it is applied
            query := 'SELECT TRUE
                      FROM pg_index
                      LEFT OUTER JOIN pg_class relclass ON (relclass.oid = pg_index.indrelid)
                      LEFT OUTER JOIN pg_namespace ON (pg_namespace.oid = relclass.relnamespace)
                      LEFT OUTER JOIN pg_class idxclass ON (idxclass.oid = pg_index.indexrelid)
                      LEFT OUTER JOIN pg_am ON (pg_am.oid = idxclass.relam)
                      WHERE relclass.relkind = ''r'' AND amname = ''gist''
                      AND lower(nspname) = lower(''' || schemaname || ''') AND lower(relclass.relname) = lower(''' || tablename || ''')';
            IF NOT idxstring IS NULL THEN
                query := query || ' AND lower(idxclass.relname) LIKE lower(''%' || idxstring || '%'');';
            END IF;
            EXECUTE QUERY query INTO hasindex;
        ELSE
            -- Otherwise we check for an index on the right column
            query := 'SELECT TRUE
                      FROM pg_index
                      LEFT OUTER JOIN pg_class relclass ON (relclass.oid = pg_index.indrelid)
                      LEFT OUTER JOIN pg_namespace ON (pg_namespace.oid = relclass.relnamespace)
                      LEFT OUTER JOIN pg_class idxclass ON (idxclass.oid = pg_index.indexrelid)
                      --LEFT OUTER JOIN pg_am ON (pg_am.oid = idxclass.relam)
                      LEFT OUTER JOIN pg_attribute ON (pg_attribute.attrelid = relclass.oid AND indkey[0] = attnum)
                      WHERE relclass.relkind = ''r''
                      AND lower(nspname) = lower(''' || schemaname || ''') AND lower(relclass.relname) = lower(''' || tablename || ''')';
            IF NOT idxstring IS NULL THEN
                query := query || ' AND lower(idxclass.relname) LIKE lower(''%' || idxstring || '%'')';
            END IF;
            IF NOT columnname IS NULL THEN
                query := query || ' AND indkey[0] != 0 AND lower(attname) = lower(''' || columnname || ''')';
            END IF;
 --RAISE NOTICE 'query = %', query;
            EXECUTE QUERY query INTO hasindex;
        END IF;
        IF hasindex IS NULL THEN
            hasindex = FALSE;
        END IF;
        RETURN hasindex;
    END;
$$;


ALTER FUNCTION public.st_hasbasicindex(schemaname name, tablename name, columnname name, idxstring text) OWNER TO postgres;

--
-- TOC entry 1505 (class 1255 OID 4441945)
-- Name: st_histogram(text, text, text, integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_histogram(schemaname text, tablename text, columnname text, nbinterval integer DEFAULT 10, whereclause text DEFAULT NULL::text) RETURNS TABLE(intervals text, cnt integer, query text)
    LANGUAGE plpgsql
    AS $$
    DECLARE
    fqtn text;
    query text;
    newschemaname name;
    findnewcolumnname boolean := FALSE;
    newcolumnname text;
    columnnamecnt int := 0;
    whereclausewithwhere text := '';
    minval double precision := 0;
    maxval double precision := 0;
    columntype text;
    BEGIN
        IF nbinterval IS NULL THEN
            nbinterval = 10;
        END IF;
        IF nbinterval <= 0 THEN
            RAISE NOTICE 'nbinterval is smaller or equal to zero. Returning nothing...';
            RETURN;
        END IF;
        IF whereclause IS NULL OR whereclause = '' THEN
            whereclause = '';
        ELSE
            whereclausewithwhere = ' WHERE ' || whereclause || ' ';
            whereclause = ' AND (' || whereclause || ')';
        END IF;
        newschemaname := '';
        IF length(schemaname) > 0 THEN
            newschemaname := schemaname;
        ELSE
            newschemaname := 'public';
        END IF;
        fqtn := quote_ident(newschemaname) || '.' || quote_ident(tablename);

        -- Build an histogram with the column values.
        IF ST_ColumnExists(newschemaname, tablename, columnname) THEN

            -- Precompute the min and max values so we can set the number of interval to 1 if they are equal
            query = 'SELECT min(' || columnname || '), max(' || columnname || ') FROM ' || fqtn || whereclausewithwhere;
            EXECUTE QUERY query INTO minval, maxval;
            IF maxval IS NULL AND minval IS NULL THEN
                query = 'WITH values AS (SELECT ' || columnname || ' val FROM ' || fqtn || whereclausewithwhere || '),
                              histo  AS (SELECT count(*) cnt FROM values)
                         SELECT ''NULL''::text intervals,
                                cnt::int,
                                ''SELECT * FROM ' || fqtn || ' WHERE ' || columnname || ' IS NULL' || whereclause || ';''::text query
                         FROM histo;';
                RETURN QUERY EXECUTE query;
            ELSE
                IF maxval - minval = 0 THEN
                    RAISE NOTICE 'maximum value - minimum value = 0. Will create only 1 interval instead of %...', nbinterval;
                    nbinterval = 1;
                END IF;

                -- We make sure double precision values are converted to text using the maximum number of digits before computing summaries involving this type of values
                query = 'SELECT pg_typeof(' || columnname || ')::text FROM ' || fqtn || ' LIMIT 1';
                EXECUTE query INTO columntype;
                IF left(columntype, 3) != 'int' THEN
                    SET extra_float_digits = 3;
                END IF;

                -- Compute the histogram
                query = 'WITH values AS (SELECT ' || columnname || ' val FROM ' || fqtn || whereclausewithwhere || '),
                              bins   AS (SELECT val, CASE WHEN val IS NULL THEN -1 ELSE least(floor((val - ' || minval || ')*' || nbinterval || '::numeric/(' || (CASE WHEN maxval - minval = 0 THEN maxval + 0.000000001 ELSE maxval END) - minval || ')), ' || nbinterval || ' - 1) END bin, ' || (maxval - minval) || '/' || nbinterval || '.0 binrange FROM values),
                              histo  AS (SELECT bin, count(*) cnt FROM bins GROUP BY bin)
                         SELECT CASE WHEN serie = -1 THEN ''NULL''::text ELSE ''['' || (' || minval || ' + serie * binrange)::float8::text || '' - '' || (CASE WHEN serie = ' || nbinterval || ' - 1 THEN ' || maxval || '::float8::text || '']'' ELSE (' || minval || ' + (serie + 1) * binrange)::float8::text || ''['' END) END intervals,
                                coalesce(cnt, 0)::int cnt,
                                (''SELECT * FROM ' || fqtn || ' WHERE ' || columnname || ''' || (CASE WHEN serie = -1 THEN '' IS NULL'' || ''' || whereclause || ''' ELSE ('' >= '' || (' || minval || ' + serie * binrange)::float8::text || '' AND ' || columnname || ' <'' || (CASE WHEN serie = ' || nbinterval || ' - 1 THEN ''= '' || ' || maxval || '::float8::text ELSE '' '' || (' || minval || ' + (serie + 1) * binrange)::float8::text END) || ''' || whereclause || ''' || '' ORDER BY ' || columnname || ''') END) || '';'')::text query
                         FROM generate_series(-1, ' || nbinterval || ' - 1) serie
                              LEFT OUTER JOIN histo ON (serie = histo.bin),
                              (SELECT * FROM bins LIMIT 1) foo
                         ORDER BY serie;';
                RETURN QUERY EXECUTE query;
                IF left(columntype, 3) != 'int' THEN
                    RESET extra_float_digits;
                END IF;
            END IF;
        ELSE
            RAISE NOTICE '''%'' does not exists. Returning nothing...',columnname::text;
            RETURN;
        END IF;

        RETURN;
    END;
$$;


ALTER FUNCTION public.st_histogram(schemaname text, tablename text, columnname text, nbinterval integer, whereclause text) OWNER TO postgres;

--
-- TOC entry 1535 (class 1255 OID 4441947)
-- Name: st_nbiggestexteriorrings(public.geometry, integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_nbiggestexteriorrings(ingeom public.geometry, nbrings integer, comptype text DEFAULT 'AREA'::text) RETURNS SETOF public.geometry
    LANGUAGE plpgsql
    AS $$
    DECLARE
    BEGIN
    IF upper(comptype) = 'AREA' THEN
        RETURN QUERY SELECT ring
                     FROM (SELECT ST_MakePolygon(ST_ExteriorRing((ST_Dump(ingeom)).geom)) ring
                          ) foo
                     ORDER BY ST_Area(ring) DESC LIMIT nbrings;
    ELSIF upper(comptype) = 'NBPOINTS' THEN
        RETURN QUERY SELECT ring
                     FROM (SELECT ST_MakePolygon(ST_ExteriorRing((ST_Dump(ingeom)).geom)) ring
                          ) foo
                     ORDER BY ST_NPoints(ring) DESC LIMIT nbrings;
    ELSE
        RAISE NOTICE 'ST_NBiggestExteriorRings: Unsupported comparison type: ''%''. Try ''AREA'' or ''NBPOINTS''.', comptype;
        RETURN;
    END IF;
    END;
$$;


ALTER FUNCTION public.st_nbiggestexteriorrings(ingeom public.geometry, nbrings integer, comptype text) OWNER TO postgres;

--
-- TOC entry 1536 (class 1255 OID 4441948)
-- Name: st_randompoints(public.geometry, integer, numeric); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_randompoints(geom public.geometry, nb integer, seed numeric DEFAULT NULL::numeric) RETURNS SETOF public.geometry
    LANGUAGE plpgsql
    AS $$
    DECLARE
        pt geometry;
        xmin float8;
        xmax float8;
        ymin float8;
        ymax float8;
        xrange float8;
        yrange float8;
        srid int;
        count integer := 0;
        gtype text;
    BEGIN
        SELECT ST_GeometryType(geom) INTO gtype;

        -- Make sure the geometry is some kind of polygon
        IF (gtype IS NULL OR (gtype != 'ST_Polygon') AND (gtype != 'ST_MultiPolygon')) THEN
            RAISE NOTICE 'Attempting to get random points in a non polygon geometry';
            RETURN NEXT NULL;
            RETURN;
        END IF;

        -- Compute the extent
        SELECT ST_XMin(geom), ST_XMax(geom), ST_YMin(geom), ST_YMax(geom), ST_SRID(geom)
        INTO xmin, xmax, ymin, ymax, srid;

        -- and the range of the extent
        SELECT xmax - xmin, ymax - ymin
        INTO xrange, yrange;

        -- Set the seed if provided
        IF seed IS NOT NULL THEN
            PERFORM setseed(seed);
        END IF;

        -- Find valid points one after the other checking if they are inside the polygon
        WHILE count < nb LOOP
            SELECT ST_SetSRID(ST_MakePoint(xmin + xrange * random(), ymin + yrange * random()), srid)
            INTO pt;

            IF ST_Contains(geom, pt) THEN
                count := count + 1;
                RETURN NEXT pt;
            END IF;
        END LOOP;
        RETURN;
    END;
$$;


ALTER FUNCTION public.st_randompoints(geom public.geometry, nb integer, seed numeric) OWNER TO postgres;

--
-- TOC entry 1537 (class 1255 OID 4441949)
-- Name: st_removeoverlaps(public.geometry[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_removeoverlaps(geomarray public.geometry[]) RETURNS SETOF public.geometry
    LANGUAGE sql
    AS $$
    WITH geoms AS (
        SELECT unnest(geomarray) geom
    )
    SELECT ST_RemoveOverlaps(array_agg((geom, null)::geomval), 'NO_MERGE') FROM geoms;
$$;


ALTER FUNCTION public.st_removeoverlaps(geomarray public.geometry[]) OWNER TO postgres;

--
-- TOC entry 1538 (class 1255 OID 4441950)
-- Name: st_removeoverlaps(public.geomval[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_removeoverlaps(gvarray public.geomval[]) RETURNS SETOF public.geometry
    LANGUAGE sql
    AS $$
    SELECT ST_RemoveOverlaps(gvarray, 'LARGEST_VALUE');
$$;


ALTER FUNCTION public.st_removeoverlaps(gvarray public.geomval[]) OWNER TO postgres;

--
-- TOC entry 1539 (class 1255 OID 4441951)
-- Name: st_removeoverlaps(public.geometry[], text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_removeoverlaps(geomarray public.geometry[], mergemethod text) RETURNS SETOF public.geometry
    LANGUAGE sql
    AS $$
    WITH geoms AS (
        SELECT unnest(geomarray) geom
    )
    SELECT ST_RemoveOverlaps(array_agg((geom, ST_Area(geom))::geomval), mergemethod) FROM geoms;
$$;


ALTER FUNCTION public.st_removeoverlaps(geomarray public.geometry[], mergemethod text) OWNER TO postgres;

--
-- TOC entry 1540 (class 1255 OID 4441952)
-- Name: st_removeoverlaps(public.geomval[], text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_removeoverlaps(gvarray public.geomval[], mergemethod text) RETURNS SETOF public.geometry
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
    DECLARE
        query text;
    BEGIN
        mergemethod = upper(mergemethod);
--RAISE NOTICE 'method = %', mergemethod;
        query = E'WITH geomvals AS (\n'
             || E'  SELECT unnest($1) gv\n'
             || E'), geoms AS (\n'
             || E'  SELECT row_number() OVER () id, ST_CollectionExtract((gv).geom, 3) geom';
        IF right(mergemethod, 4) = 'AREA' THEN
            query = query || E', ST_Area((gv).geom) val\n';
        ELSE
            query = query || E', (gv).val\n';
        END IF;
        query = query || E'  FROM geomvals\n'
                      || E'), polygons AS (\n'
                      || E'  SELECT id, (ST_Dump(geom)).geom geom\n'
                      || E'  FROM geoms\n'
                      || E'), rings AS (\n'
                      || E'  SELECT id, ST_ExteriorRing((ST_DumpRings(geom)).geom) geom\n'
                      || E'  FROM polygons\n'
                      || E'), extrings_union AS (\n'
                      || E'  SELECT ST_Union(geom) geom\n'
                      || E'  FROM rings\n'
                      || E'), parts AS (\n'
                      || E'  SELECT (ST_Dump(ST_Polygonize(geom))).geom \n'
                      || E'  FROM extrings_union\n'
                      || E'), assigned_parts AS (\n'
                      || E'  SELECT id, \n'
                      || E'         count(*) OVER (PARTITION BY ST_AsEWKB(geom)) cnt, \n'
                      || E'         val, geom\n'
                      || E'  FROM (SELECT id, val, parts.geom,\n'
                      || E'               ST_Area(ST_Intersection(ori_polys.geom, parts.geom)) intarea\n'
                      || E'        FROM parts,\n'
                      || E'             (SELECT id, val, geom FROM geoms) ori_polys\n'
                      || E'        WHERE ST_Intersects(ori_polys.geom, parts.geom)\n'
                      || E'       ) foo\n'
                      || E'  WHERE intarea > 0 AND abs(intarea - ST_Area(geom)) < 0.001\n';

         IF right(mergemethod, 5) = '_EDGE' THEN
             query = query || E'), edge_length AS (\n'
                           || E'  SELECT a.id, b.id bid, \n'
                           || E'         ST_Union(ST_AsEWKB(a.geom)::geometry) geom,\n'
                           || E'         sum(ST_Length(ST_CollectionExtract(ST_Intersection(a.geom, b.geom), 2))) val\n'
                           || E'  FROM (SELECT id, geom FROM assigned_parts WHERE cnt > 1) a \n'
                           || E'      LEFT OUTER JOIN assigned_parts b \n'
                           || E'   ON (ST_AsEWKB(a.geom) != ST_AsEWKB(b.geom) AND \n'
                           || E'       ST_Touches(a.geom, b.geom) AND\n'
                           || E'      ST_Length(ST_CollectionExtract(ST_Intersection(a.geom, b.geom), 2)) > 0)\n'
                           || E'  GROUP BY a.id, b.id, ST_AsEWKB(a.geom)\n'
                           || E'    ), keep_parts AS (\n'
                           || E'   SELECT DISTINCT ON (ST_AsEWKB(geom)) id, geom\n'
                           || E'   FROM edge_length\n'
                           || E'   ORDER BY ST_AsEWKB(geom), val ';
             IF left(mergemethod, 7) = 'LONGEST' THEN
                 query = query || E'DESC';
             END IF;
             query = query || E', abs(id - bid)\n';

         ELSEIF left(mergemethod, 8) != 'NO_MERGE' AND left(mergemethod, 4) != 'OVER' THEN
             query = query || E'), keep_parts AS (\n'
                           || E'   SELECT DISTINCT ON (ST_AsEWKB(geom)) id, val, geom\n'
                           || E'   FROM assigned_parts\n'
                           || E'   ORDER BY ST_AsEWKB(geom), val';


             IF left(mergemethod, 7) = 'LARGEST' THEN
                 query = query || E' DESC';
             END IF;
             query = query || E'\n';
         END IF;

         IF left(mergemethod, 8) = 'NO_MERGE' OR left(mergemethod, 13) = 'OVERLAPS_ONLY' THEN
             query = query || E')\n';
             IF right(mergemethod, 4) = '_DUP' THEN
                    query = query || E'(SELECT geom\n';
             ELSE
                    query = query || E'(SELECT DISTINCT ON (ST_AsEWKB(geom)) geom\n';
             END IF;
             query = query || E' FROM assigned_parts\n'
                           || E' WHERE cnt > 1)\n';
             IF left(mergemethod, 8) = 'NO_MERGE' THEN
                 query = query || E'UNION ALL\n'
                               || E'(SELECT ST_Union(geom) geom\n'
                               || E' FROM assigned_parts\n'
                               || E' WHERE cnt = 1\n'
                               || E' GROUP BY id);\n';
             END IF;

         ELSEIF right(mergemethod, 5) = '_EDGE' THEN
            query = query || E')\n'
                          || E'SELECT ST_Union(geom) geom\n'
                          || E'FROM (SELECT id, geom FROM keep_parts\n'
                          || E'      UNION ALL \n'
                          || E'      SELECT id, geom FROM assigned_parts WHERE cnt = 1) foo\n'
                          || E'GROUP BY id\n';

         ELSE -- AREA or VALUE
             query = query || E')\n'
                           || E'SELECT ST_Union(geom) geom\n'
                           || E'FROM keep_parts\n'
                           || E'GROUP BY id;\n';
         END IF;
 --RAISE NOTICE 'query = %', query;
         RETURN QUERY EXECUTE query USING gvarray;
    END;
$_$;


ALTER FUNCTION public.st_removeoverlaps(gvarray public.geomval[], mergemethod text) OWNER TO postgres;

--
-- TOC entry 1541 (class 1255 OID 4441953)
-- Name: st_splitbygrid(public.geometry, double precision, double precision, double precision, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_splitbygrid(ingeom public.geometry, xgridsize double precision, ygridsize double precision DEFAULT NULL::double precision, xgridoffset double precision DEFAULT 0.0, ygridoffset double precision DEFAULT 0.0) RETURNS TABLE(geom public.geometry, tid bigint, x integer, y integer, tgeom public.geometry)
    LANGUAGE plpgsql
    AS $$
    DECLARE
        width int;
        height int;
        xminrounded double precision;
        yminrounded double precision;
        xmaxrounded double precision;
        ymaxrounded double precision;
        xmin double precision := ST_XMin(ingeom);
        ymin double precision := ST_YMin(ingeom);
        xmax double precision := ST_XMax(ingeom);
        ymax double precision := ST_YMax(ingeom);
        x int;
        y int;
        env geometry;
        xfloor int;
        yfloor int;
    BEGIN
        IF ingeom IS NULL OR ST_IsEmpty(ingeom) THEN
            RETURN QUERY SELECT ingeom, NULL::int8;
            RETURN;
        END IF;
        IF xgridsize IS NULL OR xgridsize <= 0 THEN
            RAISE NOTICE 'Defaulting xgridsize to 1...';
            xgridsize = 1;
        END IF;
        IF ygridsize IS NULL OR ygridsize <= 0 THEN
            ygridsize = xgridsize;
        END IF;
        xfloor = floor((xmin - xgridoffset) / xgridsize);
        xminrounded = xfloor * xgridsize + xgridoffset;
        xmaxrounded = ceil((xmax - xgridoffset) / xgridsize) * xgridsize + xgridoffset;
        yfloor = floor((ymin - ygridoffset) / ygridsize);
        yminrounded = yfloor * ygridsize + ygridoffset;
        ymaxrounded = ceil((ymax - ygridoffset) / ygridsize) * ygridsize + ygridoffset;

        width = round((xmaxrounded - xminrounded) / xgridsize);
        height = round((ymaxrounded - yminrounded) / ygridsize);

        FOR x IN 1..width LOOP
            FOR y IN 1..height LOOP
                env = ST_MakeEnvelope(xminrounded + (x - 1) * xgridsize, yminrounded + (y - 1) * ygridsize, xminrounded + x * xgridsize, yminrounded + y * ygridsize, ST_SRID(ingeom));
                IF ST_Intersects(env, ingeom) THEN
                     RETURN QUERY SELECT ST_Intersection(ingeom, env), ((xfloor::int8 + x) * 10000000 + (yfloor::int8 + y))::int8, xfloor + x, yfloor + y, env
                            WHERE ST_Dimension(ST_Intersection(ingeom, env)) = ST_Dimension(ingeom) OR
                                  ST_GeometryType(ST_Intersection(ingeom, env)) = ST_GeometryType(ingeom);
                 END IF;
            END LOOP;
        END LOOP;
    RETURN;
    END;
$$;


ALTER FUNCTION public.st_splitbygrid(ingeom public.geometry, xgridsize double precision, ygridsize double precision, xgridoffset double precision, ygridoffset double precision) OWNER TO postgres;

--
-- TOC entry 1518 (class 1255 OID 4441954)
-- Name: st_trimmulti(public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.st_trimmulti(geom public.geometry, minarea double precision DEFAULT 0.0) RETURNS public.geometry
    LANGUAGE sql IMMUTABLE
    AS $_$
    SELECT ST_Union(newgeom)
    FROM (SELECT ST_CollectionExtract((ST_Dump($1)).geom, 3) newgeom
         ) foo
    WHERE ST_Area(newgeom) > $2;
$_$;


ALTER FUNCTION public.st_trimmulti(geom public.geometry, minarea double precision) OWNER TO postgres;

--
-- TOC entry 1544 (class 1255 OID 4441955)
-- Name: tables_empty(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.tables_empty(username character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
                 DECLARE
				 	 num int;
                     statements CURSOR FOR
                         SELECT tablename FROM pg_tables
                         WHERE tableowner = username AND schemaname = 'public';
                 BEGIN
                     FOR stmt IN statements LOOP
                         EXECUTE 'SELECT count(*) FROM ' || quote_ident(stmt.tablename) INTO num;
						 IF num > 0 THEN
						 	RETURN false;
						 END IF;
                 	END LOOP;
					RETURN true;
                END;
                $$;


ALTER FUNCTION public.tables_empty(username character varying) OWNER TO postgres;

--
-- TOC entry 1543 (class 1255 OID 4441956)
-- Name: truncate_tables(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.truncate_tables(username character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
				DECLARE
   					statements CURSOR FOR
        				SELECT tablename FROM pg_tables
        				WHERE tableowner = username AND schemaname = 'public';
				BEGIN
    				FOR stmt IN statements LOOP
        				EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
   					END LOOP;
			   END;
			   $$;


ALTER FUNCTION public.truncate_tables(username character varying) OWNER TO postgres;

--
-- TOC entry 1542 (class 1255 OID 4441957)
-- Name: update_array_elements(jsonb, text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_array_elements(arr jsonb, key text, value jsonb) RETURNS jsonb
    LANGUAGE sql
    AS $$
    select jsonb_agg(jsonb_build_object(k, case when k <> key then v else value end))
    from jsonb_array_elements(arr) e(e), 
    lateral jsonb_each(e) p(k, v)
$$;


ALTER FUNCTION public.update_array_elements(arr jsonb, key text, value jsonb) OWNER TO postgres;

--
-- TOC entry 2249 (class 1255 OID 4441958)
-- Name: st_areaweightedsummarystats(public.geometry); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_areaweightedsummarystats(public.geometry) (
    SFUNC = public._st_areaweightedsummarystats_statefn,
    STYPE = public.agg_areaweightedstatsstate,
    FINALFUNC = public._st_areaweightedsummarystats_finalfn
);


ALTER AGGREGATE public.st_areaweightedsummarystats(public.geometry) OWNER TO postgres;

--
-- TOC entry 2250 (class 1255 OID 4441959)
-- Name: st_areaweightedsummarystats(public.geomval); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_areaweightedsummarystats(public.geomval) (
    SFUNC = public._st_areaweightedsummarystats_statefn,
    STYPE = public.agg_areaweightedstatsstate,
    FINALFUNC = public._st_areaweightedsummarystats_finalfn
);


ALTER AGGREGATE public.st_areaweightedsummarystats(public.geomval) OWNER TO postgres;

--
-- TOC entry 2251 (class 1255 OID 4441960)
-- Name: st_areaweightedsummarystats(public.geometry, double precision); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_areaweightedsummarystats(public.geometry, double precision) (
    SFUNC = public._st_areaweightedsummarystats_statefn,
    STYPE = public.agg_areaweightedstatsstate,
    FINALFUNC = public._st_areaweightedsummarystats_finalfn
);


ALTER AGGREGATE public.st_areaweightedsummarystats(public.geometry, double precision) OWNER TO postgres;

--
-- TOC entry 2252 (class 1255 OID 4441961)
-- Name: st_bufferedunion(public.geometry, double precision); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_bufferedunion(public.geometry, double precision) (
    SFUNC = public._st_bufferedunion_statefn,
    STYPE = public.geomval,
    FINALFUNC = public._st_bufferedunion_finalfn
);


ALTER AGGREGATE public.st_bufferedunion(public.geometry, double precision) OWNER TO postgres;

--
-- TOC entry 2253 (class 1255 OID 4441962)
-- Name: st_differenceagg(public.geometry, public.geometry); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_differenceagg(public.geometry, public.geometry) (
    SFUNC = public._st_differenceagg_statefn,
    STYPE = public.geometry
);


ALTER AGGREGATE public.st_differenceagg(public.geometry, public.geometry) OWNER TO postgres;

--
-- TOC entry 2254 (class 1255 OID 4441963)
-- Name: st_removeoverlaps(public.geometry); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_removeoverlaps(public.geometry) (
    SFUNC = public._st_removeoverlaps_statefn,
    STYPE = public.geomvaltxt[],
    FINALFUNC = public._st_removeoverlaps_finalfn
);


ALTER AGGREGATE public.st_removeoverlaps(public.geometry) OWNER TO postgres;

--
-- TOC entry 2255 (class 1255 OID 4441964)
-- Name: st_removeoverlaps(public.geometry, double precision); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_removeoverlaps(public.geometry, double precision) (
    SFUNC = public._st_removeoverlaps_statefn,
    STYPE = public.geomvaltxt[],
    FINALFUNC = public._st_removeoverlaps_finalfn
);


ALTER AGGREGATE public.st_removeoverlaps(public.geometry, double precision) OWNER TO postgres;

--
-- TOC entry 2256 (class 1255 OID 4441965)
-- Name: st_removeoverlaps(public.geometry, text); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_removeoverlaps(public.geometry, text) (
    SFUNC = public._st_removeoverlaps_statefn,
    STYPE = public.geomvaltxt[],
    FINALFUNC = public._st_removeoverlaps_finalfn
);


ALTER AGGREGATE public.st_removeoverlaps(public.geometry, text) OWNER TO postgres;

--
-- TOC entry 2257 (class 1255 OID 4441966)
-- Name: st_removeoverlaps(public.geometry, double precision, text); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_removeoverlaps(public.geometry, double precision, text) (
    SFUNC = public._st_removeoverlaps_statefn,
    STYPE = public.geomvaltxt[],
    FINALFUNC = public._st_removeoverlaps_finalfn
);


ALTER AGGREGATE public.st_removeoverlaps(public.geometry, double precision, text) OWNER TO postgres;

--
-- TOC entry 2258 (class 1255 OID 4441967)
-- Name: st_splitagg(public.geometry, public.geometry); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_splitagg(public.geometry, public.geometry) (
    SFUNC = public._st_splitagg_statefn,
    STYPE = public.geometry[]
);


ALTER AGGREGATE public.st_splitagg(public.geometry, public.geometry) OWNER TO postgres;

--
-- TOC entry 2259 (class 1255 OID 4441968)
-- Name: st_splitagg(public.geometry, public.geometry, double precision); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.st_splitagg(public.geometry, public.geometry, double precision) (
    SFUNC = public._st_splitagg_statefn,
    STYPE = public.geometry[]
);


ALTER AGGREGATE public.st_splitagg(public.geometry, public.geometry, double precision) OWNER TO postgres;

--
-- TOC entry 206 (class 1259 OID 4441969)
-- Name: access_token_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.access_token_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.access_token_id_seq OWNER TO monkey;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 207 (class 1259 OID 4441971)
-- Name: access_token; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.access_token (
    id bigint DEFAULT nextval('public.access_token_id_seq'::regclass) NOT NULL,
    user_id bigint,
    token text,
    expiration_time bigint
);


ALTER TABLE public.access_token OWNER TO monkey;

--
-- TOC entry 208 (class 1259 OID 4441978)
-- Name: account_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.account_id_seq OWNER TO monkey;

--
-- TOC entry 209 (class 1259 OID 4441980)
-- Name: account; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.account (
    id bigint DEFAULT nextval('public.account_id_seq'::regclass) NOT NULL,
    name text,
    hashed_password text,
    email text,
    profile_picture text,
    created bigint NOT NULL,
    is_moderator boolean NOT NULL
);


ALTER TABLE public.account OWNER TO monkey;

--
-- TOC entry 210 (class 1259 OID 4441987)
-- Name: account_permission_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.account_permission_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.account_permission_id_seq OWNER TO monkey;

--
-- TOC entry 211 (class 1259 OID 4441989)
-- Name: account_permission; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.account_permission (
    id bigint DEFAULT nextval('public.account_permission_id_seq'::regclass) NOT NULL,
    can_remove_label boolean,
    account_id bigint,
    can_unlock_image_description boolean,
    can_unlock_image boolean,
    can_monitor_system boolean,
    can_accept_trending_label boolean
);


ALTER TABLE public.account_permission OWNER TO monkey;

--
-- TOC entry 212 (class 1259 OID 4441993)
-- Name: image_annotation_data_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_annotation_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_annotation_data_id_seq OWNER TO monkey;

--
-- TOC entry 213 (class 1259 OID 4441995)
-- Name: annotation_data; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.annotation_data (
    id bigint DEFAULT nextval('public.image_annotation_data_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    annotation jsonb,
    annotation_type_id bigint NOT NULL,
    image_annotation_revision_id bigint,
    uuid uuid NOT NULL
);


ALTER TABLE public.annotation_data OWNER TO monkey;

--
-- TOC entry 214 (class 1259 OID 4442002)
-- Name: annotation_refinements_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.annotation_refinements_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.annotation_refinements_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 215 (class 1259 OID 4442004)
-- Name: annotation_refinements_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.annotation_refinements_per_country (
    id bigint DEFAULT nextval('public.annotation_refinements_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE public.annotation_refinements_per_country OWNER TO monkey;

--
-- TOC entry 216 (class 1259 OID 4442011)
-- Name: annotation_type; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.annotation_type (
    id bigint NOT NULL,
    name text
);


ALTER TABLE public.annotation_type OWNER TO monkey;

--
-- TOC entry 217 (class 1259 OID 4442017)
-- Name: annotations_per_app_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.annotations_per_app_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.annotations_per_app_id_seq OWNER TO monkey;

--
-- TOC entry 218 (class 1259 OID 4442019)
-- Name: annotations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.annotations_per_app (
    id bigint DEFAULT nextval('public.annotations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE public.annotations_per_app OWNER TO monkey;

--
-- TOC entry 219 (class 1259 OID 4442026)
-- Name: annotations_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.annotations_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.annotations_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 220 (class 1259 OID 4442028)
-- Name: annotations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.annotations_per_country (
    id bigint DEFAULT nextval('public.annotations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE public.annotations_per_country OWNER TO monkey;

--
-- TOC entry 221 (class 1259 OID 4442035)
-- Name: api_token_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.api_token_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_token_id_seq OWNER TO monkey;

--
-- TOC entry 222 (class 1259 OID 4442037)
-- Name: api_token; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.api_token (
    id bigint DEFAULT nextval('public.api_token_id_seq'::regclass) NOT NULL,
    description text,
    token text,
    issued_at bigint,
    account_id bigint,
    revoked boolean,
    expires_at bigint
);


ALTER TABLE public.api_token OWNER TO monkey;

--
-- TOC entry 223 (class 1259 OID 4442044)
-- Name: donations_per_app_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.donations_per_app_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.donations_per_app_id_seq OWNER TO monkey;

--
-- TOC entry 224 (class 1259 OID 4442046)
-- Name: donations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.donations_per_app (
    id bigint DEFAULT nextval('public.donations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE public.donations_per_app OWNER TO monkey;

--
-- TOC entry 225 (class 1259 OID 4442053)
-- Name: donations_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.donations_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.donations_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 226 (class 1259 OID 4442055)
-- Name: donations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.donations_per_country (
    id bigint DEFAULT nextval('public.donations_per_country_id_seq'::regclass) NOT NULL,
    country_code text,
    count bigint
);


ALTER TABLE public.donations_per_country OWNER TO monkey;

--
-- TOC entry 227 (class 1259 OID 4442062)
-- Name: image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image (
    id bigint NOT NULL,
    image_provider_id bigint,
    key text,
    unlocked boolean,
    hash bigint,
    width integer NOT NULL,
    height integer NOT NULL,
    sys_period tstzrange
);


ALTER TABLE public.image OWNER TO monkey;

--
-- TOC entry 228 (class 1259 OID 4442068)
-- Name: image_annotation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_annotation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_annotation_id_seq OWNER TO monkey;

--
-- TOC entry 229 (class 1259 OID 4442070)
-- Name: image_annotation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_annotation (
    id bigint DEFAULT nextval('public.image_annotation_id_seq'::regclass) NOT NULL,
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


ALTER TABLE public.image_annotation OWNER TO monkey;

--
-- TOC entry 230 (class 1259 OID 4442078)
-- Name: image_annotation_coverage_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_annotation_coverage_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_annotation_coverage_id_seq OWNER TO monkey;

--
-- TOC entry 231 (class 1259 OID 4442080)
-- Name: image_annotation_coverage; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_annotation_coverage (
    id bigint DEFAULT nextval('public.image_annotation_coverage_id_seq'::regclass) NOT NULL,
    image_id bigint,
    area integer,
    annotated_percentage integer
);


ALTER TABLE public.image_annotation_coverage OWNER TO monkey;

--
-- TOC entry 232 (class 1259 OID 4442084)
-- Name: image_annotation_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_annotation_history (
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


ALTER TABLE public.image_annotation_history OWNER TO monkey;

--
-- TOC entry 233 (class 1259 OID 4442091)
-- Name: image_annotation_refinement_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_annotation_refinement_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_annotation_refinement_id_seq OWNER TO monkey;

--
-- TOC entry 234 (class 1259 OID 4442093)
-- Name: image_annotation_refinement; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_annotation_refinement (
    id bigint DEFAULT nextval('public.image_annotation_refinement_id_seq'::regclass) NOT NULL,
    annotation_data_id bigint,
    label_id bigint NOT NULL,
    num_of_valid integer,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    fingerprint_of_last_modification text
);


ALTER TABLE public.image_annotation_refinement OWNER TO monkey;

--
-- TOC entry 235 (class 1259 OID 4442101)
-- Name: image_annotation_refinement_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_annotation_refinement_history (
    id bigint NOT NULL,
    annotation_data_id bigint,
    label_id bigint,
    num_of_valid integer,
    sys_period tstzrange NOT NULL,
    fingerprint_of_last_modification text
);


ALTER TABLE public.image_annotation_refinement_history OWNER TO monkey;

--
-- TOC entry 236 (class 1259 OID 4442107)
-- Name: image_annotation_revision_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_annotation_revision_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_annotation_revision_id_seq OWNER TO monkey;

--
-- TOC entry 237 (class 1259 OID 4442109)
-- Name: image_annotation_revision; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_annotation_revision (
    id bigint DEFAULT nextval('public.image_annotation_revision_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    revision integer
);


ALTER TABLE public.image_annotation_revision OWNER TO monkey;

--
-- TOC entry 238 (class 1259 OID 4442113)
-- Name: image_classification_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_classification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_classification_id_seq OWNER TO monkey;

--
-- TOC entry 239 (class 1259 OID 4442115)
-- Name: image_collection_image_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_collection_image_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_collection_image_id_seq OWNER TO monkey;

--
-- TOC entry 240 (class 1259 OID 4442117)
-- Name: image_collection_image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_collection_image (
    id bigint DEFAULT nextval('public.image_collection_image_id_seq'::regclass) NOT NULL,
    image_id bigint NOT NULL,
    user_image_collection_id bigint NOT NULL
);


ALTER TABLE public.image_collection_image OWNER TO monkey;

--
-- TOC entry 241 (class 1259 OID 4442121)
-- Name: image_description_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_description_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_description_id_seq OWNER TO monkey;

--
-- TOC entry 242 (class 1259 OID 4442123)
-- Name: image_description; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_description (
    id bigint DEFAULT nextval('public.image_description_id_seq'::regclass) NOT NULL,
    image_id bigint NOT NULL,
    description text,
    uuid uuid NOT NULL,
    processed_by bigint,
    num_of_valid integer,
    num_of_invalid integer,
    state public.state_type,
    language_id bigint NOT NULL,
    sys_period tstzrange NOT NULL
);


ALTER TABLE public.image_description OWNER TO monkey;

--
-- TOC entry 243 (class 1259 OID 4442130)
-- Name: image_description_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_description_history (
    id bigint NOT NULL,
    image_id bigint NOT NULL,
    description text,
    uuid uuid NOT NULL,
    processed_by bigint,
    num_of_valid integer,
    num_of_invalid integer,
    state public.state_type,
    language_id bigint NOT NULL,
    sys_period tstzrange
);


ALTER TABLE public.image_description_history OWNER TO monkey;

--
-- TOC entry 244 (class 1259 OID 4442136)
-- Name: image_descriptions_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_descriptions_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_descriptions_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 245 (class 1259 OID 4442138)
-- Name: image_descriptions_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_descriptions_per_country (
    id bigint DEFAULT nextval('public.image_descriptions_per_country_id_seq'::regclass) NOT NULL,
    country_code text,
    count bigint
);


ALTER TABLE public.image_descriptions_per_country OWNER TO monkey;

--
-- TOC entry 246 (class 1259 OID 4442145)
-- Name: image_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_id_seq OWNER TO monkey;

--
-- TOC entry 4304 (class 0 OID 0)
-- Dependencies: 246
-- Name: image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: monkey
--

ALTER SEQUENCE public.image_id_seq OWNED BY public.image.id;


--
-- TOC entry 247 (class 1259 OID 4442147)
-- Name: image_label_suggestion_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_label_suggestion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_label_suggestion_id_seq OWNER TO monkey;

--
-- TOC entry 248 (class 1259 OID 4442149)
-- Name: image_label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_label_suggestion (
    id bigint DEFAULT nextval('public.image_label_suggestion_id_seq'::regclass) NOT NULL,
    label_suggestion_id bigint,
    image_id bigint,
    fingerprint_of_last_modification text,
    annotatable boolean NOT NULL,
    sys_period tstzrange
);


ALTER TABLE public.image_label_suggestion OWNER TO monkey;

--
-- TOC entry 249 (class 1259 OID 4442156)
-- Name: image_provider_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_provider_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_provider_id_seq OWNER TO monkey;

--
-- TOC entry 250 (class 1259 OID 4442158)
-- Name: image_provider; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_provider (
    id bigint DEFAULT nextval('public.image_provider_id_seq'::regclass) NOT NULL,
    name text
);


ALTER TABLE public.image_provider OWNER TO monkey;

--
-- TOC entry 251 (class 1259 OID 4442165)
-- Name: image_quarantine_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_quarantine_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_quarantine_id_seq OWNER TO monkey;

--
-- TOC entry 252 (class 1259 OID 4442167)
-- Name: image_quarantine; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_quarantine (
    id bigint DEFAULT nextval('public.image_quarantine_id_seq'::regclass) NOT NULL,
    image_id bigint
);


ALTER TABLE public.image_quarantine OWNER TO monkey;

--
-- TOC entry 253 (class 1259 OID 4442171)
-- Name: report_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.report_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.report_id_seq OWNER TO monkey;

--
-- TOC entry 254 (class 1259 OID 4442173)
-- Name: image_report; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_report (
    id bigint DEFAULT nextval('public.report_id_seq'::regclass) NOT NULL,
    reason text,
    image_id bigint
);


ALTER TABLE public.image_report OWNER TO monkey;

--
-- TOC entry 255 (class 1259 OID 4442180)
-- Name: image_source_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_source_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_source_id_seq OWNER TO monkey;

--
-- TOC entry 256 (class 1259 OID 4442182)
-- Name: image_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_source (
    id bigint DEFAULT nextval('public.image_source_id_seq'::regclass) NOT NULL,
    url text,
    image_id bigint
);


ALTER TABLE public.image_source OWNER TO monkey;

--
-- TOC entry 257 (class 1259 OID 4442189)
-- Name: image_validation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_validation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_validation_id_seq OWNER TO monkey;

--
-- TOC entry 258 (class 1259 OID 4442191)
-- Name: image_validation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_validation (
    id bigint DEFAULT nextval('public.image_validation_id_seq'::regclass) NOT NULL,
    image_id bigint,
    label_id bigint,
    num_of_valid integer,
    num_of_invalid integer,
    sys_period tstzrange DEFAULT tstzrange(now(), NULL::timestamp with time zone) NOT NULL,
    fingerprint_of_last_modification text,
    uuid uuid NOT NULL,
    num_of_not_annotatable integer NOT NULL
);


ALTER TABLE public.image_validation OWNER TO monkey;

--
-- TOC entry 259 (class 1259 OID 4442199)
-- Name: image_validation_history; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_validation_history (
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


ALTER TABLE public.image_validation_history OWNER TO monkey;

--
-- TOC entry 260 (class 1259 OID 4442205)
-- Name: image_validation_source_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.image_validation_source_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_validation_source_id_seq OWNER TO monkey;

--
-- TOC entry 261 (class 1259 OID 4442207)
-- Name: image_validation_source; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.image_validation_source (
    id bigint DEFAULT nextval('public.image_validation_source_id_seq'::regclass) NOT NULL,
    image_validation_id bigint,
    image_source_id bigint
);


ALTER TABLE public.image_validation_source OWNER TO monkey;

--
-- TOC entry 262 (class 1259 OID 4442211)
-- Name: imagehunt_task_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.imagehunt_task_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.imagehunt_task_id_seq OWNER TO monkey;

--
-- TOC entry 263 (class 1259 OID 4442213)
-- Name: imagehunt_task; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.imagehunt_task (
    image_validation_id bigint,
    id bigint DEFAULT nextval('public.imagehunt_task_id_seq'::regclass) NOT NULL,
    created bigint
);


ALTER TABLE public.imagehunt_task OWNER TO monkey;

--
-- TOC entry 264 (class 1259 OID 4442217)
-- Name: name_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.name_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.name_id_seq OWNER TO monkey;

--
-- TOC entry 265 (class 1259 OID 4442219)
-- Name: label; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.label (
    id bigint DEFAULT nextval('public.name_id_seq'::regclass) NOT NULL,
    name text,
    parent_id bigint,
    uuid uuid NOT NULL,
    label_type public.label_type
);


ALTER TABLE public.label OWNER TO monkey;

--
-- TOC entry 266 (class 1259 OID 4442226)
-- Name: label_accessor_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.label_accessor_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.label_accessor_id_seq OWNER TO monkey;

--
-- TOC entry 267 (class 1259 OID 4442228)
-- Name: label_accessor; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.label_accessor (
    id bigint DEFAULT nextval('public.label_accessor_id_seq'::regclass) NOT NULL,
    label_id bigint,
    accessor text
);


ALTER TABLE public.label_accessor OWNER TO monkey;

--
-- TOC entry 268 (class 1259 OID 4442235)
-- Name: label_example_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.label_example_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.label_example_id_seq OWNER TO monkey;

--
-- TOC entry 269 (class 1259 OID 4442237)
-- Name: label_example; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.label_example (
    id bigint DEFAULT nextval('public.label_example_id_seq'::regclass),
    attribution text,
    label_id bigint,
    filename text
);


ALTER TABLE public.label_example OWNER TO monkey;

--
-- TOC entry 270 (class 1259 OID 4442244)
-- Name: label_suggestion_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.label_suggestion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.label_suggestion_id_seq OWNER TO monkey;

--
-- TOC entry 271 (class 1259 OID 4442246)
-- Name: label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.label_suggestion (
    id bigint DEFAULT nextval('public.label_suggestion_id_seq'::regclass) NOT NULL,
    name text,
    proposed_by bigint
);


ALTER TABLE public.label_suggestion OWNER TO monkey;

--
-- TOC entry 272 (class 1259 OID 4442253)
-- Name: language_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.language_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.language_id_seq OWNER TO monkey;

--
-- TOC entry 273 (class 1259 OID 4442255)
-- Name: language; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.language (
    id bigint DEFAULT nextval('public.language_id_seq'::regclass) NOT NULL,
    name text,
    fullname text NOT NULL
);


ALTER TABLE public.language OWNER TO monkey;

--
-- TOC entry 274 (class 1259 OID 4442262)
-- Name: quiz_answer_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.quiz_answer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.quiz_answer_id_seq OWNER TO monkey;

--
-- TOC entry 275 (class 1259 OID 4442264)
-- Name: quiz_answer; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.quiz_answer (
    id bigint DEFAULT nextval('public.quiz_answer_id_seq'::regclass) NOT NULL,
    quiz_question_id bigint,
    label_id bigint
);


ALTER TABLE public.quiz_answer OWNER TO monkey;

--
-- TOC entry 276 (class 1259 OID 4442268)
-- Name: quiz_question_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.quiz_question_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.quiz_question_id_seq OWNER TO monkey;

--
-- TOC entry 277 (class 1259 OID 4442270)
-- Name: quiz_question; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.quiz_question (
    id bigint DEFAULT nextval('public.quiz_question_id_seq'::regclass) NOT NULL,
    question text,
    refines_label_id bigint,
    recommended_control public.control_type,
    allow_unknown boolean,
    allow_other boolean,
    browse_by_example boolean,
    multiselect boolean,
    uuid uuid NOT NULL
);


ALTER TABLE public.quiz_question OWNER TO monkey;

--
-- TOC entry 278 (class 1259 OID 4442277)
-- Name: trending_label_bot_task_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.trending_label_bot_task_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.trending_label_bot_task_id_seq OWNER TO monkey;

--
-- TOC entry 279 (class 1259 OID 4442279)
-- Name: trending_label_bot_task; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.trending_label_bot_task (
    id bigint DEFAULT nextval('public.trending_label_bot_task_id_seq'::regclass) NOT NULL,
    trending_label_suggestion_id bigint,
    branch_name text,
    state public.label_bot_state_type,
    job_url text,
    try integer,
    label_type public.label_bot_label_type NOT NULL,
    description text,
    plural text
);


ALTER TABLE public.trending_label_bot_task OWNER TO monkey;

--
-- TOC entry 280 (class 1259 OID 4442286)
-- Name: trending_label_suggestion_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.trending_label_suggestion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.trending_label_suggestion_id_seq OWNER TO monkey;

--
-- TOC entry 281 (class 1259 OID 4442288)
-- Name: trending_label_suggestion; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.trending_label_suggestion (
    id bigint DEFAULT nextval('public.trending_label_suggestion_id_seq'::regclass) NOT NULL,
    label_suggestion_id bigint,
    num_of_last_sent integer,
    github_issue_id bigint NOT NULL,
    closed boolean NOT NULL,
    productive_label_id bigint
);


ALTER TABLE public.trending_label_suggestion OWNER TO monkey;

--
-- TOC entry 282 (class 1259 OID 4442292)
-- Name: user_annotation_blacklist_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.user_annotation_blacklist_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_annotation_blacklist_id_seq OWNER TO monkey;

--
-- TOC entry 283 (class 1259 OID 4442294)
-- Name: user_annotation_blacklist; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.user_annotation_blacklist (
    id bigint DEFAULT nextval('public.user_annotation_blacklist_id_seq'::regclass) NOT NULL,
    account_id bigint,
    image_validation_id bigint
);


ALTER TABLE public.user_annotation_blacklist OWNER TO monkey;

--
-- TOC entry 284 (class 1259 OID 4442298)
-- Name: user_image_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.user_image_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_image_id_seq OWNER TO monkey;

--
-- TOC entry 285 (class 1259 OID 4442300)
-- Name: user_image; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.user_image (
    id bigint DEFAULT nextval('public.user_image_id_seq'::regclass) NOT NULL,
    image_id bigint,
    account_id bigint
);


ALTER TABLE public.user_image OWNER TO monkey;

--
-- TOC entry 286 (class 1259 OID 4442304)
-- Name: user_image_annotation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.user_image_annotation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_image_annotation_id_seq OWNER TO monkey;

--
-- TOC entry 287 (class 1259 OID 4442306)
-- Name: user_image_annotation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.user_image_annotation (
    id bigint DEFAULT nextval('public.user_image_annotation_id_seq'::regclass) NOT NULL,
    image_annotation_id bigint,
    account_id bigint,
    "timestamp" timestamp with time zone
);


ALTER TABLE public.user_image_annotation OWNER TO monkey;

--
-- TOC entry 288 (class 1259 OID 4442310)
-- Name: user_image_collection_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.user_image_collection_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_image_collection_id_seq OWNER TO monkey;

--
-- TOC entry 289 (class 1259 OID 4442312)
-- Name: user_image_collection; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.user_image_collection (
    id bigint DEFAULT nextval('public.user_image_collection_id_seq'::regclass) NOT NULL,
    account_id bigint,
    name text,
    description text
);


ALTER TABLE public.user_image_collection OWNER TO monkey;

--
-- TOC entry 290 (class 1259 OID 4442319)
-- Name: user_image_validation_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.user_image_validation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_image_validation_id_seq OWNER TO monkey;

--
-- TOC entry 291 (class 1259 OID 4442321)
-- Name: user_image_validation; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.user_image_validation (
    id bigint DEFAULT nextval('public.user_image_validation_id_seq'::regclass) NOT NULL,
    image_validation_id bigint,
    account_id bigint,
    "timestamp" timestamp with time zone
);


ALTER TABLE public.user_image_validation OWNER TO monkey;

--
-- TOC entry 292 (class 1259 OID 4442325)
-- Name: validations_per_app_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.validations_per_app_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.validations_per_app_id_seq OWNER TO monkey;

--
-- TOC entry 293 (class 1259 OID 4442327)
-- Name: validations_per_app; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.validations_per_app (
    id bigint DEFAULT nextval('public.validations_per_app_id_seq'::regclass) NOT NULL,
    app_identifier text,
    count bigint
);


ALTER TABLE public.validations_per_app OWNER TO monkey;

--
-- TOC entry 294 (class 1259 OID 4442334)
-- Name: validations_per_country_id_seq; Type: SEQUENCE; Schema: public; Owner: monkey
--

CREATE SEQUENCE public.validations_per_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.validations_per_country_id_seq OWNER TO monkey;

--
-- TOC entry 295 (class 1259 OID 4442336)
-- Name: validations_per_country; Type: TABLE; Schema: public; Owner: monkey
--

CREATE TABLE public.validations_per_country (
    id bigint DEFAULT nextval('public.validations_per_country_id_seq'::regclass) NOT NULL,
    count bigint,
    country_code text
);


ALTER TABLE public.validations_per_country OWNER TO monkey;

--
-- TOC entry 3860 (class 2604 OID 4442343)
-- Name: image id; Type: DEFAULT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image ALTER COLUMN id SET DEFAULT nextval('public.image_id_seq'::regclass);


--
-- TOC entry 3897 (class 2606 OID 4442345)
-- Name: access_token access_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.access_token
    ADD CONSTRAINT access_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3906 (class 2606 OID 4442347)
-- Name: account_permission account_permission_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.account_permission
    ADD CONSTRAINT account_permission_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3909 (class 2606 OID 4442349)
-- Name: annotation_data annotation_data_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_data
    ADD CONSTRAINT annotation_data_uuid_unique UNIQUE (uuid);


--
-- TOC entry 3916 (class 2606 OID 4442351)
-- Name: annotation_refinements_per_country annotation_refinements_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_refinements_per_country
    ADD CONSTRAINT annotation_refinements_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 3918 (class 2606 OID 4442353)
-- Name: annotation_refinements_per_country annotation_refinements_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_refinements_per_country
    ADD CONSTRAINT annotation_refinements_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3920 (class 2606 OID 4442355)
-- Name: annotation_type annotation_type_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_type
    ADD CONSTRAINT annotation_type_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3922 (class 2606 OID 4442357)
-- Name: annotation_type annotation_type_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_type
    ADD CONSTRAINT annotation_type_name_unique UNIQUE (name);


--
-- TOC entry 3924 (class 2606 OID 4442359)
-- Name: annotations_per_app annotations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotations_per_app
    ADD CONSTRAINT annotations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 3926 (class 2606 OID 4442361)
-- Name: annotations_per_app annotations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotations_per_app
    ADD CONSTRAINT annotations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3928 (class 2606 OID 4442363)
-- Name: annotations_per_country annotations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotations_per_country
    ADD CONSTRAINT annotations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 3930 (class 2606 OID 4442365)
-- Name: annotations_per_country annotations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotations_per_country
    ADD CONSTRAINT annotations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3932 (class 2606 OID 4442367)
-- Name: api_token api_token_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.api_token
    ADD CONSTRAINT api_token_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3934 (class 2606 OID 4442369)
-- Name: api_token api_token_token_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.api_token
    ADD CONSTRAINT api_token_token_unique UNIQUE (token);


--
-- TOC entry 3937 (class 2606 OID 4442371)
-- Name: donations_per_app donations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.donations_per_app
    ADD CONSTRAINT donations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 3939 (class 2606 OID 4442373)
-- Name: donations_per_app donations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.donations_per_app
    ADD CONSTRAINT donations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3941 (class 2606 OID 4442375)
-- Name: donations_per_country donations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.donations_per_country
    ADD CONSTRAINT donations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 3943 (class 2606 OID 4442377)
-- Name: donations_per_country donations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.donations_per_country
    ADD CONSTRAINT donations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3966 (class 2606 OID 4442379)
-- Name: image_annotation_coverage image_annotation_coverage_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_coverage
    ADD CONSTRAINT image_annotation_coverage_image_id_unique UNIQUE (image_id);


--
-- TOC entry 3968 (class 2606 OID 4442381)
-- Name: image_annotation_coverage image_annotation_coverage_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_coverage
    ADD CONSTRAINT image_annotation_coverage_pkey PRIMARY KEY (id);


--
-- TOC entry 3914 (class 2606 OID 4442383)
-- Name: annotation_data image_annotation_data_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_data
    ADD CONSTRAINT image_annotation_data_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3957 (class 2606 OID 4442385)
-- Name: image_annotation image_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation
    ADD CONSTRAINT image_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3960 (class 2606 OID 4442387)
-- Name: image_annotation image_annotation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation
    ADD CONSTRAINT image_annotation_image_label_uniquekey UNIQUE (image_id, label_id, auto_generated);


--
-- TOC entry 3972 (class 2606 OID 4442389)
-- Name: image_annotation_refinement image_annotation_refinement_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3974 (class 2606 OID 4442391)
-- Name: image_annotation_refinement image_annotation_refinement_label_annotation_data_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_annotation_data_unique UNIQUE (annotation_data_id, label_id);


--
-- TOC entry 3977 (class 2606 OID 4442393)
-- Name: image_annotation_revision image_annotation_revision_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3981 (class 2606 OID 4442395)
-- Name: image_collection_image image_collection_image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_collection_image
    ADD CONSTRAINT image_collection_image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3984 (class 2606 OID 4442397)
-- Name: image_collection_image image_collection_image_user_image_collection_id_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_collection_image
    ADD CONSTRAINT image_collection_image_user_image_collection_id_image_id_unique UNIQUE (image_id, user_image_collection_id);


--
-- TOC entry 3993 (class 2606 OID 4442399)
-- Name: image_description image_description_text_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_description
    ADD CONSTRAINT image_description_text_image_id_unique UNIQUE (image_id, description);


--
-- TOC entry 3995 (class 2606 OID 4442401)
-- Name: image_description image_description_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_description
    ADD CONSTRAINT image_description_uuid_unique UNIQUE (uuid);


--
-- TOC entry 3997 (class 2606 OID 4442403)
-- Name: image_descriptions_per_country image_descriptions_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_descriptions_per_country
    ADD CONSTRAINT image_descriptions_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3999 (class 2606 OID 4442405)
-- Name: image_descriptions_per_country image_descriptions_per_country_unique_country_code; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_descriptions_per_country
    ADD CONSTRAINT image_descriptions_per_country_unique_country_code UNIQUE (country_code);


--
-- TOC entry 3947 (class 2606 OID 4442407)
-- Name: image image_hash_unique_key; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image
    ADD CONSTRAINT image_hash_unique_key UNIQUE (hash);


--
-- TOC entry 3949 (class 2606 OID 4442409)
-- Name: image image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image
    ADD CONSTRAINT image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3953 (class 2606 OID 4442411)
-- Name: image image_key_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image
    ADD CONSTRAINT image_key_unique UNIQUE (image_provider_id, key);


--
-- TOC entry 4003 (class 2606 OID 4442413)
-- Name: image_label_suggestion image_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4005 (class 2606 OID 4442415)
-- Name: image_label_suggestion image_label_suggestion_image_id_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_label_suggestion_id_unique UNIQUE (label_suggestion_id, image_id);


--
-- TOC entry 4007 (class 2606 OID 4442417)
-- Name: image_provider image_provider_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_provider
    ADD CONSTRAINT image_provider_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4010 (class 2606 OID 4442419)
-- Name: image_quarantine image_quarantine_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_quarantine
    ADD CONSTRAINT image_quarantine_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4012 (class 2606 OID 4442421)
-- Name: image_quarantine image_quarantine_image_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_unique UNIQUE (image_id);


--
-- TOC entry 4018 (class 2606 OID 4442423)
-- Name: image_source image_source_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_source
    ADD CONSTRAINT image_source_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4022 (class 2606 OID 4442425)
-- Name: image_validation image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation
    ADD CONSTRAINT image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4025 (class 2606 OID 4442427)
-- Name: image_validation image_validation_image_label_uniquekey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation
    ADD CONSTRAINT image_validation_image_label_uniquekey UNIQUE (image_id, label_id);


--
-- TOC entry 4031 (class 2606 OID 4442429)
-- Name: image_validation_source image_validation_source_id; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation_source
    ADD CONSTRAINT image_validation_source_id PRIMARY KEY (id);


--
-- TOC entry 4034 (class 2606 OID 4442431)
-- Name: imagehunt_task imagehunt_task_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.imagehunt_task
    ADD CONSTRAINT imagehunt_task_pkey PRIMARY KEY (id);


--
-- TOC entry 4045 (class 2606 OID 4442433)
-- Name: label_accessor label_accessor_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_accessor
    ADD CONSTRAINT label_accessor_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4047 (class 2606 OID 4442435)
-- Name: label_accessor label_accessor_label_id_accessor_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_accessor
    ADD CONSTRAINT label_accessor_label_id_accessor_unique UNIQUE (label_id, accessor);


--
-- TOC entry 4037 (class 2606 OID 4442437)
-- Name: label label_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label
    ADD CONSTRAINT label_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4040 (class 2606 OID 4442439)
-- Name: label label_name_parent_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label
    ADD CONSTRAINT label_name_parent_id_unique UNIQUE (name, parent_id);


--
-- TOC entry 4051 (class 2606 OID 4442441)
-- Name: label_suggestion label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_suggestion
    ADD CONSTRAINT label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4053 (class 2606 OID 4442443)
-- Name: label_suggestion label_suggestion_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_suggestion
    ADD CONSTRAINT label_suggestion_name_unique UNIQUE (name);


--
-- TOC entry 4042 (class 2606 OID 4442445)
-- Name: label label_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label
    ADD CONSTRAINT label_uuid_unique UNIQUE (uuid);


--
-- TOC entry 4055 (class 2606 OID 4442447)
-- Name: language language_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.language
    ADD CONSTRAINT language_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4057 (class 2606 OID 4442449)
-- Name: language language_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.language
    ADD CONSTRAINT language_name_unique UNIQUE (name);


--
-- TOC entry 4061 (class 2606 OID 4442451)
-- Name: quiz_answer quiz_answer_label_id_quiz_question_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_answer
    ADD CONSTRAINT quiz_answer_label_id_quiz_question_unique UNIQUE (quiz_question_id, label_id);


--
-- TOC entry 4063 (class 2606 OID 4442453)
-- Name: quiz_answer quiz_id_pley; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_answer
    ADD CONSTRAINT quiz_id_pley PRIMARY KEY (id);


--
-- TOC entry 4066 (class 2606 OID 4442455)
-- Name: quiz_question quiz_question_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_question
    ADD CONSTRAINT quiz_question_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4068 (class 2606 OID 4442457)
-- Name: quiz_question quiz_question_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_question
    ADD CONSTRAINT quiz_question_uuid_unique UNIQUE (uuid);


--
-- TOC entry 4015 (class 2606 OID 4442459)
-- Name: image_report report_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_report
    ADD CONSTRAINT report_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4071 (class 2606 OID 4442461)
-- Name: trending_label_bot_task trending_label_bot_task_branch_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_bot_task
    ADD CONSTRAINT trending_label_bot_task_branch_name_unique UNIQUE (branch_name);


--
-- TOC entry 4073 (class 2606 OID 4442463)
-- Name: trending_label_bot_task trending_label_bot_task_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_bot_task
    ADD CONSTRAINT trending_label_bot_task_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4079 (class 2606 OID 4442465)
-- Name: trending_label_suggestion trending_label_suggestion_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4075 (class 2606 OID 4442467)
-- Name: trending_label_bot_task trending_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_bot_task
    ADD CONSTRAINT trending_label_suggestion_id_unique UNIQUE (trending_label_suggestion_id);


--
-- TOC entry 4081 (class 2606 OID 4442469)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_unique UNIQUE (label_suggestion_id);


--
-- TOC entry 4085 (class 2606 OID 4442471)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_image_validation_id_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_image_validation_id_unique UNIQUE (account_id, image_validation_id);


--
-- TOC entry 4097 (class 2606 OID 4442473)
-- Name: user_image_annotation user_annotation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_annotation
    ADD CONSTRAINT user_annotation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3900 (class 2606 OID 4442475)
-- Name: account user_email_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT user_email_unique UNIQUE (email);


--
-- TOC entry 3902 (class 2606 OID 4442477)
-- Name: account user_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT user_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4087 (class 2606 OID 4442479)
-- Name: user_annotation_blacklist user_image_blacklist_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_annotation_blacklist
    ADD CONSTRAINT user_image_blacklist_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4101 (class 2606 OID 4442481)
-- Name: user_image_collection user_image_collection_account_id_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_collection
    ADD CONSTRAINT user_image_collection_account_id_name_unique UNIQUE (account_id, name);


--
-- TOC entry 4103 (class 2606 OID 4442483)
-- Name: user_image_collection user_image_collection_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_collection
    ADD CONSTRAINT user_image_collection_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4092 (class 2606 OID 4442485)
-- Name: user_image user_image_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image
    ADD CONSTRAINT user_image_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4108 (class 2606 OID 4442487)
-- Name: user_image_validation user_image_validation_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_validation
    ADD CONSTRAINT user_image_validation_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3904 (class 2606 OID 4442489)
-- Name: account user_name_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT user_name_unique UNIQUE (name);


--
-- TOC entry 4110 (class 2606 OID 4442491)
-- Name: validations_per_app validations_per_app_app_identifier_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.validations_per_app
    ADD CONSTRAINT validations_per_app_app_identifier_unique UNIQUE (app_identifier);


--
-- TOC entry 4112 (class 2606 OID 4442493)
-- Name: validations_per_app validations_per_app_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.validations_per_app
    ADD CONSTRAINT validations_per_app_id_pkey PRIMARY KEY (id);


--
-- TOC entry 4114 (class 2606 OID 4442495)
-- Name: validations_per_country validations_per_country_country_code_unique; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.validations_per_country
    ADD CONSTRAINT validations_per_country_country_code_unique UNIQUE (country_code);


--
-- TOC entry 4116 (class 2606 OID 4442497)
-- Name: validations_per_country validations_per_country_id_pkey; Type: CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.validations_per_country
    ADD CONSTRAINT validations_per_country_id_pkey PRIMARY KEY (id);


--
-- TOC entry 3898 (class 1259 OID 4442498)
-- Name: fki_access_token_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_access_token_user_id_fkey ON public.access_token USING btree (user_id);


--
-- TOC entry 3907 (class 1259 OID 4442499)
-- Name: fki_account_permission_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_account_permission_account_id_fkey ON public.account_permission USING btree (account_id);


--
-- TOC entry 3910 (class 1259 OID 4442500)
-- Name: fki_annotation_data_annotation_type_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_annotation_type_fkey ON public.annotation_data USING btree (annotation_type_id);


--
-- TOC entry 3911 (class 1259 OID 4442501)
-- Name: fki_annotation_data_image_annotation_revision_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_annotation_data_image_annotation_revision_id_fkey ON public.annotation_data USING btree (image_annotation_revision_id);


--
-- TOC entry 3935 (class 1259 OID 4442502)
-- Name: fki_api_token_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_api_token_account_id_fkey ON public.api_token USING btree (account_id);


--
-- TOC entry 3963 (class 1259 OID 4442503)
-- Name: fki_image_annotation_coverage_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_coverage_image_id_fkey ON public.image_annotation_coverage USING btree (image_id);


--
-- TOC entry 3912 (class 1259 OID 4442504)
-- Name: fki_image_annotation_data_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_data_annotation_id_fkey ON public.annotation_data USING btree (image_annotation_id);


--
-- TOC entry 3955 (class 1259 OID 4442505)
-- Name: fki_image_annotation_label_id_key; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_label_id_key ON public.image_annotation USING btree (label_id);


--
-- TOC entry 3975 (class 1259 OID 4442506)
-- Name: fki_image_annotation_revision_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_annotation_revision_image_annotation_id_fkey ON public.image_annotation_revision USING btree (image_annotation_id);


--
-- TOC entry 3978 (class 1259 OID 4442507)
-- Name: fki_image_collection_image_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_collection_image_image_id_fkey ON public.image_collection_image USING btree (image_id);


--
-- TOC entry 3979 (class 1259 OID 4442508)
-- Name: fki_image_collection_image_user_image_collection_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_collection_image_user_image_collection_id_fkey ON public.image_collection_image USING btree (user_image_collection_id);


--
-- TOC entry 3986 (class 1259 OID 4442509)
-- Name: fki_image_description_language_language_id; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_description_language_language_id ON public.image_description USING btree (language_id);


--
-- TOC entry 3987 (class 1259 OID 4442510)
-- Name: fki_image_description_unlocked_by_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_description_unlocked_by_account_id_fkey ON public.image_description USING btree (processed_by);


--
-- TOC entry 3988 (class 1259 OID 4442511)
-- Name: fki_image_id_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_id_image_id_fkey ON public.image_description USING btree (image_id);


--
-- TOC entry 4000 (class 1259 OID 4442512)
-- Name: fki_image_label_suggestion_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_image_id_fkey ON public.image_label_suggestion USING btree (image_id);


--
-- TOC entry 4001 (class 1259 OID 4442513)
-- Name: fki_image_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_label_suggestion_label_suggestion_id_fkey ON public.image_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 3944 (class 1259 OID 4442514)
-- Name: fki_image_provider_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_provider_id_fkey ON public.image USING btree (image_provider_id);


--
-- TOC entry 4008 (class 1259 OID 4442515)
-- Name: fki_image_quarantine_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quarantine_image_id_fkey ON public.image_quarantine USING btree (image_id);


--
-- TOC entry 3969 (class 1259 OID 4442516)
-- Name: fki_image_quiz_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_image_annotation_id_fkey ON public.image_annotation_refinement USING btree (annotation_data_id);


--
-- TOC entry 3970 (class 1259 OID 4442517)
-- Name: fki_image_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_quiz_label_id_fkey ON public.image_annotation_refinement USING btree (label_id);


--
-- TOC entry 4013 (class 1259 OID 4442518)
-- Name: fki_image_report_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_report_image_id_fkey ON public.image_report USING btree (image_id);


--
-- TOC entry 4016 (class 1259 OID 4442519)
-- Name: fki_image_source_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_source_image_id_fkey ON public.image_source USING btree (image_id);


--
-- TOC entry 4019 (class 1259 OID 4442520)
-- Name: fki_image_validation_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_image_id_fkey ON public.image_validation USING btree (image_id);


--
-- TOC entry 4020 (class 1259 OID 4442521)
-- Name: fki_image_validation_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_label_id_fkey ON public.image_validation USING btree (label_id);


--
-- TOC entry 4028 (class 1259 OID 4442522)
-- Name: fki_image_validation_source_image_source_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_source_id_fkey ON public.image_validation_source USING btree (image_source_id);


--
-- TOC entry 4029 (class 1259 OID 4442523)
-- Name: fki_image_validation_source_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_image_validation_source_image_validation_id_fkey ON public.image_validation_source USING btree (image_validation_id);


--
-- TOC entry 4032 (class 1259 OID 4442524)
-- Name: fki_imagehunt_task_image_validation_id_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_imagehunt_task_image_validation_id_image_validation_id_fkey ON public.imagehunt_task USING btree (image_validation_id);


--
-- TOC entry 4043 (class 1259 OID 4442525)
-- Name: fki_label_accessor_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_accessor_label_id_fkey ON public.label_accessor USING btree (label_id);


--
-- TOC entry 4048 (class 1259 OID 4442526)
-- Name: fki_label_example_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_example_label_id_fkey ON public.label_example USING btree (label_id);


--
-- TOC entry 4035 (class 1259 OID 4442527)
-- Name: fki_label_parent_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_parent_id_fkey ON public.label USING btree (parent_id);


--
-- TOC entry 4049 (class 1259 OID 4442528)
-- Name: fki_label_suggestion_proposed_by_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_label_suggestion_proposed_by_fkey ON public.label_suggestion USING btree (proposed_by);


--
-- TOC entry 4058 (class 1259 OID 4442529)
-- Name: fki_quiz_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_label_id_fkey ON public.quiz_answer USING btree (label_id);


--
-- TOC entry 4064 (class 1259 OID 4442530)
-- Name: fki_quiz_question_refines_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_question_refines_label_id_fkey ON public.quiz_question USING btree (refines_label_id);


--
-- TOC entry 4059 (class 1259 OID 4442531)
-- Name: fki_quiz_quiz_question_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_quiz_quiz_question_id_fkey ON public.quiz_answer USING btree (quiz_question_id);


--
-- TOC entry 4069 (class 1259 OID 4442532)
-- Name: fki_trending_label_suggestion_id_trending_label_suggestion_id_f; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_trending_label_suggestion_id_trending_label_suggestion_id_f ON public.trending_label_bot_task USING btree (trending_label_suggestion_id);


--
-- TOC entry 4076 (class 1259 OID 4442533)
-- Name: fki_trending_label_suggestion_label_suggestion_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_trending_label_suggestion_label_suggestion_id_fkey ON public.trending_label_suggestion USING btree (label_suggestion_id);


--
-- TOC entry 4077 (class 1259 OID 4442534)
-- Name: fki_trending_label_suggestion_productive_label_id_label_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_trending_label_suggestion_productive_label_id_label_id_fkey ON public.trending_label_suggestion USING btree (productive_label_id);


--
-- TOC entry 4082 (class 1259 OID 4442535)
-- Name: fki_user_annotation_blacklist_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_account_id_fkey ON public.user_annotation_blacklist USING btree (account_id);


--
-- TOC entry 4083 (class 1259 OID 4442536)
-- Name: fki_user_annotation_blacklist_image_validation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_annotation_blacklist_image_validation_id_fkey ON public.user_annotation_blacklist USING btree (image_validation_id);


--
-- TOC entry 4088 (class 1259 OID 4442537)
-- Name: fki_user_image_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_account_id_fkey ON public.user_image USING btree (account_id);


--
-- TOC entry 4094 (class 1259 OID 4442538)
-- Name: fki_user_image_annotation_image_annotation_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_image_annotation_id_fkey ON public.user_image_annotation USING btree (image_annotation_id);


--
-- TOC entry 4095 (class 1259 OID 4442539)
-- Name: fki_user_image_annotation_user_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_annotation_user_id_fkey ON public.user_image_annotation USING btree (account_id);


--
-- TOC entry 4098 (class 1259 OID 4442540)
-- Name: fki_user_image_collection_account_id_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_collection_account_id_account_id_fkey ON public.user_image_collection USING btree (account_id);


--
-- TOC entry 4089 (class 1259 OID 4442541)
-- Name: fki_user_image_image_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_image_id_fkey ON public.user_image USING btree (image_id);


--
-- TOC entry 4105 (class 1259 OID 4442542)
-- Name: fki_user_image_validation_acccount_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_acccount_id_fkey ON public.user_image_validation USING btree (account_id);


--
-- TOC entry 4106 (class 1259 OID 4442543)
-- Name: fki_user_image_validation_account_id_fkey; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX fki_user_image_validation_account_id_fkey ON public.user_image_validation USING btree (image_validation_id);


--
-- TOC entry 3964 (class 1259 OID 4442544)
-- Name: image_annotation_coverage_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_coverage_image_id_index ON public.image_annotation_coverage USING btree (image_id);


--
-- TOC entry 3958 (class 1259 OID 4442545)
-- Name: image_annotation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_image_id_index ON public.image_annotation USING btree (image_id);


--
-- TOC entry 3961 (class 1259 OID 4442546)
-- Name: image_annotation_sys_period_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_sys_period_index ON public.image_annotation USING btree (sys_period);


--
-- TOC entry 3962 (class 1259 OID 4442547)
-- Name: image_annotation_uuid_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_annotation_uuid_index ON public.image_annotation USING btree (uuid);


--
-- TOC entry 3982 (class 1259 OID 4442548)
-- Name: image_collection_image_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_collection_image_image_id_index ON public.image_collection_image USING btree (image_id);


--
-- TOC entry 3985 (class 1259 OID 4442549)
-- Name: image_collection_image_user_image_collection_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_collection_image_user_image_collection_id_index ON public.image_collection_image USING btree (user_image_collection_id);


--
-- TOC entry 3989 (class 1259 OID 4442550)
-- Name: image_description_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_description_image_id_index ON public.image_description USING btree (image_id);


--
-- TOC entry 3990 (class 1259 OID 4442551)
-- Name: image_description_state_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_description_state_index ON public.image_description USING btree (state);


--
-- TOC entry 3991 (class 1259 OID 4442552)
-- Name: image_description_sys_period_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_description_sys_period_index ON public.image_description USING btree (sys_period);


--
-- TOC entry 3945 (class 1259 OID 4442553)
-- Name: image_hash_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_hash_index ON public.image USING btree (hash);


--
-- TOC entry 3950 (class 1259 OID 4442554)
-- Name: image_image_provider_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_image_provider_index ON public.image USING btree (image_provider_id);


--
-- TOC entry 3951 (class 1259 OID 4442555)
-- Name: image_key_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_key_index ON public.image USING btree (key);


--
-- TOC entry 3954 (class 1259 OID 4442556)
-- Name: image_unlocked_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_unlocked_index ON public.image USING btree (unlocked);


--
-- TOC entry 4023 (class 1259 OID 4442557)
-- Name: image_validation_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_image_id_index ON public.image_validation USING btree (image_id);


--
-- TOC entry 4026 (class 1259 OID 4442558)
-- Name: image_validation_label_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_label_id_index ON public.image_validation USING btree (label_id);


--
-- TOC entry 4027 (class 1259 OID 4442559)
-- Name: image_validation_sys_period_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX image_validation_sys_period_index ON public.image_validation USING btree (sys_period);


--
-- TOC entry 4038 (class 1259 OID 4442560)
-- Name: label_label_type_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX label_label_type_index ON public.label USING btree (label_type);


--
-- TOC entry 4090 (class 1259 OID 4442561)
-- Name: user_image_account_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX user_image_account_id_index ON public.user_image USING btree (account_id);


--
-- TOC entry 4099 (class 1259 OID 4442562)
-- Name: user_image_collection_account_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX user_image_collection_account_id_index ON public.user_image_collection USING btree (account_id);


--
-- TOC entry 4104 (class 1259 OID 4442563)
-- Name: user_image_collection_name_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX user_image_collection_name_index ON public.user_image_collection USING btree (name);


--
-- TOC entry 4093 (class 1259 OID 4442564)
-- Name: user_image_image_id_index; Type: INDEX; Schema: public; Owner: monkey
--

CREATE INDEX user_image_image_id_index ON public.user_image USING btree (image_id);


--
-- TOC entry 4165 (class 2620 OID 4442565)
-- Name: image_annotation_refinement image_annotation_refinement_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_annotation_refinement_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON public.image_annotation_refinement FOR EACH ROW EXECUTE PROCEDURE public.versioning('sys_period', 'image_annotation_refinement_history', 'true');


--
-- TOC entry 4164 (class 2620 OID 4442566)
-- Name: image_annotation image_annotation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_annotation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON public.image_annotation FOR EACH ROW EXECUTE PROCEDURE public.versioning('sys_period', 'image_annotation_history', 'true');


--
-- TOC entry 4166 (class 2620 OID 4442567)
-- Name: image_description image_description_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_description_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON public.image_description FOR EACH ROW EXECUTE PROCEDURE public.versioning('sys_period', 'image_description_history', 'true');


--
-- TOC entry 4167 (class 2620 OID 4442568)
-- Name: image_validation image_validation_versioning_trigger; Type: TRIGGER; Schema: public; Owner: monkey
--

CREATE TRIGGER image_validation_versioning_trigger BEFORE INSERT OR DELETE OR UPDATE ON public.image_validation FOR EACH ROW EXECUTE PROCEDURE public.versioning('sys_period', 'image_validation_history', 'true');


--
-- TOC entry 4117 (class 2606 OID 4442569)
-- Name: access_token access_token_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.access_token
    ADD CONSTRAINT access_token_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.account(id);


--
-- TOC entry 4118 (class 2606 OID 4442574)
-- Name: account_permission account_permission_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.account_permission
    ADD CONSTRAINT account_permission_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4119 (class 2606 OID 4442579)
-- Name: annotation_data annotation_data_annotation_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_data
    ADD CONSTRAINT annotation_data_annotation_type_fkey FOREIGN KEY (annotation_type_id) REFERENCES public.annotation_type(id);


--
-- TOC entry 4120 (class 2606 OID 4442584)
-- Name: annotation_data annotation_data_image_annotation_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_data
    ADD CONSTRAINT annotation_data_image_annotation_revision_id_fkey FOREIGN KEY (image_annotation_revision_id) REFERENCES public.image_annotation_revision(id);


--
-- TOC entry 4122 (class 2606 OID 4442589)
-- Name: api_token api_token_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.api_token
    ADD CONSTRAINT api_token_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4126 (class 2606 OID 4442594)
-- Name: image_annotation_coverage image_annotation_coverage_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_coverage
    ADD CONSTRAINT image_annotation_coverage_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4121 (class 2606 OID 4442599)
-- Name: annotation_data image_annotation_data_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.annotation_data
    ADD CONSTRAINT image_annotation_data_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES public.image_annotation(id);


--
-- TOC entry 4124 (class 2606 OID 4442604)
-- Name: image_annotation image_annotation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation
    ADD CONSTRAINT image_annotation_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4125 (class 2606 OID 4442609)
-- Name: image_annotation image_annotation_label_id_key; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation
    ADD CONSTRAINT image_annotation_label_id_key FOREIGN KEY (label_id) REFERENCES public.label(id);


--
-- TOC entry 4127 (class 2606 OID 4442614)
-- Name: image_annotation_refinement image_annotation_refinement_annotation_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_annotation_data_id_fkey FOREIGN KEY (annotation_data_id) REFERENCES public.annotation_data(id);


--
-- TOC entry 4128 (class 2606 OID 4442619)
-- Name: image_annotation_refinement image_annotation_refinement_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_refinement
    ADD CONSTRAINT image_annotation_refinement_label_id_fkey FOREIGN KEY (label_id) REFERENCES public.label(id);


--
-- TOC entry 4129 (class 2606 OID 4442624)
-- Name: image_annotation_revision image_annotation_revision_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_annotation_revision
    ADD CONSTRAINT image_annotation_revision_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES public.image_annotation(id);


--
-- TOC entry 4130 (class 2606 OID 4442629)
-- Name: image_collection_image image_collection_image_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_collection_image
    ADD CONSTRAINT image_collection_image_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4131 (class 2606 OID 4442634)
-- Name: image_collection_image image_collection_image_user_image_collection_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_collection_image
    ADD CONSTRAINT image_collection_image_user_image_collection_id_fkey FOREIGN KEY (user_image_collection_id) REFERENCES public.user_image_collection(id);


--
-- TOC entry 4132 (class 2606 OID 4442639)
-- Name: image_description image_description_image_id_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_description
    ADD CONSTRAINT image_description_image_id_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4133 (class 2606 OID 4442644)
-- Name: image_description image_description_language_language_id; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_description
    ADD CONSTRAINT image_description_language_language_id FOREIGN KEY (language_id) REFERENCES public.language(id);


--
-- TOC entry 4134 (class 2606 OID 4442649)
-- Name: image_description image_description_processed_by_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_description
    ADD CONSTRAINT image_description_processed_by_account_id_fkey FOREIGN KEY (processed_by) REFERENCES public.account(id);


--
-- TOC entry 4123 (class 2606 OID 4442654)
-- Name: image image_image_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image
    ADD CONSTRAINT image_image_provider_id_fkey FOREIGN KEY (image_provider_id) REFERENCES public.image_provider(id);


--
-- TOC entry 4135 (class 2606 OID 4442659)
-- Name: image_label_suggestion image_label_suggestion_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4136 (class 2606 OID 4442664)
-- Name: image_label_suggestion image_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_label_suggestion
    ADD CONSTRAINT image_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES public.label_suggestion(id);


--
-- TOC entry 4137 (class 2606 OID 4442669)
-- Name: image_quarantine image_quarantine_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_quarantine
    ADD CONSTRAINT image_quarantine_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4138 (class 2606 OID 4442674)
-- Name: image_report image_report_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_report
    ADD CONSTRAINT image_report_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4139 (class 2606 OID 4442679)
-- Name: image_source image_source_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_source
    ADD CONSTRAINT image_source_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4140 (class 2606 OID 4442684)
-- Name: image_validation image_validation_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation
    ADD CONSTRAINT image_validation_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4141 (class 2606 OID 4442689)
-- Name: image_validation image_validation_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation
    ADD CONSTRAINT image_validation_label_id_fkey FOREIGN KEY (label_id) REFERENCES public.label(id);


--
-- TOC entry 4142 (class 2606 OID 4442694)
-- Name: image_validation_source image_validation_source_image_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation_source
    ADD CONSTRAINT image_validation_source_image_source_id_fkey FOREIGN KEY (image_source_id) REFERENCES public.image_source(id);


--
-- TOC entry 4143 (class 2606 OID 4442699)
-- Name: image_validation_source image_validation_source_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.image_validation_source
    ADD CONSTRAINT image_validation_source_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES public.image_validation(id);


--
-- TOC entry 4144 (class 2606 OID 4442704)
-- Name: imagehunt_task imagehunt_task_image_validation_id_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.imagehunt_task
    ADD CONSTRAINT imagehunt_task_image_validation_id_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES public.image_validation(id);


--
-- TOC entry 4146 (class 2606 OID 4442709)
-- Name: label_accessor label_accessor_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_accessor
    ADD CONSTRAINT label_accessor_label_id_fkey FOREIGN KEY (label_id) REFERENCES public.label(id);


--
-- TOC entry 4147 (class 2606 OID 4442714)
-- Name: label_example label_example_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_example
    ADD CONSTRAINT label_example_label_id_fkey FOREIGN KEY (label_id) REFERENCES public.label(id);


--
-- TOC entry 4145 (class 2606 OID 4442719)
-- Name: label label_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label
    ADD CONSTRAINT label_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.label(id);


--
-- TOC entry 4148 (class 2606 OID 4442724)
-- Name: label_suggestion label_suggestion_proposed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.label_suggestion
    ADD CONSTRAINT label_suggestion_proposed_by_fkey FOREIGN KEY (proposed_by) REFERENCES public.account(id);


--
-- TOC entry 4149 (class 2606 OID 4442729)
-- Name: quiz_answer quiz_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_answer
    ADD CONSTRAINT quiz_label_id_fkey FOREIGN KEY (label_id) REFERENCES public.label(id);


--
-- TOC entry 4151 (class 2606 OID 4442734)
-- Name: quiz_question quiz_question_refines_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_question
    ADD CONSTRAINT quiz_question_refines_label_id_fkey FOREIGN KEY (refines_label_id) REFERENCES public.label(id);


--
-- TOC entry 4150 (class 2606 OID 4442739)
-- Name: quiz_answer quiz_quiz_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.quiz_answer
    ADD CONSTRAINT quiz_quiz_question_id_fkey FOREIGN KEY (quiz_question_id) REFERENCES public.quiz_question(id);


--
-- TOC entry 4152 (class 2606 OID 4442744)
-- Name: trending_label_bot_task trending_label_suggestion_id_trending_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_bot_task
    ADD CONSTRAINT trending_label_suggestion_id_trending_label_suggestion_id_fkey FOREIGN KEY (trending_label_suggestion_id) REFERENCES public.trending_label_suggestion(id);


--
-- TOC entry 4153 (class 2606 OID 4442749)
-- Name: trending_label_suggestion trending_label_suggestion_label_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_label_suggestion_id_fkey FOREIGN KEY (label_suggestion_id) REFERENCES public.label_suggestion(id);


--
-- TOC entry 4154 (class 2606 OID 4442754)
-- Name: trending_label_suggestion trending_label_suggestion_productive_label_id_label_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.trending_label_suggestion
    ADD CONSTRAINT trending_label_suggestion_productive_label_id_label_id_fkey FOREIGN KEY (productive_label_id) REFERENCES public.label(id);


--
-- TOC entry 4155 (class 2606 OID 4442759)
-- Name: user_annotation_blacklist user_annotation_blacklist_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4156 (class 2606 OID 4442764)
-- Name: user_annotation_blacklist user_annotation_blacklist_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_annotation_blacklist
    ADD CONSTRAINT user_annotation_blacklist_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES public.image_validation(id);


--
-- TOC entry 4157 (class 2606 OID 4442769)
-- Name: user_image user_image_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image
    ADD CONSTRAINT user_image_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4159 (class 2606 OID 4442774)
-- Name: user_image_annotation user_image_annotation_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_annotation
    ADD CONSTRAINT user_image_annotation_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4160 (class 2606 OID 4442779)
-- Name: user_image_annotation user_image_annotation_image_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_annotation
    ADD CONSTRAINT user_image_annotation_image_annotation_id_fkey FOREIGN KEY (image_annotation_id) REFERENCES public.image_annotation(id);


--
-- TOC entry 4161 (class 2606 OID 4442784)
-- Name: user_image_collection user_image_collection_account_id_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_collection
    ADD CONSTRAINT user_image_collection_account_id_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4158 (class 2606 OID 4442789)
-- Name: user_image user_image_image_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image
    ADD CONSTRAINT user_image_image_id_fkey FOREIGN KEY (image_id) REFERENCES public.image(id);


--
-- TOC entry 4162 (class 2606 OID 4442794)
-- Name: user_image_validation user_image_validation_acccount_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_validation
    ADD CONSTRAINT user_image_validation_acccount_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(id);


--
-- TOC entry 4163 (class 2606 OID 4442799)
-- Name: user_image_validation user_image_validation_image_validation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: monkey
--

ALTER TABLE ONLY public.user_image_validation
    ADD CONSTRAINT user_image_validation_image_validation_id_fkey FOREIGN KEY (image_validation_id) REFERENCES public.image_validation(id);


--
-- TOC entry 4299 (class 0 OID 0)
-- Dependencies: 6
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO monkey;


-- Completed on 2019-07-27 16:47:00 CEST

--
-- PostgreSQL database dump complete
--

