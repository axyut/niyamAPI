FROM golang:1.24.2-alpine AS builder

WORKDIR /app

ENV GO111MODULE=on \
    CGO_ENABLED=0

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# -ldflags "-s -w": Strips debug symbols and DWARF tables, significantly reducing binary size.
RUN go build -o /app/niyam -ldflags "-s -w" ./main.go

FROM scratch

WORKDIR /app

# Copy only the compiled binary from the 'builder' stage
COPY --from=builder /app/niyam .

# Hugging Face Spaces often expose port 7860 by default for Gradio/Streamlit, still exposing...
EXPOSE 7860

CMD ["./niyam"]