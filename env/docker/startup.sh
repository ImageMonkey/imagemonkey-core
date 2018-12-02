#!/bin/bash
sleep infinity & PID=$!
trap "kill $PID" INT TERM
trap "kill 0" EXIT


run_tests=false
if [ "$1" ]; then
	if [ "$1" == "--run-tests" ]; then
		run_tests=true
	fi
fi

if [ "$run_tests" = true ] ; then
	echo -e "host\t all\t all\t 127.0.0.1/32\t trust" > /etc/postgresql/9.6/main/pg_hba.conf
fi


echo "Starting PostgreSQL..."
#start postgres
/root/imagemonkey-core/env/docker/start_postgres.sh 

echo "Starting redis-server..."
#start redis server
service redis-server start


#replace api_base_url with API_BASE_URL from env variable (use @ as delimiter)
sed -i.bak 's@-api_base_url=xxxxxx@-api_base_url='"$API_BASE_URL"'@g' /etc/supervisor/conf.d/imagemonkey-web.conf


#replace api_base_url with API_BASE_URL from env variable (use @ as delimiter)
sed -i.bak 's@-api_base_url=xxxxxx@-api_base_url='"${API_BASE_URL}/"'@g' /etc/supervisor/conf.d/imagemonkey-api.conf

echo "Starting netdata..."
/usr/sbin/netdata

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

if [ "$run_tests" = true ] ; then
	echo "Running Tests"
	go get -u gopkg.in/resty.v1
	cd /root/imagemonkey-core/tests/
	go test
else
	echo "You can now connect to the webserver via <machine ip>:8080 and to the REST API via <machine ip>:8081."
	echo "This docker image is for development only - do NOT use it in production!"

	wait

	#shutting down
	echo "Exited"
fi


