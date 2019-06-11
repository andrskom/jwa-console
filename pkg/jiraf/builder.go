package jiraf

import (
	"github.com/andygrunwald/go-jira"

	"github.com/andrskom/jwa-console/pkg/creds"
)

type Factory struct {
	credsComponent *creds.Component
}

func NewFactory(credsComponent *creds.Component) *Factory {
	return &Factory{credsComponent: credsComponent}
}

func (b *Factory) GetClient() (*jira.Client, error) {
	model, err := b.credsComponent.Get()
	if err != nil {
		return nil, err
	}
	return BuildByCredsModel(model)
}

func BuildByCredsModel(model *creds.Model) (*jira.Client, error) {
	tp := jira.BasicAuthTransport{
		Username: model.Username,
		Password: model.Password,
	}

	return jira.NewClient(tp.Client(), model.Addr)
}
