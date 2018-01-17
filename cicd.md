### Harbor Compose and CI/CD

The `up` command is currently not very CI/CD friendly since it requires login credentials and uses internal-facing APIs.  That's where the `deploy` command comes in.  The `deploy` command can be used to trigger a deployment of new versions of Docker images specified in compose files (one or many shipments with one or many containers).  

This works from public build services (e.g.; Circle CI, Codeship, Travis CI, etc.) by using the shipment/environment build token specified with environment variables with the naming convention, `${SHIPMENT}_${ENV}_TOKEN`.

So, for example, to deploy an app with two shipments named, "mss-app-web" and "mss-app-worker" to your dev environment, you would add the following environment variables to your build environment (see the [buildtoken command](#the-buildtoken-command) below for more info).

```
$ harbor-compose buildtoken ls

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-web      dev           MSS_POC_SQS_WEB_DEV_TOKEN      3xFVlltLZ7JwPH20Km75DrpMwOk2a4yq
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3yuA8	
```

And then simply run the following to deploy all containers in all shipments specified in your compose files.

```
harbor-compose deploy
```

A shortcut for this is to use the `env` command.

```
eval "$(harbor-compose buildtoken env)"
harbor-compose deploy
```

If you wanted to conditionally deploy to a different environment (e.g., QA) in your build (maybe based on branch) using the same set of compose files, you could add another set of environment variables to your build.

```
$ harbor-compose buildtoken ls -e qa

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-web      qa            MSS_POC_SQS_WEB_QA_TOKEN       ihtvPrAH84ULVm6IC7LjWvXUgEhr7cnQ
mss-poc-sqs-worker   qa            MSS_POC_SQS_WORKER_QA_TOKEN    Y3Jk0DmMaUsoWO8mbI2Edn9Ixhwj14Vd
```

And then run

```
harbor-compose deploy -e qa
```

This allows for a clean CI/CD work flow in your build scripts...

```
docker-compose build
docker-compose push
harbor-compose deploy
```

If you're just doing CI and not CD, you can use the `catalog` command to catalog all of the built docker images but not deploy them.

```
docker-compose build
docker-compose push
harbor-compose catalog
```


#### generate --build-provider

The `generate` command has a `--build-provider` flag that can help with scenarios where teams want to take existing applications running on Harbor and migrate them to various third-party build CI/CD providers.  The idea is that a `build provider` can output the compose files along with any other necessary files required to do CI/CD using a particular provider.  The following is a list of supported providers.

##### local

Simply adds a docker-compose [build](https://github.com/turnerlabs/harbor-compose/blob/master/compose-reference.md#build) directive.


##### circleciv1

This provider will output a docker-compose.yml file with a [build](https://github.com/turnerlabs/harbor-compose/blob/master/compose-reference.md#build) directive and an image tagged with a Circle CI build number.  Note that environment variables updates in Harbor are not currently supported via the public API, and are therefore not outputted.  A `harbor-compose.yml` file and a [`circle.yml`](https://circleci.com/docs/1.0/configuration/) file are also outputted and are already setup to be able to catalog and deploy new images.  You can run this command in the root of your source code repo, and after linking your repo to Circle CI, you can commit/push the files and get basic CI/CD working.  For example:

```
$ harbor-compose generate mss-my-shipment dev --build-provider circleciv1

Be sure to supply the following environment variables in your Circle CI build:
DOCKER_USER (registry user)
DOCKER_PASS (registry password)
MSS_MY_SHIPMENT_DEV_TOKEN=2N3QFkQkdilwj34ezS2JKxwt6Fn3yuA7

done
```

circle.yml
```yaml
machine:
  pre:
    # install newer docker and docker-compose
    - curl -sSL https://s3.amazonaws.com/circle-downloads/install-circleci-docker.sh | bash -s -- 1.10.0
    - pip install docker-compose==1.11.2

    # install harbor-compose
    - sudo wget -O /usr/local/bin/harbor-compose https://github.com/turnerlabs/harbor-compose/releases/download//ncd_linux_amd64 && sudo chmod +x /usr/local/bin/harbor-compose
  services:
    - docker

dependencies:
  override:
    # login to quay registry
    - docker login -u="${DOCKER_USER}" -p="${DOCKER_PASS}" -e="." quay.io

compile:
  override:
    - docker-compose build

test:
  override:
    - docker-compose up -d
    - echo "tests run here"
    - docker-compose down

deployment:
  CI:
    branch: master
    commands:
      # push image to registry and catalog in harbor
      - docker-compose push
      - harbor-compose catalog
  CD:
    branch: develop
    commands:
      # push image to registry and deploy to harbor
      - docker-compose push
      - harbor-compose deploy		
```

##### circleciv2

Same as circleciv1 but outputs the v2 format.

```
$ harbor-compose generate mss-my-shipment dev --build-provider circleciv2
```

.circle/config.yml
```yaml
version: 2
jobs:
  build:
    docker:
      - image: quay.io/turner/harbor-cicd-image:v0.12.1
    working_directory: ~/app
    steps:
      - checkout
      - setup_remote_docker:
          version: 17.06.0-ce
      - run:
          name: Build app image
          command: docker-compose build
      - run:        
          name: Login to registry
          command: docker login -u="${DOCKER_USER}" -p="${DOCKER_PASS}" quay.io
      - run:
          name: Push app image to registry
          command: docker-compose push
      - run:
          name: Catalog in Harbor
          command: harbor-compose catalog
      - run:
          name: Deploy to Harbor
          command: |
            if [ "${CIRCLE_BRANCH}" == "develop" ]; then 
              harbor-compose deploy;
            fi
```

##### codeship

Outputs files to support CI/CD using [Codeship](https://documentation.codeship.com/).

```
$ harbor-compose generate mss-my-shipment dev -b codeship

Now you just need to:

- add your quay.io registry credentials to codeship.env
- download your AES key from your codeship project and put it in codeship.aes
- encrypt your codeship.env by running 'jet encrypt codeship.env codeship.env.encrypted'
- check in codeship.env.encrypted but don't check in codeship.env
```

codeship-steps.yml
```yaml
- service: cicd
  name: build image
  command: docker-compose build

- service: cicd
  name: push image to registry
  command: ./docker-push.sh

- service: cicd
  name: catalog image in harbor
  command: harbor-compose catalog

- service: cicd
  tag: develop
  name: deploy develop branch to harbor
  command: harbor-compose deploy
```

Other files that are outputted:
- codeship-services.yml
- codeship.env
- codeship.aes
- docker-push.sh
- .gitignore


#### The buildtoken command

The `buildtoken` command has three sub commands.
```
Available Commands:
  env         display the commands to set up the environment for the deploy command
  get         displays a build token for the requested shipment and environment
  list        list harbor build tokens for shipment environments
```

The `env` command configures your shell for deployment.
```
$ eval "$(harbor-compose buildtoken env)"

export MSS_POC_SQS_WORKER_DEV_TOKEN=2N3QFkQkdilwj34ezS2JTxwt6Fn3asdf
# Run this command to configure your shell:
# eval "$(harbor-compose buildtoken env)"
```

The `get` command prompts you for a shipment and environment.
```
$ harbor-compose buildtoken get

Shipment: mss-poc-sqs-worker
Environment: dev

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3asdf
```

Or, alternatively, you can pass the shipment and environment as args.
```
$ harbor-compose buildtoken get mss-poc-sqs-worker dev

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3asdf
```

The `list` (or `ls`) command provides the build tokens for the shipments in your harbor-compose.yml.
```
$ harbor-compose buildtoken ls

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-web      dev           MSS_POC_SQS_WEB_DEV_TOKEN      3xFVlltLZ7JwPH20Km75DrpMwOk2a4yq
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3yuA8	
```

Or, if you want to deploy your shipments to an environment different from the one that's specified in your harbor-compose.yml, you can use the following.
```
$ harbor-compose buildtoken ls -e qa

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-web      qa            MSS_POC_SQS_WEB_QA_TOKEN       ihtvPrAH84ULVm6IC7LjWvXUgEhr7cnQ
mss-poc-sqs-worker   qa            MSS_POC_SQS_WORKER_QA_TOKEN    Y3Jk0DmMaUsoWO8mbI2Edn9Ixhwj14Vd
```
