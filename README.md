sudo docker run --name elton_mysql -e MYSQL_ROOT_PASSWORD=mysql -p 3306:3306 -d mysql:5.6
sleep 10
mysql -uroot -h127.0.0.1 -p -e "create database elton_test"
mysql -uroot -h127.0.0.1 -p elton_test < examples/schema.sql
sudo docker run --name elton_backup -p 56789:56789 -d elton/server backup
sudo docker run --name elton_server --link elton_mysql:db --link elton_backup:backup -p 52339:52339 -d elton/server
