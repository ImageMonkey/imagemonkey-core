The following document briefly describes how to start up your own ImageMonkey instance in a Docker container. 

# Scenario #1
This is the most common scenario and the easiest to set up. Choose this option, if your web browser and your docker container will run on the same machine. 

* install docker
* run `docker pull bbernhard/imagemonkey-core`
* start docker instance with `docker run --ulimit nofile=90000:90000 -p 8080:8080 -p 8081:8081 imagemonkey`

This will start a new ImageMonkey docker instance on your machine. After your docker instance is up and running, you will see the following screen: 

Now open your browser and navigate to `http://127.0.0.1:8080`

# Scenario #2
As docker acquires a significant portion of your systems resources, one might want to run the docker instance on a different machine. 

Let's assume your workstation has the private IP `192.168.1.9`. As your workstation is quite old, you want to run the ImageMonkey docker container on a different machine (e.q Raspberry Pi) which is in the same subnet and has the IP `192.168.1.16`. 

* install docker
* run `docker pull bbernhard/imagemonkey-core`
* start docker instance with `docker run -e API_BASE_URL=http://192.168.1.16:8081 --ulimit nofile=90000:90000 -p 8080:8080 -p 8081:8081 imagemonkey`

The docker run command looks almost identical to the one in Scenario #1, except that we are setting the `API_BASE_URL` environmental variable inside the docker container to the host systems IP (i.e `192.168.1.16`).


**Detailed description of the docker run command** 

`-p 8080:8080 -p 8081:8081`

Both the ImageMonkey webservice and the ImageMonkey API listen on port 8080 resp. port 8081 inside the docker container.
In order to make ImageMonkey easily accessible on the host system, we are mapping the hosts ports 8080 and 8081 
to the corresponding ports inside the docker container. 

The docker port mapping is also helpful if you already have a service running on the hosts system, that is listening on port 8080 or 8081. In that case you would need to choose different host ports and map those to port 8080 and 8081 inside the docker container. e.q: the commandline options `-p 8082:8080 -p 8083:8081` map the host port 8082 to the docker container port 8080 and the host port 8083 to the docker container port 8081. 

`--ulimit nofile=90000:90000` This commandline option changes the number of available file descriptors within the docker container. Without that option `redis` will not be able to run inside the docker container. 

