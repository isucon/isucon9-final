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

8000番ポートでアプリケーションが動作するようにコンテナイメージを作成してください。

また、ソースコードは実行時にVolumeでマウントするようにし、イメージにはソースコードを含めないでください。

## 起動と動作確認

docker-composeでコンテナを起動すると、webapp、MySQL、Nginx、外部決済サービスが起動します。

```bash
export LANGUAGE=go
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml build
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml up
```

コードの変更時にwebappが自動で再起動しない場合には以下のコマンドでwebappコンテナだけ再起動することができます。

```bash
export LANGUAGE=go
docker-compose -f webapp/docker-compose.yml -f webapp/docker-compose.${LANGUAGE}.yml restart webapp
```

起動すると、 http://localhost:8080 でフロントエンドとwebapp両方にアクセスできます。

## ベンチマークの実行

### ベンチマーカーのディレクトリに移動し、バイナリをビルドする

```bash
cd bench
make
```

### ベンチマーカーが必要とする情報を集める

以下に、案内通りの起動を行なった場合の情報を記載します

* 課金のアドレス
  * http://localhost:5000
* webappのアドレス
  * http://localhost:8080
* 静的ファイルの配置ディレクトリ
  * webapp/frontend/dist

### 初期状態のwebappはあまりにも遅いので、インデックスを貼る

これをしないと、ベンチマーカーが即失格判定を出してしまい、ちゃんとした検証ができません

まずは、コンテナのシェルを立ち上げ、mysqlログインします

```bash
$ docker exec -it webapp_mysql_1 mysql -uroot -ppassword isutrain
```

以下のようにインデックスを貼ります (数分くらいで終わります)

```bash
mysql> create index train_timetable_master01 ON train_timetable_master (date, train_class, train_name, station);
```

### ベンチマーカーを起動する

```bash
bench/bin/bench_darwin run --payment=http://localhost:5000 --target=http://localhost:8080 --assetdir=webapp/frontend/dist
```

## 実装時の注意点

### 環境変数の利用

データベースの接続情報や、外部APIのURLは必ず [.env](../webapp/.env) ファイルに存在する環境変数を使ってください。

環境変数が設定されていない場合のデフォルト値はGolangの参考実装を元にしてください。

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


PAYMENT_APIは環境変数が入っていない場合、webappからのリクエストは http://payment:5000 へ投げ、　`/settings` で応答するコンテンツは `http://localhost:5000` を返してください。



### データベースのスキーム

データベースのスキームは変更しないでください。


#### fixtureのインポート方法

docker-composeを利用する場合、初期データが投入された状態で起動します。

docker-composeを使わない場合、 `CREATE DATABSE isutrain` してから `webapp/sql/` 以下の.sqlファイルを順にmysqlコマンドで取り込んでいくことでfixtureを投入できる

```bash
sudo mysql < webapp/sql/01_schema.sql
sudo mysql < webapp/sql/90_train.sql
sudo mysql < webapp/sql/91_station.sql
sudo mysql < webapp/sql/92_fare.sql
sudo mysql < webapp/sql/93_seat.sql
sudo mysql < webapp/sql/94_0_train_timetable.sql
sudo mysql < webapp/sql/94_1_train_timetable.sql
sudo mysql < webapp/sql/94_2_train_timetable.sql
sudo mysql < webapp/sql/94_3_train_timetable.sql
sudo mysql < webapp/sql/94_4_train_timetable.sql
sudo mysql < webapp/sql/94_5_train_timetable.sql
sudo mysql < webapp/sql/99_fixture.sql
```
