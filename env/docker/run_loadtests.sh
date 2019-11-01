#!/bin/bash

/tmp/wait-for-it.sh $DB_HOST:$DB_PORT -- echo "Waiting for database...database ($DB_HOST:$DB_PORT) is up"
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

/tmp/wait-for-it.sh 127.0.0.1:8081 -- echo "Waiting for ImageMonkey API service..ImageMonkey API service is up"

echo "Starting stresstest(after 5 sec delay)"
sleep 5

echo "Terminating existing imagemonkey database connections"
psql -U postgres -h $DB_HOST -p $DB_PORT -v "ON_ERROR_STOP=1" -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'imagemonkey';"
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Couldn't terminate existing imagemonkey database sessions...aborting"
	exit 1
fi

echo "Dropping existing imagemonkey database"
psql -U postgres -h $DB_HOST -p $DB_PORT -v "ON_ERROR_STOP=1" -c "DROP DATABASE IF EXISTS imagemonkey;"
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Couldn't drop imagemonkey database...aborting"
	exit 1
fi

#echo "Creating new imagemonkey database with owner 'monkey'"
#psql -U postgres -h $DB_HOST -p $DB_PORT -v "ON_ERROR_STOP=1" -c "CREATE DATABASE imagemonkey OWNER monkey;" 
#retVal=$?
#if [ $retVal -ne 0 ]; then
#    echo "Couldn't create imagemonkey database...aborting"
#	exit 1
#fi

echo "Applying imagemonkey database dump"
#psql -U postgres -h $DB_HOST -p $DB_PORT -d imagemonkey < /tmp/imagemonkey_dump.sql
psql -U postgres -h $DB_HOST -p $DB_PORT < /tmp/imagemonkey_dump.sql
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Couldn't apply imagemonkey database dump...aborting"
	exit 1
fi

echo "Vegeta attack!"
cat /tmp/tests/loadtests.txt | vegeta attack -duration=5s | vegeta report

exit 0
