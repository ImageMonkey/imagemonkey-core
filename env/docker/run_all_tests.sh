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

echo "Starting integration tests (after 5 sec delay)"
sleep 5


MODE="all"
if [[ -z "${TEST_MODE}" ]]; then
	MODE="all"
else
	MODE="${TEST_MODE}"
fi

if [ ${MODE} == "all" ]  || [ ${MODE} == "only-api" ]; then
	echo "Starting API tests"
	./test -test.v -test.parallel 1 -donations_dir=/tmp/data/donations/ -unverified_donations_dir=/tmp/data/unverified_donations/ -test.timeout=600m
	retVal=$?
	if [ $retVal -ne 0 ]; then
		echo "Aborting due to error"
		exit $retVal
	fi
fi

if [ ${MODE} == "all" ]  || [ ${MODE} == "only-unittest" ]; then
	echo "Starting unittests"
	echo "Parser tests"
	./parser_test -test.v
	retVal=$?
	if [ $retVal -ne 0 ]; then
		echo "Aborting due to error"
		exit $retVal
	fi


	echo "Parser v2 tests"
	./parserv2_test -test.v
	retVal=$?
	if [ $retVal -ne 0 ]; then
		echo "Aborting due to error"
		exit $retVal
	fi
fi

if [ ${MODE} == "all" ] || [ ${MODE} == "only-ui" ]; then
	#echo "Starting UI tests"
	#echo "Run selenium tests"
	#cd ui/
	#python3 -m unittest
	#retVal=$?
	#if [ $retVal -ne 0 ]; then
	#	echo "Aborting due to error"
	#	exit $retVal
	#fi

	echo "Run cypress tests"
	cd /tmp/tests/
	DEBUG=cypress:server:util:process_profiler /tmp/node_modules/.bin/cypress run --project ui/ --browser chromium
	retVal=$?
	if [ $retVal -ne 0 ]; then
		echo "Aborting due to error"
		echo "Copy cypress screenshots to /tmp/test_output"
		cp -r /tmp/tests/ui/cypress/screenshots /tmp/test_output/
		echo "Copy cypress videos to /tmp/test_output"
		cp -r /tmp/tests/ui/cypress/videos /tmp/test_output/
		exit $retVal
	fi
fi

echo "All tests successful"

exit 0

