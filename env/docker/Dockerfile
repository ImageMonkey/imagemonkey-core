FROM debian:9

MAINTAINER bbernhard version: 0.3

RUN apt-get update && apt-get install -y postgresql-9.6 nginx nginx-extras redis-server git supervisor wget build-essential postgresql-server-dev-9.6 postgresql-contrib-9.6 dos2unix neovim postgis
RUN mkdir -p /root/imagemonkey-core
RUN git clone https://github.com/bbernhard/imagemonkey-core.git /root/imagemonkey-core \
	&& cd /root/imagemonkey-core \
	&& git checkout develop

RUN chmod +x /root/imagemonkey-core/env/docker/start_postgres.sh
RUN chmod +x /root/imagemonkey-core/env/docker/startup.sh

RUN cd /tmp && wget https://github.com/arkhipov/temporal_tables/archive/v1.2.0.tar.gz \
 && mkdir -p /tmp/temporal_table \
 && cd /tmp/temporal_table && tar xvf /tmp/v1.2.0.tar.gz \
 && cd /tmp/temporal_table/temporal_tables-1.2.0/ && make && make install && cd /root/

RUN cp /root/imagemonkey-core/env/postgres/schema.sql /tmp/schema.sql \
 && chown postgres:postgres /tmp/schema.sql \
 && chmod u+rx /tmp/schema.sql

RUN echo CREATE EXTENSION uuid-ossp; >> /tmp/create_extension \
 && chown -R postgres:postgres /tmp/create_extension \
 && chmod -R u+rx /tmp/create_extension

RUN cp /root/imagemonkey-core/env/postgres/defaults.sql /tmp/defaults.sql \
 && chown postgres:postgres /tmp/defaults.sql \
 && chmod u+rx /tmp/defaults.sql

RUN cp /root/imagemonkey-core/env/postgres/indexes.sql /tmp/indexes.sql \
 && chown postgres:postgres /tmp/indexes.sql \
 && chmod u+rx /tmp/indexes.sql

RUN cp -r /root/imagemonkey-core/env/postgres/functions /tmp/postgres_functions \
 && chown -R postgres:postgres /tmp/postgres_functions \
 && chmod -R u+rx /tmp/postgres_functions

RUN cp -r /root/imagemonkey-core/env/postgres/stored_procs /tmp/postgres_stored_procs \
 && chown -R postgres:postgres /tmp/postgres_stored_procs \
 && chmod -R u+rx /tmp/postgres_stored_procs

RUN /root/imagemonkey-core/env/docker/start_postgres.sh && /bin/su - postgres -c "psql -c \"CREATE database imagemonkey;\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -c \"CREATE USER monkey WITH PASSWORD 'dbRuwMUo4Nfhs5hmMxhk';\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -c \"CREATE EXTENSION \"temporal_tables\";\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -c \"CREATE EXTENSION \"postgis\";\"" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/schema.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/create_extension" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/defaults.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/indexes.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/postgres_functions/fn_ellipse.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/postgres_functions/third_party/postgis_addons/postgis_addons.sql" \
	&& /bin/su - postgres -c "psql -d imagemonkey -f /tmp/postgres_stored_procs/sp_get_image_annotation_coverage.sql"

RUN rm -rf /tmp/create_extension \
 && rm -rf /tmp/schema.sql \
 && rm -rf /tmp/temporal_table \
 && rm -rf /tmp/defaults.sql \
 && rm -rf /tmp/indexes.sql \
 && rm -rf /tmp/postgres_functions \
 && rm -rf /tmp/postgres_stored_procs

RUN adduser imagemonkey --disabled-password --gecos "First Last,RoomNumber,WorkPhone,HomePhone" --home /home/imagemonkey
#RUN adduser imagemonkey supervisor

RUN /bin/su - imagemonkey -c "mkdir -p /home/imagemonkey/go"
ENV GOPATH="/home/imagemonkey/go"
RUN /bin/su - imagemonkey -c "mkdir -p /home/imagemonkey/bin"
ENV GOBIN="/home/imagemonkey/bin"


