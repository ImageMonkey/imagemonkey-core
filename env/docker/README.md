The following document briefly describes how to start up your own `ImageMonkey` instance in a Docker container. 

* pull imagemonkey-core image via `docker pull imagemonkey-core`
* start docker instance with `docker run --ulimit nofile=90000:90000 -p 8080:8080 8081:8081 imagemonkey`

This will start a new ImageMonkey docker instance on your machine. 

**Detailed description of the docker run command**

`-p 8080:8080 8081:8081`

Both the ImageMonkey webservice and the ImageMonkey API listen on port 8080 resp. port 8081 inside the docker container.
In order to make ImageMonkey easily accessible on the host system, we are mapping the hosts ports 8080 and 8081 
to the corresponding ports inside the docker container. 

