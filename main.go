package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
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

	// Insert the breakpoint line before the specified line
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

	tempFile := "tempfile"
	if err := writeLines(lines, tempFile); err != nil {
		log.Fatalf("Error writing temp file: %v", err)
	}
	defer os.Remove(tempFile)

	if err := runBuildah(tempFile, lines); err != nil {
		log.Fatalf("error building image: %v", err)
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

func runBuildah(file string, lines []string) error {
	for i, line := range lines {
		if strings.TrimSpace(line) == "# BREAKPOINT" {
			fmt.Printf("Breakpoint found at line %d in %s\n", i+1, file)
			fmt.Println("Entering interactive shell. Type 'exit' to continue.")
			cmd := exec.Command("./buildah-run.sh", "from", "scratch")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error running shell: %v", err)
			}
			fmt.Println("Exiting interactive shell. Continuing build...")
		} else if strings.HasPrefix(line, "RUN ") {
			command := strings.Fields(strings.TrimPrefix(line, "RUN "))
			cmd := exec.Command("./buildah-run.sh", append([]string{"run", "working-container"}, command...)...)
			fmt.Printf("Running command: %s\n", strings.Join(cmd.Args, " "))
			output, err := cmd.CombinedOutput()
			fmt.Printf("Command output: %s\n", string(output))
			if err != nil {
				return fmt.Errorf("error running command: %v", err)
			}
		}
	}

	cmd := exec.Command("./buildah-run.sh", "commit", "working-container", "my-debug-image:latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Running command: %s\n", strings.Join(cmd.Args, " "))
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Command output: %s\n", string(output))
		return fmt.Errorf("error committing image: %v", err)
	} else {
		fmt.Printf("Command output: %s\n", string(output))
	}

	fmt.Printf("Successfully built image with reference: my-debug-image:latest\n")
	return nil
}
