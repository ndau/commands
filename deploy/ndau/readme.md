## docker files

These files are organized in a way that allows dependencies to be downloaded once, and then passed around for multiple builds.

In the CircleCI environment, for example, this means you `glide install` one time, and then build for ndaunode tests, ndaunode, and ndauapi.

`.bin/build.sh` builds the containers using `docker-compose up` with default/overridden variables.

## caveat

Since the `ndau-deps` image sticks around, there are commits in that image that contain a copy of our github key. Care must be taken to not expose this key. This is prevented by two measures 1) never uploading the ndau-deps image anywhere and 2) never copying the key out of the container to another container. The image itself is always built where it is used. It is no more exposed than the key file sitting on your machine and `docker rmi ndau-deps` removes it entirely.

## VSCode file associations

We have to deal with the fact that "dockerfiles" eschew the industry standard practice of file extensions. By convention, these are suffixed with ".docker". You can configure VSCode to associate *.docker files with the following user setting:

```json
    "files.associations": {
        "*.docker":"dockerfile"
    },
```
