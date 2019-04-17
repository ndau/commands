# BeefCI

Where's the beef?

I find it hard to reason about CircleCI's filter ignore tag build system, so I made a tool. This way you don't have to wait for the CircleCI workflow to complete to test what jobs will run.

This assumes you're using CircleCI version `2.1` and using `workflows`.

```
yarn install
./beef.js /path/to/.circle/config.yml BRANCH TAG
```

Testing what a commit to the `josh-feature` branch will do.

```
$ ./beef.js ../config.yml josh-feature
Workflow: master-build
Workflow: tagged-build
 - build-deps
 - test
 - build
```

Testing what a commit to the `josh-feature` branch with the tag `josh-deploy` will do.

```
$ ./beef.js ../config.yml josh-feature josh-deploy
Workflow: master-build
Workflow: tagged-build
 - build-deps
 - test
 - build
 - push
 - deploy-devnet
 - integration
```

One commit to master.

```
$ ./beef.js ../config.yml master
Workflow: master-build
 - build-deps
 - test
 - build
 - push
 - deploy-devnet
 - integration
Workflow: tagged-build
```
