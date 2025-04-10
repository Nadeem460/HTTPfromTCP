package response

import (
	"fmt"
	"html/template"
	"httpfromtcp/internal/headers"
	"io"
	"net/http"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const (
	WriterStateStatusLine = iota
	WriterStateHeaders
	WriterStateBody
)

// Define the template structure for dynamic content
const tmpl = `<html>
	<head>
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>{{.Heading}}</h1>
		<p>{{.Message}}</p>
	</body>
</html>`

const lenTmpl = len(tmpl) - len("{{.Title}}") - len("{{.Heading}}") - len("{{.Message}}")

// Struct to hold data for dynamic population
type PageData struct {
	Title   string
	Heading string
	Message string
}

func (p *PageData) ContentLength() int {
	return lenTmpl + len(p.Title) + len(p.Heading) + len(p.Message)
}

type Writer struct {
	io.Writer
	writerState int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer:      w,
		writerState: WriterStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	// Check if the writer is in the correct state
	if w.writerState != WriterStateStatusLine {
		return fmt.Errorf("incorrect writer state, should write status line first")
	}

	// Write the status line
	var reasonPhrase string
	if statusCode == StatusOK || statusCode == StatusBadRequest || statusCode == StatusInternalServerError {
		reasonPhrase = http.StatusText(int(statusCode))
	} else {
		reasonPhrase = "" //TODO: MAY NEED TO CHANGE TO SPACE to follow the HTTP spec
	}

	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	if err == nil {
		// Set the writer state to headers after writing the status line
		w.writerState = WriterStateHeaders
	}
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	// Check if the writer is in the correct state
	if w.writerState != WriterStateHeaders {
		return fmt.Errorf("incorrect writer state, should write headers second")
	}

	// Write the headers
	for key, value := range headers {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	if err == nil {
		// Set the writer state to body after writing the headers
		w.writerState = WriterStateBody
	}
	return err
}

func (w *Writer) WriteBody(data PageData) error {
	// Check if the writer is in the correct state
	if w.writerState != WriterStateBody {
		return fmt.Errorf("incorrect writer state, should write body last")
	}

	// Create a new template and parse it
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		return err
	}

	// Reset the writer state to status line after writing the body
	w.writerState = WriterStateStatusLine

	// Write the populated template to the custom writer (e.g., http.ResponseWriter)
	return t.Execute(w, data)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen), // fmt.Sprintf("%d", contentLen) is generally prefered but strconv.Itoa is faster
		"Content-Type":   "text/html",
		"Connection":     "close",
	}
}
