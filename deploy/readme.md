# deploy

This folder contains files that are used in circle for building and testing all of the go source code.

## Circle CI builds

There are three modes regarding git tags and git branches that are captured in `./circleci/config.yml`, and each triggers different actions.

Firstly, all commits run the build job (which also runs unit tests in all `oneiro-ndev` repos under the `commands` vendor directory).

Secondly, tagged builds respond to commits that are tagged. See the README at the root of this repo for information on supported tags. They are useful for running "catchup", "integration", "push" and "deploy" jobs specifically.

The third case is a commit to master, this will run all steps and deploy to devnet.

## VSCode file associations

We have to deal with the fact that "dockerfiles" eschew the industry standard practice of file extensions. By convention, the docker files in this repository are suffixed with ".docker". You can configure VSCode to associate *.docker files with the following user setting:

```json
    "files.associations": {
        "*.docker":"dockerfile"
    },
```
