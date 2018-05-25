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


#copy original imagemonkey-web.conf file and replace it with API_BASE_URL from env variable
cp /tmp/imagemonkey-web.conf.bak /etc/supervisor/conf.d/imagemonkey-web.conf
sed -i.bak 's/-api_base_url=xxxxxx/-api_base_url='"$API_BASE_URL"'/g' /etc/supervisor/conf.d/imagemonkey-web.conf


echo "Starting supervisord..."
#start supervisord
service supervisor start && supervisorctl reread && supervisorctl update && supervisorctl restart all

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