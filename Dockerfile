FROM docker.io/golang:1.26-alpine AS build

WORKDIR /usr/local/go/src/github.com/arizon-dread/secret-syncer
LABEL MAINTAINER=github.com/arizon-dread
COPY . .

RUN apk update && apk add --no-cache git
RUN go build -v -o /usr/local/bin/secret-syncer/ ./...


FROM docker.io/alpine:3.23 AS final
WORKDIR /go/bin
ARG VERSION
ENV GENERAL_VERSION=${VERSION}
#RUN apk add --no-cache libc6-compat musl-dev
COPY --from=build /usr/local/bin/secret-syncer/ /go/bin/
EXPOSE 8080
CMD [ "./secret-syncer" ]
