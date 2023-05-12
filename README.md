# xk6-gtp
Client extension for interacting with a use GTP proto of your k6 test.
🚧 This project WIP... 🚧

## Prepair
require asdf installed.
[how to asdf install](https://asdf-vm.com/guide/getting-started.html#_2-download-asdf)

Install tools required for development.
```shell=
make install-dev-pkg
```

## Build
```shell=
make install-go-tools
make build
```

## Test Running
```shell
./out/bin/xk6gtp run example/echo-stress.js

./out/bin/pgw
```

## Support scenario
### GTPv2-C
- [x] Node monitoring (Echo Request/Echo Response)
- [ ] Create Session  (Create Session Request/Create Session Response)
- [ ] Delete Session (Delete Session Request/Delete Session Response)
- [ ] Modify Bearer (Modify Bearer Request/Modify Bearer Response)
- [ ] Delete Bearer (Delete Bearer Request/Delete Bearer Response)


##　Special thanks
This PoC takes full advantage of [go-gtp](https://github.com/wmnsk/go-gtp). Thanks to the @wmnsk and developer all.

## Developers Settings

```shell
pre-commit install
```
