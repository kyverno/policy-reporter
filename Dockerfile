FROM golang:1.16 as builder

WORKDIR /app
COPY . .

RUN go get -d -v \
    && go install -v

RUN make build

FROM alpine:latest
LABEL MAINTAINER "Frank Jogeleit <frank.jogeleit@gweb.de>"

WORKDIR /app

RUN apk add --update --no-cache ca-certificates

RUN addgroup -S policyreporter && adduser -u 1234 -S policyreporter -G policyreporter

USER 1234

COPY --from=builder /app/LICENSE.md .
COPY --from=builder /app/build/policyreporter /app/policyreporter

EXPOSE 2112

ENTRYPOINT ["/app/policyreporter", "run"]