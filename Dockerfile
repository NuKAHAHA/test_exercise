FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk --no-cache add ca-certificates

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=build /main .
COPY .env .

EXPOSE 8000

CMD ["./main"]
