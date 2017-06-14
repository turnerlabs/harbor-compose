# How to contribute

Thanks for your interest in the project!  We want to welcome contributors so we put together the following set of guidelines to help participate.


## Getting Started

* If you have an idea for a new feature or would like to report a bug, please first check if your issue already exists, and if not, then open an issue.
  * Clearly describe the issue including steps to reproduce when it is a bug.
  * Make sure you fill in the earliest version that you know has the issue.
* Fork the repository on GitHub

## Developing Against Different APIs
* By default harbor-compose hits production apis. You can set an environment variable called `HC_CONFIG` or create file named
`~/.harbor/config` and overwrite any of the harbor apis.

example of the config file looks like so.

```
{
  "shipit": "http://shipit.foo.com",
  "catalogit": "http://catalogit.foo.com",
  "trigger": "http://trigger.foo.com",
  "authn": "http://auth.foo.com",
  "helmit": "http://helmit.foo.com",
  "harbor": "http://harbor.foo.com",
  "customs": "http://customs.foo.com"
}
```


## Making Changes

* Create a feature branch from where you want to base your work.
  * This is usually the `develop` branch.
  * To quickly create a feature branch; `git checkout -b feature/my-feature`. Please avoid working directly on the
    `master` branch.
* Make commits of logical units.
* Run `go fmt ./cmd` before committing.
* Make sure you have added the necessary tests for your changes.
* Run _all_ the tests to assure nothing else was accidentally broken (`go test ./cmd`).


## Submitting Changes

* Push your changes to a feature branch in your fork of the repository.
* Submit a pull request to the `develop` branch to the repository in the turnerlabs organization.


## Release Process

* After a feature pull request has been merged into the `develop` branch, a CI build will automatically kicked off.  The CI build will run unit tests, do a multi-platform build and automatically deploy the build to the [Github releases](releases) page as a pre-release using the latest tag (`git describe --tags`) as the version number.
* After the core team decides which features will be included in the next release, a release branch is created (e.g., `release/v0.5`) from develop.
* The `CHANGELOG.md` file is updated to document the release in the release branch.
* The release branch is merged to `master`, tagged, and pushed (along with tags).
* This will kick off a build that builds using the latest tag and deploys as a Github release.
* The release branch is then merged back to `develop`, tagged for pre-release (to start next version, e.g. v0.6.0-pre) and pushed.
