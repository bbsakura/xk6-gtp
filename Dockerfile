ARG K6_VERSION
ARG XK6_VERSION
ARG K6_PROMTHEUS_VERSION

FROM golang:1.24.5 AS builder
ARG K6_VERSION=v0.47.0
ARG XK6_VERSION=v0.9.2

WORKDIR $GOPATH/src/go.k6.io/k6
COPY . .
RUN go install -trimpath go.k6.io/xk6/cmd/xk6@${XK6_VERSION}
RUN xk6 build \
  --with github.com/grafana/xk6-dashboard@latest \
  --with github.com/bbsakura/xk6-gtp@latest=.

RUN cp -r k6 $GOPATH/bin/k6
WORKDIR /go/src/app

USER k6:k6
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder --chown=k6:k6 /go/bin/k6 ./
COPY --from=builder --chown=k6:k6 /go/src/app ./

ENTRYPOINT [ "/app/k6" ]
