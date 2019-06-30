#!/bin/bash

echo "Initializing ImageMonkey database"

echo "Installing temporal tables extension" \
    && echo "Creating user" \
	&& cat /tmp/imagemonkeydb/create_user.sql \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/create_user.sql \
    && psql -d imagemonkey -c "CREATE EXTENSION \"temporal_tables\";" \
    && echo "Installing Postgis extension" \
    && psql -d imagemonkey -c "CREATE EXTENSION \"postgis\";" \
    && echo "Applying schema" \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/schema.sql \
    && echo "Applying database defaults" \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/defaults.sql \
    && echo "Applying indexes" \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/indexes.sql \
    && echo "Applying functions" \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/postgres_functions/fn_ellipse.sql \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/postgres_functions/third_party/postgis_addons/postgis_addons.sql \
    && echo "Applying stored procedures" \
    && psql -d imagemonkey -f /tmp/imagemonkeydb/postgres_stored_procs/sp_get_image_annotation_coverage.sql

