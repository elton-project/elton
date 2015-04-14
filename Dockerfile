FROM centos:centos7
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install golang git make

COPY . /elton
WORKDIR /elton

RUN mkdir -p /vendor
ENV GOPATH /elton/vendor

RUN make
