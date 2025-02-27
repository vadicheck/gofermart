FROM golang:1.22-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc musl-dev gettext vim

# dependencies
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

# build
COPY ./ ./
RUN go build -o ./bin/app cmd/gophermart/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY --from=builder /usr/local/src/internal/app/migration/migrations /migrations
COPY [".env", "/"]

EXPOSE 8082
