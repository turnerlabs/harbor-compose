## 0.18.0 (2018-09-20)

Features:

- Adds `migrate` command


## 0.17.0 (2018-02-07)

Features:

- Add support for `roles` to `init` command ([#174](https://github.com/turnerlabs/harbor-compose/issues/174))

- adds `loadbalancer` and `iam_role` to terraform output ([#177](https://github.com/turnerlabs/harbor-compose/issues/177))

- Add an `events` command to surface the k8s events in the helmit api ([#118](https://github.com/turnerlabs/harbor-compose/issues/118))

- adds fields to telemetry payload ([#173](https://github.com/turnerlabs/harbor-compose/issues/173))


Bug fixes:

- fixes file permissions with circle ci build provider ([#178](https://github.com/turnerlabs/harbor-compose/issues/178))



## 0.16.0 (2018-01-17)

Features:

- Adds new `env` command (ls, push, pull) ([#165](https://github.com/turnerlabs/harbor-compose/issues/165), [#137](https://github.com/turnerlabs/harbor-compose/issues/137), [#121](https://github.com/turnerlabs/harbor-compose/issues/121))

- `ps` and `logs` commands now work without yaml files ([#163](https://github.com/turnerlabs/harbor-compose/issues/163), [#150](https://github.com/turnerlabs/harbor-compose/issues/150))

- Adds `buildtoken env` command ([#164](https://github.com/turnerlabs/harbor-compose/issues/164))

- commands that output terraform should omit duplicate config in harbor-compose.yml ([#167](https://github.com/turnerlabs/harbor-compose/issues/167))

- telemetry integration ([#151](https://github.com/turnerlabs/harbor-compose/issues/151))


Bug fixes:

- `generate` should write log shipping env vars to harbor-compose.yml ([#141](https://github.com/turnerlabs/harbor-compose/issues/141))

- The `up` command should require and validate that env vars are not empty ([#138](https://github.com/turnerlabs/harbor-compose/issues/138))

- The `generate` command should properly escape env vars with $ characters ([#153](https://github.com/turnerlabs/harbor-compose/issues/153))

- healthcheckIntervalSeconds must be > healthcheckTimeoutSeconds ([#140](https://github.com/turnerlabs/harbor-compose/issues/140))


## 0.15.0 (2017-11-10)

Features:

- Add support for generating terraform source files ([#142](https://github.com/turnerlabs/harbor-compose/issues/142))

- Generate a terraform main.tf from 'init' command ([#143](https://github.com/turnerlabs/harbor-compose/issues/143))

- friendly dns install ([#149](https://github.com/turnerlabs/harbor-compose/issues/149))


## 0.14.0 (2017-9-25)

Features:

- Support for `hidden` environment variables ([#19](https://github.com/turnerlabs/harbor-compose/issues/19))

- Output a `hidden.env` in the `init` command ([#125](https://github.com/turnerlabs/harbor-compose/issues/125))

- More efficient environment variable updates in `up` ([#126](https://github.com/turnerlabs/harbor-compose/issues/126))

- Support for `enableMonitoring` ([#79](https://github.com/turnerlabs/harbor-compose/issues/79))

- Support for `healthcheckIntervalSeconds` ([#129](https://github.com/turnerlabs/harbor-compose/issues/129))

- Support for `healthcheckTimeoutSeconds` ([#78](https://github.com/turnerlabs/harbor-compose/issues/78))

- Adds `barge` to `ps` command output ([#134](https://github.com/turnerlabs/harbor-compose/issues/134))


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

## 0.9.1 (2017-2-1)

- bump new version because build failed to use correct tag

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
