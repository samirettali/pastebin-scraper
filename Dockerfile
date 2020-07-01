FROM golang:alpine AS build

RUN apk add --no-cache git

# Set necessary environment variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build
RUN go build .

# Use scratch image to run
FROM ubuntu AS bin

# Move to /app directory as the place for resulting binary folder
WORKDIR /app

# Copy binary from build to main folder
COPY --from=build /build/pastebin-scraper .

# Command to run when starting the container
CMD ["/app/pastebin-scraper"]
