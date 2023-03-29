FROM golang:alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o golang-example .

EXPOSE 8080

CMD  ./golang-example --config . database migrate -m ./migrations && \
     ./golang-example --config . database seed && \
     ./golang-example --config . serve \