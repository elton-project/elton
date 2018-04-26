FROM centos:latest
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install golang git make tar curl gcc-c++ && yum clean all

RUN mkdir -p /vendor/src
ENV GOPATH /vendor
ENV PATH $PATH:$GOPATH/bin
RUN go get github.com/kr/godep

RUN curl -kL -O https://github.com/google/protobuf/releases/download/v3.0.0-beta-2/protobuf-cpp-3.0.0-beta-2.tar.gz
RUN tar zxvf protobuf-cpp-3.0.0-beta-2.tar.gz
RUN cd protobuf-3.0.0-beta-2 && ./configure && make && make install
RUN rm -rf protobuf*

COPY . /vendor/src/gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton
WORKDIR /vendor/src/gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton
RUN make elton

RUN chmod +x examples/start.sh
