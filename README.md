# ISUCON9 本戦問題

## ドキュメント


* [当日の参加者向けマニュアル](docs/MANUAL.md)
* [移植作業用マニュアル](docs/IMPLEMENT.md)


## スペシャルサンクス

### 各言語移植

時間の無い中、移植にご協力いただき、誠にありがとうございます。

* @ykzts 氏 (Ruby)
* @kazeburo 氏 (Perl)
* @shoma 氏 (PHP)

(GolangとPythonはさくらインターネット実装です。)


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
export LANGUAGE=go
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml build
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml up
```

http://127.0.0.1:8080

## ローカルでのベンチマーカーの起動

### 必要な環境

- [Golang (1.13)](https://golang.org/dl/)

### ベンチマーカーの起動方法

リポジトリをCloneして、ビルドを行う

```bash
git clone git@github.com:chibiegg/isucon9-final.git
cd isucon9-final
(cd bench && make)
```

ビルドされたバイナリは bench/bin/ 配下に配置されるため、ベンチマーカーのバイナリだけ起動する (bin/bench~~~)

(Mac の場合は バイナリ名のサフィックスに `_darwin` が、Linuxの場合はバイナリ名のサフィックスに `_linux` がつきます)

```bash
bench/bin/bench_darwin run --payment=<課金APIのアドレス> --target=<webappのアドレス> --assetdir=<フロントエンドのビルド結果が配置されたディレクトリ>
```

* assetdirについて補足
  * assetdirは、ベンチマーカーがアプリケーションが正常動作しているかテストする際に用いられます
  * 当問題ではフロントエンドがVue.jsで実装されているため、webapp/frontend/dist を指定することになります

* 出力が大変多いので、詳しく調査したい場合は teeコマンドなどでログに書き出すことをお勧めいたします


## デプロイ

```bash
make archive
cd ansible
ansible-playbook -i hosts -u root -s -c paramiko -D playbook.yml
```

## 既知の問題

コンテスト開催中にも残っており、準備中に発覚しなかった問題。

* 座席情報が誤っている #209

### 修正済み

* Ruby実装で、データベースの環境変数参照が誤っていた #211
