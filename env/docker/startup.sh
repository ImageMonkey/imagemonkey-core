#!/bin/bash
sleep infinity & PID=$!
trap "kill $PID" INT TERM
trap "kill 0" EXIT

echo "Starting PostgreSQL..."
#start postgres
/root/imagemonkey-core/env/docker/start_postgres.sh 

echo "Starting redis-server..."
#start redis server
service redis-server start

echo "Starting supervisord..."
#start supervisord
#service supervisor start && 
supervisorctl reread && supervisorctl update && supervisorctl restart all

echo ""
echo ""
echo ""
echo "#############################################################"
echo "################ ImageMonkey is ready #######################"
echo "#############################################################"

echo ""
echo ""
echo "You can now connect to the webserver via <machine ip>:8080 and to the REST API via <machine ip>:8081."
echo "This docker image is for development only - do NOT use it in production!"

wait

#shutting down
echo "Exited"