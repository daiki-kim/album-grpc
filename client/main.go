package main

import (
	"album-grpc/pb"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
)

var (
	serverAddr = "localhost:50051"

	timeoutDuration = 10 * time.Second
	timeSleep       = 0 * time.Second
)

// Unary RPCの実装例
// サーバーにtitleを送り、ファイルに存在するかの確認結果をAlbum型で受け取る
func callGetAlbum(client pb.AlbumServiceClient, title string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	resp, err := client.GetAlbum(ctx, &pb.GetAlbumRequest{Title: title})
	if err != nil {
		log.Fatalf("client.GetAlbum failed: %v", err)
	}

	log.Printf("response: %v", resp.Album)
}

// ServerStreaming RPCの実装例
// サーバーにartistを送り、artistが一致するAlbumをすべてAlbum型で受け取る
func callListAlbums(client pb.AlbumServiceClient, artist string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	stream, err := client.ListAlbums(ctx, &pb.ListAlbumsRequest{Artist: artist})
	if err != nil {
		log.Fatalf("client.ListAlbums failed: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.ListAlbums failed: %v", err)
		}

		log.Printf("response: %v", resp.Album)
	}
}

// ClientStreaming RPCの実装例
// サーバーに複数のtitleを送り、ファイルに存在したAlbumの総数とその合計金額、メッセージを受け取る
func callGetTotalAmount(client pb.AlbumServiceClient) {
	titles := []string{
		"Blue Train",
		"Giant Steps",
		"Speak No Evil",
		"Weather Report",
		"A Portrait in Jazz",
		"Chet Baker Sings",
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	stream, err := client.GetTotalAmount(ctx)
	if err != nil {
		log.Fatalf("client.GetTotalAmount failed: %v", err)
	}

	for _, title := range titles {
		if err := stream.Send(&pb.GetTotalAmountRequest{Title: title}); err != nil {
			log.Fatalf("client.GetTotalAmount: stream.Send(%s) failed: %v", title, err)
		}

		time.Sleep(timeSleep) // 動作確認用
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("client.GetTotalAmount: stream.CloseAndRecv failed: %v", err)
	}

	log.Printf("response: %v", resp)
}

// BidirectionalStreaming RPCの実装例
// サーバーにアップロードしたいAlbumを連続で送り、その度にサーバーからメッセージを受け取る
func callUploadAndNotify(client pb.AlbumServiceClient) {
	albums := []*pb.Album{
		{Title: "New Album", Artist: "New Artist", Price: 10.99},
		{Title: "New Album 2", Artist: "New Artist 2", Price: 20.99},
		{Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
		{Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	stream, err := client.UploadAndNotify(ctx)
	if err != nil {
		log.Fatalf("client.UploadAndNotify failed: %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			recp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("client.UploadAndNotify: stream.Recv() failed: %v", err)
			}

			log.Printf("response: %v", recp)
		}
	}()

	for _, album := range albums {
		req := &pb.UploadAndNotifyRequest{Album: album}
		if err := stream.Send(req); err != nil {
			log.Fatalf("client.UploadAndNotify: stream.Send(%v) failed: %v", album, err)
		}

		time.Sleep(timeSleep) // 動作確認用
	}

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("client.UploadAndNotify: stream.CloseSend() failed: %v", err)
	}
	<-waitc
}

func main() {
	conn, err := grpc.NewClient(serverAddr, grpc.WithInsecure()) // gRPCクライアントを作成
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewAlbumServiceClient(conn) // gRPCクライアントをAlbumServiceClientに変換

	// Unary RPCの実行例
	fmt.Println("Unary RPC started")
	fmt.Println("valid request")
	callGetAlbum(client, "Blue Train")
	fmt.Println("invalid request")
	callGetAlbum(client, "Not Exist Title")
	fmt.Println("")

	// ServerStreaming RPCの実行例
	fmt.Println("ServerStreaming RPC started")
	callListAlbums(client, "Miles Davis")
	fmt.Println("")

	// ClientStreaming RPCの実行例
	fmt.Println("ClientStreaming RPC started")
	callGetTotalAmount(client)
	fmt.Println("")

	// BidirectionalStreaming RPCの実行例
	fmt.Println("BidirectionalStreaming RPC started")
	callUploadAndNotify(client)
	fmt.Println("")
}
