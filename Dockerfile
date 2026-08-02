FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/deployctl ./cmd/deployctl

FROM alpine:3.20
RUN apk add --no-cache ca-certificates docker-cli kubectl
COPY --from=build /out/deployctl /usr/local/bin/deployctl
ENTRYPOINT ["deployctl"]
