DROP VIEW IF EXISTS image_annotation_coverage;

CREATE OR REPLACE VIEW image_annotation_coverage AS

WITH all_annotations AS (
    SELECT an.image_id as image_id, d.id as annotation_data_id, d.annotation as annotation, t.name as annotation_type
    FROM image_annotation an 
    JOIN annotation_data d ON d.image_annotation_id = an.id
    JOIN annotation_type t ON t.id = d.annotation_type_id
    JOIN image i ON i.id = an.image_id
    WHERE i.unlocked = true AND an.auto_generated = false
), 
ellipse_annotations AS (
    SELECT a.image_id, a.annotation_data_id as id, 
    Ellipse( (a.annotation->'left')::text::float, 
             (a.annotation->'top')::text::float, 
             2* (a.annotation->'rx')::text::float, 
             2* (a.annotation->'ry')::text::float, 
             CASE 
                WHEN a.annotation->'angle' is null THEN 0 
                ELSE (a.annotation->'angle')::text::float
             END
           ) as geom
    FROM all_annotations a 
    WHERE annotation_type = 'ellipse'
),
polygon_annotations AS (
  -- ST_MakePolygon might return a polygon with intersecting points. In order to fix that, one needs to call ST_MakeValid on the resulting polygon.
  --Unfortunately, this is _really_ slow (especially, if a lot of polygons are affected). In order to circumvent that, we create a ConvexHull around the
  --polygon. This works way faster and should also be precise enough for our purpose.
	SELECT q.image_id, q.annotation_data_id as id, ST_ConvexHull(ST_MakePolygon(ST_GeomFromText('LINESTRING(' || 
                                                                  string_agg((((q.annotation->'x')::text) || ' ' || ((q.annotation->'y')::text)), ',') 
                                                                  || ',' || (array_agg((q.annotation->'x')::text))[1] || ' ' || (array_agg((q.annotation->'y')::text))[1] 
                                                                  || ')'))) as geom
    FROM
    (
        SELECT a.image_id, a.annotation_data_id, jsonb_array_elements(a.annotation->'points') as  annotation
        FROM all_annotations a 
        WHERE a.annotation_type = 'polygon' AND jsonb_array_length(a.annotation->'points') > 2
    ) q
    GROUP BY q.image_id, q.annotation_data_id
),
rectangle_annotations AS (
    SELECT a.image_id, a.annotation_data_id as id, ST_MakePolygon(ST_MakeLine(
       ARRAY[
             ST_MakePoint((a.annotation->'left')::text::integer, (a.annotation->'top')::text::integer), 
             ST_MakePoint((a.annotation->'left')::text::float + (a.annotation->'width')::text::float, (a.annotation->'top')::text::float),
             ST_MakePoint((a.annotation->'left')::text::float + (a.annotation->'width')::text::float, 
                                                    (a.annotation->'top')::text::float + (a.annotation->'height')::text::float),
             ST_MakePoint((a.annotation->'left')::text::float, (a.annotation->'top')::text::float + (a.annotation->'height')::text::float),
             ST_MakePoint((a.annotation->'left')::text::float, (a.annotation->'top')::text::float)
            ])) as geom
    FROM all_annotations a 
    WHERE a.annotation_type = 'rect'
    --GROUP BY a.annotation_data_id, a.annotation
),
all_annotation_areas AS (
    SELECT id, image_id, geom from polygon_annotations
    UNION 
    SELECT id, image_id, geom from rectangle_annotations
    UNION
    SELECT id, image_id, geom from ellipse_annotations
)
SELECT i.id as image_id, i.key as image_key, i.width as image_width, i.height as image_height, 
    (SUM(q.area)/(i.width * i.height)) * 100 as annotated_percentage
    FROM
    (                                                                                   
        SELECT a.id, a.image_id, ST_Area(ST_DifferenceAgg(a.geom, b.geom)) as area
        FROM all_annotation_areas a
        LEFT JOIN all_annotation_areas b
        ON (ST_Contains(a.geom, b.geom) OR -- Select all the containing, contained and overlapping polygons
            ST_Contains(b.geom, a.geom) OR
            ST_Overlaps(a.geom, b.geom)) AND
            (ST_Area(a.geom) < ST_Area(b.geom) OR -- Make sure bigger polygons are removed from smaller ones
            (ST_Area(a.geom) = ST_Area(b.geom) AND -- If areas are equal, arbitrarily remove one from the other but in a determined order so it's not done twice.
              a.id < b.id)) AND (a.image_id = b.image_id)
        GROUP BY a.id, a.image_id
        HAVING ST_Area(ST_DifferenceAgg(a.geom, b.geom)) > 0 AND NOT ST_IsEmpty(ST_DifferenceAgg(a.geom, b.geom))
    ) q
    JOIN image i ON q.image_id = i.id
    GROUP BY i.id;
ALTER VIEW image_annotation_coverage OWNER TO monkey; --change owner