FROM --platform=$BUILDPLATFORM golang:1.25 AS build

ARG VERSION
ARG GOTARGET
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.work go.work.sum ./

COPY apps/finder/go.mod apps/finder/go.sum ./apps/finder/
COPY apps/reader/go.mod apps/reader/go.sum ./apps/reader/

COPY pkg/kafka/go.mod pkg/kafka/go.sum ./pkg/kafka/
COPY pkg/otel/go.mod pkg/otel/go.sum ./pkg/otel/
COPY pkg/mongo/go.mod pkg/mongo/go.sum ./pkg/mongo/

RUN go work sync

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-X 'main._version=${VERSION}'" \
    -o /bin/service \
    ${GOTARGET}

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/service /bin/service

CMD ["/bin/service"]
