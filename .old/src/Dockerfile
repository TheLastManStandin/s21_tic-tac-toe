FROM golang:1.26 AS build

WORKDIR /app

COPY . .

RUN go build -o /app/bin/tictactoe ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=build /app/bin/tictactoe /app/tictactoe

EXPOSE 5005

CMD ["/app/tictactoe"]
