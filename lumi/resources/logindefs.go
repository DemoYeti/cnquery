package resources

import (
	"errors"
	"strings"

	"go.mondoo.io/mondoo/lumi"
	"go.mondoo.io/mondoo/lumi/resources/logindefs"
)

func (s *lumiLogindefs) init(args *lumi.Args) (*lumi.Args, Logindefs, error) {
	if x, ok := (*args)["path"]; ok {
		path, ok := x.(string)
		if !ok {
			return nil, nil, errors.New("Wrong type for 'path' in logindefs initialization, it must be a string")
		}

		f, err := s.MotorRuntime.CreateResource("file", "path", path)
		if err != nil {
			return nil, nil, err
		}
		(*args)["file"] = f
		delete(*args, "path")
	}

	return args, nil, nil
}

const defaultLoginDefsConfig = "/etc/login.defs"

func (s *lumiLogindefs) id() (string, error) {
	r, err := s.File()
	if err != nil {
		return "", err
	}
	return r.Path()
}

func (s *lumiLogindefs) GetFile() (File, error) {
	f, err := s.MotorRuntime.CreateResource("file", "path", defaultLoginDefsConfig)
	if err != nil {
		return nil, err
	}
	return f.(File), nil
}

// borrowed from ssh resource
func (s *lumiLogindefs) GetContent(file File) (string, error) {
	// TODO: this can be heavily improved once we do it right, since this is constantly
	// re-registered as the file changes
	err := s.MotorRuntime.WatchAndCompute(file, "content", s, "content")
	if err != nil {
		return "", err
	}

	return file.Content()
}

func (s *lumiLogindefs) GetParams(content string) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	params := logindefs.Parse(strings.NewReader(content))

	for k, v := range params {
		res[k] = v
	}

	return res, nil
}
