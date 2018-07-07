CREATE TRIGGER image_validation_versioning_trigger
BEFORE INSERT OR UPDATE OR DELETE ON image_validation
FOR EACH ROW EXECUTE PROCEDURE versioning(
  'sys_period', 'image_validation_history', true
);


CREATE TRIGGER image_annotation_versioning_trigger
BEFORE INSERT OR UPDATE OR DELETE ON image_annotation
FOR EACH ROW EXECUTE PROCEDURE versioning(
  'sys_period', 'image_annotation_history', true
);