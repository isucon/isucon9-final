# PaymentAPI

* 1. install protoeasy -> https://christina04.hatenablog.com/entry/2017/11/12/060726
* 2. write .proto file(add/del/mod API endpoint)
* 3. compile .proto file
example : `protoeasy --go --go-import-path=github.com/chibiegg/isucon9-final/blackbox/payment/pb --grpc --grpc-gateway ./`
* 4. implement server-side code according to `pb.go` and `pb.gw.go` file.
* 5. go build
* 6. run

build
```
make build
```

test(needs root privilege)
```
make test
```
