package instance

type DefaultFactory struct {
	sdk           SDK
	executor      Executor
	browserOpener BrowserOpener
	imageReader   ImageReader
}

func NewDefaultFactory(
	sdk SDK,
	executor Executor,
	browserOpener BrowserOpener,
	imageReader ImageReader,
) *DefaultFactory {
	return &DefaultFactory{
		sdk:           sdk,
		executor:      executor,
		browserOpener: browserOpener,
		imageReader:   imageReader,
	}
}

func (f *DefaultFactory) Create(ctxID string) Instance {
	return NewInstance(ctxID, f.executor, f.sdk, f.browserOpener, f.imageReader)
}
