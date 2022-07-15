FROM ubuntu:20.04

COPY . ./
RUN echo "hi there"
RUN apt update
RUN apt install -y curl tar gcc
RUN curl -L -o  /tmp/go1.18.4.linux-amd64.tar.gz  https://go.dev/dl/go1.18.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf /tmp/go1.18.4.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

RUN go version
RUN go build -o ~/bible_api cmd/deploy.go
ENTRYPOINT ["./docker/entry.sh"]