#!/bin/bash

/tmp/wait-for-it.sh $DB_HOST:$DB_PORT -- echo "Starting integration tests"

./test -test.v -test.parallel 1 -donations_dir=/tmp/data/donations/ -unverified_donations_dir=/tmp/data/unverified_donations/ -test.timeout=600m
retVal=$?
if [ $retVal -ne 0 ]; then
	echo "Aborting due to error"
	exit $retVal
fi
echo "All tests successful"

exit 0

