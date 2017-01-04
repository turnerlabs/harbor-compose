# harbor-compose

A tool for defining and running multi-container Docker applications on Harbor.

[![CircleCI](https://circleci.com/gh/turnerlabs/harbor-compose/tree/master.svg?style=shield)](https://circleci.com/gh/turnerlabs/harbor-compose/tree/master)  

With Harbor Compose, you start with a standard [Docker Compose](https://docs.docker.com/compose/) file to configure your application’s services. You then add a Harbor Compose file to configure Harbor-specific settings.  Then, using a single command, you create and start all the services from your configuration.

Using Harbor Compose is basically a four-step process.

1. Define your app’s environment with a `Dockerfile` so it can be reproduced anywhere.

2. Define the services that make up your app in `docker-compose.yml` so they can be run together in an isolated environment.  You can use the standard Docker Compose commands (like `docker-compose build`, `docker-compose push`, `docker-compose up`, etc.) to build/run/test your Docker app locally.

3. When you're ready to launch your Docker app on Harbor, you define the Harbor-specifc parameters in a [`harbor-compose.yml`](compose-reference.md) file.

4. Run `harbor-compose up` and Harbor Compose will start and run your entire app on a managed barge.


Just like `docker-compose`, `harbor-compose` has similar commands for managing the lifecycle of your app on Harbor:

- Start and stop services
- View the status of running services
- Stream the log output of running services


A simple `docker-compose.yml` might look like this:

```yaml
version: "2"
services:
  web-app:
    image: registry.services.dmtio.net/my-web-app:1.0.0
    ports:
      - "80:5000"
    environment:
      PORT: 5000
      HEALTHCHECK: /hc
      CONTAINER_LEVEL: foo
```

A [`harbor-compose.yml`](compose-reference.md) might look like this:

```yaml
version: "1"
shipments:
  my-web-app:    
    env: dev
    barge: corp-sandbox
    containers:
      - web-app    
    replicas: 2
    group: mss
    property: turner.com
    project: my-web-app
    product: my-web-app    
```

Then to start your application...

```
$ harbor-compose up
```

Access your app logs...

```
$ harbor-compose logs
```

Get the status of your shipment...

```
$ harbor-compose ps
```

To stop your application, remove all running containers and delete your load balancer...

```
$ harbor-compose down
```

#### Getting Started

There are currently two installation options for Harbor Compose.

1. Download the binary from the [Github releases section](https://github.com/turnerlabs/harbor-compose/releases).

- You can use the following script to download and install (update the URL for your desired platform and version).

```
$ sudo wget -O /usr/local/bin/harbor-compose https://github.com/turnerlabs/harbor-compose/releases/download/v0.8.2/ncd_darwin_amd64 && sudo chmod +x /usr/local/bin/harbor-compose
```

2. Run as a docker container.

```
$ docker run -it --rm -v `pwd`:/work quay.io/turner/harbor-compose up
```

- or if you want to reuse your session (and use a specific version):

```
$ docker run -it —rm -v `pwd`:/work -v ${HOME}/.harbor:/root/.harbor quay.io/turner/harbor-compose:0.8.2 up
```


To get started with an existing shipment, you can run the following to generate `docker-compose.yml` and [`harbor-compose.yml`](compose-reference.md) files, by specifying the shipment name and environment as args.  Note that you will be prompted to login if you don't already have a token or if your token has expired.  For example:

```
$ harbor-compose generate my-shipment dev
```

To create new shipments and environments, you can use the `init` command to generate [`harbor-compose.yml`](compose-reference.md) files.  `init` will ask you questions to build your compose file.  Note that you use the `--yes` flag to accept defaults and generate one quickly.

```
$ harbor-compose init
```

This will output the files in the current directory.  You can then run a bunch of useful commands, for example...

Run your shipment locally in Docker (practically identically to how it runs in Harbor)...

```
$ docker-compose up
```

Scale your shipment by changing the replicas in `harbor-compose.yml`, or change your environment variables and re-deploy, or deploy a new image, etc....

```
$ harbor-compose up
```

Get the status of your shipment(s) using the `ps` command.  With this command you can see the status of each container replica, when it started and the last known state.  For example:

```
$ harbor-compose ps

SHIPMENT:      mss-poc-multi-container   
ENVIRONMENT:   dev                       
STATUS:        Running                   
CONTAINERS:    2                         
REPLICAS:      2

ID        IMAGE                                                        STATUS    STARTED      RESTARTS   LAST STATE              
ab97cef   registry.services.dmtio.net/mss-poc-multi-container:1.0.0    running   1 week ago   1          terminated 1 week ago   
873a390   registry.services.dmtio.net/mss-poc-multi-container2:1.0.0   running   1 week ago   1          terminated 1 week ago   
73fad42   registry.services.dmtio.net/mss-poc-multi-container:1.0.0    running   5 days ago   2          terminated 5 days ago   
db93650   registry.services.dmtio.net/mss-poc-multi-container2:1.0.0   running   5 days ago   2          terminated 5 days ago   
```


To stop your application, delete your load balancer, and delete the environment.  Note that this is equivalent to setting replicas = 0 and triggering.

```
$ harbor-compose down --delete
```

You can also manage multiple shipments using Harbor Compose by listing them in your harbor-compose.yml file.  This is particularly useful if you have a web/worker, or microservices type application where each shipment can be scaled independently.

#### Authentication

Some commands (`up`, `down`, `generate`) require authentication and will automatically prompt you for your credentials.  A temporary (6 hours) authentication token is stored on your machine so that you don't have to login when running each command.  If you want to logout and remove the authentication token, you can run the `logout` command.  You can also explicitly login by running the `login` command.


### Compose file reference

See the [full harbor-compose.yml reference](compose-reference.md) along with which [docker-compose.yml](https://docs.docker.com/compose/) properties are supported by Harbor Compose.