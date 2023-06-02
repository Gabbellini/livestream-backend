package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/streaming", wsEndpoint)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Process and convert the blob data using FFMPEG
		cmd := exec.Command(
			"FFMPEG",
			"-loglevel",
			"debug",
			"-f",
			"mjpeg",
			"-i",
			"pipe:0",
			"-c:v",
			"libx264",
			"-c:a",
			"copy",
			"-preset",
			"veryfast",
			"-f",
			"flv",
			"rtmp://localhost/live/stream",
		) // RTMP server URL

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Connect FFMPEG's stdin to the WebSocket connection
		cmd.Stdin = bytes.NewReader(data)

		// Start the FFMPEG command
		if err = cmd.Start(); err != nil {
			log.Println(err)
			return
		}

		// Wait for the FFMPEG command to finish
		err = cmd.Wait()
		if err != nil {
			fmt.Println("Error executing FFMPEG command:", err)
		}

	}
}
