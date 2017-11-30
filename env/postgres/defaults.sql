INSERT INTO annotation_type(id, name) VALUES(1, 'rect');
INSERT INTO annotation_type (id, name) VALUES(2, 'ellipse');
INSERT INTO annotation_type (id, name) VALUES(3, 'polygon');

CREATE TYPE control_type AS ENUM ('dropdown', 'checkbox', 'radio');