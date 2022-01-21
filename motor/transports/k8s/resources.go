package k8s

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	api "go.mondoo.io/mondoo/cosmo/resources"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
)

func (t *Transport) ServerVersion() *version.Info {
	return t.d.ServerVersion
}

// discover api and resources that have a list method
func (t *Transport) SupportedResources() (*api.ApiResourceIndex, error) {
	// TODO: this should likely be cached
	return t.d.SupportedResourceTypes()
}

type ResourceResult struct {
	Name          string
	Kind          string
	ResourceType  *api.ApiResource // resource type that matched kind
	AllResources  []runtime.Object
	RootResources []runtime.Object
	Namespace     string
	AllNs         bool
}

func (t *Transport) Resources(kind string, name string) (*ResourceResult, error) {
	ctx := context.Background()
	ns := t.opts["namespace"]
	allNs := false
	if len(ns) == 0 {
		allNs = true
	}

	var err error
	var resourceObjects []runtime.Object

	// TODO: this should only apply for api calls
	resTypes, err := t.SupportedResources()
	if err != nil {
		return nil, err
	}
	log.Debug().Msg("completed querying resource types")

	if len(t.manifestFile) > 0 {
		log.Debug().Str("file", t.manifestFile).Msg("load resources from manifest files")
		var input io.Reader

		// if content is piped
		if t.manifestFile == "-" {
			input = os.Stdin
		} else {
			// return all resources from manifest
			filenames := []string{}

			fi, err := os.Stat(t.manifestFile)
			if err != nil {
				return nil, err
			}

			if fi.IsDir() {
				err = filepath.WalkDir(t.manifestFile, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if path != "." {
						return nil
					}

					// only load yaml files for now
					ext := filepath.Ext(path)
					if ext == "yaml" || ext == "yml" {
						filenames = append(filenames, path)
					} else {
						log.Debug().Str("file", path).Msg("ignore file")
					}

					return nil
				})
				if err != nil {
					return nil, err
				}
			} else {
				filenames = append(filenames, t.manifestFile)
			}

			buf := bytes.NewBuffer(nil)
			for _, filename := range filenames {
				f, _ := os.Open(filename) // Error handling elided for brevity.
				io.Copy(buf, f)           // Error handling elided for brevity.
				f.Close()
				buf.WriteString("---")
			}

			input = buf
		}

		resourceObjects, err = api.ResourcesFromManifest(input)
		if err != nil {
			return nil, errors.Wrap(err, "could not query resource objects")
		}
		log.Debug().Msgf("found %d resource objects", len(resourceObjects))
	} else {
		// return all resources for specified resource tpyes and namespace
		log.Debug().Msg("fetch all resource objects")
		resourceObjects, err = t.d.GetAllResources(ctx, resTypes, ns, allNs)
		if err != nil {
			return nil, errors.Wrap(err, "could not query resource objects")
		}
		log.Debug().Msgf("found %d resource objects", len(resourceObjects))
	}

	// find root nodes
	resType, rootResources, err := t.d.FilterResource(resTypes, resourceObjects, kind, name)

	return &ResourceResult{
		Name:          name,
		Kind:          kind,
		ResourceType:  resType,
		AllResources:  resourceObjects,
		RootResources: rootResources,
		Namespace:     ns,
		AllNs:         allNs,
	}, err
}
