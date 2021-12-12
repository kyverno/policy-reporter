FROM golang:1.17.2 as builder

ARG LD_FLAGS
ARG TARGETPLATFORM

WORKDIR /app
COPY . .

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN go get -d -v \
    && go install -v

RUN CGO_ENABLED=1 go build -ldflags="${LD_FLAGS}" -o /app/build/policyreporter -v

FROM scratch
LABEL MAINTAINER="Frank Jogeleit <frank.jogeleit@gweb.de>"

WORKDIR /app

USER 1234

COPY --from=builder /app/LICENSE.md .
COPY --from=builder /app/build/policyreporter /app/policyreporter
# copy the debian's trusted root CA's to the final image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 2112

ENTRYPOINT ["/app/policyreporter", "run"]
