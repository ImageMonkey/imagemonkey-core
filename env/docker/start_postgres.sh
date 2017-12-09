#!/bin/bash

service postgresql start

while ! pg_isready 
do
    echo "$(date) - waiting for database to start"
    sleep 5
done

