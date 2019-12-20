# Build, Deploy
## Requirement
* Go 1.13.5
* Ansible 2.9.2
* Docker version 19.03
* make

## build
1. `make all` で `frontend, webapp, payment, bench` を全部ビルドできる
1. ビルドに必要なものは make, docker, go1.13 あたりっぽい
1. ビルドしたソースは ansible 配下に置かれる

## deploy
1. ビルドした成果物を含め競技用サーバに設置する
1. `ansible/hosts` を編集する
1. `make -C isucon9-final/ansible deploy USER={user} KEY={ssh-key}`
