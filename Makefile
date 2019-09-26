.PHONY: frontend webapp payment bench

all: frontend webapp payment bench

frontend:
	cd webapp/frontend && make
	cd webapp/frontend/dist && tar zcvf ../../../ansible/files/frontend.tar.gz .

webapp:
	tar zcvf ansible/files/webapp.tar.gz \
	--exclude webapp/frontend \
	webapp

payment:
	cd blackbox/payment && make && cp bin/payment_linux ../../ansible/roles/benchmark/files/payment

bench:
	cd bench && make && cp -av bin/bench_linux ../ansible/roles/benchmark/files/bench && cp -av bin/benchworker_linux ../ansible/roles/benchmark/files/benchworker
