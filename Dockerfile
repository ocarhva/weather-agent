FROM golang:1.27-bookworm AS build
RUN apt-get update && apt-get install -y --no-install-recommends build-essential && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o /out/agent ./cmd/agent

FROM debian:bookworm-slim
ARG ORT_VERSION=1.19.2
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
 && curl -fsSL https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/onnxruntime-linux-x64-${ORT_VERSION}.tgz | tar xz -C /tmp \
 && cp /tmp/onnxruntime-linux-x64-${ORT_VERSION}/lib/libonnxruntime.so.${ORT_VERSION} /usr/lib/libonnxruntime.so \
 && rm -rf /tmp/onnxruntime-* /var/lib/apt/lists/* && apt-get purge -y curl && apt-get autoremove -y
ENV ONNXRUNTIME_LIB=/usr/lib/libonnxruntime.so
WORKDIR /app
COPY --from=build /out/agent ./agent
COPY models ./models
COPY web ./web
CMD ["./agent"]
