package chat

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"gitverse.ru/udonetsm/aiserver/contentreader"
	"google.golang.org/api/iterator"
)

type FileManager interface {
	SendFile(ctx context.Context) (link, ctype string, err error)
	LisFiles(ctx context.Context) ([]string, error)
	Configure(ctx context.Context, contentreader contentreader.ContentReader) error
	DeleteFileByFilename(ctx context.Context, filename string) error
}

func (c *client) Configure(ctx context.Context, contentreader contentreader.ContentReader) error {
	if contentreader == nil {
		return fmt.Errorf("content reader is nil")
	}
	c.contentReader = contentreader
	return nil
}

func (c *client) contentTypeSupported(contentType string) bool {
	return !strings.Contains(contentType, "octet") && !strings.Contains(contentType, "zip")
}

func (c *client) SendFile(ctx context.Context) (link, ctype string, err error) {
	if c.contentReader == nil {
		return "", "", fmt.Errorf("client.SendFile(): content reader is nil")
	}
	read, err := c.contentReader.ReadContent(ctx)
	if err != nil {
		return "", "", fmt.Errorf("client.SendFile(): %w", err)
	}
	contentType, err := c.contentReader.DetectContentType(ctx, read)
	if err != nil {
		return "", "", fmt.Errorf("client.SendFile(): %w", err)
	}
	if !c.contentTypeSupported(contentType) {
		return "", "", fmt.Errorf("client.SendFile(): content type not supported")
	}
	name := uuid.NewString()
	opts := &genai.UploadFileOptions{
		DisplayName: name,
		MIMEType:    contentType,
	}
	genaiFile, err := c.Client.UploadFile(ctx, name, bytes.NewBuffer(read), opts)
	if err != nil {
		return "", "", fmt.Errorf("client.SendFile(): %w", err)
	}
	return genaiFile.URI, contentType, nil
}

func (c *client) LisFiles(ctx context.Context) ([]string, error) {
	fileInfos := make([]string, 0)
	iter := c.Client.ListFiles(ctx)
	for {
		file, err := iter.Next()
		if err == iterator.Done {
			if len(fileInfos) < 1 {
				return nil, fmt.Errorf("no files found for associated user")
			}
			return fileInfos, nil
		}
		if err != nil {
			return fileInfos, fmt.Errorf("%w", err)
		}
		fileInfos = append(fileInfos, file.Name)
	}
}

func (c *client) DeleteFileByFilename(ctx context.Context, filename string) error {
	c.logger.Infof("removing [%s]", filename)

	err := c.Client.DeleteFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	c.logger.Infof("removed [%s]", filename)
	return nil
}
