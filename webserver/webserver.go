package main

import (
	"net/http"
	// "net/http/httputil"
	// "net/url"
	"log"
	"os"

	"bufio"
	// "strings"
	// "path/filepath"
	"bytes"
)

var (
	serverIP    string // IP address for WebSocket connection
	targetPort  string // Port number for WebSocket connection
	title       string // Title of the application to capture
)

var logFilePaths = map[string]string{
	"video":  "../logs/video.log",
	"audio":   "../logs/audio.log",
	"client":  "../logs/client.log",
	"server":   "../logs/server.log",
	"sign":  "../logs/sign.log",
}

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

	serverPort := "80"
	// streamTarget := "http://" + serverIP + ":" + targetPort
	
	// Serve the static files from the React build directory
	http.Handle("/", http.FileServer(http.Dir("../ui/build")))

	// Start the server
	addr := serverIP + ":" + serverPort
	log.Printf("Server listening on %s\n", addr)
	// log.Fatal(http.ListenAndServe(addr, nil))

	// // Handle requests for static files for the client page
	// http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	log.Printf("Client request: %s %s\n", r.Method, r.URL.Path)
	// 	http.FileServer(http.Dir("client")).ServeHTTP(w, r)
	// }))

	// Handle requests for static files for the admin page
	// http.Handle("/admin/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	log.Printf("Admin request: %s %s\n", r.Method, r.URL.Path)
	// 	http.StripPrefix("/admin/", http.FileServer(http.Dir("admin"))).ServeHTTP(w, r)
	// }))

	// proxyTarget, _ := url.Parse(streamTarget)
	// proxy := httputil.NewSingleHostReverseProxy(proxyTarget)

	// http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Printf("Stream request: %s %s\n", r.Method, r.URL.Path)
	// 	proxy.ServeHTTP(w, r)
	// })

	// Add a new handler for serving the log file content to the admin page
	http.HandleFunc("/admin/logs/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the log type from the request URL
		logType := strings.TrimPrefix(r.URL.Path, "/admin/logs/")
	
		log.Printf("Admin log request: %s %s (Log type: %s)\n", r.Method, r.URL.Path, logType)
	
		// Get the log file path based on the log type
		logFilePath, ok := logFilePaths[logType]
		if !ok {
			http.Error(w, "Invalid log type", http.StatusBadRequest)
			return
		}

		log.Println(logFilePath)

		absPath, err := filepath.Abs(logFilePath)
		if err != nil {
			http.Error(w, "Can't find file path", http.StatusBadRequest)
			return
		}

		// Open the log file
		file, err := os.Open(absPath)
		if err != nil {
			http.Error(w, "Error opening log file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Read the last 10 lines of the log file
		lines, err := readLastLines(file, 10)
		if err != nil {
			http.Error(w, "Error reading log file", http.StatusInternalServerError)
			return
		}

		// Set the content type as "text/plain" to display plain text content
		w.Header().Set("Content-Type", "text/plain")

		// Write the log file content to the response
		for _, line := range lines {
			_, err := w.Write([]byte(line + "\n"))
			if err != nil {
				http.Error(w, "Error writing response", http.StatusInternalServerError)
				return
			}
		}
	})

	// http.ListenAndServe(serverIP+":"+serverPort, nil)

	log.Fatal(http.ListenAndServe(addr, nil))
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

