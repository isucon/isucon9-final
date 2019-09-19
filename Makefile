.PHONY: archive

all: archive

archive:
	tar zcvf ansible/roles/challenge/files/webapp.tar.gz webapp
