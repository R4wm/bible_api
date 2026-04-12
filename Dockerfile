FROM node:20 AS ui
WORKDIR /ui
COPY web/package.json /ui/
RUN npm install
COPY web /ui
RUN npm run build

#FROM golang:1.21-alpine AS build
FROM golang:latest AS build

# RUN apt update
# RUN apt install wget

# RUN update-ca-certificates

RUN mkdir -p /go/src/bible_api
WORKDIR /go/src/bible_api

COPY go.mod go.sum /go/src/bible_api/
RUN go mod download

COPY . /go/src/bible_api
COPY --from=ui /ui/dist /go/src/bible_api/web/dist

ENV CGO_ENABLED=1
ENV GOOS=linux

RUN go build -o /bible_api ./cmd/bible_api.go

################
# WHEN IN PROD #
################
# FROM scratch
WORKDIR /go/src/bible_api
EXPOSE 8000

# Set default Redis connection to the linked Redis container
ENV REDIS_ADDR=redis:6379
ENV REDIS_PASSWORD=

CMD sh -c '/bible_api -createDB -dbPath /data/kjv.db; /bible_api -dbPath /data/kjv.db'

################
# WHEN TESTING #
################
# docker run -it -d --mount type=bind,source="$(pwd)",target=/app bible_api

# RUN alias ll='ls -la --color=yes'
# WORKDIR /
# EXPOSE 8000
# CMD ["tail", "-F", "/dev/null"]
