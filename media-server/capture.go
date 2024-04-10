package main

import (
	"os"
	"os/exec"
	"fmt"
	"time"
	"strconv"
    "syscall"
	"unicode/utf16"
	"unsafe"
	"io"
)

var (
	moduser32        = syscall.NewLazyDLL("user32.dll")
	enumWindowsProc  = moduser32.NewProc("EnumWindows")
	getWindowTextW   = moduser32.NewProc("GetWindowTextW")
	isWindowVisible  = moduser32.NewProc("IsWindowVisible")
)

var ffmpegCmdVideo *exec.Cmd
var ffmpegCmdAudio *exec.Cmd

/**
 * captureStream captures a video stream from a window specified by its title using FFMPEG.
 * It launches FFMPEG processes to capture video and audio streams from the specified window title.
 *
 * @param windowTitle The title string of the window to capture.
 */
func captureStream(windowTitle string) {
	fmt.Println("Trying to capture video stream")
	var err error

	// Used for debugging
	if windowTitle == "" {
		titles := getWindowTitles()
		fmt.Println("[Window Titles (visible)]")
		for i, title := range titles {
			fmt.Printf("%d: %s\n", i, title)
		}

		// Wait for index input after the loop
		var index int
		index = -1
		fmt.Print("Enter the index of the window you want to select: ")
		fmt.Scanln(&index)

		// Perform further operations based on the selected index
		if index >= 0 && index < len(titles) {
			fmt.Println("You selected window title:", titles[index])
			windowTitle = titles[index]
		} else {
			fmt.Println("Invalid index. Please enter a valid index.")
			panic(0)
		}
	}

	fmt.Println("Got a window to capture")
	fmt.Println("#"+windowTitle+"#")

	// Start FFMPEG process to generate raw video frames
	ffmpegCmdVideo, err = launchCommand(
		"ffmpeg",
		[]string{
			"-stats_period", "10",
			"-f", "gdigrab",
			"-thread_queue_size", "1024",
			"-framerate", "30",
			"-i", "title="+windowTitle, // desktop works tooZ
			// "-i", "desktop",
			"-vf", "scale=-1:720",
			// "-vf", "scale=1280:720",
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-tune", "zerolatency",
			"-crf", "25",
			"-pix_fmt", "yuv420p",
			"-an",
			"-f", "rtp", "rtp://127.0.0.1:5004?pkt_size=1200",
		},
		"../logs/video.log",
	)
	if err != nil {
		panic(err)
	}

	// ffmpegCmdAudio, err = launchCommand(
	// 	"ffmpeg",
	// 	[]string{
	// 		"-stats_period", "10",
	// 		"-f", "dshow",					// Use something else ?
	// 		"-i", "audio=Mixage stéréo (Realtek(R) Audio)",
	// 		"-c:a", "libopus",
	// 		"-application", "lowdelay",    	// Enable low-delay mode for Opus
	// 		"-vbr", "off",
	// 		"-compression_level", "0",
	// 		"-frame_duration", "20",      	// Set the Opus frame duration to 20 ms for lower latency
	// 		"-vn",
	// 		"-f", "rtp", "rtp://127.0.0.1:5005",
	// 	},
	// 	"../logs/audio.log",
	// )
	// if err != nil {
	// 	panic(err)
	// }

	time.Sleep(500 * time.Millisecond)
}

/**
 * launchCommand launches a command as a subprocess with redirected output.
 * It starts the command specified by the 'command' parameter with the provided arguments.
 * The subprocess's standard output and standard error streams are redirected to the specified file path.
 *
 * @param command The name or path of the executable command to run.
 * @param args An array of strings representing the command-line arguments for the executable.
 * @param filePath The path of the log file for the process output redirection.
 * @return (*exec.Cmd, error) A pointer to the Cmd struct representing the running command and an error (if any).
 */
