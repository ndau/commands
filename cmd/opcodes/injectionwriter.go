package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// InjectionWriter is an io.WriteCloser that searches for start and end markers in a file
// and injects the contents of its write calls between those markers.
// Close() MUST be called!
type InjectionWriter struct {
	filename    string
	startmarker string
	endmarker   string
	input       *bytes.Buffer
	output      *bytes.Buffer
	endline     string
}

// NewInjectionWriter builds an InjectionWriter and loads the specified
// file, looking for the given markers
func NewInjectionWriter(fn, start, end string) (*InjectionWriter, error) {
	iw := &InjectionWriter{
		filename:    fn,
		startmarker: start,
		endmarker:   end,
		input:       &bytes.Buffer{},
		output:      &bytes.Buffer{},
	}
	// open the file and read it all into the input buffer
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(iw.input, f)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}

	// now scan the buffer looking for the start point, and copy everything up to
	// the start point into the output buffer
	skipping := false
	for {
		line, err := iw.input.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !skipping {
			iw.output.WriteString(line)
		}
		if strings.HasPrefix(line, iw.startmarker) {
			skipping = true
		}
		if strings.HasPrefix(line, iw.endmarker) {
			iw.endline = line
			break
		}
	}

	// if we didn't find the markers we shouldn't do this
	if !skipping {
		return nil, errors.New("start marker not found")
	}
	if iw.endline == "" {
		return nil, errors.New("end marker not found")
	}

	return iw, nil
}

// SetMarkers lets an InjectionWriter's client change its set of markers -- this
// only works before the first Write call.
func (iw *InjectionWriter) SetMarkers(start, end string) {
	iw.startmarker = start
	iw.endmarker = end
}

// Write implements the io.Writer interface by just writing the
// contents to the output buffer
func (iw *InjectionWriter) Write(b []byte) (int, error) {
	return iw.output.Write(b)

}

// Close implements the io.Closer interface
func (iw *InjectionWriter) Close() error {
	// copy the rest of the input buffer to the output buffer
	iw.output.WriteString(iw.endline)
	_, err := io.Copy(iw.output, iw.input)
	if err != nil {
		return err
	}
	// now write it to a temp file
	tmpfile, err := ioutil.TempFile("", "injection")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name()) // clean up if necessary

	_, err = io.Copy(tmpfile, iw.output)
	if err != nil {
		return err
	}
	err = tmpfile.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tmpfile.Name(), iw.filename)
	return err
}

// assert that we have properly implemented io.WriteCloser
var _ io.WriteCloser = (*InjectionWriter)(nil)
