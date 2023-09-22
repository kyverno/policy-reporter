FROM golang:1.21 as builder

ARG LD_FLAGS='-s -w -linkmode external -extldflags "-static"'
ARG TARGETPLATFORM

WORKDIR /app

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

COPY go.* ./
RUN go env && go mod download

COPY . .

RUN CGO_ENABLED=1 go build -ldflags="${LD_FLAGS}" -tags="sqlite_unlock_notify" -o /app/build/policyreporter -v

FROM scratch
LABEL MAINTAINER="Frank Jogeleit <frank.jogeleit@gweb.de>"

WORKDIR /app

USER 1234

COPY --from=builder /app/LICENSE.md .
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/build/policyreporter /app/policyreporter
# copy the debian's trusted root CA's to the final image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080

ENTRYPOINT ["/app/policyreporter", "run"]
