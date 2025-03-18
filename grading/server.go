package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/yokeTH/our-grader-backend/api/pkg/config"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
	"github.com/yokeTH/our-grader-backend/api/pkg/repository"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
	"github.com/yokeTH/our-grader-backend/grading/pkg/result"
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
	db, err := database.NewPostgresDB(config.PSQL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	submissionRepo := repository.NewSubmissionRepository(db)

	submission, err := submissionRepo.GetSubmissionsByID(uint(in.SubmissionID))
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

	// write file to dir
	for _, file := range submission.SubmissionFile {
		fileKey := fmt.Sprintf("submissions/%d/%d", submission.ID, file.TemplateFileID)
		fileContent, err := store.GetFile(ctx, fileKey)
		if err != nil {
			return &verilog.VerilogResponse{Msg: err.Error()}, nil
		}
		defer fileContent.Close()

		localFilePath := fmt.Sprintf("%s/%s", unzipDir, file.TemplateFile.Name)
		localFile, err := os.Create(localFilePath)
		if err != nil {
			return &verilog.VerilogResponse{Msg: err.Error()}, nil
		}
		defer localFile.Close()

		_, err = io.Copy(localFile, fileContent)
		if err != nil {
			return &verilog.VerilogResponse{Msg: err.Error()}, nil
		}
	}

	// run
	simDir := fmt.Sprintf("%s/%s", unzipDir, "cocotb")
	cmd := exec.Command("make")
	cmd.Dir = simDir
	stdOut, _ := cmd.CombinedOutput()

	// save stdout
	fileKey := fmt.Sprintf("submissions/%d/stdOut.txt", submission.ID)
	if err := store.UploadFile(ctx, fileKey, "text/plain", strings.NewReader(string(stdOut))); err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	resultPath := fmt.Sprintf("%s/%s", simDir, "results.xml")
	simResult, err := result.GetResult(resultPath)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	testcaseRepo := repository.NewTestcaseRepository(db)

	for _, testsuite := range simResult.Testsuite {
		for i, testcase := range testsuite.Testcase {
			wg.Add(1)
			go func(testcase result.Testcase) {
				defer wg.Done()
				var result domain.TestcaseResult
				if testcase.Failure != nil {
					result = domain.TestResultFail
				} else {
					result = domain.TestResultPass
				}

				testcaseResult := submission.Testcases[i]
				testcaseResult.Result = result
				err := testcaseRepo.UpdateTestcase(&testcaseResult)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				}
			}(testcase)
		}
	}

	wg.Wait()

	if firstErr != nil {
		return &verilog.VerilogResponse{Msg: firstErr.Error()}, nil
	}

	// fileKey = fmt.Sprintf("submissions/%d/stdOut.txt", submission.ID)

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
