# README #

## INFRASTRUCTURE ##

The following section contains some notes on how to set up your own instance to host `imagemonkey` yourself.
This should only give you an idea how you *could* configure your system. Of course you are totally free in choosing 
a different linux distribution, tools and scripts. 

Info: Some commands are distribution (Debian 9.1) specific and may not work on your system. 

* create a new user `imagemonkey`  with `adduser imagemonkey` 
* disable root login via ssh by changing the `PermitRootLogin` line in `/etc/ssh/sshd_config` to `PermitRootLogin no`)
* block all ports except port 22, 443 and 80 (on eth0) with: 
```
#!bash

iptables -P INPUT DROP && iptables -A INPUT -i eth0 -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -i eth0 -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -i eth0 -p tcp --dport 80 -j ACCEPT
```

* allow all established connections with:

```
#!bash

iptables -A INPUT  -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
```

* allow all loopback access with:
```
#!bash
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT
```

* install `iptables-persistent` to load firewall rules at startup
* save firewall rules with: `iptables-save > /etc/iptables/rules.v4`
* verify that rules are loaded with `iptables -L`
* install PostgreSQL
* edit `/etc/postgresql/9.6/main/postgresql.conf` and set `listen_addresses = 'localhost'`
* restart PostgreSQL service with `service postgresql restart` to apply changes
* create database by applying schema `/env/postgres/schema.sql` with `psql -f schema.sql`
* create new postgres user `monkey` by executing the following in psql: 
```
CREATE USER monkey WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE imagemonkey to monkey;

```
* test if newly created user works with: `psql -d imagemonkey -U monkey -h 127.0.0.1`

* install nginx with `apt-get install nginx`
* install nginx-extras with `apt-get install nginx-extras`
* install letsencrypt certbot with `apt-get install certbot`
* add a A-Record DNS entry which points to the IP address of your instance
* run `certbot certonly` to obtain a certificate for your registered domain
* modify `conf/nginx/nginx.conf` and replace `imagemonkey.io` with your own domain name, copy it to `/etc/nginx/nginx.conf` and reload nginx with `service nginx reload`
* install supervisor with `apt-get install supervisor`
* add `imagemonkey` user to supervisor group with `adduser imagemonkey supervisor`
* create logging directories with `mkdir -p /var/log/imagemonkey-api` and `mkdir -p /var/log/imagemonkey-web`

### Building Application ###
* install git with `apt-get install git`
* install golang with `apt-get install golang`
* clone repository
* set GOPATH with `export GOPATH=$HOME/go`
* set GOBIN with `export GOBIN=$HOME/bin`
* build application with `go build api.go api_secrets.go common.go imagedb.go -release`
* install all dependencies with `go get -d ./... 
* copy `wordlists/en/misc.txt` to `/home/imagemonkey/wordlists/en/misc.txt`
* create donations directory with: `mkdir -p /home/imagemonkey/donations`
* copy `conf/supervisor/imagemonkey-api.conf` to `/etc/supervisor/conf.d/imagemonkey-api.conf`
* copy `conf/supervisor/imagemonkey-web.conf` to `/etc/supervisor/conf.d/imagemonkey-web.conf`
* run `supervisorctl reread && supervisorctl update && supervisorctl restart all`
