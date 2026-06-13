package dhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
	"sync"
)

type File struct {
	Name        string
	Reader      io.Reader
	ContentType string
}

func NewFile(name string, reader io.Reader) *File {
	return &File{
		Name:        name,
		Reader:      reader,
		ContentType: detectContentType(name),
	}
}

func detectContentType(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		ext := name[i:]
		if t := mime.TypeByExtension(ext); t != "" {
			return t
		}
	}
	return "application/octet-stream"
}

func (c *Client) SendFile(ctx context.Context, name string, reader io.Reader, msg *Message) ([]*Response, error) {
	file := &File{
		Name:        name,
		Reader:      reader,
		ContentType: detectContentType(name),
	}
	return c.SendFiles(ctx, msg, file)
}

func (c *Client) SendFiles(ctx context.Context, msg *Message, files ...*File) ([]*Response, error) {
	if msg == nil {
		msg = &Message{}
	}

	body, contentType, err := buildMultipart(msg, files)
	if err != nil {
		return nil, err
	}

	var (
		responses []*Response
		mu        sync.Mutex
		errs      []error
		wg        sync.WaitGroup
	)

	for _, url := range c.urls {
		wg.Add(1)
		go func(webhookURL string) {
			defer wg.Done()
			resp, err := c.doPost(ctx, webhookURL, contentType, body.Bytes())
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				c.fireError(err)
				return
			}
			mu.Lock()
			responses = append(responses, resp)
			mu.Unlock()
			c.fireSuccess(resp)
		}(url)
	}

	wg.Wait()

	if len(errs) > 0 {
		return responses, fmt.Errorf("dhook: %w", errs[0])
	}
	return responses, nil
}

func buildMultipart(msg *Message, files []*File) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, "", err
	}

	jsonPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="payload_json"`},
		"Content-Type":       {"application/json"},
	})
	if err != nil {
		return nil, "", err
	}
	if _, err := jsonPart.Write(jsonData); err != nil {
		return nil, "", err
	}

	for i, file := range files {
		header := textproto.MIMEHeader{
			"Content-Disposition": {fmt.Sprintf(`form-data; name="files[%d]"; filename="%s"`, i, file.Name)},
		}
		if file.ContentType != "" {
			header.Set("Content-Type", file.ContentType)
		}

		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(part, file.Reader); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return &buf, writer.FormDataContentType(), nil
}
