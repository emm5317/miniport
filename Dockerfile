FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /miniport ./cmd/miniport

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /miniport /miniport
EXPOSE 8092
ENV MINIPORT_HOST=0.0.0.0
ENTRYPOINT ["/miniport"]
