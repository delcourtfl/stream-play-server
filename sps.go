package main // Stream Play Server

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	stdInputHandle = uint32(-10 & 0xFFFFFFFF)
)

var signCmd *exec.Cmd
var serverCmd *exec.Cmd
var clientCmd *exec.Cmd

var (
	ipAddress  string
	web_port   string
	sign_port  string
	input_port string
)

/**
 * main is the main entry point of the SPS program.
 * It reads configuration from a JSON file, launches various processes, and handles user input.
 */
func main() {
	// args := os.Args[1:] // Exclude the first argument, which is the program name

	jsonFile := "config.json"

	// Read the JSON file
	jsonBytes, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the IP address using string manipulation
	jsonStr := string(jsonBytes)

	ipAddress = extractFieldValue(jsonStr, `"ip_address"`)
	web_port = extractFieldValue(jsonStr, `"web_port"`)
	sign_port = extractFieldValue(jsonStr, `"sign_port"`)
	input_port = extractFieldValue(jsonStr, `"input_port"`)

	fmt.Println(ipAddress + ":" + web_port)
	fmt.Println(ipAddress + ":" + sign_port)
	fmt.Println("localhost:" + input_port)

	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Add Config json file to the webserver
	configWebserverPath := filepath.Join(workingDir, "webserver/wconfig.json")

	// Create the destination file
	configWebserverFile, err := os.Create(configWebserverPath)
	if err != nil {
		log.Fatal(err)
	}
	defer configWebserverFile.Close()

	// Write the content to the destination file
	_, err = configWebserverFile.Write(jsonBytes)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Config File copied successfully!")

	// Set the folder to be appended to the working directory
	folderSign := "signaling"
	folderServer := "media-server"
	folderWeb := "webserver"

	////////////////////////////
	// Start signaling server //
	////////////////////////////

	// Append the folder to the working directory
	folderSignPath := filepath.Join(workingDir, folderSign)

	// Start signalling server process to generate raw video frames
	signCmd, err = launchCommand(
		"go",
		[]string{
			"run", ".",
			ipAddress, sign_port,
		},
		"logs/sign.log",
		folderSignPath,
	)
	if err != nil {
		panic(err)
	}

	// Use a defer statement to ensure the command process is killed when the main function exits
	defer func() {
		stopProcess(signCmd)
	}()

	/////////////////////////
	// Start WebRTC server //
	/////////////////////////

	// Append the folder to the working directory
	folderServerPath := filepath.Join(workingDir, folderServer)

	// Start signalling server process to generate raw video frames
	serverCmd, err = launchCommand(
		"go",
		[]string{
			"run", ".",
			input_port,
		},
		"logs/server.log",
		folderServerPath,
	)
	if err != nil {
		panic(err)
	}

	// Use a defer statement to ensure the command process is killed when the main function exits
	defer func() {
		stopProcess(serverCmd)
	}()

	/////////////////////
	// Start WebServer //
	/////////////////////

	// Append the folder to the working directory
	folderWebPath := filepath.Join(workingDir, folderWeb)

	// Start signalling server process to generate raw video frames
	clientCmd, err = launchCommand(
		"go",
		[]string{
			"run", ".",
			ipAddress, web_port,
		},
		"logs/client.log",
		folderWebPath,
	)
	if err != nil {
		panic(err)
	}

	// Use a defer statement to ensure the command process is killed when the main function exits
	defer func() {
		stopProcess(clientCmd)
	}()

	// Monitoring with goroutine every 10 sec

	go func() {
		for {
			// Wait for 10 seconds before checking again
			time.Sleep(10 * time.Second)

			status := "[status]"

			if signCmd != nil {
				if signCmd.ProcessState != nil && signCmd.ProcessState.Exited() {
					status += " sign: off "
				} else {
					status += " sign: on "
				}
			}

			if serverCmd != nil {
				if serverCmd.ProcessState != nil && serverCmd.ProcessState.Exited() {
					status += " server: off "
				} else {
					status += " server: on "
				}
			}

			if clientCmd != nil {
				if clientCmd.ProcessState != nil && clientCmd.ProcessState.Exited() {
					status += " client: off "
				} else {
					status += " client: on "
				}
			}

			fmt.Println(status)
		}
	}()

	// Set the terminal to raw mode to capture keypresses immediately
	setRawMode()

	reader := bufio.NewReader(os.Stdin)

	for {
		// Read a line of input (including newline characters)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		line = strings.TrimRight(line, "\r\n")

		switch line {
		case "exit":
			fmt.Println("Stopping SPS...")
			return

		case "stop":
			stopProcess(signCmd)
			stopProcess(clientCmd)
			stopProcess(serverCmd)

		case "sign":
			fmt.Println("Restart sign")
			stopProcess(signCmd)

			signCmd, err = launchCommand(
				"go",
				[]string{
					"run", ".",
					ipAddress, sign_port,
				},
				"logs/sign.log",
				folderSignPath,
			)
			if err != nil {
				panic(err)
			}

		case "client":
			fmt.Println("Restart client")
			stopProcess(clientCmd)

			clientCmd, err = launchCommand(
				"go",
				[]string{
					"run", ".",
					ipAddress, web_port,
				},
				"logs/client.log",
				folderWebPath,
			)
			if err != nil {
				panic(err)
			}

		case "server":
			fmt.Println("Restart server")
			stopProcess(serverCmd)

			serverCmd, err = launchCommand(
				"go",
				[]string{
					"run", ".",
					input_port,
				},
				"logs/server.log",
				folderServerPath,
			)
			if err != nil {
				panic(err)
			}

		default:
			// Handle the case when an invalid command is provided.
			fmt.Println("Invalid command : " + line)
		}
	}

}

