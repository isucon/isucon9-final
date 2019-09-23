# 移植作業マニュアル


## 新規言語の実装開始

### ディレクトリを作成し、言語固有のdocker-composeファイルを作成する。

```bash
mkdir webapp/${LANGUAGE}
cp webapp/docker-compose.go.yml webapp/docker-compose.${LANGUAGE}.yml
```

### Dockerfileを作成する

```bash
${EDITOR} webapp/${LANGUAGE}/Dockerfile
```

## 実装時の注意点

### 環境変数の利用

データベースの接続情報や、外部APIのURLは必ず [.env](../webapp/.env) ファイルに存在する環境変数を使ってください。

#### 利用すべき環境変数

* MYSQL_HOSTNAME
  * MySQLサーバのホスト名
* MYSQL_DATABASE
  * MySQLサーバのデータベース名
* MYSQL_USER
  * MySQLサーバへの接続ユーザー名
* MYSQL_PASSWORD
  * MySQLサーバへの接続パスワード
* PAYMENT_API
  * 決済代行サービスURL


### データベースのスキーム

データベースのスキームは変更しないでください。


#### fixtureのインポート方法

docker-composeを使わない場合、 `CREATE DATABSE isutrain` してから `webapp/sql/` 以下の.sqlファイルを順にmysqlコマンドで取り込んでいくことでfixtureを投入できる

```bash
sudo mysql < webapp/sql/01_schema.sql
sudo mysql < webapp/sql/90_train.sql
sudo mysql < webapp/sql/91_station.sql
sudo mysql < webapp/sql/92_fare.sql
sudo mysql < webapp/sql/99_fixture.sql
```
