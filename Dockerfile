# Stage 1: Builder
# Use a specific, stable Go version with Alpine for a smaller build environment
FROM golang:1.24.2-alpine AS builder

# Set the working directory inside the builder container
WORKDIR /app

# Enable Go modules and disable CGO for a statically linked binary.
# CGO_ENABLED=0 is important for scratch or Alpine base images.
ENV GO111MODULE=on \
    CGO_ENABLED=0

# Copy go.mod and go.sum first to leverage Docker's build cache.
# This ensures dependencies are downloaded only if these files change.
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of your application source code
COPY . .

# Build the Go application.
# -o /app/niyam: Specifies the output path and name of the executable.
# -ldflags "-s -w": Strips debug symbols and DWARF tables, significantly reducing binary size.
# ./main.go: Assumes your main package is in main.go at the root. Adjust if it's elsewhere (e.g., ./cmd/niyam/main.go).
RUN go build -o /app/niyam -ldflags "-s -w" ./main.go

# ---

# Stage 2: Runner
# Use a minimal base image for the final production image.
# 'scratch' is the smallest possible image, containing only your executable.
FROM scratch

# Set the working directory in the final image
WORKDIR /app

# Copy only the compiled binary from the 'builder' stage
COPY --from=builder /app/niyam .

# If your Go application serves HTTP requests, you should expose the port.
# Hugging Face Spaces often expose port 7860 by default for Gradio/Streamlit,
# but for a custom Go app, you might use 8080 or another common port.
# Make sure your Go app listens on this port.
EXPOSE 7860

# Define the command to run your application.
# This should be your compiled executable.
CMD ["./niyam"]