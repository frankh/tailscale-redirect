FROM golang:1.22-alpine as build
WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download -x

COPY *.go .
RUN CGO_ENABLED=0 go build -v -o tailscale-redirect main.go


FROM alpine:latest
WORKDIR /data
COPY --from=build /work/tailscale-redirect /tailscale-redirect

ENTRYPOINT ["/tailscale-redirect"]
