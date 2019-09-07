# 実装の手引き


## 利用すべき環境変数

* MYSQL_HOSTNAME
  * MySQLサーバのホスト名
* MYSQL_DATABASE
  * MySQLサーバのデータベース名
* MYSQL_USER
  * MySQLサーバへの接続ユーザー名
* MYSQL_PASSWORD
  * MySQLサーバへの接続パスワード
* PAYMENT_URL
  * 決済代行サービスURL

## fixtureのインポート方法
CREATE DATABSE isutrainしてからwebapp/sql/ 以下の.sqlファイルを順にmysqlコマンドで取り込んでいくことでfixtureを投入できる

sudo mysql < webapp/sql/01_schema.sql
sudo mysql < webapp/sql/90_train.sql
sudo mysql < webapp/sql/91_station.sql
sudo mysql < webapp/sql/92_fare.sql
sudo mysql < webapp/sql/99_fixture.sql
