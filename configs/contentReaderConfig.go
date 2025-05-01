package configs

type contentReaderConfig struct {
	source string
}

type ContentReaderConfig interface {
	ContentSource() string
}

func (frc *contentReaderConfig) ContentSource() string {
	return frc.source
}

func NewFileReaderConfig(source string) ContentReaderConfig {
	return &contentReaderConfig{source: source}
}
