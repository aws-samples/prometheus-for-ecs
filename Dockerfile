FROM golang:1.15 as builder
WORKDIR /src
RUN go mod vendor
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -mod=vendor -tags=netgo -o config-reloader cmd/main.go 


FROM scratch AS final
WORKDIR /home/prometheus-for-ecs
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/config-reloader .
ENV GO111MODULE=on
ENTRYPOINT ["./config-reloader"]
