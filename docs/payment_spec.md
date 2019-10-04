## payment service

決済サービスAPI。クレジットカード情報の非保持化にも対応しているので安心して利用できます。
### `POST /card`

* カード情報(番号/Cvv/有効期限)を送るとクレジットカード番号の代わりに使えるトークンが発行されます。
* それぞれの形式は以下の通りです。
    *  card_number: `[0-9]{8}`
    *  cvv: `[0-9]{3}`
    *  expiry_date: `[0-9]{2}/[0-9]{2}`
*  有効期限が実際に本戦開催月(2019/10)より前のものだとエラーになります。

#### API仕様

- request: application/json
  - card_information
    - card_number
    - cvv
    - expiry_date
- response: application/json
  - http status code: 200
    - card_token
    - is_ok
  - http status code: 400
    - error: invalid card information
  - http status code: 500
    - error: token generate error

```
example:

# request
{
	"card_information": {
		"card_number":"11111111",
		"cvv": "111",
      	"expiry_date": "11/22"
	}
}

# response
{
"card_token": "f042a6e3-a7cf-4511-5f96-694ea9b177eb",
"is_ok": true
}

{
"error": "Invalid CardNumber Length",
"message": "Invalid CardNumber Length",
"code": 3,
"details": [],
}
```

### `POST /payment`

* トークン・予約ID・金額を送ると決済登録されます。
* トークンが間違っているとエラーになります。
* 決済されると決済IDが発行されます。決済後のキャンセルは決済IDが必要になります。

#### API仕様

- request: application/json
  - payment_information
    - card_token
    - reservation_id
    - amount
- response: application/json
  - http status code: 200
    - payment_id
    - is_ok
  - http status code: 404
    - error: card token not found

```
example:

# request
{
	"payment_information": {
		"card_token": "0faa90fc-61a7-47ed-685c-805a4527e831",
		"reservation_id": 123,
		"amount": 12345
	}
}

# response
{
"payment_id": "bm83su1f8ltcqscrcdk0",
"is_ok": true
}

{
"error": "Card_Token Not Found",
"message": "Card_Token Not Found",
"code": 5,
"details": [],
}
```

### `DELETE /payment/:payment_id`

* 決済IDを送るとキャンセル処理されます。
* 決済IDが間違っているとエラーになります。

#### API仕様

- request: URI
- response: application/json
  - http status code: 200
    - is_ok
  - http status code: 404
    - error: card token not found
```
example:

# request
curl -X DELETE http://localhost:5000/payment/bm83su1f8ltcqscrcdk0

# response
{
"is_ok": true
}

{
"error": "PaymentID Not Found",
"message": "PaymentID Not Found",
"code": 5,
"details": [],
}
```

### `POST /payment/_bulk`

* 決済IDを配列で送るとまとめてキャンセル処理されます。
* 配列の途中に誤った決済IDがあると無視し、正しい決済IDのみキャンセル処理します。
* リクエストが成功すると、キャンセルした決済IDの数を返します。
* エラーはありません。

#### API仕様

- request: URI
- response: application/json
  - http status code: 200
    - deleted
```
example:

# request
{
	"payment_id": [
		"bm849shf8ltcqmi2qc8g",
		"bm84afhf8ltcqmi2qc90"
	]
}

# response
{
"deleted": 2
}
```