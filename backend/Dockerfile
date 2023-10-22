# Stage build: will build the go into single binary
FROM golang:1.20-alpine AS build
WORKDIR /go/src/app
COPY ./app ./app
COPY ./public ./public
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./conf.env ./
COPY ./main.go ./

RUN go env -w CGO_ENABLED=0
RUN go env -w GO111MODULE="on"
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64
RUN ls -al

# Install the dependencies with docker cache.
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build the service binary with docker cache.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -v .

CMD ["/go/src/app/nhcafe-backend", "serve"]
