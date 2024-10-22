FROM golang:1.23.1 as go
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOBIN=/bin
ARG BUILDARCH=amd64
RUN go install github.com/go-delve/delve/cmd/dlv@v1.22.0
ADD https://github.com/spiffe/spire/releases/download/v1.10.3/spire-1.10.3-linux-${BUILDARCH}-musl.tar.gz .
RUN tar xzvf spire-1.10.3-linux-${BUILDARCH}-musl.tar.gz -C /bin --strip=2 spire-1.10.3/bin/spire-server spire-1.10.3/bin/spire-agent

FROM go as build
WORKDIR /build
COPY go.mod go.sum ./
COPY ./internal/imports ./internal/imports
RUN go build ./internal/imports
COPY . .
RUN go build -o /bin/app .

FROM build as test
CMD go test -test.v ./...

FROM test as debug
CMD dlv -l :40000 --headless=true --api-version=2 test -test.v ./...

FROM alpine as runtime
COPY --from=build /bin/app /bin/app
ENTRYPOINT ["/bin/app"]