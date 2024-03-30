FROM golang:1.22-alpine AS build

WORKDIR /app

COPY . .
RUN go mod download && go mod verify

WORKDIR /app/cmd/app
RUN go build -o severws .

FROM busybox:1.36.1

WORKDIR /app

COPY --from=build /app/cmd/app/severws .
COPY .env .
COPY cert ./cert/

EXPOSE 8082
EXPOSE 2112

CMD ["./severws"]
