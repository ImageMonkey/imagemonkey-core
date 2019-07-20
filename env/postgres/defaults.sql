INSERT INTO annotation_type(id, name) VALUES(1, 'rect');
INSERT INTO annotation_type (id, name) VALUES(2, 'ellipse');
INSERT INTO annotation_type (id, name) VALUES(3, 'polygon');

CREATE TYPE control_type AS ENUM ('dropdown', 'checkbox', 'radio', 'color tags');

insert into image_provider(name) values('donation');
insert into image_provider(name) values('labelme');

CREATE TYPE label_type AS ENUM ('normal', 'refinement', 'refinement_category', 'meta');

CREATE TYPE state_type AS ENUM ('unknown', 'locked', 'unlocked');



insert into language(name, fullname) values('en', 'English');
insert into language(name, fullname) values('ger', 'German');


CREATE TYPE label_bot_state_type AS ENUM ('pending', 'building', 'build-failed', 'build-success', 'waiting for moderator approval', 'accepted', 'retry', 'build-canceled');
