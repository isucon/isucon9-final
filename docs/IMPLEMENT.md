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

### データベースのスキーム

データベースのスキームは変更しないでください。

### 
