FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && CGO_ENABLED=0 go build -o /out/worker ./cmd/worker
FROM alpine:3.21
RUN adduser -D app
COPY --from=build /out/api /app/api
COPY --from=build /out/worker /app/worker
USER app
EXPOSE 8080
