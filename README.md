# xk6-gtp

A client extension for interacting with the GTP protocol in your k6 tests.


## Preparation

Required packages and tools:

- [mise](https://github.com/jdx/mise)

Install the tools required for development:

```shell=
make install-dev-pkg
```

## Build

```shell=
make install-go-tools
make build
```

## Running Tests

```shell
./out/bin/xk6 run example/echo-stress.js

./out/bin/pgw
```

## Supported Scenarios

### GTPv2-C

- [x] Node monitoring (Echo Request/Echo Response)
- [x] Create Session  (Create Session Request/Create Session Response)
  - [x] sgw->pgw scenario
- [x] Delete Session (Delete Session Request/Delete Session Response)
  - [x] sgw->pgw scenario
- [x] Modify Bearer (Modify Bearer Request/Modify Bearer Response)
  - [x] sgw->pgw scenario
- [ ] Delete Bearer (Delete Bearer Request/Delete Bearer Response)

## Special Thanks

This PoC takes full advantage of [go-gtp](https://github.com/wmnsk/go-gtp). Thanks to @wmnsk and all the developers.

## Developer Settings

```shell
# Format, lint, commit message validation, etc.
pre-commit install

# Mob programming
co-author hook > .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg

# Create Docker image
make docker-build
make docker-release
```
