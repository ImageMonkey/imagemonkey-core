
--Image table
CREATE INDEX image_image_provider_index ON image (image_provider_id);
CREATE INDEX image_key_index ON image (key);
CREATE INDEX image_hash_index ON image (hash);
CREATE INDEX image_unlocked_index ON image (unlocked);

--Image Validation table
CREATE INDEX image_validation_image_id_index ON image_validation (image_id);
CREATE INDEX image_validation_label_id_index ON image_validation (label_id);
CREATE INDEX image_validation_uuid_index ON image_validation (uuid);
CREATE INDEX image_validation_sys_period_index ON image_validation(sys_period);

-- Label table
CREATE INDEX label_name_index ON label (name);
CREATE INDEX label_parent_id_index ON label(parent_id);
CREATE INDEX label_label_type_index ON label(label_type);

-- annotation_data table
CREATE INDEX annotation_data_image_annotation_id_idx ON annotation_data(image_annotation_id);


-- user_image table
CREATE INDEX user_image_image_id_index ON user_image (image_id);
CREATE INDEX user_image_account_id_index ON user_image(account_id);

-- image_quarantine table
CREATE INDEX image_quarantine_image_id_index ON image_quarantine(image_id);

-- image_annotation_coverage table
CREATE INDEX image_annotation_coverage_image_id_index ON image_annotation_coverage(image_id);

--image_description table

CREATE INDEX image_description_image_id_index ON image_description(image_id);
CREATE INDEX image_description_state_index ON image_description(state);
CREATE INDEX image_description_language_id_index ON image_description(language_id);
CREATE INDEX image_description_sys_period_index ON image_description(sys_period);

--image_annotation table
CREATE INDEX image_annotation_sys_period_index ON image_annotation(sys_period);


--user_image_collection table
CREATE INDEX user_image_collection_account_id_index ON user_image_collection(account_id);
CREATE INDEX user_image_collection_name_index ON user_image_collection(name);

--image_collection_image table
CREATE INDEX image_collection_image_image_id_index ON image_collection_image(image_id);
CREATE INDEX image_collection_image_user_image_collection_id_index ON image_collection_image(user_image_collection_id);





