FROM centos:latest
MAINTAINER Taku MIZUNO <dev@nashio-lab.info>

RUN yum -y upgrade && yum -y install epel-release cronie httpd && yum -y install munin munin-node && yum clean all
RUN sed -i 's/graph_strategy cron/graph_strategy cgi/g' /etc/munin/munin.conf
ADD munin.conf /etc/httpd/conf.d/munin.conf
ADD munin.cron /etc/cron.d/munin
ADD run.sh /usr/local/bin/run
RUN chmod +x /usr/local/bin/run
CMD ["/usr/local/bin/run"]
