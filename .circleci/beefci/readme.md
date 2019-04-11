# BeefCI

Where's the beef?

I find it hard to reason about CircleCI's filter ignore tag build system, so I made a tool. This way you don't have to

This assumes you're using CircleCI version `2.1` and using `workflows`.

```
yarn install
./beef.js /path/to/.circle/config.yml BRANCH TAG
```


