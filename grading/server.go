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

const port = ":50051"

type server struct {
	verilog.UnimplementedSomeServiceServer
}

func (s *server) Run(ctx context.Context, in *verilog.VerilogRequest) (*verilog.VerilogResponse, error) {
	// Load configurations and setup storage and database
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

	// Retrieve submission from database
	submission, err := submissionRepo.GetSubmissionsByID(uint(in.SubmissionID))
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Retrieve and store the zip file from cloud storage
	body, err := store.GetFile(ctx, submission.Problem.ProjectZipFile)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer body.Close()

	zipLocation := fmt.Sprintf("/Users/yoketh/Repo/our-grader-backend/bin/%s", submission.Problem.ProjectZipFile)
	os.MkdirAll(zipLocation[:len(zipLocation)-len("/zip.zip")], os.ModePerm)
	outputFile, err := os.Create(zipLocation)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer outputFile.Close()

	// Copy the content of the zip file to the local file system
	_, err = io.Copy(outputFile, body)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Unzip the contents to a temporary directory
	unzipDir := "/Users/yoketh/Repo/our-grader-backend/bin/tmp/run"
	os.MkdirAll(unzipDir, os.ModePerm)
	if err := unzip.UnzipFile(zipLocation, unzipDir); err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Write submission files to the unzip directory
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

	// Run the simulation using `make`
	simDir := fmt.Sprintf("%s/%s", unzipDir, "cocotb")
	cmd := exec.Command("make")
	cmd.Dir = simDir
	stdOut, _ := cmd.CombinedOutput()

	// Save stdout to cloud storage
	fileKey := fmt.Sprintf("submissions/%d/stdOut.txt", submission.ID)
	if err := store.UploadFile(ctx, fileKey, "text/plain", strings.NewReader(string(stdOut))); err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Parse the simulation result
	resultPath := fmt.Sprintf("%s/%s", simDir, "results.xml")
	simResult, err := result.GetResult(resultPath)
	if err != nil {
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Process the test case results concurrently
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

	// Handle any errors encountered during the concurrent processing
	if firstErr != nil {
		return &verilog.VerilogResponse{Msg: firstErr.Error()}, nil
	}

	return &verilog.VerilogResponse{Msg: "Success"}, nil
}

func main() {
	// Start the gRPC server
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
