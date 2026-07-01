package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		log.Println("Failed to bind to port 9999")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading:", err)
		return
	}
	req := strings.Split(string(buf[:n]), "\r\n")
	fmt.Println(req)
	method := strings.Split(req[0], " ")[0]
	path := strings.Split(req[0], " ")[1]
	headers := req[1:]

	var res string

	switch {
	case path == "/":
		res = getStatus(200, "OK") + "\r\n\r\n"
	case strings.HasPrefix(path, "/echo/"):
		params := path[6:]
		acceptEncodingIndex := slices.IndexFunc(headers, func(h string) bool { return strings.Contains(h, "Accept-Encoding: ") })
		if acceptEncodingIndex == -1 {
			res = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(params), params)
		} else {
			acceptedEncoding := strings.TrimPrefix(headers[acceptEncodingIndex], "Accept-Encoding: ")
			fmt.Println(acceptedEncoding)
			if strings.Contains(acceptedEncoding, "gzip") {
				var b bytes.Buffer
				enc := gzip.NewWriter(&b)
				_, err := enc.Write([]byte(params))
				if err != nil {
					fmt.Println("error of encoding data:", err.Error())
				}
				enc.Close()

				res = fmt.Sprintf("%s\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(b.String()), b.String())
			} else {
				res = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(params), params)
			}
		}
	case path == "/user-agent":
		userAgent := strings.Split(req[2], " ")[1]
		res = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(userAgent), userAgent)
	case strings.HasPrefix(path, "/files/"):
		dir := os.Args[2]
		filename := strings.TrimPrefix(path, "/files/")
		if  method == "GET" {
			data, err := os.ReadFile(dir + filename)
			if err != nil {
				res = getStatus(404, "Not Found") + "\r\n\r\n"
			} else {
				res = fmt.Sprintf("%s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(data), data)
			}
		} else if method == "POST" {
			body := strings.Trim(req[len(req)-1], "\x00")
			if err := os.WriteFile(dir + filename, []byte(body), 0644); err  != nil {
				log.Println("error creating a file", err.Error())
				return
			}
			res = getStatus(201, "Created") + "\r\n\r\n"
		}
	default:
		res = getStatus(404, "Not Found") + "\r\n\r\n"
	}

	_, err = conn.Write([]byte(res))
	if err != nil {
		log.Println("error on writing response:", err.Error())
	}
}

func getStatus(statusCode int, statusText string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s", statusCode, statusText)
}
