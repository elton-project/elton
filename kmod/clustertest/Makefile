build:

generate: generate-ltp-fails

generate-ltp-fails:
	./ltp-fail.generate.py

clean:
	rm -f -- ./ltp-fail.generated.*.yaml

test-all: \
	test-fast \
	test-fails \
	test-apps \
	test-long-time

test-fast:
	clustertest task run ./fast-test.yaml

test-ltp-fails: generate-ltp-fails
	clustertest task run ./ltp-fail.generated.*.yaml

test-apps:
	cd apps && $(MAKE) test

test-long-time:
	clustertest task run ./long-test.yaml