func launchCommand(command string, args []string, filePath string) (*exec.Cmd, error) {
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

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

/**
 * killProcess forcefully terminates a process and its subprocesses on Windows.
 * It sends a SIGTERM signal to the process group associated with the given cmd parameter.
 *
 * @param cmd A pointer to the exec.Cmd struct representing the running command to be terminated.
 * @return error An error (if any) encountered during the termination process. Returns nil if successful.
 */
func killProcess(cmd *exec.Cmd) error {
	pgid := -cmd.Process.Pid
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(-pgid))
	return killCmd.Run()
}

/**
 * restartFFMPEG stops and restarts the FFMPEG processes for capturing streams.
 * It terminates the existing FFMPEG processes and relaunches them with the new configuration.
 *
 * @param windowTitle The title string of the window to capture after restarting FFMPEG.
 */
func restartFFMPEG(windowTitle string) {
	// Stop the existing FFMPEG processes if they are running
	if ffmpegCmdVideo != nil && ffmpegCmdVideo.Process != nil {
		if err := killProcess(ffmpegCmdVideo); err != nil {
			panic(err)
		}
	}

	if ffmpegCmdAudio != nil && ffmpegCmdAudio.Process != nil {
		if err := killProcess(ffmpegCmdAudio); err != nil {
			panic(err)
		}
	}

	// Delay a bit to ensure the previous processes are terminated
	time.Sleep(500 * time.Millisecond)

	// Relaunch the FFMPEG processes
	captureStream(windowTitle)
}

/**
 * getWindowTitles retrieves a list of visible window titles using the Windows API.
 *
 * @return []string A slice containing the titles of visible windows.
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
 * stopProcess stops a running process by killing its associated process group.
 *
 * @param cmd A pointer to the exec.Cmd struct representing the running process to be stopped.
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
 * stopAllCapture stops the FFMPEG processes capturing video and audio streams.
 */
func stopAllCapture() {
	stopProcess(ffmpegCmdVideo)
	stopProcess(ffmpegCmdAudio)
}

func RunCommand(name string, arg ...string) (io.ReadCloser, error) {
	cmd := exec.Command(name, arg...)
	fmt.Println(cmd)

	dataPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return dataPipe, nil
}

func newCaptureStream(windowTitle string) (io.ReadCloser, error) {
	fmt.Println("Trying to capture video stream")
	var err error

	// Used for debugging
	if windowTitle == "" {
		titles := getWindowTitles()
		fmt.Println("[Window Titles (visible)]")
		for i, title := range titles {
			fmt.Printf("%d: %s\n", i, title)
		}

		// Wait for index input after the loop
		var index int
		index = -1
		fmt.Print("Enter the index of the window you want to select: ")
		fmt.Scanln(&index)

		// Perform further operations based on the selected index
		if index >= 0 && index < len(titles) {
			fmt.Println("You selected window title:", titles[index])
			windowTitle = titles[index]
		} else {
			fmt.Println("Invalid index. Please enter a valid index.")
			panic(0)
		}
	}

	fmt.Println("Got a window to capture")
	fmt.Println("#"+windowTitle+"#")

	command := "ffmpeg"
	// args := []string{
	// 	"-stats_period", "10",
	// 	"-f", "gdigrab",
	// 	"-thread_queue_size", "1024",
	// 	"-framerate", "10",
	// 	// "-i", "title="+windowTitle, // desktop works tooZ
	// 	"-vf", "scale=-1:720",
	// 	"-i", "desktop",
	// 	"-c:v", "libx264",
	// 	"-preset", "ultrafast",
	// 	"-tune", "zerolatency",
	// 	"-crf", "25",
	// 	"-pix_fmt", "yuv420p",
	// 	"-an",
	// 	"-f", "h264",
	// 	"-",
	// }

	args := []string{
		"-video_size", "1920x1080", "-framerate", "30", "-f", "gdigrab", "-i", "desktop", "-c:v", "libx264", "-preset", "ultrafast", "-color_range", "2", "-f", "h264", "-",
	}

	filePath := "../logs/video.log"

	fmt.Println(args)

	cmd := exec.Command(command, args...)
	fmt.Println(cmd)

	outFile, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	// cmd.Stdout = outFile
	cmd.Stderr = outFile

	dataPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	return dataPipe, nil
}