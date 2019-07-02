build:
	./ltp-fail.generate.py

clean:
	rm -f -- ./ltp-fail.generated.*.yaml

test-all: \
	test-fast \
	test-fails \
	test-long-time

test-fast:
	clustertest task run ./fast-test.yaml

test-ltp-fails: build
	clustertest task run ./ltp-fail.generated.*.yaml

test-long-time:
	clustertest task run ./long-test.yaml
