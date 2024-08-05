package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/containers/buildah"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/reexec"
)

func main() {
	if reexec.Init() {
		return
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command> [<args>]", os.Args[0])
	}

	command := os.Args[1]

	switch command {
	case "set-breakpoint":
		if len(os.Args) != 4 {
			log.Fatalf("Usage: %s set-breakpoint <file> <line>", os.Args[0])
		}
		setBreakpoint(os.Args[2], os.Args[3])
	case "build":
		if len(os.Args) != 3 {
			log.Fatalf("Usage: %s build <file>", os.Args[0])
		}
		build(os.Args[2])
	case "continue":
		continueBuild()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func setBreakpoint(file, lineStr string) {
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		log.Fatalf("Invalid line number: %s", lineStr)
	}

	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	lines = append(lines[:line-1], append([]string{"# BREAKPOINT"}, lines[line-1:]...)...)

	if err := writeLines(lines, file); err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Printf("Breakpoint set at line %d in %s\n", line, file)
}

func build(file string) {
	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	for i, line := range lines {
		if strings.TrimSpace(line) == "# BREAKPOINT" {
			fmt.Printf("Breakpoint at line %d\n", i+1)
			fmt.Println("Entering interactive shell. Type 'exit' to continue.")
			cmd := exec.Command("/bin/sh")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else {
			tempFile := fmt.Sprintf("tempfile-%d", i)
			if err := writeLines(lines[:i+1], tempFile); err != nil {
				log.Fatalf("Error writing temp file: %v", err)
			}
			runBuildah(tempFile)
			os.Remove(tempFile)
		}
	}
}

func continueBuild() {
	fmt.Println("Resuming build...")
	// Implement continuation logic if needed
}

func readLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func runBuildah(file string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting user home directory: %v", err)
	}

	storeOptions := storage.StoreOptions{
		GraphDriverName:    "vfs",
		GraphRoot:          homeDir + "/.local/share/containers/storage",
		RunRoot:            homeDir + "/.local/share/containers/run",
		GraphDriverOptions: []string{},
	}

	store, err := storage.GetStore(storeOptions)
	if err != nil {
		log.Fatalf("error getting store: %v", err)
	}

	ctx := context.Background()
	systemContext := &types.SystemContext{
		ArchitectureChoice: "arm64",
		OSChoice:           "linux",
		VariantChoice:      "v8",
	}

	// Create a new builder with specified architecture and OS
	builder, err := buildah.NewBuilder(ctx, store, buildah.BuilderOptions{
		FromImage:     "docker://alpine:latest",
		SystemContext: systemContext,
	})
	if err != nil {
		log.Fatalf("error creating builder: %v", err)
	}

	// Read and execute each step in the temp file
	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("Error reading temp file: %v", err)
	}

	for _, line := range lines {
		// Parse and execute build steps
		// This is a simplified example; you would need to handle different Dockerfile/Containerfile instructions
		if strings.HasPrefix(line, "RUN ") {
			command := strings.Fields(strings.TrimPrefix(line, "RUN "))
			if err := builder.Run(command, buildah.RunOptions{}); err != nil {
				log.Fatalf("error running command: %v", err)
			}
		}
	}

	// Commit the image
	imageRef, err := alltransports.ParseImageName("docker-daemon:my-debug-image:latest")
	if err != nil {
		log.Fatalf("error parsing image reference: %v", err)
	}

	_, _, _, err = builder.Commit(ctx, imageRef, buildah.CommitOptions{})
	if err != nil {
		log.Fatalf("error committing image: %v", err)
	}

	fmt.Printf("Successfully built image with reference: %s\n", imageRef.StringWithinTransport())
}
