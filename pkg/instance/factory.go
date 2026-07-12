package instance

type DefaultFactory struct {
	sdk           SDK
	executor      Executor
	browserOpener BrowserOpener
}

func NewDefaultFactory(
	sdk SDK,
	executor Executor,
	browserOpener BrowserOpener,
) *DefaultFactory {
	return &DefaultFactory{
		sdk:           sdk,
		executor:      executor,
		browserOpener: browserOpener,
	}
}

func (f *DefaultFactory) Create(ctxID string) Instance {
	return NewInstance(ctxID, f.executor, f.sdk, f.browserOpener)
}
