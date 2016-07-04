# harbor-compose

A tool for defining and running multi-container Docker applications on Harbor.  With Harbor Compose, you start with a standard [Docker Compose](https://docs.docker.com/compose/) file to configure your application’s services. You then add a Harbor Compose file to configure Harbor-specific settings.  Then, using a single command, you create and start all the services from your configuration.

Using Harbor Compose is basically a four-step process.

1. Define your app’s environment with a `Dockerfile` so it can be reproduced anywhere.

2. Define the services that make up your app in `docker-compose.yml` so they can be run together in an isolated environment.  You can use the standard Docker Compose commands (like `docker-compose build` and `docker-compose up`) to build/run/test your Docker app locally.

3. When you're ready to launch your Docker app on Harbor, you define the Harbor-specifc parameters in a `harbor-compose.yml` file.

4. Run `harbor-compose up` and Harbor Compose will start and run your entire app on a managed barge.


Just like `docker-compose`, `harbor-compose` has commands for managing the lifecycle of your application:

- Start and stop services
- View the status of running services
- Stream the log output of running services


A simple `docker-compose.yml` might look like this:

```yaml
version: '2'
services:
  web:
    image: registry.services.dmtio.net/my-app-web:1.0.0
    ports:
     - "80:5000"
    environment:
      PORT: 5000
      HEALTHCHECK: "/hc"
      FOO: bar
```

A `harbor-compose.yml` might look like this:

```yaml
shipments:
  my-app:    
    env: staging
    replicas: 3
    group: mss  
    property: turner
    project: radius
    product: mss-radius-app    
    environment:
      BARGE: corp-sandbox
    containers:
      - web
```


Then to start your application...

```
$ harbor-compose up --user foo
```

To stop your application and remove all running containers...

```
$ harbor-compose down --user foo
```

