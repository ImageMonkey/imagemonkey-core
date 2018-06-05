
  <img src="https://raw.githubusercontent.com/bbernhard/imagemonkey-core/develop/img/logo.png" align="left" width="180" >


ImageMonkey is a free, public open source dataset. With all the great machine learning frameworks available it's pretty easy to train pre-trained Machine Learning models with your own image dataset. However, in order to do so you need a lot of images. And that's usually the point where it get's tricky. You either have to create the training images yourself or scrape them together from various datasources. ImageMonkey aims to solve this problem, by providing a platform where users can drop their photos, tag them with a label, and put them into public domain.

---
![Alt Text](https://github.com/bbernhard/imagemonkey-core/raw/master/img/animation.gif)

# Getting started #

There are basically two ways to set up your own `ImageMonkey` instance. You can either set up everything by hand, which gives you the flexibility to choose your own linux distribution, monitoring tools and scrips or you could use our `Dockerfile` to spin up a new `ImageMonkey` instance within just a few minutes. 

## Docker ## 

[Run ImageMonkey inside Docker](https://github.com/bbernhard/imagemonkey-core/blob/develop/env/docker/README.md)

The docker image is for development only - do **NOT** use it in production!

## Manual Setup ##

The following section contains some notes on how to set up your own instance to host ImageMonkey yourself.
This should only give you an idea how you *could* configure your system. Of course you are totally free in choosing 
a different linux distribution, tools and scripts. If you are only interested in how to compile ImageMonkey, then you can jump directly to the *Build Application* section 

*Info:* Some commands are distribution (Debian 9.1) specific and may not work on your system. 

### Base System Configuration ###

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

### Database ###

* install PostgreSQL
* edit `/etc/postgresql/9.6/main/postgresql.conf` and set `listen_addresses = 'localhost'`
* restart PostgreSQL service with `service postgresql restart` to apply changes
* create database by applying schema `/env/postgres/schema.sql` with `psql -f schema.sql`
* create new postgres user `monkey` by executing the following in psql: 
```
CREATE USER monkey WITH PASSWORD 'your_password';

\connect imagemonkey 
GRANT ALL PRIVILEGES ON DATABASE imagemonkey to monkey;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO monkey;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO monkey;
GRANT USAGE ON SCHEMA blog TO monkey;

```
* test if newly created user works with: `psql -d imagemonkey -U monkey -h 127.0.0.1`

* populate labels with `go run populate_labels.go common.go web_secrets.go`
* add donation image provider with `insert into image_provider(name) values('donation');`

* build `temporal_table` extension, as described here: https://github.com/arkhipov/temporal_tables
* connect to imagemonkey database and execute `CREATE EXTENSION temporal_tables;`
* connect to imagemonkey database and execute `CREATE EXTENSION uuid-ossp;`
* apply `defaults.sql`

### Webserver & SSL ###

* install nginx with `apt-get install nginx`
* install nginx-extras with `apt-get install nginx-extras`
* install letsencrypt certbot with `apt-get install certbot`
* add a A-Record DNS entry which points to the IP address of your instance
* run `certbot certonly` to obtain a certificate for your registered domain
* modify `conf/nginx/nginx.conf` and replace `imagemonkey.io` and `api.imagemonkey.io` with your own domain names, copy it to `/etc/nginx/nginx.conf` and reload nginx with `service nginx reload`

### Build Application ###
**Minimal** required Go version: v1.9.2

* install git with `apt-get install git`
* install golang with `apt-get install golang`
* clone repository
* set GOPATH with `export GOPATH=$HOME/go`
* set GOBIN with `export GOBIN=$HOME/bin`
* install all dependencies with `go get -d ./... `
* install API application with `go install api.go api_secrets.go common.go imagedb.go`
* install API application with `go install web.go web_secrets.go common.go imagedb.go` 

### Miscellaneous ###
* copy `wordlists/en/misc.txt` to `/home/imagemonkey/wordlists/en/misc.txt`
* create donation directories with: 
```
mkdir -p /home/imagemonkey/donations
mkdir -p /home/imagemonkey/unverified_donations
```

### Watchdog ###
* install supervisor with `apt-get install supervisor`
* add `imagemonkey` user to supervisor group with `adduser imagemonkey supervisor`
* create logging directories with `mkdir -p /var/log/imagemonkey-api` and `mkdir -p /var/log/imagemonkey-web`
* copy `conf/supervisor/imagemonkey-api.conf` to `/etc/supervisor/conf.d/imagemonkey-api.conf`
* copy `conf/supervisor/imagemonkey-web.conf` to `/etc/supervisor/conf.d/imagemonkey-web.conf`
* run `supervisorctl reread && supervisorctl update && supervisorctl restart all`


### Datasync ###
**on imagemonkey-playground instance**
* install `rsync` with `apt-get install rsync`
* create a new user `backupuser` with `adduser backupuser` (use a strong password)
* change to user `backupuser` with `su backupuser` and create a new SSH key with `ssh-keygen -t ed25519 -a 100`
* copy SSH public key to imagemonkey instance with: `ssh-copy-id -i ~/.ssh/your_generated_id.pub backupuser@imagemonkey-host`
* give `backupuser` permissions to write to `/home/playground/donations` with: `chgrp backupuser /home/playground/donations && chmod g+rwx /home/playground/donations`
* add a new cronjob for the user `backupuser` with: `crontab -u backupuser -e` and add the following line (runs rsync every 15min):

```
*/15 * * * * rsync -a backupuser@imagemonkey.io:/home/imagemonkey/donations/ /home/playground/donations/
```
