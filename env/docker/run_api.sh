#!/bin/bash

/usr/bin/wait-for-it.sh $DB_HOST:$DB_PORT -- echo "Database ($DB_HOST:$DB_PORT) is up"
while true
	do
		db_initialized=$(psql -h $DB_HOST -p $DB_PORT -U postgres -lqt | cut -d \| -f 1 | grep "imagemonkey" | xargs)
		if [[ $db_initialized = "imagemonkey" ]]
		then
			echo "Database is initialized"
			break;
		else
			echo "Waiting for the database to be initialized ($db_initialized)"
		fi
	done

echo "Starting api (after 5 sec delay)"
sleep 5

merge_labels=false
if [ "$1" ]; then
	if [ "$1" == "--merge-labels-before-start" ]; then
		merge_labels=true
	fi
fi

if [ "$merge_labels" = true ] ; then
./labels_merger
fi

./api -use_sentry=$USE_SENTRY -redis_address=$REDIS_ADDRESS -donations_dir=/home/imagemonkey/data/donations/ -unverified_donations_dir=/home/imagemonkey/data/unverified_donations/ -image_quarantine_dir=/home/imagemonkey/data/quarantine/
