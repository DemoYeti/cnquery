// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

package provider

import (
	"errors"
	"os"
	"strconv"

	"go.mondoo.com/cnquery/providers-sdk/v1/vault"

	"go.mondoo.com/cnquery/llx"
	"go.mondoo.com/cnquery/providers-sdk/v1/inventory"
	"go.mondoo.com/cnquery/providers-sdk/v1/plugin"
	"go.mondoo.com/cnquery/providers-sdk/v1/upstream"
	"go.mondoo.com/cnquery/providers/gcp/connection"
	"go.mondoo.com/cnquery/providers/gcp/resources"
)

type Service struct {
	runtimes         map[uint32]*plugin.Runtime
	lastConnectionID uint32
}

func Init() *Service {
	return &Service{
		runtimes:         map[uint32]*plugin.Runtime{},
		lastConnectionID: 0,
	}
}

// returns only the env vars that have a set value
func readEnvs(envs ...string) []string {
	vals := []string{}
	for i := range envs {
		val := os.Getenv(envs[i])
		if val != "" {
			vals = append(vals, val)
		}
	}

	return vals
}

// to be used by gcp/googleworkspace cmds, fetches the creds from either the env vars provided or from a flag in the provided cmd
func getGoogleCreds(credentialPath string, envs ...string) []byte {
	var credsPaths []string
	// env vars have precedence over the --credentials-path arg
	credsPaths = readEnvs(envs...)

	if credentialPath != "" {
		credsPaths = append(credsPaths, credentialPath)
	}

	for i := range credsPaths {
		path := credsPaths[i]

		serviceAccount, err := os.ReadFile(path)
		if err == nil {
			return serviceAccount
		}
	}
	return nil
}

func (s *Service) ParseCLI(req *plugin.ParseCLIReq) (*plugin.ParseCLIRes, error) {
	flags := req.Flags
	if flags == nil {
		flags = map[string]*llx.Primitive{}
	}

	if len(req.Args) != 2 {
		return nil, errors.New("missing argument, use `gcp project id` or `gcp organization id` or `gcp folder id`")
	}

	conf := &inventory.Config{
		Type:    req.Connector,
		Options: map[string]string{},
	}

	// custom flag parsing
	var credentialsPath string
	if x, ok := flags["credentials-path"]; ok && len(x.Value) != 0 {
		credentialsPath = string(x.Value)
	}

	envVars := []string{
		"GOOGLE_APPLICATION_CREDENTIALS",
		"GOOGLE_CREDENTIALS",
		"GOOGLE_CLOUD_KEYFILE_JSON",
		"GCLOUD_KEYFILE_JSON",
	}
	serviceAccount := getGoogleCreds(credentialsPath, envVars...)
	if serviceAccount != nil {
		conf.Credentials = append(conf.Credentials, &vault.Credential{
			Type:   vault.CredentialType_json,
			Secret: serviceAccount,
		})
	}

	switch req.Args[0] {
	case "org":
		conf.Options["organization-id"] = req.Args[1]
	case "project":
		conf.Options["project-id"] = req.Args[1]
	case "folder":
		conf.Options["folder-id"] = req.Args[1]
	}

	asset := inventory.Asset{
		Connections: []*inventory.Config{conf},
	}

	// parse discovery flags
	conf.Discover = &inventory.Discovery{
		Targets: []string{},
	}
	if x, ok := flags["discover"]; ok && len(x.Array) != 0 {
		for i := range x.Array {
			entry := string(x.Array[i].Value)
			conf.Discover.Targets = append(conf.Discover.Targets, entry)
		}
	}

	return &plugin.ParseCLIRes{Asset: &asset}, nil
}

func (s *Service) Connect(req *plugin.ConnectReq, callback plugin.ProviderCallback) (*plugin.ConnectRes, error) {
	if req == nil || req.Asset == nil {
		return nil, errors.New("no connection data provided")
	}

	conn, err := s.connect(req, callback)
	if err != nil {
		return nil, err
	}

	// We only need to run the detection step when we don't have any asset information yet.
	if req.Asset.Platform == nil {
		if err := s.detect(req.Asset, conn); err != nil {
			return nil, err
		}
	}

	// discovery assets for further scanning
	inventory, err := s.discover(conn)
	if err != nil {
		return nil, err
	}

	return &plugin.ConnectRes{
		Id:        conn.ID(),
		Name:      conn.Name(),
		Asset:     req.Asset,
		Inventory: inventory,
	}, nil
}

