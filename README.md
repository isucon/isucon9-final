# ISUCON9 本戦問題

## ドキュメント


* [当日の参加者向けマニュアル](docs/MANUAL.md)
* [移植作業用マニュアル](docs/IMPLEMENT.md)


## 本番当日の動作環境

TBD

## ローカルでのアプリケーションの起動

### 必要な環境

- [Docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/)
- [Golang](https://golang.org/)

### ウェブアプリケーションの起動方法

リポジトリをCloneし、

```bash
git clone git@github.com:chibiegg/isucon9-final.git
cd isucon9-final
(cd webapp/frontend && make)
```

実装言語を指定してBuildとUpをする。

```bash
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml build
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml up
```

http://127.0.0.1:8080


## デプロイ

```bash
make archive
cd ansible
ansible-playbook -i hosts -u root -s -c paramiko -D playbook.yml
```
