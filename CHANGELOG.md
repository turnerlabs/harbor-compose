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