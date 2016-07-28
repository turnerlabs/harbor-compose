# Compose file reference

The Harbor Compose file is a YAML file defining shipments, containers, and environment variables. The default path for a Harbor Compose file is ./harbor-compose.yml.

## Shipment configuration reference

This section contains a list of all configuration options supported by Harbor Compose.


### version

There is currently only 1 version available.


### shipments

This defines a list of one or more shipments that are part of your application.  

`harbor-compose up` will then bring up all of the shipments listed here.

```yaml
version: "1"
shipments:
  shipment1:    
    ...
  shipment2:    
    ...    
```

### env

This defines which environment your shipment will run in (e.g; dev, qa, prod, etc.).  Must be a string.   


### barge

This defines which managed cluster your shipment will run on.  Must be a string that matches one of our supported clusters.

Note that you must run `harbor-compose down` when changing this value.


### environment

Add environment variables at the shipment/environment level. This must be a dictionary. Any boolean values; true, false, yes no, need to be enclosed in quotes to ensure they are not converted to True or False by the YML parser.

Note that these environment variables get injected into all containers that are associated with the shipment.

```yaml
environment:
  RACK_ENV: development
  SHOW: "true"
```

### containers

This is a list of containers that are part of the shipment.  It is an array of string values that must be a valid service that exists in the `docker-compose.yml` file.  

This also means that you can have containers in your docker-compose.yml that get started when you dev locally, but are not referenced in your harbor-compose.yml, and therefore, not deployed to Harbor.

For example if you have a `docker-compose.yml` like...

```yaml
version: "2"
services:
  service1:
  ...
  service2:
  ...  
```

Then your `harbor-compose.yml` could be:

```yaml
my-shipment:    
  containers:
    - service1
    - service2 
...   
```

but the following would be invalid:

```yaml
my-shipment:    
  containers:
    - some-other-service
...   
```

### replicas

This value defines how many instances of your container will be run behind a load balancer.  Must be a positive integer.  Note that the value applies to all containers in a shipment.  If you want to scale your set of containers differently, you can move them onto different shipments.

```yaml
replicas: 2
```

### group

This value defines which customer group your shipment belongs to.  The group determines who can access your shipment.  It must be a  valid group that is defined in Harbor.

```yaml
group: mss
```

### property

This value defines which property your shipment serves.

```yaml
property: cnn
```

### project

This value defines which project your shipment is a part of.

```yaml
project: expansion
```

### product

This value defines your shipment as a product.

```yaml
product: mss-my-app-web
```

## Docker Compose configuration options

The following options are currently supported by Harbor Compose.  Note that you are free to use all of the Docker Compose options when working with Docker Compose, however, only the following options are used by Harbor Compose.

### [build](https://docs.docker.com/compose/compose-file/#build)

Can be used to build images locally (`docker-compose build`) that can be pushed (`docker-compose push`) and run on Harbor.

```yaml
build: .
```

### [environment](https://docs.docker.com/compose/compose-file/#environment)

These environment variables get injected into your container.  Any values here will override values that are also specified in `harbor-compose.yml`.

```yaml
environment:
  RACK_ENV: development
  SHOW: "true"
```

### [image](https://docs.docker.com/compose/compose-file/#image)

The tagged Docker image that is deployed to Harbor.

```yaml
image: registry.services.dmtio.net/my-web-app:1.0.0
```

### [ports](https://docs.docker.com/compose/compose-file/#ports)

The docker exposed ports (HOST:CONTAINER) will be mapped to the (FRONT-END:BACK-END) ports on the load balancer.  Note the currently only a single Harbor port is supported.

```yaml
ports:
  - "80:5000"
```