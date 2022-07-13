FROM golang:1.18-alpine3.16 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata alpine-sdk bash && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser
WORKDIR $GOPATH/app/

# Using go mod.
COPY . $GOPATH/app/
RUN GO111MODULE=on go mod download

# Copy source code
COPY . $GOPATH/app/

# Compile statically, CGO is required for sqlite to work, we don't add support here.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-w -extldflags '-static'" -o /app dockerlog2web.go
RUN chmod +x /app

FROM scratch
COPY --from=builder /app /app

# Import from builder - needed for accessing https apis
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# so we can run as 'appuser' created in builder
COPY --from=builder /etc/passwd /etc/passwd

# copy timezone infos from builder so we can use the databases time functions and reliably run cron jobs
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY . .
ENV TZ Europe/Berlin

# Use an unprivileged user
#USER appuser

EXPOSE 8080

# Run the main binary
CMD ["/app"]
