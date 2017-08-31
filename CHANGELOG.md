## 0.13.1 (2017-8-31)

Bug fixes:

- Codeship fix ([#122](https://github.com/turnerlabs/harbor-compose/issues/122))

- Restart fix ([#123](https://github.com/turnerlabs/harbor-compose/issues/123))


## 0.13.0 (2017-8-28)

Features:

- Adds Codeship build provider (`generate --build-provider codeship`) ([#101](https://github.com/turnerlabs/harbor-compose/issues/101))

- Adds `restart` command ([#109](https://github.com/turnerlabs/harbor-compose/issues/109))

- Asterisks are now displayed when entering password for login. ([#113](https://github.com/turnerlabs/harbor-compose/issues/113))

- Adds richer validation/feedback for `up` command ([#111](https://github.com/turnerlabs/harbor-compose/issues/111))

- Adds environment name validation for `up` command ([#110](https://github.com/turnerlabs/harbor-compose/issues/110))


Bug fixes:

- The `ps` command doesn't work when using expanded docker compose build context ([#112](https://github.com/turnerlabs/harbor-compose/issues/112))


## 0.12.1 (2017-7-24)

Bug fixes:

- generate --build-provider circleciv2 now uses `docker 17.06.0-ce` (to support multi-stage build)


## 0.12.0 (2017-7-12)

Features:

- Harbor Compose now scaffolds out `Circle CI v2 config files` that enable `CI/CD pipelines` into Harbor (with more providers to come). ([#95](https://github.com/turnerlabs/harbor-compose/issues/95))

- Added a new `buildtoken` command (and sub commands) for managing Harbor build tokens ([#100](https://github.com/turnerlabs/harbor-compose/issues/100))

- The `init` command now outputs `docker-compose.yml` files to make it easier to Dockerize/Harborize ([#35](https://github.com/turnerlabs/harbor-compose/issues/35))

- The `ps` command now includes service endpoint(s) ([#94](https://github.com/turnerlabs/harbor-compose/issues/94))

- All commands should return a useful exit code should ([#96](https://github.com/turnerlabs/harbor-compose/issues/96))

- Added ability to configure harbor-compose via a config file ([#97](https://github.com/turnerlabs/harbor-compose/issues/97))


Bug fixes:

- The up command continues to run even if user is not authenticated ([#102](https://github.com/turnerlabs/harbor-compose/issues/102))


## 0.11.0 (2017-5-16)

Features:

- `generate --build-provider` feature along with `circleciv1` implementation ([#84](https://github.com/turnerlabs/harbor-compose/issues/84))

- Use of variable substitution in docker-compose.yml for things other than envvars ([#82](https://github.com/turnerlabs/harbor-compose/issues/82))

Bug fixes:

- Environment variables containing equals signs ("=") are not parsed correctly ([#77](https://github.com/turnerlabs/harbor-compose/issues/77))

- HEALTHCHECK environment variable not found when in an included env_file ([#89](https://github.com/turnerlabs/harbor-compose/issues/89))

- up now checks catalog before cataloging ([#93](https://github.com/turnerlabs/harbor-compose/issues/93))

- Switched download scripts from wget to curl ([#80](https://github.com/turnerlabs/harbor-compose/issues/80))


## 0.10.1 (2017-3-2)

Bug fixes:

- Fixed bug in `deploy -e` command.


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
