# album-grpc

gRPCのクライアントとサーバーを実装し、AlbumデータをgRPCの各通信方式でやりとりする簡単なプロジェクト。

## gRPCの概要

- Googleが開発した高性能なRPC（Remote Procedure Call）フレームワーク。

- プロトコルバッファを使用してインターフェースを定義し、異なる言語間での通信が容易である。

- HTTP/2を使用しており、以下4つの通信方式をサポートしている。

### Unary RPC

クライアントが1つのリクエストを送り、サーバーが1つのレスポンスを返す最も基本的な通信方式。

### Server Streaming RPC

クライアントが1つのリクエストを送り、サーバーが複数のレスポンスをストリーミングで送信する通信方式。クライアントはサーバーがすべてのデータを送り終えるまでレスポンスを受け取る。

### Client Streaming RPC

クライアントが複数のリクエストをストリーミング形式で送信し、サーバーが1つのレスポンスを返す通信方式。サーバーは、クライアントからすべてのデータを受け取った後に処理を行う。

### Bidirectional Streaming RPC

クライアントとサーバーが同時にストリーミングでデータを送受信する通信方式。クライアントとサーバーは独立してデータを送信し続け、任意のタイミングで処理を完了できる。

## ディレクトリ構成

```bash
.
├── client # gRPCクライアントの実装
│   └── main.go
│
├── db # アルバム情報
│   └── album.json
│
├── pb # protocで自動生成されたGoコード
│   ├── album.pb.go
│   └── album_grpc.pb.go
│
├── proto # プロトコルバッファの定義
│   └── album.proto
│
└── server # gRPCサーバーの実装
    └── main.go
```

### client

`client/main.go`にはgRPCクライアントを実装。

- Unary RPC: サーバーにアルバムのタイトルを送り、存在するか確認する。
- Server Streaming RPC: サーバーにアーティスト名を送り、そのアーティストのアルバム一覧を受け取る。
- Client Streaming RPC: サーバーに複数のアルバムタイトルを送り、合計金額とアルバム数を受け取る。
- Bidirectional Streaming RPC: サーバーに複数のアルバム情報を送り、その度にアップロード結果のメッセージを受け取る。

### server

`server/main.go`には、gRPCサーバーを実装。

- Unary RPC: アルバムの存在確認と情報の提供。
- Server Streaming RPC: 指定されたアーティストのアルバム一覧の送信。
- Client Streaming RPC: 受け取ったアルバムタイトルの合計金額とアルバム数の計算。
- Bidirectional Streaming RPC: クライアントから受け取ったアルバム情報の保存と結果通知。

### proto

`proto/album.proto`は、サービスとメッセージの定義を行うプロトコルバッファのファイルで、ここでgRPCのインターフェースを定義している。

### pb

protocコマンドで生成されたGoのコード`album.pb.go`と`album_grpc.pb.go`が格納される。

### db

`db/album.json`にJSON形式のアルバムのデータを保存している。

## クイックスタート

1. サーバーの起動

    ```bash
    go run server/main.go
    ```

2. クライアントの起動

    ```bash
    go run client/main.go
    ```

## その他開発で必要なパッケージやコマンド

(以下[gRPC公式ドキュメント](https://grpc.io/docs/languages/go/quickstart/)より抜粋)

`.proto`ファイルからgRPCのGoコードを生成するには、以下二つをインストールする必要がある。

- Protocol Buffer

    ```bash
    brew install protobuf
    ```

- GoのProtocolコンパイラプラグイン

    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```

`.proto`で定義したgRPCのインターフェースからGoコードを生成するには、以下のコマンドを実行する。

```bash
protoc -I. --go_out=. --go-grpc_out=. proto/*.proto
```
