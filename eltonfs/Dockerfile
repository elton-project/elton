FROM centos:centos7

RUN yum -y upgrade && yum -y install golang fuse fuse-devel git

RUN mkdir -p /vendor/src
ENV GOPATH /vendor
ENV PATH $PATH:/vendor/bin

COPY . /vendor/src/eltonfs
WORKDIR /vendor/src/eltonfs
RUN go get -d -v && go install

RUN mkdir -p /eltonfs
