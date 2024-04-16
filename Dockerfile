FROM golang:alpine as build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s"

FROM scratch

COPY --from=build build/file-server file-server

COPY --from=build build/favicon.ico favicon.ico

COPY --from=build build/index.html index.html

EXPOSE 1323

ENTRYPOINT ["./file-server"]