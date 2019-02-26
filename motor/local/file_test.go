package local_test

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"go.mondoo.io/mondoo/motor/local"

	"github.com/stretchr/testify/assert"
)

func TestFileResource(t *testing.T) {
	path := "/tmp/test"

	trans, err := local.New()
	assert.Nil(t, err)

	fs := trans.FS()
	f, err := fs.Open(path)
	assert.Nil(t, err)

	afutil := afero.Afero{Fs: fs}

	// create the file and set the content
	err = ioutil.WriteFile(path, []byte("hello world"), 0666)
	assert.Nil(t, err)

	if assert.NotNil(t, f) {
		assert.Equal(t, path, f.Name(), "they should be equal")
		c, err := afutil.ReadFile(f.Name())
		assert.Nil(t, err)
		assert.Equal(t, "hello world", string(c), "content should be equal")
	}
}
