package instance

type DefaultFactory struct {
	sdk      SDK
	Executor Executor
}

func NewDefaultFactory(
	sdk SDK,
	executor Executor,
) *DefaultFactory {
	return &DefaultFactory{
		sdk:      sdk,
		Executor: executor,
	}
}

func (f *DefaultFactory) Create(ctxID string) Instance {
	return NewInstance(ctxID, f.Executor, f.sdk)
}
