.PHONY: frontend archive

all: frontend archive

frontend:
	cd webapp/frontend && make

archive:
	tar zcvf ansible/roles/challenge/files/webapp.tar.gz \
	--exclude webapp/frontend \
	webapp

	cd webapp/frontend/dist && tar zcvf ../../../ansible/roles/challenge/files/frontend.tar.gz .
