package vsphere

import (
	"context"
	"errors"
	"net/url"

	"github.com/spf13/afero"
	"github.com/vmware/govmomi"
	"go.mondoo.io/mondoo/motor/transports"
	"go.mondoo.io/mondoo/motor/transports/fsutil"
)

func vSphereURL(hostname string, port string, user string, password string) (*url.URL, error) {
	host := hostname
	if len(port) > 0 {
		host = hostname + ":" + port
	}

	u, err := url.Parse("https://" + host + "/sdk")
	if err != nil {
		return nil, err
	}
	u.User = url.UserPassword(user, password)
	return u, nil
}

func New(endpoint *transports.TransportConfig) (*Transport, error) {
	// derive vsphere connection url from Transport Config
	vsphereUrl, err := vSphereURL(endpoint.Host, endpoint.Port, endpoint.User, endpoint.Password)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := govmomi.NewClient(ctx, vsphereUrl, true)
	if err != nil {
		return nil, err
	}

	return &Transport{
		client: client,
	}, nil
}

type Transport struct {
	client *govmomi.Client
}

func (t *Transport) RunCommand(command string) (*transports.Command, error) {
	return nil, errors.New("vsphere does not implement RunCommand")
}

func (t *Transport) FileInfo(path string) (transports.FileInfoDetails, error) {
	return transports.FileInfoDetails{}, errors.New("vsphere does not implement FileInfo")
}

func (t *Transport) FS() afero.Fs {
	return &fsutil.NoFs{}
}

func (t *Transport) Close() {}

func (t *Transport) Capabilities() transports.Capabilities {
	return transports.Capabilities{}
}

func (t *Transport) Client() *govmomi.Client {
	return t.client
}
