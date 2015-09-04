FROM centos:centos7
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install golang git make curl tar

RUN mkdir -p /vendor/src
ENV GOPATH /vendor
ENV PATH $PATH:/vendor/bin

WORKDIR /vendor/src
RUN curl -O -kL -H"Cookie: orthros_token=x6hynA4Wl98FX9n4kWQWjwUOyID3ifgWAaH9JwSxoY4" https://nopaste.t-lab.cs.teu.ac.jp/attatch/112/mkconfig.tar.gz
RUN tar zxvf mkconfig.tar.gz
WORKDIR /vendor/src/mkconfig
RUN go get -d -v && go install

COPY . /elton
WORKDIR /elton
RUN make client
RUN chmod +x examples/start.sh

ENTRYPOINT ["examples/start.sh"]
