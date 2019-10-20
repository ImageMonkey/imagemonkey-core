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

echo "Starting web (after 5 sec delay)"
sleep 5


start_after_api=false
if [ "$1" ]; then
	if [ "$1" == "--start-after-api" ]; then
		start_after_api=true
	fi
fi

if [ "$start_after_api" = true ] ; then
	/usr/bin/wait-for-it.sh $API_HOST:$API_PORT -- echo "API ($API_HOST:$API_PORT) is up"		
fi

./web -use_sentry=$USE_SENTRY -redis_address=$REDIS_ADDRESS -donations_dir=/home/imagemonkey/data/donations/ 
