FROM golang:1.23.4-alpine AS builder
RUN mkdir /src /app
WORKDIR /src
COPY go.mod .
COPY *.go ./
RUN go get -u && go mod tidy
RUN go build -o /app/resolver *.go

FROM scratch
EXPOSE 5354
COPY --from=builder /app/resolver /app/
WORKDIR /app
CMD ["/app/resolver"]
