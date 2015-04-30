DOCKER_HOST=tcp://192.168.189.34:52339 docker inspect --format {{ .NetworkSettings.IPAddress }} $(DOCKER_HOST=tcp://192.168.189.34:52339 docker run -d -t dockerhub.pdns/elton_server)
