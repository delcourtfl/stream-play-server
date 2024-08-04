package main

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"os"
)

var (
	serverIP   string // IP address for WebSocket connection
	targetPort string // Port number for WebSocket connection
	// title      string // Title of the application to capture
)

// var logFilePaths = map[string]string{
// 	"video":  "../logs/video.log",
// 	"audio":  "../logs/audio.log",
// 	"client": "../logs/client.log",
// 	"server": "../logs/server.log",
// 	"sign":   "../logs/sign.log",
// }

/**
 * main is the main entry point of the Webserver program. It sets up HTTP handlers to serve static files,
 * handle WebSocket connections, and provide log file content to the admin page.
 */
func main() {
	// Get IP, PORT from the args
	args := os.Args[1:] // Exclude the first argument, which is the program name
	if len(args) < 2 {
		log.Println("Please provide both IP address, port as arguments")
		return
	}

	serverIP = args[0]
	targetPort = args[1]
	log.Println(serverIP)
	log.Println(targetPort)

	// Start the server
	addr := serverIP + ":" + targetPort
	addrLocal := "localhost:" + targetPort
	log.Printf("Server listening on %s\n", addr)

	// Handle requests for static files for the client page
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Client request: %s %s\n", r.Method, r.URL.Path)

		// Set headers to prevent caching
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		http.FileServer(http.Dir("./")).ServeHTTP(w, r)
	}))

	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	log.Fatal(http.ListenAndServe(addrLocal, nil))
}

/**
 * readLastLines reads the last 'n' lines from an open file, handling various line endings.
 *
 * @param file An open os.File representing the log file.
 * @param n The number of lines to read from the end of the file.
 * @return A slice of strings containing the last 'n' lines from the file.
 */
func readLastLines(file *os.File, n int) ([]string, error) {
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)

	// Define a custom split function that handles both \n and \r\n line endings
	splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
			// Found a line ending
			return i + 1, data[:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		// Request more data
		return 0, nil, nil
	}

	scanner.Split(splitFunc)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	startIdx := len(lines) - n
	if startIdx < 0 {
		startIdx = 0
	}

	return lines[startIdx:], nil
}
