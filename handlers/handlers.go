package handlers

import (
	"context"
	"strings"
	"sync"
	"time"

	ai_ "gitverse.ru/udonetsm/aiserver/aipack"
	"gitverse.ru/udonetsm/aiserver/chat"
	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/logger"
	"gitverse.ru/udonetsm/aiserver/semaphore"
	"gitverse.ru/udonetsm/aiserver/sessions"
)

type handlers struct {
	logger          logger.Logger
	sessionStorage  sessions.SessionStorage
	semaphoreConfig configs.SemaphoreConfig
	ai_.TransmitServiceServer
}

type Handlers interface {
	ai_.TransmitServiceServer
}

func (h *handlers) CreateSession(ctx context.Context, request *ai_.Payload) (*ai_.Status, error) {
	llmConfig, err := configs.NewLLMConfig(h.logger, request.APIKey, request.ModelVersion)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	client, err := chat.NewClient(ctx, llmConfig, h.logger, h.semaphoreConfig)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	model := client.Generative()
	chat := model.Start(h.logger)
	chat.SetClient(client)

	err = h.sessionStorage.NewSession(request.APIKey, chat)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	return &ai_.Status{Success: true, Message: "created"}, nil
}

func (h *handlers) TransmitText(ctx context.Context, request *ai_.TextWithPayload) (*ai_.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(request.Payload.CTXLen))
	defer cancel()
	chat, err := h.sessionStorage.SessionByKey(request.Payload.APIKey)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	builder := &strings.Builder{}
	response := make(chan string)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for resp := range response {
			builder.WriteString(resp)
		}
	}()
	err = chat.SendMessage(ctx, request.Text.Text, response)
	if err != nil {
		h.logger.Info(err)
		return &ai_.Status{Message: err.Error()}, err
	}
	wg.Wait()
	return &ai_.Status{Success: true, Message: builder.String()}, nil
}

func (h *handlers) DeleteFiles(ctx context.Context, request *ai_.Payload) (*ai_.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(request.CTXLen))
	defer cancel()
	chat, err := h.sessionStorage.SessionByKey(request.APIKey)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	client := chat.Client()

	fileList, err := client.LisFiles(ctx)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}

	wg := sync.WaitGroup{}
	semaphore := semaphore.NewSemaphore(h.semaphoreConfig)
	builder := strings.Builder{}
	defer builder.Reset()
	for _, fileName := range fileList {
		wg.Add(1)
		go func() {
			semaphore.Acquire()
			defer semaphore.Release()

			defer wg.Done()

			defer time.Sleep(time.Second)
			select {
			case <-ctx.Done():
				builder.WriteString(ctx.Err().Error() + "\n")
				return
			default:
				err := client.DeleteFileByFilename(ctx, fileName)
				if err != nil {
					builder.WriteString(err.Error() + "\n")
					return
				}
			}
		}()
	}
	wg.Wait()
	return &ai_.Status{Success: true, Message: "deleted:\n" + builder.String()}, nil
}

func (h *handlers) TransmitFiles(ctx context.Context, request *ai_.FilesWithPayload) (*ai_.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(request.Payload.CTXLen))
	defer cancel()
	chat, err := h.sessionStorage.SessionByKey(request.Payload.APIKey)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	semaphore := semaphore.NewSemaphore(h.semaphoreConfig)
	builder := strings.Builder{}
	defer builder.Reset()
	wg := sync.WaitGroup{}
	client := chat.Client()
	for _, absPath := range request.Files.Files {
		select {
		case <-ctx.Done():
			return &ai_.Status{Message: ctx.Err().Error()}, ctx.Err()
		default:
			wg.Add(1)
			go func() {
				semaphore.Acquire()
				defer semaphore.Release()

				defer wg.Done()

				defer time.Sleep(time.Second)
				select {
				case <-ctx.Done():
					builder.WriteString(ctx.Err().Error() + "\n")
					return
				default:
					link, ctype, err := client.SendFile(ctx, absPath)
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						return
					}
					indx, err := chat.HistManager().AddMessageToHistory(ctx, link, "user", ctype)
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						return
					}
					err = chat.HistManager().SaveFileIndex(indx)
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						return
					}
					h.logger.Infof("pinned file [%s]", absPath)
					return
				}
			}()
		}
	}
	wg.Wait()
	return &ai_.Status{Success: true, Message: builder.String()}, nil
}

func (h *handlers) DeleteChat(ctx context.Context, request *ai_.Payload) (*ai_.Status, error) {
	chat, err := h.sessionStorage.SessionByKey(request.APIKey)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	err = chat.HistManager().ClearHistory(ctx)
	if err != nil {
		return &ai_.Status{Message: err.Error()}, err
	}
	return &ai_.Status{Success: true, Message: "cleared"}, nil
}

func NewHandlers(logger logger.Logger, sessionStorage sessions.SessionStorage, semaphoreConfig configs.SemaphoreConfig) Handlers {
	return &handlers{sessionStorage: sessionStorage, logger: logger, semaphoreConfig: semaphoreConfig}
}
