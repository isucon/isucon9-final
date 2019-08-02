package endpoint

type Endpoint struct {
	HTTPMethod string
	URI        string
}


// TODO: ベンチマーカーにEndpoint一覧を差し込めるようにする
// モックの際にも、それでアクセスしてもらう