/**
 * extractFieldValue extracts the field value from a JSON string.
 *
 * @param jsonStr The JSON string to extract the field value from.
 * @param field The field name to extract the value for.
 * @return The extracted field value as a string.
 */
func extractFieldValue(jsonStr string, field string) string {
	startIndex := strings.Index(jsonStr, field)
	if startIndex == -1 {
		log.Fatalf("Field '%s' not found in JSON", field)
	}

	startIndex += len(field) + 3 // Adjust start index to skip the field and the following characters `": `
	endIndex := strings.Index(jsonStr[startIndex:], `"`)
	if endIndex == -1 {
		log.Fatalf("Field value not found for '%s'", field)
	}

	return jsonStr[startIndex : startIndex+endIndex]
}

/**
 * setRawMode sets the terminal to raw mode for capturing keypresses immediately.
 */
func setRawMode() {
	stdinHandle, _, _ := procGetStdHandle.Call(uintptr(stdInputHandle))
	var mode uint32
	_, _, _ = procGetConsoleMode.Call(stdinHandle, uintptr(unsafe.Pointer(&mode)))

	mode &^= 0x0004 // Disable ECHO

	_, _, _ = procSetConsoleMode.Call(stdinHandle, uintptr(mode))
}

/**
 * launchCommand launches a command as a subprocess in a new process group.
 *
 * @param command The name or path of the executable command to run.
 * @param args An array of strings representing the command-line arguments for the executable.
 * @param filePath Path of the log file for the process output redirection.
 * @param workingDir Directory on which to launch the process.
 * @return A pointer to the Cmd struct representing the running command and an error (if any).
 *         If there is an error while starting the command, the returned pointer will be nil, and the error will be non-nil.
 */
func launchCommand(command string, args []string, filePath string, workingDir string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	fmt.Println(cmd)

	outFile, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	cmd.Stdout = outFile
	cmd.Stderr = outFile

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Used to get the process state
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	return cmd, nil
}

/**
 * killProcess forcefully terminates a process and its subprocesses on Windows.
 *
 * @param cmd A pointer to the exec.Cmd struct representing the running command to be terminated.
 * @return An error encountered during the termination process. Returns nil if successful.
 */
func killProcess(cmd *exec.Cmd) error {
	pgid := -cmd.Process.Pid
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(-pgid))
	return killCmd.Run()
}

/**
 * stopProcess forcefully terminates a process and handles its termination status.
 * If the process has already exited, prints a message indicating so.
 * If the process is still running, attempts to kill it and prints a message about the action taken.
 *
 * @param cmd A pointer to the exec.Cmd struct representing the running command to be stopped.
 */
func stopProcess(cmd *exec.Cmd) {
	if cmd != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			fmt.Println("Process is already stopped")
			return
		} else {
			if cmd.Process != nil {
				if err := killProcess(cmd); err != nil {
					fmt.Println("Error killing process:", err)
					return
				}
				fmt.Println("Process killed :", cmd)
				return
			}
		}
	}
	fmt.Println("Should not happen...")
}
