#FROM golang:1.21-alpine AS build
FROM golang:latest AS build

# RUN apt update
# RUN apt install wget

# RUN update-ca-certificates

RUN mkdir -p /go/src/bible_api
WORKDIR /go/src/bible_api

COPY . /go/src/bible_api

ENV CGO_ENABLED=1
ENV GOOS=linux

RUN go build -o /bible_api ./cmd/bible_api.go
RUN /bible_api -createDB

################
# WHEN IN PROD #
################
# FROM scratch
WORKDIR /
EXPOSE 8000
# COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# COPY --from=build /go/src/bible_api/data/kjv.db /kjv.db
# COPY --from=build /go/src/bible_api /bible_api
CMD ["./bible_api", "-dbPath", "/tmp/kjv.db"]

################
# WHEN TESTING #
################
# docker run -it -d --mount type=bind,source="$(pwd)",target=/app bible_api

# RUN alias ll='ls -la --color=yes'
# WORKDIR /
# EXPOSE 8000
# CMD ["tail", "-F", "/dev/null"]