FROM golang:1.23-alpine

WORKDIR /app
COPY . .

# Install dependencies & build script
RUN go mod tidy && \
    go build -o token-finder .

CMD ["./token-finder"]
