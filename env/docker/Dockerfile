FROM debian:9

MAINTAINER bbernhard version: 0.3

RUN apt-get update && apt-get install --no-install-recommends -y postgresql-9.6 nginx nginx-extras redis-server git supervisor wget build-essential postgresql-server-dev-9.6 postgresql-contrib-9.6 dos2unix neovim postgresql-9.6-postgis-2.3 postgresql-9.6-postgis-2.3-scripts curl autoconf ca-certificates gcc make automake pkg-config uuid-dev zlib1g-dev lsb-release sudo gir1.2-glib-2.0 python3 python3-pip && rm -rf /var/lib/apt/lists/*

RUN apt-get update \
	&& wget https://my-netdata.io/kickstart.sh --directory-prefix=/tmp/ \
    && chmod u+rx /tmp/kickstart.sh \
    && /tmp/kickstart.sh --non-interactive \
    && rm /tmp/kickstart.sh \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /root/imagemonkey-core
RUN git clone https://github.com/bbernhard/imagemonkey-core.git /root/imagemonkey-core \
	&& cd /root/imagemonkey-core \
	&& git checkout develop

RUN chmod +x /root/imagemonkey-core/env/docker/start_postgres.sh \
	&& chmod +x /root/imagemonkey-core/env/docker/startup.sh

RUN cd /tmp && wget https://github.com/arkhipov/temporal_tables/archive/v1.2.0.tar.gz \
 && mkdir -p /tmp/temporal_table \
 && cd /tmp/temporal_table && tar xvf /tmp/v1.2.0.tar.gz \
 && cd /tmp/temporal_table/temporal_tables-1.2.0/ && make && make install && cd /root/ \
 && cp /root/imagemonkey-core/env/postgres/schema.sql /tmp/schema.sql \
 && chown postgres:postgres /tmp/schema.sql \
 && chmod u+rx /tmp/schema.sql \
 && echo CREATE EXTENSION uuid-ossp; >> /tmp/create_extension \
 && chown -R postgres:postgres /tmp/create_extension \
 && chmod -R u+rx /tmp/create_extension \
 && cp /root/imagemonkey-core/env/postgres/defaults.sql /tmp/defaults.sql \
 && chown postgres:postgres /tmp/defaults.sql \
 && chmod u+rx /tmp/defaults.sql \
 && cp /root/imagemonkey-core/env/postgres/indexes.sql /tmp/indexes.sql \
 && chown postgres:postgres /tmp/indexes.sql \
 && chmod u+rx /tmp/indexes.sql \
 && cp -r /root/imagemonkey-core/env/postgres/functions /tmp/postgres_functions \
 && chown -R postgres:postgres /tmp/postgres_functions \
 && chmod -R u+rx /tmp/postgres_functions \
 && cp -r /root/imagemonkey-core/env/postgres/stored_procs /tmp/postgres_stored_procs \
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

RUN /bin/su - imagemonkey -c "mkdir -p /home/imagemonkey/go"
ENV GOPATH="/home/imagemonkey/go"
RUN /bin/su - imagemonkey -c "mkdir -p /home/imagemonkey/bin"
ENV GOBIN="/home/imagemonkey/bin"


RUN cd /tmp/ \
   && wget https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz \
   && tar -C /usr/local -xzf go1.12.1.linux-amd64.tar.gz \
   && cd /root/

RUN cp /root/imagemonkey-core/env/docker/conf/supervisor/* /etc/supervisor/conf.d/ \
 && cp /root/imagemonkey-core/env/docker/src/* /root/imagemonkey-core/src/

RUN cp -r /root/imagemonkey-core/src /tmp/imagemonkey-core-src \
 && cp /tmp/imagemonkey-core-src/api_secrets.template /tmp/imagemonkey-core-src/api_secrets.go \
 && cp /tmp/imagemonkey-core-src/web_secrets.template /tmp/imagemonkey-core-src/web_secrets.go \
 && cp /tmp/imagemonkey-core-src/api_secrets.template /root/imagemonkey-core/src/api_secrets.go \
 && cp /tmp/imagemonkey-core-src/web_secrets.template /root/imagemonkey-core/src/web_secrets.go \
 && chown -R imagemonkey:imagemonkey /tmp/imagemonkey-core-src \
 && chmod -R u+rwx /tmp/imagemonkey-core-src \
 && cp /root/imagemonkey-core/tests/secrets.go.template /root/imagemonkey-core/tests/secrets.go \
 && cp /tmp/imagemonkey-core-src/shared_secrets.go.template /root/imagemonkey-core/src/shared_secrets.go \
 && cp /tmp/imagemonkey-core-src/api_secrets.template /root/imagemonkey-core/src/api_secrets.go

#add sudo (we need that that to install opencv via the gocv script)
RUN useradd -m docker && echo "docker:docker" | chpasswd && adduser docker sudo

ENV PATH $PATH:/usr/local/go/bin/

#install gocv
RUN /usr/local/go/bin/go get -u -d gocv.io/x/gocv \
	&& cd /home/imagemonkey/go/src/gocv.io/x/gocv \
	&& make install 

# until this pull request (https://github.com/h2non/bimg/pull/266) is merged and https://github.com/h2non/bimg/issues/269 is resolved, use the fork
RUN curl -s https://raw.githubusercontent.com/bbernhard/bimg/master/preinstall.sh | bash -

# that is a requirement for go-jsonnet
RUN go get github.com/fatih/color

RUN chown -R imagemonkey:imagemonkey /home/imagemonkey/go && chmod -R u+rwx /home/imagemonkey/go \
 && /bin/su - imagemonkey -c "cd /tmp/imagemonkey-core-src/ && export GOPATH=/home/imagemonkey/go && export GOBIN=/home/imagemonkey/bin && /usr/local/go/bin/go get -d && /usr/local/go/bin/go install api.go api_secrets.go auth.go label_graph.go && /usr/local/go/bin/go install web.go web_secrets.go auth.go && /usr/local/go/bin/go install statworker.go web_secrets.go && /usr/local/go/bin/go install populate_labels.go web_secrets.go && /usr/local/go/bin/go install auto_unlocker.go api_secrets.go && /usr/local/go/bin/go install data_processor.go api_secrets.go" \
 && rm -rf /tmp/imagemonkey-core/src

#create directories + set permissions
RUN mkdir -p /home/imagemonkey/public_backups \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/public_backups \
 && chmod -R u+rwx /home/imagemonkey/public_backups \
 && touch /home/imagemonkey/public_backups/public_backups.json \
 && mkdir -p /home/imagemonkey/donations \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/donations \
 && chmod -R u+rwx /home/imagemonkey/donations \
 && mkdir -p /home/imagemonkey/unverified_donations \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/unverified_donations \
 && chmod -R u+rwx /home/imagemonkey/unverified_donations \
 && mkdir -p /home/imagemonkey/quarantine \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/quarantine \
 && chmod -R u+rwx /home/imagemonkey/quarantine \
 && mkdir -p /var/log/imagemonkey-api \
 && mkdir -p /var/log/imagemonkey-web \
 && mkdir -p /var/log/imagemonkey-statworker \
 && mkdir -p /var/log/imagemonkey-auto-unlocker \
 && mkdir -p /var/log/imagemonkey-data-processor \
 && cp -r /root/imagemonkey-core/wordlists /home/imagemonkey/ \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/wordlists/ \
 && chmod -R u+rx /home/imagemonkey/wordlists/


# copy to final destination
RUN cp -r /root/imagemonkey-core/html /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/html \
 && chmod -R u+rx /home/imagemonkey/html \
 && cp -r /root/imagemonkey-core/js /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/js \
 && chmod -R u+rx /home/imagemonkey/js \
 && cp -r /root/imagemonkey-core/css /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/css \
 && chmod -R u+rx /home/imagemonkey/css \
 && cp -r /root/imagemonkey-core/img /home/imagemonkey \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/img \
 && chmod -R u+rx /home/imagemonkey/img \
 && mkdir -p /home/imagemonkey/geoip_database/ \
 && cp -r /root/imagemonkey-core/geoip_database/GeoLite2-Country.mmdb /home/imagemonkey/geoip_database/ \
 && chown -R imagemonkey:imagemonkey /home/imagemonkey/geoip_database \
 && chmod -R u+rx /home/imagemonkey/geoip_database

RUN echo "[]" > /home/imagemonkey/public_backups/public_backups.json

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


ENTRYPOINT ["/root/imagemonkey-core/env/docker/startup.sh"]
