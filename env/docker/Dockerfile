# A basic apache server. To use either add or bind mount content under /var/www
FROM debian:9

MAINTAINER bbernhard version: 0.1

RUN apt-get update && apt-get install -y postgresql-9.6 nginx nginx-extras redis-server git supervisor wget build-essential postgresql-server-dev-9.6 postgresql-contrib-9.6 dos2unix neovim
RUN mkdir -p /root/imagemonkey-core
RUN git clone https://github.com/bbernhard/imagemonkey-core.git /root/imagemonkey-core
RUN cd /root/imagemonkey-core && git checkout develop 
RUN cp /root/imagemonkey-core/conf/supervisor/* /etc/supervisor/conf.d/

RUN chmod +x /root/imagemonkey-core/env/docker/start_postgres.sh
RUN chmod +x /root/imagemonkey-core/env/docker/startup.sh

RUN cd /tmp && wget https://github.com/arkhipov/temporal_tables/archive/v1.2.0.tar.gz
RUN mkdir -p /tmp/temporal_table
RUN cd /tmp/temporal_table && tar xvf /tmp/v1.2.0.tar.gz
RUN cd /tmp/temporal_table/temporal_tables-1.2.0/ && make && make install && cd /root/

RUN cp /root/imagemonkey-core/env/postgres/schema.sql /tmp/schema.sql
RUN chown postgres:postgres /tmp/schema.sql
RUN chmod u+rx /tmp/schema.sql

RUN echo CREATE EXTENSION uuid-ossp; >> /tmp/create_extension
RUN chown -R postgres:postgres /tmp/create_extension
RUN chmod -R u+rx /tmp/create_extension

RUN cp /root/imagemonkey-core/env/postgres/defaults.sql /tmp/defaults.sql
RUN chown postgres:postgres /tmp/defaults.sql
RUN chmod u+rx /tmp/defaults.sql 

RUN cp /root/imagemonkey-core/env/postgres/indexes.sql /tmp/indexes.sql
RUN chown postgres:postgres /tmp/indexes.sql
RUN chmod u+rx /tmp/indexes.sql 

RUN /root/imagemonkey-core/env/docker/start_postgres.sh && /bin/su - postgres -c "psql -c \"CREATE database imagemonkey;\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -c \"CREATE USER monkey WITH PASSWORD 'dbRuwMUo4Nfhs5hmMxhk';\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -c \"CREATE EXTENSION \"temporal_tables\";\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/schema.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/create_extension" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/defaults.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/indexes.sql"

RUN rm -rf /tmp/create_extension
RUN rm -rf /tmp/schema.sql
RUN rm -rf /tmp/temporal_table
RUN rm -rf /tmp/defaults.sql
RUN rm -rf /tmp/indexes.sql

RUN adduser imagemonkey --disabled-password --gecos "First Last,RoomNumber,WorkPhone,HomePhone" --home /home/imagemonkey
#RUN adduser imagemonkey supervisor

RUN /bin/su - imagemonkey -c "mkdir -p /home/imagemonkey/go"
ENV GOPATH="/home/imagemonkey/go"
RUN /bin/su - imagemonkey -c "mkdir -p /home/imagemonkey/bin"
ENV GOBIN="/home/imagemonkey/bin"


RUN cd /tmp/ \
   && wget https://redirector.gvt1.com/edgedl/go/go1.9.2.linux-amd64.tar.gz \
   && tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz \
   && cd /root/

RUN cp /root/imagemonkey-core/env/docker/conf/supervisor/* /etc/supervisor/conf.d/
RUN cp /root/imagemonkey-core/env/docker/src/* /root/imagemonkey-core/src/

RUN cp -r /root/imagemonkey-core/src /tmp/imagemonkey-core-src
RUN mv /tmp/imagemonkey-core-src/api_secrets.template /tmp/imagemonkey-core-src/api_secrets.go
RUN mv /tmp/imagemonkey-core-src/web_secrets.template /tmp/imagemonkey-core-src/web_secrets.go
RUN chown -R imagemonkey:imagemonkey /tmp/imagemonkey-core-src
RUN chmod -R u+rwx /tmp/imagemonkey-core-src

RUN /bin/su - imagemonkey -c "cd /tmp/imagemonkey-core-src/ && export GOPATH=/home/imagemonkey/go && export GOBIN=/home/imagemonkey/bin && /usr/local/go/bin/go get -d && /usr/local/go/bin/go install api.go api_secrets.go common.go imagedb.go parser.go label_graph.go auth.go && /usr/local/go/bin/go install web.go web_secrets.go common.go imagedb.go auth.go parser.go && /usr/local/go/bin/go install statworker.go web_secrets.go common.go imagedb.go parser.go && /usr/local/go/bin/go install populate_labels.go common.go web_secrets.go && /usr/local/go/bin/go install auto_unlocker.go api_secrets.go"

RUN rm -rf /tmp/imagemonkey-core/src

RUN mkdir -p /home/imagemonkey/donations
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/donations
RUN chmod -R u+rwx /home/imagemonkey/donations

RUN mkdir -p /home/imagemonkey/unverified_donations
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/unverified_donations
RUN chmod -R u+rwx /home/imagemonkey/unverified_donations

RUN mkdir -p /var/log/imagemonkey-api
RUN mkdir -p /var/log/imagemonkey-web
RUN mkdir -p /var/log/imagemonkey-statworker
RUN mkdir -p /var/log/imagemonkey-auto-unlocker

RUN cp -r /root/imagemonkey-core/wordlists /home/imagemonkey/
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/wordlists/
RUN chmod -R u+rx /home/imagemonkey/wordlists/

# copy html directory to final destination
RUN cp -r /root/imagemonkey-core/html /home/imagemonkey
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/html
RUN chmod -R u+rx /home/imagemonkey/html

# copy js directory to final destination
RUN cp -r /root/imagemonkey-core/js /home/imagemonkey
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/js
RUN chmod -R u+rx /home/imagemonkey/js

# copy css directory to final destination
RUN cp -r /root/imagemonkey-core/css /home/imagemonkey
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/css
RUN chmod -R u+rx /home/imagemonkey/css

# copy img directory to final destination
RUN cp -r /root/imagemonkey-core/img /home/imagemonkey
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/img
RUN chmod -R u+rx /home/imagemonkey/img

# copy geoip database to final destination
RUN mkdir -p /home/imagemonkey/geoip_database/
RUN cp -r /root/imagemonkey-core/geoip_database/GeoLite2-Country.mmdb /home/imagemonkey/geoip_database/
RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/geoip_database
RUN chmod -R u+rx /home/imagemonkey/geoip_database

# change listen address in postgres config file to localhost
RUN echo listen_addresses='localhost' >> /etc/postgresql/9.6/main/postgresql.conf

# do not add blog subscription worker...we do not need that in the docker image
RUN rm /etc/supervisor/conf.d/imagemonkey-blog-subscription-worker.conf

#do not add trending labels worker...we do not need that in the docker image
RUN rm /etc/supervisor/conf.d/imagemonkey-trending-labels-worker.conf

#populate labels in database
RUN /root/imagemonkey-core/env/docker/start_postgres.sh && cd /home/imagemonkey/bin/ && ./populate_labels --dryrun=false && cd /root/


#replace playground url and api url in supervisor config file

ENV API_BASE_URL http://127.0.0.1:8081

RUN sed -i.bak 's/-api_base_url=https:\/\/api.imagemonkey.io/-api_base_url=xxxxxx/g' /etc/supervisor/conf.d/imagemonkey-web.conf
RUN cp /etc/supervisor/conf.d/imagemonkey-web.conf /tmp/imagemonkey-web.conf.bak
RUN sed -i.bak 's/-playground_base_url=https:\/\/playground.imagemonkey.io//g' /etc/supervisor/conf.d/imagemonkey-web.conf


EXPOSE 8080
EXPOSE 8081


CMD ["/root/imagemonkey-core/env/docker/startup.sh"]
