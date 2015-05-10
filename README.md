docker inspect --format {{ .NetworkSettings.IPAddress }} $(docker run -d -t dockerhub.pdns/elton_server)
docker run --name elton_mysql -e MYSQL_ROOT_PASSWORD=mysql -d -p 13306:3306 mysql:5.6
