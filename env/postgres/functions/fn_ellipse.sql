CREATE OR REPLACE FUNCTION Ellipse(
    x double precision,
    y double precision,
    rx double precision,
    ry double precision,
    rotation double precision)
  RETURNS geometry AS
$BODY$
   SELECT ST_Translate( ST_Rotate( ST_Scale( ST_Buffer(ST_MakePoint(0,0)::geometry, 0.5)::geometry, rx, ry), rotation), x, y)       
$BODY$
  LANGUAGE sql;