package main

import (
	"album-grpc/pb"
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
)

var (
	timeoutDuration = 10 * time.Second
	serverAddr      = "localhost:50051"
)

func callGetAlbum(client pb.AlbumServiceClient, title string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	album, err := client.GetAlbum(ctx, &pb.GetAlbumRequest{Title: title})
	if err != nil {
		log.Fatalf("client.GetAlbum failed: %v", err)
	}

	log.Println(album)
}

func callListAlbums(client pb.AlbumServiceClient, artist string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	stream, err := client.ListAlbums(ctx, &pb.ListAlbumsRequest{Artist: artist})
	if err != nil {
		log.Fatalf("client.ListAlbums failed: %v", err)
	}

	for {
		album, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.ListAlbums failed: %v", err)
		}

		log.Println(album)
	}
}

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
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("client.GetTotalAmount: stream.CloseAndRecv failed: %v", err)
	}

	log.Println(resp)
}

func main() {
	conn, err := grpc.NewClient(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewAlbumServiceClient(conn)

	callGetAlbum(client, "Blue Train")
	callGetAlbum(client, "Not Exist Title")

	callListAlbums(client, "Miles Davis")

	callGetTotalAmount(client)
}
