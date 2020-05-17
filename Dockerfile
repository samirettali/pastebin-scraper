FROM golang:alpine

RUN apk add --no-cache git

# Set necessary environment variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .

# Download dependencies
RUN go get ./...

RUN go build .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/pastebin-scraper .

# Command to run when starting the container
CMD ["/dist/pastebin-scraper"]
