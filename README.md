docker run --name elton_mysql -e MYSQL_ROOT_PASSWORD=mysql -p 3306:3306 -d mysql:5.6
mysql -uroot -h127.0.0.1 -p -e "create database elton_test"
mysql -uroot -hlocalhost -p elton_test < examples/schema.sql
docker run --name elton_backup -d elton/server backup
docker run --name elton_server --link elton_mysql:db --link elton_backup:backup -d elton/server
