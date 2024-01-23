FROM golang:1.21-alpine as build

WORKDIR /project
COPY ./src ./

RUN go mod download

RUN go mod vendor

RUN go build -mod=vendor -o ./user_service ./cmd

FROM golang:1.21-alpine

COPY ./src/database ./database
COPY --from=build /project/user_service  ./user_service

CMD ["./user_service"]
