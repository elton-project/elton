FROM centos:latest
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install epel-release && yum -y install golang fuse fuse-devel git make tar curl gcc-c++ && yum clean all

RUN mkdir -p /vendor/src
ENV GOPATH /vendor

RUN curl -kL -O https://github.com/google/protobuf/releases/download/v3.0.0-beta-1/protobuf-cpp-3.0.0-beta-1.tar.gz
RUN tar zxvf protobuf-cpp-3.0.0-beta-1.tar.gz
RUN cd protobuf-3.0.0-beta-1 && ./configure && make && make install
RUN rm -rf protobuf*

ADD eltonfs/run.sh /usr/local/bin/run
RUN chmod +x /usr/local/bin/run

COPY . /eltonfs
WORKDIR /eltonfs
RUN make eltonfs

WORKDIR /eltonfs/bin
RUN mkdir upper lower mountpoint

CMD ["/usr/local/bin/run"]
