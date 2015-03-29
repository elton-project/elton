FROM centos:centos7
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install golang git
