FROM golang:latest as builder

WORKDIR /app
ADD go.mod .
ADD go.sum .
RUN go mod download
RUN go mod verify
ADD . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ./cmd ./letsencrypt/bin/cmd

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cmd /producer

CMD ["/producer"]
