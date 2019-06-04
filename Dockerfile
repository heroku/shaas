FROM golang as builder

WORKDIR $GOPATH/src/github.com/heroku/shaas
ADD . $GOPATH/src/github.com/heroku/shaas
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go install

FROM ubuntu

RUN apt-get update && apt-get install -y bash && apt-get install -y curl
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/shaas .
EXPOSE 5000

CMD ["./shaas"]
