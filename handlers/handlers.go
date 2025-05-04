package handlers

import (
	"context"
	"strings"
	"sync"
	"time"

	ai_ "gitverse.ru/udonetsm/aiserver/aipack"
	"gitverse.ru/udonetsm/aiserver/chat"
	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/contentreader"
	"gitverse.ru/udonetsm/aiserver/historystorage"
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
	llmConfig, err := configs.NewLLMConfig(request.APIKey, request.ModelVersion)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	client, err := chat.NewClient(ctx, llmConfig, h.logger, h.semaphoreConfig)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	model := client.Generative()
	chat := model.Start(h.logger)
	chat.SaveClient(client)

	err = h.sessionStorage.NewSession(request.APIKey, chat)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	h.logger.Infof("session created for %s", request.APIKey)
	return &ai_.Status{Success: true, Message: "created"}, nil
}

func (h *handlers) TransmitText(ctx context.Context, request *ai_.TextWithPayload) (*ai_.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(request.Payload.CTXLen))
	defer cancel()
	chat, err := h.sessionStorage.SessionByKey(request.Payload.APIKey)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.Payload.APIKey)
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
		h.logger.Infof("%v for %s", err, request.Payload.APIKey)
		return &ai_.Status{Message: builder.String()}, err
	}
	wg.Wait()
	h.logger.Infof("ok message for %s", request.Payload.APIKey)
	return &ai_.Status{Success: true, Message: builder.String()}, nil
}

func (h *handlers) DeleteFiles(ctx context.Context, request *ai_.Payload) (*ai_.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(request.CTXLen))
	defer cancel()
	chat, err := h.sessionStorage.SessionByKey(request.APIKey)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	client := chat.Client()

	fileList, err := client.FileManager().LisFiles(ctx)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
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
				h.logger.Infof("%v for %s", err, request.APIKey)
				return
			default:
				err := client.FileManager().DeleteFileByFilename(ctx, fileName)
				if err != nil {
					builder.WriteString(err.Error() + "\n")
					h.logger.Infof("%v for %s", err, request.APIKey)
					return
				}
			}
		}()
	}
	wg.Wait()
	h.logger.Infof("deleted files for %s", request.APIKey)
	return &ai_.Status{Success: true, Message: "deleted:\n" + builder.String()}, nil
}

func (h *handlers) TransmitFiles(ctx context.Context, request *ai_.FilesWithPayload) (*ai_.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(request.Payload.CTXLen))
	defer cancel()
	chat, err := h.sessionStorage.SessionByKey(request.Payload.APIKey)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.Payload.APIKey)
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
					h.logger.Infof("%v for %s", err, request.Payload.APIKey)
					return
				default:
					err := client.FileManager().Configure(ctx, contentreader.NewContentReader(h.logger, configs.NewFileReaderConfig(absPath)))
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						h.logger.Infof("%v for %s", err, request.Payload.APIKey)
						return
					}
					link, ctype, err := client.FileManager().SendFile(ctx)
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						h.logger.Infof("%v for %s", err, request.Payload.APIKey)
						return
					}
					indx, err := chat.HistManager().AddMessageToHistory(ctx, link, "user", ctype)
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						h.logger.Infof("%v for %s", err, request.Payload.APIKey)
						return
					}
					err = chat.HistManager().SaveFileIndex(indx)
					if err != nil {
						builder.WriteString(err.Error() + "\n")
						h.logger.Infof("%v for %s", err, request.Payload.APIKey)
						return
					}
					h.logger.Infof("pinned file [%s]", absPath)
					return
				}
			}()
		}
	}
	wg.Wait()
	h.logger.Infof("uploaded files for %s", request.Payload.APIKey)
	return &ai_.Status{Success: true, Message: builder.String()}, nil
}

func (h *handlers) DeleteChat(ctx context.Context, request *ai_.Payload) (*ai_.Status, error) {
	chat, err := h.sessionStorage.SessionByKey(request.APIKey)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	err = chat.HistManager().ClearHistory(ctx)
	if err != nil {
		h.logger.Infof("%v for %s", err, request.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	return &ai_.Status{Success: true, Message: "cleared"}, nil
}

func (h *handlers) SaveHistory(ctx context.Context, payload *ai_.Payload) (*ai_.Status, error) {
	chat, err := h.sessionStorage.SessionByKey(payload.APIKey)
	if err != nil {
		h.logger.Infof("%v for %s", err, payload.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	historystorageConfig := configs.NewHistoryStorageConfig(payload.APIKey)
	err = historystorageConfig.Configure(payload.HistorySource)
	if err != nil {
		h.logger.Infof("%v for %s", err, payload.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}

	historyStorage := historystorage.NewHistoryStorage(h.logger, historystorageConfig)
	err = historyStorage.Configure(ctx)
	if err != nil {
		h.logger.Infof("%v for %s", err, payload.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	err = chat.HistManager().SaveHistory(ctx, historyStorage)
	if err != nil {
		h.logger.Infof("%v for %s", err, payload.APIKey)
		return &ai_.Status{Message: err.Error()}, err
	}
	return &ai_.Status{Success: true, Message: "OK"}, err
}

func NewHandlers(logger logger.Logger, sessionStorage sessions.SessionStorage, semaphoreConfig configs.SemaphoreConfig) Handlers {
	return &handlers{
		sessionStorage:  sessionStorage,
		logger:          logger,
		semaphoreConfig: semaphoreConfig,
	}
}
