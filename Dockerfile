FROM ubuntu:20.04

RUN apt update && apt install -y wget gcc build-essential && wget https://golang.org/dl/go1.16.6.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.16.6.linux-amd64.tar.gz

COPY . /opt/bible_api

ENV LANG=en_US.UTF-8 PATH=$PATH:/usr/local/go/bin

EXPOSE 8000

RUN cd /opt/bible_api/cmd && go build -o ~/bible_api deploy.go && ~/bible_api -createDB ~/kjv.db
