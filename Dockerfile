FROM golang:1.23.2 AS builder

# Copy the code from the host and compile it
WORKDIR /go/src/github.com/nurdsoft/nurd-commerce-core
COPY . ./
RUN go test -race ./...
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix nocgo -o /app .

FROM alpine:latest AS prod
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app ./
COPY --from=builder /go/src/github.com/nurdsoft/nurd-commerce-core/config.yaml ./config.yaml
COPY --from=builder /go/src/github.com/nurdsoft/nurd-commerce-core/migrations ./migrations

EXPOSE 8080

CMD ./app migrate; ./app api;
