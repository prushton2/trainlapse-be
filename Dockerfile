FROM alpine:latest
# RUN apt-get update && apt-get install -y

WORKDIR /app

# RUN pip install --upgrade pip
RUN apk add go
RUN apk add gcc

COPY . .
RUN go build -o main .

EXPOSE 3000

CMD ["./main"]