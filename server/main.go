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

	"google.golang.org/grpc"
)

var (
	filePath = "db/album.json"
	port     = "50051"
)

type AlbumServer struct {
	pb.UnimplementedAlbumServiceServer

	savedAlbums []*pb.Album
}

func (s *AlbumServer) GetAlbum(ctx context.Context, req *pb.GetAlbumRequest) (*pb.GetAlbumResponse, error) {
	for _, album := range s.savedAlbums {
		if album.Title == req.Title {
			log.Printf("album found: %s", req.Title)
			return &pb.GetAlbumResponse{Album: album}, nil
		}
	}

	log.Printf("album not found: %s", req.Title)
	return &pb.GetAlbumResponse{Album: &pb.Album{}}, nil
}

func (s *AlbumServer) ListAlbums(req *pb.ListAlbumsRequest, stream pb.AlbumService_ListAlbumsServer) error {
	for _, album := range s.savedAlbums {
		if album.Artist == req.Artist {
			log.Printf("%s's album found: %s", req.Artist, album.Title)
			if err := stream.Send(&pb.ListAlbumsResponse{Album: album}); err != nil {
				return err
			}
		}
	}

	log.Printf("%s's album not found", req.Artist)
	return nil
}

func (s *AlbumServer) GetTotalAmount(stream pb.AlbumService_GetTotalAmountServer) error {
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
				log.Printf("%s's price: %f", req.Title, album.Price)
				totalAmount += album.Price
			}
		}
	}
}

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
	pb.RegisterAlbumServiceServer(grpcServer, newServer())

	log.Println("server started")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
