# Compose file reference

The Harbor Compose file is a YAML file defining shipments, containers, and environment variables. The default path for a Harbor Compose file is ./harbor-compose.yml.

## Shipment configuration reference

This section contains a list of all configuration options supported by a service definition.


### version

There is currently only 1 version available.


### shipments

This defines a list of one or more shipments that are part of your application.  

`harbor-compose up` will then bring up all of the shipments listed here.

```yaml
version: "1"
shipments:
  shipment1:    
    env: dev
    ...
  shipment2:    
    env: dev
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

