FROM centos:latest
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install golang fuse fuse-devel git make && yum clean all

RUN mkdir -p /vendor/src
ENV GOPATH /vendor

COPY . /eltonfs
WORKDIR /eltonfs
RUN make eltonfs

WORKDIR /eltonfs/bin
RUN mkdir upper lower mountpoint
#RUN ./eltonfs elton1:12345 --upperdir=upper --lowerdir=lower --host=eltonfs mountpoint & 
