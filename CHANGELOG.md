## 0.8.0 (2016-10-31)

Features:

  - Added [full support for docker compose environment variables](https://github.com/turnerlabs/harbor-compose/issues/32).  This means you can now manage harbor environment variables in .env files or pass them in via the shell.  See the [Docker Compose docs](https://docs.docker.com/compose/environment-variables/) for more info.

  Example:

  Put your environment variables in a file, for example `web.env` that looks like:

```
VALUE_1=foo
VALUE_2=bar
```

  and put your secrets in a git-ignored file named `.env` that looks:

```
DB_PASSWORD=asdf
```

  and then reference them in your `docker-compose.yml` file:

```yaml
version: "2"
services:
  my-app:
    image: registry.services.dmtio.net/my-app:1.0
    ports:
    - 80:3000
    env_file:
    - web.env    
    environment:      
      DB_PASSWORD: ${DB_PASSWORD}
```  

  then run `harbor-compose up` to upload the environment variables to Harbor and restart your shipments.

  - Added support for [streaming one or many shipments' container logs](https://github.com/turnerlabs/harbor-compose/issues/11) to the console via the `harbor-compose logs -t` "tail" flag.

  - Added ability to filter on a specific container replica's logs using the container id.
  
  Example:
```
$ harbor-compose ps

SHIPMENT:      mss-poc-sqs-web   
ENVIRONMENT:   dev               
STATUS:        Running           
CONTAINERS:    1                 
REPLICAS:      2

ID        IMAGE                                           STATUS    STARTED        RESTARTS   LAST STATE   
54bb5c8   registry.services.dmtio.net/poc-sqs-web:1.0.0   running   1 minute ago   0                       
cb6bd89   registry.services.dmtio.net/poc-sqs-web:1.0.0   running   1 minute ago   0                       
-----

SHIPMENT:      mss-poc-sqs-worker   
ENVIRONMENT:   dev                  
STATUS:        Running              
CONTAINERS:    1                    
REPLICAS:      3

ID        IMAGE                                              STATUS    STARTED          RESTARTS   LAST STATE   
4458f9b   registry.services.dmtio.net/poc-sqs-worker:1.0.0   running   1 minute ago     0                       
78bb34c   registry.services.dmtio.net/poc-sqs-worker:1.0.0   running   28 seconds ago   0                       
719e295   registry.services.dmtio.net/poc-sqs-worker:1.0.0   running   26 seconds ago   0


$ harbor-compose logs -t 719e295

Logs For:  mss-poc-sqs-worker dev
[719e295]
poc-sqs-worker:719e295  | npm info it worked if it ends with ok
poc-sqs-worker:719e295  | npm info using npm@3.9.5
poc-sqs-worker:719e295  | npm info using node@v6.2.2
poc-sqs-worker:719e295  | npm info lifecycle worker@1.0.0~prestart: worker@1.0.0
poc-sqs-worker:719e295  | npm info lifecycle worker@1.0.0~start: worker@1.0.0
poc-sqs-worker:719e295  | > worker@1.0.0 start /usr/src/app
poc-sqs-worker:719e295  | > node .
poc-sqs-worker:719e295  | Example app listening on port 6000!
poc-sqs-worker:719e295  | no messages
poc-sqs-worker:719e295  | no messages
poc-sqs-worker:719e295  | no messages
poc-sqs-worker:719e295  | no messages
poc-sqs-worker:719e295  | no messages 
```

- Added `ignoreImageVersion` flag to shipment config in `harbor-compose.yml` that makes `harbor-compose up` ignore the image defined in `docker-compose.yml`. 

```
shipments:
  my-shipment:
    env: dev
    barge: corp-sandbox
    ignoreImageVersion: true
```


## 0.7.0 (2016-10-07)

Features:

  - Added [`ps` command](https://github.com/turnerlabs/harbor-compose/issues/14) for getting shipment status

Bug fixes:

  - Fixed `logs` command which addresses [bug #40](https://github.com/turnerlabs/harbor-compose/issues/40)


## 0.6.1 (2016-09-06)

Bug fixes:

  - Error creating shipment with multiple containers [#47](https://github.com/turnerlabs/harbor-compose/issues/47)
  - Added additional logging when in verbose (-v) mode 


## 0.6.0 (2016-08-26)

Features:

  - Removed `--user` flag
  - Added [`login` and `logout` commands](https://github.com/turnerlabs/harbor-compose/issues/10) for managing temporary authentication token

Bug fixes:

  - The `generate` command no longer outputs environment variables to harbor-compose.yml which addresses [bug #38](https://github.com/turnerlabs/harbor-compose/issues/38) 


## 0.5.0 (2016-08-9)

Features:

  - Changed `logs` command to merge container logs by default and added `--separate` (or `-s`) flag to keep separate
  - Added support for creating shipment environments to `up` command  
  - Added `--delete` (or `-d`) flag to `down` command to delete a shipment environment after brining it down
  - Implemented `init` command to facilitate creation of `harbor-compose.yml` files

Other:

- Start using [godep](https://github.com/tools/godep) to manage dependencies
- Documented compose yaml formats
- Added changelog


## 0.4.0 (2016-07-18)

Features:

  - Implemented `generate` command


## 0.3.0 (2016-07-08)

Features:

  - Implemented `up` command


## 0.2.0 (2016-07-07)

Features:

  - Implemented `logs` command


## 0.1.0 (2016-07-01)

Features:

  - Implemented `down` command (doesn't delete shipment)  