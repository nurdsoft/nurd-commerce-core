FROM golang:1.24-alpine AS builder

# Copy the code from the host and compile it
WORKDIR /go/src/github.com/nurdsoft/nurd-commerce-core
COPY . ./

# This allows for caching of dependencies
RUN go mod download
RUN go build -o nurd-commerce .

FROM alpine:latest AS prod
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /go/src/github.com/nurdsoft/nurd-commerce-core/entrypoint.sh /
COPY --from=builder /go/src/github.com/nurdsoft/nurd-commerce-core/nurd-commerce /
COPY --from=builder /go/src/github.com/nurdsoft/nurd-commerce-core/config.yaml /config.yaml
COPY --from=builder /go/src/github.com/nurdsoft/nurd-commerce-core/migrations /migrations

EXPOSE 8080

CMD [ "/entrypoint.sh" ]
