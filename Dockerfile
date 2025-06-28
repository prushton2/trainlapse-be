FROM alpine:latest

WORKDIR /app

RUN apk add go
RUN apk add gcc
RUN apk add curl

COPY . .
RUN go build -o main .

EXPOSE 3000

CMD ["./main"]