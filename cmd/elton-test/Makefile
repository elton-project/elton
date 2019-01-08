
start-controller:
	go build -o /tmp/elton-test-c .
	ELTON_CONTROLLER_LISTEN_ADDR=0.0.0.0:19500 /tmp/elton-test-c; echo exit: $?

start-node:
	go build -o /tmp/elton-test-s .
	ELTON_CONTROLLERS=127.0.0.1:19500 /tmp/elton-test-s; echo exit: $?
