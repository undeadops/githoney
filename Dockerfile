FROM golang:alpine as base

# All these steps will be cached
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

FROM base as test

# These are failing because of libhoney... 
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -v ./...


FROM base as build
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/githoney *.go


FROM alpine:3.10 as release

COPY --from=build /go/bin/githoney /usr/bin/githoney

CMD ["githoney"]
