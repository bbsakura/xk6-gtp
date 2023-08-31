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
./out/bin/xk6 run example/echo-stress.js

./out/bin/pgw
```

## Support scenario
### GTPv2-C
- [x] Node monitoring (Echo Request/Echo Response)
- [x] Create Session  (Create Session Request/Create Session Response)
  - [x] sgw->pgw scenario
- [x] Delete Session (Delete Session Request/Delete Session Response)
  - [x] sgw->pgw scenario
- [x] Modify Bearer (Modify Bearer Request/Modify Bearer Response)
  - [x] sgw->pgw scenario
- [ ] Delete Bearer (Delete Bearer Request/Delete Bearer Response)


## Special thanks
This PoC takes full advantage of [go-gtp](https://github.com/wmnsk/go-gtp). Thanks to the @wmnsk and developer all.

## Developers Settings

```shell
# fmt, lint, commitmessage validate...etc checker
pre-commit install

# mob programing
co-author hook > .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg

# create docker image
make docker-build
make docker-release
```