func (s *Service) connect(req *plugin.ConnectReq, callback plugin.ProviderCallback) (*connection.GcpConnection, error) {
	if len(req.Asset.Connections) == 0 {
		return nil, errors.New("no connection options for asset")
	}

	asset := req.Asset
	conf := asset.Connections[0]
	var conn *connection.GcpConnection
	var err error

	switch conf.Type {
	default:
		s.lastConnectionID++
		conn, err = connection.NewGcpConnection(s.lastConnectionID, asset, conf)
	}

	if err != nil {
		return nil, err
	}

	var upstream *upstream.UpstreamClient
	if req.Upstream != nil {
		upstream, err = req.Upstream.InitClient()
		if err != nil {
			return nil, err
		}
	}

	asset.Connections[0].Id = conn.ID()
	s.runtimes[conn.ID()] = &plugin.Runtime{
		Connection:     conn,
		Resources:      map[string]plugin.Resource{},
		Callback:       callback,
		HasRecording:   req.HasRecording,
		CreateResource: resources.CreateResource,
		Upstream:       upstream,
	}

	return conn, err
}

func (s *Service) detect(asset *inventory.Asset, conn *connection.GcpConnection) error {
	// TODO: adjust asset detection
	asset.Id = conn.Conf.Type
	asset.Name = conn.Conf.Host

	asset.Platform = &inventory.Platform{
		Name:   "gcp",
		Family: []string{"gcp"},
		Kind:   "api",
		Title:  "GCP Cloud",
	}

	// TODO: Add platform IDs
	asset.PlatformIds = []string{"//platformid.api.mondoo.app/runtime/gcp/"}
	return nil
}

func (s *Service) discover(conn *connection.GcpConnection) (*inventory.Inventory, error) {
	if conn.Conf.Discover == nil {
		return nil, nil
	}

	runtime, ok := s.runtimes[conn.ID()]
	if !ok {
		// no connection found, this should never happen
		return nil, errors.New("connection " + strconv.FormatUint(uint64(conn.ID()), 10) + " not found")
	}

	return resources.Discover(runtime)
}

func (s *Service) GetData(req *plugin.DataReq) (*plugin.DataRes, error) {
	runtime, ok := s.runtimes[req.Connection]
	if !ok {
		return nil, errors.New("connection " + strconv.FormatUint(uint64(req.Connection), 10) + " not found")
	}

	args := plugin.PrimitiveArgsToRawDataArgs(req.Args, runtime)

	if req.ResourceId == "" && req.Field == "" {
		res, err := resources.NewResource(runtime, req.Resource, args)
		if err != nil {
			return nil, err
		}

		rd := llx.ResourceData(res, res.MqlName()).Result()
		return &plugin.DataRes{
			Data: rd.Data,
		}, nil
	}

	resource, ok := runtime.Resources[req.Resource+"\x00"+req.ResourceId]
	if !ok {
		// Note: Since resources are internally always created, there are only very
		// few cases where we arrive here:
		// 1. The caller is wrong. Possibly a mixup with IDs
		// 2. The resource was loaded from a recording, but the field is not
		// in the recording. Thus the resource was never created inside the
		// plugin. We will attempt to create the resource and see if the field
		// can be computed.
		if !runtime.HasRecording {
			return nil, errors.New("resource '" + req.Resource + "' (id: " + req.ResourceId + ") doesn't exist")
		}

		args, err := runtime.ResourceFromRecording(req.Resource, req.ResourceId)
		if err != nil {
			return nil, errors.New("attempted to load resource '" + req.Resource + "' (id: " + req.ResourceId + ") from recording failed: " + err.Error())
		}

		resource, err = resources.CreateResource(runtime, req.Resource, args)
		if err != nil {
			return nil, errors.New("attempted to create resource '" + req.Resource + "' (id: " + req.ResourceId + ") from recording failed: " + err.Error())
		}
	}

	return resources.GetData(resource, req.Field, args), nil
}

func (s *Service) StoreData(req *plugin.StoreReq) (*plugin.StoreRes, error) {
	return nil, errors.New("not yet implemented")
}
