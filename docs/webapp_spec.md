# ウェブアプリケーション API仕様書

## 競技関係
### `POST /initialize`

- DBの次のテーブルをTRUNCATEします。
  - `seat_reservations`
  - `reservations`
  - `users`

### `GET /api/settings`

- 支払いAPIの情報を取得するためのAPIです

## 予約関連
### `GET /api/stations`

- DBの `station_master` (駅マスタ) から駅一覧を返します。

- サンプルリクエスト
  - `GET /api/stations`

### `GET /api/train/search`

- 列車の検索APIです。
  - 日時・乗車駅・降車駅で検索すると、料金・空席情報・発駅と着駅の到着時刻を返します。
    - 日時の表現は `ISO8601` 形式です
    - 指定された時刻以降に発車する列車を検索し、10件返します。
    - 本APIのレスポンスは、特定の列車の予約や、詳細な座席検索に有用です。

- サンプルリクエスト
  - `GET /api/train/search?use_at=2019-12-31T21:00:00.000Z&from=東京&to=大阪&adult=1&child=0`

### `GET /api/train/seats`

- 指定した列車の詳細な空き座席を列挙するAPIです。
  - 日時・列車クラス・列車名・号車・乗車駅・降車駅で検索すると、座席の行・列・予約クラス(自由席・指定席・プレミアム席)・喫煙席付近の有無・予約状況の有無を返します。

- サンプルリクエスト
  - `GET /api/train/seats?date=2019-12-31T15:00:00.000Z&from=東京&to=東京&train_class=最速&train_name=1&car_number=4`

### `POST /api/train/reserve`

- 列車の仮予約を行うAPIです。
  - 仮予約すると座席が確保されます。料金が算出され、DBに未払いとして登録されます。
  - 未払いでも座席は確保されるため、キャンセルされない限り他の予約で再度同じ座席を予約することはできません。
  - リクエストの内容を変えることで、座席を指定しない場合 `あいまい予約モード` となり、予約人数に応じて適当な座席が選択されます。
  - あいまい予約は、号車内に希望の席数が見つからないとエラーとなり、座席は予約されません。
  - リクエストの内容と、DBのマスタ登録されている情報に差異がある (指定席座席なのにプレミアム座席に相当する座席を予約しようとした等の) 場合は、エラーを返し座席は予約されません。
  - 座席確保はログインユーザに紐づく処理を行うため、ログイン・認証を経ないセッション非保持状態ではユーザ識別ができず予約されません。
  - 予約確定のレスポンスに `予約ID` が含まれており、予約IDは支払いに必要となります。

- サンプルリクエスト
  - 遅いやつ10号、8号車、芋呉川→葉千、プレミアム座席で大人2人、子供1人の計3席をあいまい予約するリクエスト
  - ```
    {
        "date": "2020-01-06T10:33:57+09:00",
        "train_name": "10",
        "train_class": "遅いやつ",
        "car_number": 8,
        "is_smoking_seat": false,
        "seat_class": "premium",
        "departure": "芋呉川",
        "arrival": "葉千",
        "adult": 2,
        "child": 1,
        "column": "",
        "seats": []
		}
    ```
  - 遅いやつ10号、8号車、芋呉川→葉千、プレミアム座席で大人2人、子供1人の計3席、窓側席(A)を必ず1席は含むあいまい予約するリクエスト
  - ```
    {
        "date": "2020-01-06T10:33:57+09:00",
        "train_name": "10",
        "train_class": "遅いやつ",
        "car_number": 8,
        "is_smoking_seat": false,
        "seat_class": "premium",
        "departure": "芋呉川",
        "arrival": "葉千",
        "adult": 2,
        "child": 1,
        "column": "A", // 窓側・真ん中・通路側希望により変わる
        "seats": []
		}
    ```
  - 遅いやつ10号、8号車、芋呉川→葉千、プレミアム座席で大人、子供でそれぞれ2番A席と2番B席を予約するリクエスト
  - ```
    {
        "date": "2020-01-06T10:33:57+09:00",
        "train_name": "10",
        "train_class": "遅いやつ",
        "car_number": 8,
        "is_smoking_seat": false,
        "seat_class": "premium",
        "departure": "芋呉川",
        "arrival": "葉千",
        "child": 1,
        "adult": 1,
        "column": "",
        "seats": [{
                "row": 2,
                "column": "A"
            },
            {
                "row": 2,
                "column": "B"
            }
        ]
    }
    ```

### `POST /api/train/reservation/commit`

- 仮予約に支払いを行い、確定を行うAPIです。
  - カードトークンと予約IDを渡すと支払いが確定します。
  - カードトークンは、別途 `payment_spec.md` 中のカードトークン発行により入手してください。
  - 支払い確定のレスポンスは成功or失敗のみを返します。

- サンプルリクエスト
  - 予約ID1番、支払いAPIへカード登録時に発行されたトークンで支払いを行うリクエスト
  - ```
    {
		"card_token": "161b2f8f-791b-4798-42a5-ca95339b852b",
		"reservation_id": "1"
	}
    ```

## 認証関連
### `GET /api/auth`

- ログイン中のユーザに関連する情報を返すAPIです。

### `POST /api/auth/signup`

- ユーザ登録を行うAPIです。

### `POST /api/auth/login`

- ログインを行うAPIです。セッションが発行されます。

### `POST /api/auth/logout`

- ログアウトを行うAPIです。セッションが削除されます。

### `GET /api/user/reservations`

- ログイン中のユーザが登録した予約一覧を返します。

### `GET /api/user/reservations/:item_id`

- ログイン中のユーザが登録した特定の予約の詳細な情報を返します。

### `POST /api/user/reservations/:item_id/cancel`

- ログイン中のユーザが登録した特定の予約をキャンセルします。
  - キャンセルには仮予約APIで発行された `予約ID` が必要です。
