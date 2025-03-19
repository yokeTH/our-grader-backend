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
	config := config.Load()
	store, err := storage.NewR2Storage(config.R2)
	if err != nil {
		fmt.Println("storage.NewR2Storage failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
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

	body, err := store.GetFile(ctx, submission.Problem.ProjectZipFile)
	if err != nil {
		fmt.Println("store.GetFile failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	defer body.Close()

	zipLocation := fmt.Sprintf("/app/%d/tmp/%s", in.SubmissionID, submission.Problem.ProjectZipFile)
	if err := os.MkdirAll(zipLocation[:len(zipLocation)-len("/zip.zip")], os.ModePerm); err != nil {
		fmt.Println("os.MkdirAll failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	outputFile, err := os.Create(zipLocation)
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

	unzipDir := "/app/%d/run"
	unzipDir = fmt.Sprintf(unzipDir, submission.ID)
	if err := os.MkdirAll(unzipDir, os.ModePerm); err != nil {
		fmt.Println("os.MkdirAll failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	if err := unzip.UnzipFile(zipLocation, unzipDir); err != nil {
		fmt.Println("unzip.UnzipFile failed:", err.Error())
		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}

	for _, file := range submission.SubmissionFile {
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

	simDir := fmt.Sprintf("%s/%s", unzipDir, "cocotb")
	cmd := exec.Command("make")
	cmd.Dir = simDir
	stdOut, _ := cmd.CombinedOutput()

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
			submission.Testcases[i].Result = domain.TestResultCompile
			err := testcaseRepo.UpdateTestcase(&submission.Testcases[i])

			if err != nil {
				fmt.Println("testcaseRepo.UpdateTestcase failed:", err.Error())
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}

		return &verilog.VerilogResponse{Msg: err.Error()}, nil
	}
	wg.Wait()

	if firstErr != nil {
		return &verilog.VerilogResponse{Msg: firstErr.Error()}, nil
	}

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
					fmt.Println("testcaseRepo.UpdateTestcase failed:", err.Error())
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

	return &verilog.VerilogResponse{Msg: "Success"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("net.Listen failed:", err.Error())
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	verilog.RegisterSomeServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println("grpcServer.Serve failed:", err.Error())
		log.Fatalf("failed to serve: %v", err)
	}
}
