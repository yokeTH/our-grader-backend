package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"

	"github.com/yokeTH/our-grader-backend/api/pkg/config"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
	"github.com/yokeTH/our-grader-backend/grading/pkg/unzip"
	"github.com/yokeTH/our-grader-backend/proto/verilog"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	verilog.UnimplementedSomeServiceServer
}

func (s *server) Run(ctx context.Context, in *verilog.VerilogRequest) (*verilog.VerilogResponse, error) {
	config := config.Load()
	store, err := storage.NewR2Storage(config.R2)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	body, err := store.GetFile(ctx, in.ZipKey)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// zipLocation := fmt.Sprintf("/Users/yoketh/Repo/our-grader-backend/bin/%s", in.ZipKey)
	outputFile, err := os.Create("/problems/21/zip/1742237831461.zip")
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer outputFile.Close()
	defer body.Close()

	_, err = io.Copy(outputFile, body)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	if err := unzip.UnzipFile("/Users/yoketh/Repo/our-grader-backend/bin/problems/21/zip/1742237831461.zip", "/Users/yoketh/Repo/our-grader-backend/bin/unzip"); err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	return &verilog.VerilogResponse{Msg: "Success"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	verilog.RegisterSomeServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
