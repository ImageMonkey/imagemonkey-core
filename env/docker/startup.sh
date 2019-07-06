#!/bin/bash
sleep infinity & PID=$!
trap "kill $PID" INT TERM
trap "kill 0" EXIT


run_tests=false
run_stresstest=false
if [ "$1" ]; then
	if [ "$1" == "--run-tests" ]; then
		run_tests=true
	fi

	if [ "$1" == "--run-stresstest" ]; then
		run_stresstest=true
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
	pip3 install requests
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

	go test -v -p 1 -timeout=100m -donations_dir="/home/imagemonkey/donations/" -unverified_donations_dir="/home/imagemonkey/unverified_donations/"
	retVal=$?
	if [ $retVal -ne 0 ]; then
    	echo "Aborting due to error"
    	exit $retVal
	fi

	echo "Running UI Tests"
	supervisorctl start imagemonkey-web:imagemonkey-web0
	cd /root/imagemonkey-core/tests/ui/
	python3 -m unittest
	retVal=$?
	if [ $retVal -ne 0 ]; then
    	echo "Aborting due to error"
    	exit $retVal
	fi
fi

if [ "$run_stresstest" = true ] ; then
	if [ ! -f /tmp/stresstest/imagemonkey_data.zip ]; then
		echo "Couldn't run stresstest: /tmp/stresstest/imagemonkey_data.zip doesn't exist!"
		exit 1
	fi

	cd /tmp/stresstest/
	unzip -o /tmp/stresstest/imagemonkey_data.zip
	cp -r /tmp/stresstest/donations /home/imagemonkey/

	su - postgres -c "echo \"select pg_terminate_backend(pid) from pg_stat_activity where datname='imagemonkey';drop database imagemonkey;\" | psql" > /dev/null
	su - postgres -c "echo \"create database imagemonkey OWNER monkey;\" | psql"
	su - postgres -c "psql -v ON_ERROR_STOP=1 --single-transaction -d imagemonkey -f /tmp/stresstest/imagemonkey.sql"

	cd /tmp/stresstest
	wget https://github.com/tsenart/vegeta/releases/download/cli%2Fv12.1.0/vegeta-12.1.0-linux-amd64.tar.gz .
	tar xvf vegeta-12.1.0-linux-amd64.tar.gz
	rm -f vegeta-12.1.0-linux-amd64.tar.gz
	chmod u+rx vegeta

	cat /root/imagemonkey-core/tests/stresstest/requests.txt | ./vegeta attack -duration=5s | ./vegeta report
fi


if [ "$run_stresstest" = false ] && [ "$run_tests" = false ] ; then
	echo "You can now connect to the webserver via <machine ip>:8080 and to the REST API via <machine ip>:8081."
	echo "This docker image is for development only - do NOT use it in production!"

	wait

	#shutting down
	echo "Exited"
fi


