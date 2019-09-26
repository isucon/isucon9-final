# ベンチマーカー

## ビルド

```
$ make
```


## テスト

```
$ make test
```

## シナリオ開発者向け
### シナリオ作成の流れ

* `scenario/template.go` をコピーする。
    * ファイルをコピーしないでください
    * あくまで、関数をテキストコピーし、適切なファイル(正常テストならnormal.go)に配置してください
* `AwesomeScenario()` 関数の名前を変える
    * シナリオは `normal(正常)` `abnormal(異常)` `attack(攻撃)` などのシナリオ種別があります
    * `func <シナリオ種別><何をするか>Scenario` のように関数名を変更してください
* シナリオの中身をかく
* シナリオのテストを書く
* cmd/bench/benchmarker.go で定義されている load(ctx context.Context) にシナリオ実行する実装をする
* プルリクエストをだす

### シナリオに関するFAQ

* 負荷レベルをあげたい
    * デフォルトでは、ベンチマーカーが１ワークロード（シナリオ複数含む）が終わるごとレベルアップし、レベル数分だけgoroutineを生成します
    * そのほかに、シナリオ内で負荷をあげる選択肢もあり、これには scenario/attack.go のコードなどを参考にしてください

* isutrainやpaymentにリクエストを送りたい
    * isutrain.Client, payment.Clientを用います (NewClient()でClientを生成できます)
    * Clientが提供する関数を用いると、HTTPリクエスト失敗やJSON Unmarshal失敗などのエラーを検知し、そのエラーを返してくれます. シナリオ側で、これをbencherrorに追加する必要があります
    * レスポンスの具体的な中身についてチェックをし、エラーとしたい場合はscenario/assertion.goで提供される関数群を用いて、 bencherrorに適宜エラーを追加してください

* ランダムデータが欲しい
    * xrandom.GetXXX を用いてください
    * なければ定義するようにお願いします
        * DBからそのまま引っこ抜いてきたようなランダムデータはxrandomパッケージ内で閉じて定義されていますが、それ以外のパラメータはconfigパッケージに集められています

* エラーを追加したい
    * bencherrorパッケージを用います
    * エラーを追加することは、「ユーザにそのエラーメッセージを見せる」、「ペナルティとして計上する」ことを意味します
    * エラーは全部で４種類ありますが、シナリオで用いるエラーは２種類しかありません
        * アプリケーションエラー ... webappの挙動が正しくない場合やサービスとして安定していない場合のエラー. 重みと掛け合わせてペナルティ算出されます
        * クリティカルエラー ... このエラーが発生したと判断された場合、即失格となります. ベンチマークは止まりますし、スコアは０になります
    * 書き方は以下のようなものがあります
        * エラーを得て、それにメッセージを付加してbencherrorに追加したい
            * bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "メッセージ: %d", 123))
        * エラーがないが、メッセージのみでbencherrorに追加したい
            * bencherror.BenchmarkErrs.AddError(bencherror.NewSimpleApplicationError("メッセージ: %d", 123))

* ユーザには見せないが、ポータルから確認できるメッセージを書き込みたい
    * デバッグメッセージなどは `必ず` 標準エラー出力に出すようにするべく、zapロガーを使ってください
    * 標準出力に不用意に文字列を書き込んでしまうと、ベンチマーク結果のUnmarshalに失敗し、事故になります
    * log.PrintX や fmt.PrintXを用いていいのは benchworkerのみです. シナリオ定義は benchで行うので、行わないでください

* 予約状況に応じて、座席を選択したい
    * すみません、未実装です
    * internal/cache パッケージ内にて実装中で、これを用いて今予約しようとしている座席が予約可能かどうか判定できるようにしようと考えています
    * 予約できるなら 正常HTTPステータスコードが返るはずだし、そうでないなら異常HTTPステータスコードが返るはずという具合です

## 外観

![bench](https://user-images.githubusercontent.com/7540775/65664038-5dbc6780-e073-11e9-9ba3-5a07913dc880.png)

## パッケージごとの役割

```
├── assets // 静的ファイルをローカルファイルシステムから読み出す
├── bin // make build により、このディレクトリに実行ファイルが生成される
├── cmd
│   ├── bench // ベンチマークコマンド. benchworkerにより実行される
│   │   ├── bench.go
│   │   ├── benchmarker.go // ベンチマーカーの本体. ここでシナリオを追加できます
│   │   ├── benchmarker_test.go
│   │   └── main.go
│   └── benchworker // 常駐benchworker. dequeueしたジョブに応じてベンチを実行し、結果を報告する
│       ├── bench-worker.go
│       ├── main.go
│       ├── portal.go
│       └── util.go
├── internal
│   ├── bencherror // ベンチマーク(Initialize, PreTest, Benchmark, PostTest) に関するエラーを集める
│   ├── cache // ベンチマーク実行中、ベンチマーカーが覚えておかなくてはならない情報を格納し、判定関数を提供する
│   ├── config // 設定情報はconstでここに定義し、バイナリに埋め込む
│   ├── endpoint // ここでエンドポイントが定義され、基本スコア算出関数を提供する
│   ├── logger // zapロガーの定義
│   ├── util // 細かいユーティリティ
│   │   ├── random.go
│   │   ├── string.go
│   │   ├── time.go
│   │   └── url.go
│   └── xrandom // ランダムデータ生成関数を提供する
│       ├── random.go // 関数定義
│       ├── random_data.go // ランダムデータ(DBから引っ張ってきたやつとか)
│       └── random_data_test.go // 簡易テスト
├── isutrain // Isutrainウェブアプリへリクエストを送ったり、結果をUnmarshalしたりする諸々
│   ├── client.go // クライアントの本体定義
│   ├── initialize.go // /initialize のレスポンス定義
│   ├── reservation.go // 予約周りの構造体定義
│   ├── seats.go // 座席周りの構造体定義
│   ├── session.go // クライアントに用いられるセッション(認証状態などを覚えておく)
│   ├── settings.go // /settings のレスポンス定義
│   ├── station.go // 駅周りの構造体定義
│   ├── train.go // 列車周りの構造体定義
│   └── user.go // ユーザの構造体定義
├── mock // テストに用いるモックサーバ. isutrain, paymentの両方利用できる
├── payment // 課金にリクエストを送ったりする諸々
│   ├── client.go
│   └── payment_result.go
├── scenario // シナリオ定義. ここにシナリオファイルを追加していく
│   ├── abnormal.go // 異常テスト (不正なログイン情報でログインしてみるとか)
│   ├── assertion.go // アサーション関数群 (レスポンスのデータが正しいか検証する関数などの定義)
│   ├── attack.go // 攻撃テスト (やたらめったら検索をかけたり、ログイン試行したり)
│   ├── normal.go // 正常テスト (シンプルなシナリオ、ユースケースをある程度網羅するシナリオ)
│   └── template.go // シナリオのテンプレート (これを参考にシナリオを作る)
│   ├── pretest.go // Pretest用のシナリオ. (webapp周りのシナリオ作成では触れることはないです)
│   ├── finalcheck.go // FinalCheck(PostTestに改名予定)用のシナリオ. Pretestと似たような扱い
```