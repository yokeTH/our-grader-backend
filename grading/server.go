package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

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

	zipLocation := fmt.Sprintf("/Users/yoketh/Repo/our-grader-backend/bin/%s", in.ZipKey)
	os.MkdirAll(zipLocation[:len(zipLocation)-len("/zip.zip")], os.ModePerm)
	outputFile, err := os.Create(zipLocation)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer outputFile.Close()
	defer body.Close()

	_, err = io.Copy(outputFile, body)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	unzipDir := "/Users/yoketh/Repo/our-grader-backend/bin/tmp/run"
	os.MkdirAll(unzipDir, os.ModePerm)
	if err := unzip.UnzipFile(zipLocation, unzipDir); err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	workingDir := fmt.Sprintf("%s/%s", unzipDir, "cocotb")
	cmd := exec.Command("make")
	cmd.Dir = workingDir
	stdOut, err := cmd.CombinedOutput()
	if err != nil {
		return &verilog.VerilogResponse{Msg: "Success"}, nil
	}
	fmt.Println(string(stdOut))

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
