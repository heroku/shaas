FROM golang as builder

WORKDIR $GOPATH/src/github.com/heroku/shaas
ADD . $GOPATH/src/github.com/heroku/shaas
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go install -mod=vendor

FROM ubuntu

RUN apt-get update && apt-get install -y bash && apt-get install -y curl
RUN groupadd -g 1000 app
RUN useradd -s /bin/bash -u 1000 -g 1000 -d /app app
RUN mkdir -p /app && chown app:app /app

USER app
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/shaas .
COPY --from=builder /go/src/github.com/heroku/shaas/bin/pseudo-interactive-bash /app/bin/pseudo-interactive-bash

EXPOSE 5000

ENTRYPOINT ["/app/shaas"]
