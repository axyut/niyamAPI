# Stage 1: Builder
# Uses a Debian-based Go image for building the application.
# This makes it easier to install C development libraries required by gosseract.
FROM golang:1.24.2-bookworm AS builder

# Set working directory inside the container.
WORKDIR /app

# Install system dependencies required for Tesseract/gosseract at build time.
# tesseract-ocr: The main OCR engine.
# libtesseract-dev: Development headers (crucial for CGO compilation).
# libleptonica-dev: Leptonica development headers (dependency for Tesseract).
# pkg-config: Tool used by Go's CGO to find libraries.
# git: Needed for Go's VCS stamping during build.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    tesseract-ocr \
    libtesseract-dev \
    libleptonica-dev \
    pkg-config \
    git \
    && rm -rf /var/lib/apt/lists/*

# IMPORTANT: CGO_ENABLED MUST BE 1 for gosseract to work.
ENV CGO_ENABLED=1
# PKG_CONFIG_PATH helps CGO locate the installed C libraries.
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig

# Copy go.mod and go.sum first to leverage Docker's build cache.
COPY go.mod go.sum ./

# Download all Go module dependencies.
# This step only runs if go.mod or go.sum change.
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the application.
# -o /app/niyam: Specifies the output binary name and path.
# -ldflags "-s -w": Strips debug symbols and DWARF tables, reducing binary size.
RUN go build -o /app/niyam -ldflags "-s -w" ./main.go

# --- Stage 2: Runner ---
# Uses a minimal Debian-based image for the final production container.
# This image will contain only the necessary runtime components for your Go app and Tesseract.
FROM debian:bookworm-slim AS runner

# Set working directory.
WORKDIR /app

# Install Tesseract OCR runtime, language data, and CA certificates for TLS.
# tesseract-ocr: The Tesseract engine itself (runtime binaries).
# tesseract-ocr-eng: English language data.
# tesseract-ocr-nep: Nepali language data.
# tesseract-ocr-hin: Hindi language data.
# tesseract-ocr-script-deva: Devanagari script data (useful for both Hindi and Nepali).
# ca-certificates: Provides the SSL/TLS root certificates needed to verify secure connections. <--- ADDED
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    tesseract-ocr \
    tesseract-ocr-eng \
    tesseract-ocr-nep \
    tesseract-ocr-hin \
    tesseract-ocr-script-deva \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the 'builder' stage into the 'runner' stage.
COPY --from=builder /app/niyam /app/niyam

# Expose the port your application listens on.
EXPOSE 7860

# Set the default command to run your compiled application when the container starts.
CMD ["./niyam"]
