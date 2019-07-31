# ISUCON9 本戦マニュアル

## はじめに

### スケジュール

TBD

### ISUCON9 本戦ポータルサイト

本選は以下のポータルサイトからベンチマーク走行のリクエスト・結果チェックを行って進行します。
参加登録したGitHubアカウントを利用し、事前にログインを行ってください。

*このページは18:00を過ぎると即座に閲覧不可能になります。ご注意ください。*

https://portal.isucon.net/

ポータルサイトでは、ベンチマーカーが負荷をかける対象となるサーバーを1台選択することができます。 後述する競技後の追試でもこの設定を利用します。

*追試が実行できない場合は失格になるので、競技終了までにこの情報が正しいことを必ず確認してください。*

ポータルサイトでは、ベンチマーク走行の処理状況も確認できます。 ベンチマーク走行が待機中もしくは実行中の間はリクエストは追加できません。


## 作業開始

### 1. サーバへのログイン


### 2. アプリケーションの動作確認


### 3. 負荷走行


### 参照実装の切り替え方法

`/etc/systemd/system/isucon.service` の中の以下の行の、 `go` の部分を各言語 (`perl,ruby,python,php`) に修正します。

```
ExecStartPre = /usr/local/bin/docker-compose -f docker-compose.yml -f docker-compose.go.yml build
ExecStart = /usr/local/bin/docker-compose -f docker-compose.yml -f docker-compose.go.yml up
ExecStop = /usr/local/bin/docker-compose -f docker-compose.yml -f docker-compose.go.yml down
```

例) Perl に変更する場合。

```
ExecStartPre = /usr/local/bin/docker-compose -f docker-compose.yml -f docker-compose.perl.yml build
ExecStart = /usr/local/bin/docker-compose -f docker-compose.yml -f docker-compose.perl.yml up
ExecStop = /usr/local/bin/docker-compose -f docker-compose.yml -f docker-compose.perl.yml down
```

その後、再起動を行います。

```console
$ sudo systemctl daemon-reload
$ sudo systemctl restart isucon
```

### リカバリ方法

参照実装で起動している MySQL のデータは Docker の local volume (webapp_mysql) に保存されています。Docker の外に出すなどの場合は mysqldump 等でデータを dump して、それを使用するのがよいでしょう。

```console
$ mysqldump -uroot -proot --host 127.0.0.1 --port 13306 isucon > isucon.dump
```

配布される4台のサーバすべてに同じデータは存在しますが、競技開始時に mysqldump 実行してバックアップしておくことを強くお勧めします。
競技中にデータをロストしても、運営からの救済は基本的に行いません。


## アプリケーションについて

### ストーリー

TBD
