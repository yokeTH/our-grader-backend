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
	"time"

	"github.com/yokeTH/our-grader-backend/api/pkg/config"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
	"github.com/yokeTH/our-grader-backend/api/pkg/repository"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
	"github.com/yokeTH/our-grader-backend/grading/pkg/result"
	"github.com/yokeTH/our-grader-backend/grading/pkg/unzip"
	"github.com/yokeTH/our-grader-backend/proto/verilog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port             = ":50051"
	executionTimeout = 5 * time.Minute // Set an appropriate timeout value
)

type server struct {
	verilog.UnimplementedSomeServiceServer
}

func (s *server) Run(ctx context.Context, in *verilog.VerilogRequest) (*verilog.VerilogResponse, error) {
	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(ctx, executionTimeout)
	defer cancel()

	// Create a channel to track if context is done before function completes
	done := make(chan struct{})
	var resp *verilog.VerilogResponse
	var execErr error

	go func() {
		resp, execErr = s.processVerilog(ctx, in)
		close(done)
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "request timed out after %v", executionTimeout)
	case <-done:
		return resp, execErr
	}
}

func (s *server) processVerilog(ctx context.Context, in *verilog.VerilogRequest) (*verilog.VerilogResponse, error) {
	config := config.Load()
	store, err := storage.NewR2Storage(config.R2)
	if err != nil {
		fmt.Println("storage.NewR2Storage failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	db, err := database.NewPostgresDB(config.PSQL)
	if err != nil {
		fmt.Println("database.NewPostgresDB failed:", err.Error())
		log.Fatalf("failed to connect database: %v", err)
	}

	submissionRepo := repository.NewSubmissionRepository(db)

	submission, err := submissionRepo.GetSubmissionsByID(uint(in.SubmissionID))
	if err != nil {
		fmt.Println("submissionRepo.GetSubmissionsByID failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	body, err := store.GetFile(ctx, submission.Problem.ProjectZipFile)
	if err != nil {
		fmt.Println("store.GetFile failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer body.Close()

	basePath := fmt.Sprintf("/tmp/verilog/%d", in.SubmissionID)

	zipPath := fmt.Sprintf("%s/%s", basePath, "zip.zip")
	zipDir := zipPath[:len(zipPath)-len("/zip.zip")]
	if err := os.MkdirAll(zipDir, os.ModePerm); err != nil {
		fmt.Println("os.MkdirAll failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	outputFile, err := os.Create(zipPath)
	if err != nil {
		fmt.Println("os.Create failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, body)
	if err != nil {
		fmt.Println("io.Copy failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	unzipDir := fmt.Sprintf("%s/prj", basePath)
	if err := os.MkdirAll(unzipDir, os.ModePerm); err != nil {
		fmt.Println("os.MkdirAll failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	if err := unzip.UnzipFile(zipPath, unzipDir); err != nil {
		fmt.Println("unzip.UnzipFile failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	for _, file := range submission.SubmissionFile {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		fileKey := fmt.Sprintf("submissions/%d/%d", submission.ID, file.TemplateFileID)
		fileContent, err := store.GetFile(ctx, fileKey)
		if err != nil {
			fmt.Println("store.GetFile failed:", err.Error())
			return &verilog.VerilogResponse{Msg: err.Error()}, nil
		}
		defer fileContent.Close()

		localFilePath := fmt.Sprintf("%s/%s", unzipDir, file.TemplateFile.Name)
		localFile, err := os.Create(localFilePath)
		if err != nil {
			fmt.Println("os.Create failed:", err.Error())
			return &verilog.VerilogResponse{Msg: err.Error()}, nil
		}
		defer localFile.Close()

		_, err = io.Copy(localFile, fileContent)
		if err != nil {
			fmt.Println("io.Copy failed:", err.Error())
			return &verilog.VerilogResponse{Msg: err.Error()}, nil
		}
	}

	// Run simulation with context
	simDir := fmt.Sprintf("%s/%s", unzipDir, "cocotb")

	// Use CommandContext to respect context timeouts
	cmd := exec.CommandContext(ctx, "make")
	cmd.Dir = simDir
	stdOut, err := cmd.CombinedOutput()

	// Check if the error was due to context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// First handle the command execution error if there is one
	// Note: This doesn't necessarily mean the simulation failed as the exit code
	// might be non-zero but we still want to capture the output
	if err != nil {
		fmt.Printf("make command execution error: %v\n", err)
		// We continue processing to capture the output regardless
	}

	fileKey := fmt.Sprintf("submissions/%d/stdOut.txt", submission.ID)
	if err := store.UploadFile(ctx, fileKey, "text/plain", strings.NewReader(string(stdOut))); err != nil {
		fmt.Println("store.UploadFile failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	submission.StdoutObjectKey = fileKey
	if err := submissionRepo.Update(&submission); err != nil {
		fmt.Println("submissionRepo.Update failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	testcaseRepo := repository.NewTestcaseRepository(db)

	resultPath := fmt.Sprintf("%s/%s", simDir, "results.xml")
	simResult, err := result.GetResult(resultPath)
	if err != nil {
		fmt.Println("result.GetResult failed:", err.Error())

		for i := range submission.Testcases {
			// Check if context is cancelled
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			submission.Testcases[i].Result = domain.TestResultCompile
			updateErr := testcaseRepo.UpdateTestcase(&submission.Testcases[i])

			if updateErr != nil {
				fmt.Println("testcaseRepo.UpdateTestcase failed:", updateErr.Error())
				mu.Lock()
				if firstErr == nil {
					firstErr = updateErr
				}
				mu.Unlock()
			}
		}

		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	// Create a context-aware goroutine pool
	processDone := make(chan struct{})

	go func() {
		for _, testsuite := range simResult.Testsuite {
			for i, testcase := range testsuite.Testcase {
				// Check if context is cancelled
				if ctx.Err() != nil {
					wg.Wait()
					close(processDone)
					return
				}

				wg.Add(1)
				go func(testcase result.Testcase, index int) {
					defer wg.Done()

					// Check context cancellation within goroutine
					if ctx.Err() != nil {
						return
					}

					var result domain.TestcaseResult
					if testcase.Failure != nil {
						result = domain.TestResultFail
					} else {
						result = domain.TestResultPass
					}

					testcaseResult := submission.Testcases[index]
					testcaseResult.Result = result
					updateErr := testcaseRepo.UpdateTestcase(&testcaseResult)

					if updateErr != nil {
						fmt.Println("testcaseRepo.UpdateTestcase failed:", updateErr.Error())
						mu.Lock()
						if firstErr == nil {
							firstErr = updateErr
						}
						mu.Unlock()
					}
				}(testcase, i)
			}
		}

		wg.Wait()
		close(processDone)
	}()

	// Wait for either processing to complete or context to be cancelled
	select {
	case <-processDone:
		// Processing completed normally
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if firstErr != nil {
		return &verilog.VerilogResponse{Msg: firstErr.Error()}, nil
	}

	return &verilog.VerilogResponse{Msg: "Success"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("net.Listen failed:", err.Error())
		log.Fatalf("failed to listen: %v", err)
	}

	// Configure server options with timeout
	var serverOptions []grpc.ServerOption
	serverOptions = append(serverOptions, grpc.ConnectionTimeout(30*time.Second))

	// Create gRPC server with the configured options
	grpcServer := grpc.NewServer(serverOptions...)
	verilog.RegisterSomeServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println("grpcServer.Serve failed:", err.Error())
		log.Fatalf("failed to serve: %v", err)
	}
}
