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
	"unicode/utf16"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

var (
	moduser32       = syscall.NewLazyDLL("user32.dll")
	enumWindowsProc = moduser32.NewProc("EnumWindows")
	getWindowTextW  = moduser32.NewProc("GetWindowTextW")
	isWindowVisible = moduser32.NewProc("IsWindowVisible")
)

const (
	stdInputHandle = uint32(-10 & 0xFFFFFFFF)
)

var signCmd *exec.Cmd
var serverCmd *exec.Cmd
var clientCmd *exec.Cmd

var (
	ipAddress string
	web_port  string
	sign_port string
	title     string
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
	title := ""
	// if len(args) > 0 && args[0] == "-ui" {
	// 	title = ""
	// } else {
	// 	title = extractFieldValue(jsonStr, `"window_name"`)
	// }

	// if title == "" {
	// 	title = setTitleManually()
	// 	if title == "" {
	// 		panic(0)
	// 	}
	// }

	fmt.Println(ipAddress)
	fmt.Println(web_port)
	fmt.Println(sign_port)
	fmt.Println(title)

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
	// folderServer := "media-server"
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

	// // Append the folder to the working directory
	// folderServerPath := filepath.Join(workingDir, folderServer)

	// // Start signalling server process to generate raw video frames
	// serverCmd, err = launchCommand(
	// 	"go",
	// 	[]string{
	// 		"run", ".",
	// 		ipAddress, port, title,
	// 	},
	// 	"logs/server.log",
	// 	folderServerPath,
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// // Use a defer statement to ensure the command process is killed when the main function exits
	// defer func() {
	// 	stopProcess(serverCmd)
	// }()

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

			// if serverCmd != nil {
			// 	if serverCmd.ProcessState != nil && serverCmd.ProcessState.Exited() {
			// 		status += " server: off "
			// 	} else {
			// 		status += " server: on "
			// 	}
			// }

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
			// stopProcess(serverCmd)

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

		// case "server":
		// 	fmt.Println("Restart server")
		// 	stopProcess(serverCmd)

		// 	serverCmd, err = launchCommand(
		// 		"go",
		// 		[]string{
		// 			"run", ".",
		// 			ipAddress, port, title,
		// 		},
		// 		"logs/server.log",
		// 		folderServerPath,
		// 	)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// case "change":
		// 	title = setTitleManually()
		// 	if title == "" {
		// 		continue
		// 	}

		// 	fmt.Println("Restart server")
		// 	stopProcess(serverCmd)

		// 	serverCmd, err = launchCommand(
		// 		"go",
		// 		[]string{
		// 			"run", ".",
		// 			ipAddress, port, title,
		// 		},
		// 		"logs/server.log",
		// 		folderServerPath,
		// 	)
		// 	if err != nil {
		// 		panic(err)
		// 	}

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
 * getCurrentDirectory returns the current directory.
 *
 * @return The current directory as a string.
 */
func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
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

/**
 * findWindowTitle retrieves the window title of a given executable string.
 *
 * @param execString The executable string used to identify the window.
 * @return The window title as a string if found, or an error if the title is not found.
 */
func findWindowTitle(execString string) (string, error) {
	// Run the tasklist command to get the window titles
	tasklistCmd := exec.Command("cmd.exe", "/C", "tasklist", "/v", "/fi", "imagename eq "+execString, "/fo", "list", "|", "findstr", "Titre")
	output, err := tasklistCmd.Output()
	if err != nil {
		return "", err
	}
	// Convert bytes to string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Split the windowString using ":" as the delimiter
		parts := strings.Split(line, ": ")
		if len(parts) > 1 && len(parts[1]) > 0 {
			//  Need to remove useless byte at the end : [13] => (carriage return)
			return parts[1][:len(parts[1])-1], nil
		}
	}

	return "", fmt.Errorf("window title not found")
}

/**
 * getWindowTitles retrieves a list of visible window titles.
 *
 * @return A slice of strings representing the visible window titles.
 */
func getWindowTitles() []string {
	var titles []string

	enumWindowsProc.Call(syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		var buf [256]uint16
		if _, _, err := getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))); err.(syscall.Errno) == 0 {
			length := 0
			for buf[length] != 0 {
				length++
			}
			title := string(utf16.Decode(buf[:length]))
			visible, _, _ := isWindowVisible.Call(uintptr(hwnd))
			if visible != 0 && len(title) > 0 {
				titles = append(titles, title)
			}
		}
		return 1 // Continue enumeration
	}), 0)

	return titles
}

/**
 * setTitleManually interactively prompts the user to manually select a window title.
 *
 * @return The selected window title as a string. If the index is invalid, it returns an empty string.
 */
func setTitleManually() string {
	titles := getWindowTitles()
	fmt.Println("[Window Titles (visible)]")
	for i, t := range titles {
		fmt.Printf("%d: %s\n", i, t)
	}

	// Wait for index input after the loop
	var index int
	index = -1
	fmt.Print("Enter the index of the window you want to select: ")
	fmt.Scanln(&index)

	// Perform further operations based on the selected index
	if index >= 0 && index < len(titles) {
		fmt.Println("You selected window title:", titles[index])
		return titles[index]
	} else {
		fmt.Println("Invalid index. Please enter a valid index.")
		return ""
	}
}
