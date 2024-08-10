FROM golang:latest AS build

WORKDIR /file-server

COPY go.mod go.sum main.go ./

RUN mkdir files && CGO_ENABLED=0 go build -ldflags="-w -s"

FROM gcr.io/distroless/static:nonroot

WORKDIR /file-server

COPY --from=build --chown=nonroot file-server/files files
COPY --from=build file-server/file-server file-server
COPY favicon.ico index.html ./

EXPOSE 1323

ENTRYPOINT ["./file-server"]