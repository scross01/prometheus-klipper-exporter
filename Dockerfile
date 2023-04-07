# build stage
FROM golang:latest as builder
LABEL maintainer="Stephen Cross <scross01@gmail.com>"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
COPY collector ./collector
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o main .

# run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 9101
CMD ["./main"]