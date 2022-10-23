# xk6-gtp
Client extension for interacting with a use GTP proto of your k6 test.

## Build
```shell=
go install github.com/dmarkham/enumer@latest
go install go.k6.io/xk6/cmd/xk6@latest
make
```

## support scenario
### gtpv2
- [x] Node monitoring (Echo Request/Echo Response)
- [ ] Create Session  (Create Session Request/Create Session Response)
- [ ] Delete Session (Delete Session Request/Delete Session Response)
- [ ] Modify Bearer (Modify Bearer Request/Modify Bearer Response)
- [ ] Delete Bearer (Delete Bearer Request/Delete Bearer Response)