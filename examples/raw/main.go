package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	conn, err := net.Dial("tcp", "0.0.0.0:1491")
	if err != nil {
		fmt.Printf("error: %v+", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	query := fmt.Sprintf("START %s %s", "search", "SecretPassword")
	n, err := conn.Write([]byte(query + "\r\n"))
	if err != nil {
		fmt.Printf("error: %v+\n", err)
	}

	fmt.Printf("Start: %d\n", n)
	buffer := bytes.Buffer{}

	goto selam

	//-------------------------------------------------
ping:
	n, err = conn.Write([]byte("PING" + "\r\n"))
	if err != nil {
		fmt.Printf("error: %v+\n", err)
	}

	fmt.Printf("Ping: %d\n", n)

	goto selam
	//-------------------------------------------------
search:
	n, err = conn.Write([]byte("QUERY messages user:0dcde3a6 \"valerian saliou\" LIMIT(10)" + "\r\n"))
	if err != nil {
		fmt.Printf("error: %v+\n", err)
	}

	fmt.Printf("Ping: %d\n", n)

	goto selam

	//-------------------------------------------------
quit:
	n, err = conn.Write([]byte("QUIT" + "\r\n"))
	if err != nil {
		fmt.Printf("error: %v+\n", err)
	}

	fmt.Printf("Quit: %d\n", n)

	goto selam
	//-------------------------------------------------

selametle:
	fmt.Println("selametle")

selam:

	var marker string

	buffer.Reset()
	for {
		endEOF := false
		for {
			line, isPrefix, err := reader.ReadLine()
			buffer.Write(line)
			if err != nil {
				if err == io.EOF {
					conn.Close()
				}
				fmt.Println("end eof")
				endEOF = true
			}
			if !isPrefix {
				fmt.Println("isPrefix geldi")
				break
			}
		}

		if endEOF {
			break
		}

		str := buffer.String()
		fmt.Printf("Str: %s\n", str)
		if strings.HasPrefix(str, "ERR ") {
			fmt.Printf("error: %v+\n", err)
			panic(err)
		}

		if strings.HasPrefix(str, "STARTED ") {

			ss := strings.FieldsFunc(str, func(r rune) bool {
				if unicode.IsSpace(r) || r == '(' || r == ')' {
					return true
				}
				return false
			})
			bufferSize, err := strconv.Atoi(ss[len(ss)-1])
			if err != nil {
				fmt.Printf("Unable to parse STARTED response: %s", str)
				panic(err)
			}

			fmt.Printf("Buffersize: %d\n", bufferSize)
			goto ping
			//break
		} else if strings.HasPrefix(str, "PONG") {
			goto search
			//break
		} else if strings.HasPrefix(str, "PENDING") {
			// Durmak yok yola devam
			marker = strings.Replace(str, "PENDING ", "", 1)
			fmt.Printf("Follow Marker: %s\n", marker)

			//break
		} else if marker != "" && strings.HasPrefix(str, "EVENT QUERY "+marker) {
			// Durmak yok yola devam
			fmt.Printf("Search Results: %s\n", strings.Split(strings.Replace(str, "EVENT QUERY "+marker, "", 1), " "))

			marker = ""
			goto quit
			//break
		} else if strings.HasPrefix(str, "ENDED") {
			goto selametle
			//break
		}

		buffer.Reset()
	}

}
