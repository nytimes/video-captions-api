package providers

import "github.com/NYTimes/video-captions-api/database"

type trint struct {
}

func (t *trint) DispatchJob(_ *database.Job) error {
	panic("not implemented") // TODO: Implement
}

func (t *trint) Download(_ *database.Job, _ string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (t *trint) GetProviderJob(_ *database.Job) (*database.ProviderJob, error) {
	panic("not implemented") // TODO: Implement
}

func (t *trint) GetName() string {
	panic("not implemented") // TODO: Implement
}

func (t *trint) CancelJob(_ *database.Job) (bool, error) {
	panic("not implemented") // TODO: Implement
}
