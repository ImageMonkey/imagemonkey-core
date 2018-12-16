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
	echo -e "local\t all\t postgres\t ident" >> /etc/postgresql/9.6/main/pg_hba.conf
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
	echo "Running test suite"

	echo "Installing additional requirements"
	go get -u gopkg.in/resty.v1
	pip3 install selenium
	wget https://chromedriver.storage.googleapis.com/2.44/chromedriver_linux64.zip --directory-prefix=/tmp/
	cd /tmp \
		&& unzip /tmp/chromedriver_linux64.zip \
		&& cp /tmp/chromedriver /root/imagemonkey-core/tests/ui/ \
		&& rm /tmp/chromedriver \
		&& rm /tmp/chromedriver_linux64.zip \
		&& wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb --directory-prefix=/tmp/ \
		&& dpkg -i google-chrome-stable_current_amd64.deb \
		&& rm /tmp/google-chrome-stable_current_amd64.deb
	apt-get install -y -f 

	mkdir -p /root/imagemonkey-core/unverified_donations
	mkdir -p /root/imagemonkey-core/donations
	
	echo "Running unittests"
	cd /root/imagemonkey-core/src/parser/
	supervisorctl stop all
	go test
	retVal=$?
	if [ $retVal -ne 0 ]; then
    	echo "Aborting due to error"
    	exit $retVal
	fi

	echo "Running Integration Tests"
	cd /root/imagemonkey-core/tests/
	supervisorctl stop all
	supervisorctl start imagemonkey-api:imagemonkey-api0
	
	go test -v -timeout=100m -donations_dir="/home/imagemonkey/donations/" -unverified_donations_dir="/home/imagemonkey/unverified_donations/"
	retVal=$?
	if [ $retVal -ne 0 ]; then
    	echo "Aborting due to error"
    	exit $retVal
	fi

	#echo "Running UI Tests"
	#supervisorctl start imagemonkey-web:imagemonkey-web0
	#cd /root/imagemonkey-core/tests/ui/
	#python3 -m unittest
	#retVal=$?
	#if [ $retVal -ne 0 ]; then
    #	echo "Aborting due to error"
    #	exit $retVal
	#fi

else
	echo "You can now connect to the webserver via <machine ip>:8080 and to the REST API via <machine ip>:8081."
	echo "This docker image is for development only - do NOT use it in production!"

	wait

	#shutting down
	echo "Exited"
fi


