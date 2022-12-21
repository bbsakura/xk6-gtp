# xk6-gtp
Client extension for interacting with a use GTP proto of your k6 test.

## Build
```shell=
go install github.com/dmarkham/enumer@latest
go install go.k6.io/xk6/cmd/xk6@latest
make
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
