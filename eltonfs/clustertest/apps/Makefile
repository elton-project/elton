test: \
	test-tar \
	test-git \
	test-nginx \
	test-mysql

test-tar:
	clustertest task run ./tar-*.yaml

test-git:
	clustertest task run ./git-*.yaml

test-nginx:
	clustertest task run ./nginx.yaml

test-mysql:
	clustertest task run ./mysql.yaml
