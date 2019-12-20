USER=root
KEY=~/.ssh/id_rsa
deploy:
	ansible-playbook -i hosts -u $(USER) -c paramiko -D playbook.yml --private-key=$(KEY)
.PHONY: deploy

print:
	@echo ansible-playbook -i hosts -u $(USER) -c paramiko -D playbook.yml --private-key=$(KEY)
.PHONY: print