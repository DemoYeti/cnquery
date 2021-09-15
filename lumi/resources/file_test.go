package resources_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mondoo.io/mondoo/lumi"
)

const passwdContent = `root:x:0:0::/root:/bin/bash
chris:x:1000:1001::/home/chris:/bin/bash
christopher:x:1000:1001::/home/christopher:/bin/bash
chris:x:1002:1003::/home/chris:/bin/bash
bin:x:1:1::/:/usr/bin/nologin
`

func TestResource_File(t *testing.T) {
	runSimpleTests(t, []simpleTest{
		{
			"file(\"/etc/passwd\").exists",
			0, true,
		},
		{
			"file(\"/etc/passwd\").basename",
			0, "passwd",
		},
		{
			"file(\"/etc/passwd\").dirname",
			0, "/etc",
		},
		{
			"file(\"/etc/passwd\").size",
			0, int64(193),
		},
		{
			"file(\"/etc/passwd\").permissions.mode",
			0, int64(420),
		},
		{
			"file(\"/etc/passwd\").content",
			0, passwdContent,
		},
	})
}

func TestResource_File_NotExist(t *testing.T) {
	res := testQuery(t, "file('Nope').content")
	assert.ErrorIs(t, res[0].Data.Error, lumi.NotFound)
}
