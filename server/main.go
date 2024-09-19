package main

import (
	"album-grpc/pb"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

var (
	filePath = "db/album.json"
	port     = "50051"

	timeSleep = 0 * time.Second
)

// pb.AlbumServiceServerインターフェースを満たすサーバーを定義する
// 未実装のメソッドはpb.UnimplementedAlbumServiceServerのメソッドが使用される
type AlbumServer struct {
	pb.UnimplementedAlbumServiceServer

	savedAlbums []*pb.Album
}

// Unary RPCの実装例
// クライアントからtitleを受け取り、ファイルにデータが存在するかの確認結果をAlbum型で返す
func (s *AlbumServer) GetAlbum(ctx context.Context, req *pb.GetAlbumRequest) (*pb.GetAlbumResponse, error) {
	fmt.Println("")
	fmt.Println("Unary RPC started")

	log.Printf("request: %s", req.Title)
	for _, album := range s.savedAlbums {
		if album.Title == req.Title {
			return &pb.GetAlbumResponse{Album: album}, nil
		}
	}

	log.Printf("album not found: %s", req.Title)
	return &pb.GetAlbumResponse{Album: &pb.Album{}}, nil
}

// ServerStreaming RPCの実装例
// クライアントからartistを受け取り、artistが一致するAlbumをすべてAlbum型で返す
func (s *AlbumServer) ListAlbums(req *pb.ListAlbumsRequest, stream pb.AlbumService_ListAlbumsServer) error {
	fmt.Println("")
	fmt.Println("ServerStreaming RPC started")

	log.Printf("request: %s", req.Artist)
	for _, album := range s.savedAlbums {
		if album.Artist == req.Artist {
			if err := stream.Send(&pb.ListAlbumsResponse{Album: album}); err != nil {
				return err
			}
			time.Sleep(timeSleep) // 確認用
		}
	}

	return nil
}

// ClientStreaming RPCの実装例
// クライアントから複数のtitleを受け取り、ファイルに存在したAlbumの総数とその合計金額、メッセージを返す
func (s *AlbumServer) GetTotalAmount(stream pb.AlbumService_GetTotalAmountServer) error {
	fmt.Println("")
	fmt.Println("ClientStreaming RPC started")

	var (
		albumCount  int32
		totalAmount float32
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(
				&pb.GetTotalAmountResponse{
					AlbumCount:  albumCount,
					TotalAmount: totalAmount,
					Message:     "success to get total amount",
				})
		}

		if err != nil {
			return err
		}

		albumCount++

		log.Printf("request: %s", req.Title)
		for _, album := range s.savedAlbums {
			if album.Title == req.Title {
				totalAmount += album.Price
				break
			}
		}
	}
}

// BidirectionalStreaming RPCの実装例
// クライアントから複数のAlbumを受け取り、その度にメッセージを返す
// 受け取ったAlbumがファイルに存在しない場合はファイルに追加する
func (s *AlbumServer) UploadAndNotify(stream pb.AlbumService_UploadAndNotifyServer) error {
	fmt.Println("")
	fmt.Println("BidirectionalStreaming RPC started")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		log.Printf("request: %s", req.Album.Title)
		res := &pb.UploadAndNotifyResponse{}

		for _, album := range s.savedAlbums {
			if album.Title == req.Album.Title {
				res.Message = fmt.Sprintf("%s is already exists", req.Album.Title)
				break
			}
		}

		if res.Message == "" {
			s.savedAlbums = append(s.savedAlbums, req.Album)
			s.UpdateAlbums(s.savedAlbums, filePath)
			res.Message = fmt.Sprintf("%s is uploaded", req.Album.Title)
		}

		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

// JSONファイルからAlbumのリストをロードするメソッド
func (s *AlbumServer) loadAlbums(filePath string) error {
	var (
		data []byte
		err  error
	)
	data, err = os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.savedAlbums); err != nil {
		return err
	}

	return nil
}

// AlbumのリストをJSONファイルに書き込むメソッド
func (s *AlbumServer) UpdateAlbums(albums []*pb.Album, filePath string) error {
	newFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer newFile.Close()

	newData, err := json.MarshalIndent(albums, "", "	")
	if err != nil {
		return err
	}

	_, err = newFile.Write(newData)
	if err != nil {
		return err
	}

	return nil
}

// サーバーを作成するメソッド
func newServer() *AlbumServer {
	s := &AlbumServer{}
	s.loadAlbums(filePath)

	return s
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAlbumServiceServer(grpcServer, newServer()) // 作成したサーバーをgrpcServerに登録

	log.Println("server started")
	if err := grpcServer.Serve(lis); err != nil { // grpcServerを起動
		log.Fatalf("failed to serve: %v", err)
	}
}
