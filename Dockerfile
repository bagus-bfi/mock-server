# Use the official Golang image as the base image
FROM golang:1.23.4-bookworm

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 9090 to the outside world
EXPOSE 9090

# Command to run the executable
CMD ["./main"]