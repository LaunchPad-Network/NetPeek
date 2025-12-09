package bird

import (
	"bytes"
	"io"
	"net"

	"github.com/lfcypo/viperx"
)

const MAX_LINE_SIZE = 1024

// Check if a byte is character for number
func isNumeric(b byte) bool {
	return b >= byte('0') && b <= byte('9')
}

// Read a line from bird socket, removing preceding status number, output it.
// Returns if there are more lines.
func birdReadln(bird io.Reader, w io.Writer) bool {
	// Read from socket byte by byte, until reaching newline character
	c := make([]byte, MAX_LINE_SIZE)
	pos := 0
	for {
		// Leave one byte for newline character
		if pos >= MAX_LINE_SIZE-1 {
			break
		}
		_, err := bird.Read(c[pos : pos+1])
		if err != nil {
			w.Write([]byte(err.Error()))
			return false
		}
		if c[pos] == byte('\n') {
			break
		}
		pos++
	}

	c = c[:pos+1]
	c[pos] = '\n'
	// print(string(c[:]))

	// Remove preceding status number, different situations
	if pos > 4 && isNumeric(c[0]) && isNumeric(c[1]) && isNumeric(c[2]) && isNumeric(c[3]) {
		// There is a status number at beginning, remove first 5 bytes
		if w != nil && pos > 6 {
			pos = 5
			w.Write(c[pos:])
		}
		return c[0] != byte('0') && c[0] != byte('8') && c[0] != byte('9')
	} else {
		if w != nil {
			w.Write(c[1:])
		}
		return true
	}
}

// Write a command to a bird socket
func birdWriteln(bird io.Writer, s string) {
	bird.Write([]byte(s + "\n"))
}

func getBirdSocket() (io.ReadWriter, error) {
	bird, err := net.Dial("unix", viperx.GetString("bird.socket", "/var/run/bird/bird.ctl"))
	if err != nil {
		return nil, err
	}
	return bird, nil
}

func CallBirdRestricted(query string, output io.Writer) error {
	bird, err := getBirdSocket()
	if err != nil {
		return err
	}

	// Read initial greeting
	birdReadln(bird, nil)

	// Send restrict command
	birdWriteln(bird, "restrict")
	var restrictedConfirmation bytes.Buffer
	birdReadln(bird, &restrictedConfirmation)

	// Send actual query
	birdWriteln(bird, query)
	// Read all output lines
	for birdReadln(bird, output) {
	}
	return nil
}

func CallBirdUnrestricted(query string, output io.Writer) error {
	bird, err := getBirdSocket()
	if err != nil {
		return err
	}

	// Read initial greeting
	birdReadln(bird, nil)

	// Send actual query
	birdWriteln(bird, query)
	// Read all output lines
	for birdReadln(bird, output) {
	}
	return nil
}
