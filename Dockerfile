FROM golang:1.22-alpine AS builder
# Pass the GitHub token as a build argument
ARG GITHUB_TOKEN

# Install necessary tools
RUN apk update && apk add --no-cache make git ca-certificates build-base

# Create a temporary .netrc file for Git authentication
# This file is only used during the build and won't be copied to the final image.
RUN echo "machine github.com login ${GITHUB_TOKEN} password x-oauth-basic" > /root/.netrc && \
    chmod 600 /root/.netrc

WORKDIR /go/src/github.com/forbole/callisto
COPY . ./

# Download the wasmvm libraries for the current architecture
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.5.2/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.5.2/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN cp /lib/libwasmvm_muslc.$(uname -m).a /lib/libwasmvm_muslc.a

# Download dependencies and build the binary
RUN go mod download
RUN LINK_STATICALLY=true BUILD_TAGS="muslc" make build

FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates build-base
WORKDIR /callisto
# Copy the built binary from the builder stage.
COPY --from=builder /go/src/github.com/forbole/callisto/build/callisto /usr/bin/callisto
CMD [ "callisto" ]