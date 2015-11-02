FROM centos:latest
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install golang git make tar curl gcc-c++ && yum clean all

RUN mkdir -p /vendor/src
ENV GOPATH /vendor
ENV PATH $PATH:$GOPATH/bin
RUN go get github.com/kr/godep

RUN curl -kL -O https://github.com/google/protobuf/releases/download/v3.0.0-beta-1/protobuf-cpp-3.0.0-beta-1.tar.gz
RUN tar zxvf protobuf-cpp-3.0.0-beta-1.tar.gz
RUN cd protobuf-3.0.0-beta-1 && ./configure && make && make install
RUN rm -rf protobuf*

RUN go get github.com/golang/protobuf/proto
RUN go get github.com/golang/protobuf/protoc-gen-go

COPY . /elton
WORKDIR /elton/grpc/proto
RUN protoc --go_out=plugins=grpc:. elton_service.proto

WORKDIR /elton
CMD make
