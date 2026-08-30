FROM golang:1.27.0-alpine3.24 AS build

ARG version=unknown

RUN if [ -z "$version" ]; then \
			echo "version is not set"; \
			exit 1; \
    fi

RUN apk update && \
		apk add git

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download && \
		go mod verify

COPY . .

RUN go build  \
    -o ../bin/livebuffer \
    -ldflags "-X github.com/matthiasharzer/livebuffer/cmd/version.version=$version"  \
    ./main.go

FROM alpine:3.24

RUN apk update && \
		apk add --no-cache ffmpeg streamlink

COPY --from=build /go/bin/livebuffer /usr/local/bin/livebuffer

WORKDIR /var/lib/livebuffer

ENTRYPOINT ["/usr/local/bin/livebuffer"]

