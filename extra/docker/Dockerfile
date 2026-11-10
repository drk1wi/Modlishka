# Build from the repository root folder:
#   docker build -f extra/docker/Dockerfile -t modlishka .

###############################################################################
# Build stage
###############################################################################
FROM golang:alpine AS builder

# Cross-compile Go static binary
ENV CGO_ENABLED=0

RUN apk add --no-cache make git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

# Copy the source tree and build
COPY . ./
RUN make build

###############################################################################
# Runtime stage
#
# Alpine's supplies everything run-server.sh needs: sh, env, sed, awk, cut and tr
###############################################################################
FROM alpine

# Keep outbound TLS working if upstream verification is enabled.
RUN apk add --no-cache ca-certificates

ENV MODLISHKA_BIN="/bin/proxy"

COPY --from=builder /src/dist/proxy ${MODLISHKA_BIN}
COPY --from=builder /src/extra/docker/run-server.sh /bin/run-server.sh

RUN chmod +x /bin/run-server.sh

VOLUME [ "/data" ]
WORKDIR /data
CMD ["run-server.sh"]

EXPOSE 80 443
