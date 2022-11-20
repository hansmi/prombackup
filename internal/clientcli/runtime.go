package clientcli

import "github.com/hansmi/prombackup/api"

type Runtime struct {
	NewClient func() (api.Interface, error)
}

func (r *Runtime) WithClient(fn func(api.Interface) error) error {
	cl, err := r.NewClient()
	if err != nil {
		return err
	}

	return fn(cl)
}