RUN cd /tmp/ \
   && wget https://redirector.gvt1.com/edgedl/go/go1.10.3.linux-amd64.tar.gz \
   && tar -C /usr/local -xzf go1.10.3.linux-amd64.tar.gz \
   && cd /root/

RUN cp /root/imagemonkey-core/env/docker/conf/supervisor/* /etc/supervisor/conf.d/
RUN cp /root/imagemonkey-core/env/docker/src/* /root/imagemonkey-core/src/

RUN cp -r /root/imagemonkey-core/src /tmp/imagemonkey-core-src
RUN mv /tmp/imagemonkey-core-src/api_secrets.template /tmp/imagemonkey-core-src/api_secrets.go \
 && mv /tmp/imagemonkey-core-src/web_secrets.template /tmp/imagemonkey-core-src/web_secrets.go
RUN chown -R imagemonkey:imagemonkey /tmp/imagemonkey-core-src
RUN chmod -R u+rwx /tmp/imagemonkey-core-src

RUN /bin/su - imagemonkey -c "cd /tmp/imagemonkey-core-src/ && export GOPATH=/home/imagemonkey/go && export GOBIN=/home/imagemonkey/bin && /usr/local/go/bin/go get -d && /usr/local/go/bin/go install api.go api_secrets.go common.go imagedb.go parser.go label_graph.go auth.go && /usr/local/go/bin/go install web.go web_secrets.go common.go imagedb.go auth.go parser.go && /usr/local/go/bin/go install statworker.go web_secrets.go common.go imagedb.go parser.go && /usr/local/go/bin/go install populate_labels.go common.go web_secrets.go && /usr/local/go/bin/go install auto_unlocker.go api_secrets.go && /usr/local/go/bin/go install data_processor.go api_secrets.go"

RUN rm -rf /tmp/imagemonkey-core/src

RUN mkdir -p /home/imagemonkey/donations \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/donations \
 && chmod -R u+rwx /home/imagemonkey/donations

RUN mkdir -p /home/imagemonkey/unverified_donations \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/unverified_donations \
 && chmod -R u+rwx /home/imagemonkey/unverified_donations

RUN mkdir -p /home/imagemonkey/quarantine \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/quarantine \
 && chmod -R u+rwx /home/imagemonkey/quarantine

RUN mkdir -p /var/log/imagemonkey-api \
 && mkdir -p /var/log/imagemonkey-web \
 && mkdir -p /var/log/imagemonkey-statworker \
 && mkdir -p /var/log/imagemonkey-auto-unlocker \
 && mkdir -p /var/log/imagemonkey-data-processor

RUN cp -r /root/imagemonkey-core/wordlists /home/imagemonkey/ \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/wordlists/ \
 && chmod -R u+rx /home/imagemonkey/wordlists/

# copy html directory to final destination
RUN cp -r /root/imagemonkey-core/html /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/html \
 && chmod -R u+rx /home/imagemonkey/html

# copy js directory to final destination
RUN cp -r /root/imagemonkey-core/js /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/js \
 && chmod -R u+rx /home/imagemonkey/js

# copy css directory to final destination
RUN cp -r /root/imagemonkey-core/css /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/css \
 && chmod -R u+rx /home/imagemonkey/css

# copy img directory to final destination
RUN cp -r /root/imagemonkey-core/img /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/img \
 && chmod -R u+rx /home/imagemonkey/img

# copy geoip database to final destination
RUN mkdir -p /home/imagemonkey/geoip_database/ \
 && cp -r /root/imagemonkey-core/geoip_database/GeoLite2-Country.mmdb /home/imagemonkey/geoip_database/ \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/geoip_database \
 && chmod -R u+rx /home/imagemonkey/geoip_database

# change listen address in postgres config file to localhost
RUN echo listen_addresses='localhost' >> /etc/postgresql/9.6/main/postgresql.conf

#populate labels in database
RUN /root/imagemonkey-core/env/docker/start_postgres.sh \
	&& cd /home/imagemonkey/bin/ \
	&& ./populate_labels --dryrun=false \
	&& cd /root/



ENV API_BASE_URL http://127.0.0.1:8081

EXPOSE 8080
EXPOSE 8081


CMD ["/root/imagemonkey-core/env/docker/startup.sh"]
