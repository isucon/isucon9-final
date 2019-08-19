# ISUCON9 本戦問題


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
```

実装言語を指定してBuildとUpをする。

```bash
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml build
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml up
```

フロントエンドが必要な場合は、 `cd frontend && npm run serve` でフロントエンドを起動させる。
