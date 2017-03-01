## 0.10.0 (2017-3-1)

Features:

- Support for CI/CD via the new `deploy` and `catalog` commands.

## 0.9.0 (2017-2-1)

Features:

- catalog command has been added ([#67](https://github.com/turnerlabs/harbor-compose/issues/67)).
- all up commands catalog images

## 0.8.3 (2017-1-9)

Features:

- Now publishing a docker image to quay.io as part of the CI build.


## 0.8.2 (2017-1-3)

Bug fixes:

- Fixes overly restrictive authentication validation issues ([#58](https://github.com/turnerlabs/harbor-compose/issues/58) and [#60](https://github.com/turnerlabs/harbor-compose/issues/60)).


## 0.8.1 (2016-12-20)

Features:

- Improved error handling/logging around trigger.


## 0.8.0 (2016-10-31)

Features:

- Added [full support for docker compose environment variables](https://github.com/turnerlabs/harbor-compose/issues/32).  This means you can now manage harbor environment variables in .env files or pass them in via the shell.  See the [Docker Compose docs](https://docs.docker.com/compose/environment-variables/) for more info.

- Added support for [streaming one or many shipments' container logs](https://github.com/turnerlabs/harbor-compose/issues/11) to the console via the `harbor-compose logs -t` "tail" flag.

- Added ability to filter on a specific container replica's logs using the container id.
  
- Added `ignoreImageVersion` flag to shipment config in `harbor-compose.yml` that makes `harbor-compose up` ignore the image defined in `docker-compose.yml`. 


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
