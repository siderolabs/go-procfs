// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//nolint:testpackage
package procfs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type KernelSuite struct {
	suite.Suite
}

func (suite *KernelSuite) TestNewCmdline() {
	for _, t := range []struct {
		params   string
		expected Parameters
	}{
		{"", Parameters{}},
		{
			"boot=xyz root=/dev/abc nogui console=tty0 console=ttyS0,9600",
			Parameters{
				&Parameter{key: "boot", values: []string{"xyz"}},
				&Parameter{key: "root", values: []string{"/dev/abc"}},
				&Parameter{key: "nogui", values: []string{""}},
				&Parameter{key: "console", values: []string{"tty0", "ttyS0,9600"}},
			},
		},
		{
			" root=/dev/abc=1  nogui  \n",
			Parameters{
				&Parameter{key: "root", values: []string{"/dev/abc=1"}},
				&Parameter{key: "nogui", values: []string{""}},
			},
		},
		{
			"root=/dev/sda root=/dev/sdb",
			Parameters{
				&Parameter{key: "root", values: []string{"/dev/sda", "/dev/sdb"}},
			},
		},
	} {
		cmdline := NewCmdline(t.params)
		suite.Assert().Equal(t.expected, cmdline.Parameters)
	}
}

func (suite *KernelSuite) TestCmdlineGet() {
	for _, t := range []struct { //nolint:govet
		params   string
		arg      string
		expected *Parameter
	}{
		{
			"root=/dev/sda root=/dev/sdb",
			"root",
			&Parameter{key: "root", values: []string{"/dev/sda", "/dev/sdb"}},
		},
		{
			" root=/dev/sda nogui \n",
			"nogui",
			&Parameter{key: "nogui", values: []string{""}},
		},
	} {
		cmdline := NewCmdline(t.params)
		suite.Assert().Equal(t.expected, cmdline.Get(t.arg))
	}
}

func (suite *KernelSuite) TestCmdlineSet() {
	for _, t := range []struct {
		params         string
		k              string
		v              *Parameter
		expected       *Parameter
		expectedParams string
	}{
		{
			"root=/dev/sda root=/dev/sdb",
			"root",
			&Parameter{key: "root", values: []string{"/dev/sdc"}},
			&Parameter{key: "root", values: []string{"/dev/sdc"}},
			"root=/dev/sdc",
		},
		{
			"boot=xyz root=/dev/abc nogui console=tty0 console=ttyS0,9600",
			"console",
			&Parameter{key: "console", values: []string{""}},
			&Parameter{key: "console", values: []string{""}},
			"boot=xyz root=/dev/abc nogui console",
		},
		{
			"initrd=initramfs.xz",
			"initrd",
			&Parameter{key: "initrd", values: []string{"/ROOT-A/initramfs.xz"}},
			&Parameter{key: "initrd", values: []string{"/ROOT-A/initramfs.xz"}},
			"initrd=/ROOT-A/initramfs.xz",
		},
		{
			"",
			"initrd",
			&Parameter{key: "initrd", values: []string{"/ROOT-A/initramfs.xz"}},
			&Parameter{key: "initrd", values: []string{"/ROOT-A/initramfs.xz"}},
			"initrd=/ROOT-A/initramfs.xz",
		},
	} {
		cmdline := NewCmdline(t.params)
		cmdline.Set(t.k, t.v)
		suite.Assert().Equal(t.expected, cmdline.Get(t.k))
		suite.Assert().Equal(t.expectedParams, cmdline.String())
	}
}

func (suite *KernelSuite) TestCmdlineDelete() {
	for _, t := range []struct {
		params   string
		p        *Parameter
		expected string
	}{
		{
			"console=tty0",
			&Parameter{key: "console", values: []string{""}},
			"console=tty0",
		},
		{
			"console=tty0",
			&Parameter{key: "console", values: []string{"tty"}},
			"console=tty0",
		},
		{
			"console=tty0",
			&Parameter{key: "console", values: []string{"tty0"}},
			"",
		},
		{
			"console=tty0 console=ttyS0,9600",
			&Parameter{key: "console", values: []string{"ttyS0,9600"}},
			"console=tty0",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sda panic=0 console=ttyS0,115000",
			&Parameter{key: "console", values: []string{"ttyS0,9600"}},
			"console=tty0 console=ttyS0,115000 root=/dev/sda panic=0",
		},
		{
			"root=/dev/sda panic=0",
			&Parameter{key: "console", values: []string{"ttyS0,9600"}},
			"root=/dev/sda panic=0",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sda panic=0 console=ttyS0,115000",
			&Parameter{key: "console", values: []string{"ttyS0,96000"}},
			"console=tty0 console=ttyS0,9600 console=ttyS0,115000 root=/dev/sda panic=0",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sda panic=0 console=ttyS0,115000",
			&Parameter{key: "consoles", values: []string{"ttyS0,9600"}},
			"console=tty0 console=ttyS0,9600 console=ttyS0,115000 root=/dev/sda panic=0",
		},
		{
			"console=tty0 console=ttyS0,9600",
			&Parameter{key: "console", values: []string{"ttyS0,9600", "tty0"}},
			"",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sda panic=0 console=ttyS0,115000",
			&Parameter{key: "console", values: []string{"ttyS0,9600", "tty0"}},
			"console=ttyS0,115000 root=/dev/sda panic=0",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sda panic=0 console=ttyS0,115000",
			&Parameter{key: "console", values: []string{"ttyS0,9600", "tty0"}},
			"console=ttyS0,115000 root=/dev/sda panic=0",
		},
	} {
		cmdline := NewCmdline(t.params)
		cmdline.Delete(t.p)
		suite.Assert().Equal(t.expected, cmdline.String())
	}
}

func (suite *KernelSuite) TestCmdlineDeleteAll() {
	for _, t := range []struct {
		params   string
		key      string
		expected string
	}{
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sda panic=0 console=ttyS0,115000",
			"console",
			"root=/dev/sda panic=0",
		},
	} {
		cmdline := NewCmdline(t.params)
		cmdline.DeleteAll(t.key)
		suite.Assert().Equal(t.expected, cmdline.String())
	}
}

func (suite *KernelSuite) TestCmdlineAppend() {
	for _, t := range []struct { //nolint:govet
		params   string
		k        string
		v        string
		expected *Parameter
	}{
		{
			"root=/dev/sda root=/dev/sdb",
			"root",
			"/dev/sdc",
			&Parameter{key: "root", values: []string{"/dev/sda", "/dev/sdb", "/dev/sdc"}},
		},
		{
			"boot=xyz root=/dev/abc nogui",
			"console",
			"tty0",
			&Parameter{key: "console", values: []string{"tty0"}},
		},
	} {
		cmdline := NewCmdline(t.params)
		cmdline.Append(t.k, t.v)
		suite.Assert().Equal(t.expected, cmdline.Get(t.k))
	}
}

func (suite *KernelSuite) TestCmdlineAppendAll() {
	for _, t := range []struct { //nolint:govet
		initial  string
		params   []string
		opts     []AppendAllOption
		expected string
	}{
		{
			"ip=dhcp console=x root=/dev/sdc",
			[]string{"root=/dev/sda", "root=/dev/sdb"},
			nil,
			"ip=dhcp console=x root=/dev/sdc root=/dev/sda root=/dev/sdb",
		},
		{
			"root=/dev/sdb",
			[]string{"this=that=those"},
			nil,
			"root=/dev/sdb this=that=those",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"console=tty0", "console=ttyS1,115200", "nogui"},
			nil,
			"console=tty0 console=ttyS0 console=tty0 console=ttyS1,115200 root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"console=tty0", "console=ttyS1,115200", "nogui"},
			[]AppendAllOption{WithOverwriteArgs("console")},
			"console=tty0 console=ttyS1,115200 root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"-console", "nogui"},
			[]AppendAllOption{WithOverwriteArgs("console"), WithDeleteNegatedArgs()},
			"root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sdb",
			[]string{"-console", "nogui"},
			[]AppendAllOption{WithDeleteNegatedArgs()},
			"root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0,9600 root=/dev/sdb",
			[]string{"-console=ttyS0,9600", "nogui"},
			[]AppendAllOption{WithDeleteNegatedArgs()},
			"console=tty0 root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"-console=tty0", "nogui"},
			[]AppendAllOption{WithOverwriteArgs("console"), WithDeleteNegatedArgs()},
			"console=ttyS0 root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"-console", "nogui", "console=ttyAMA0"},
			[]AppendAllOption{WithOverwriteArgs("console"), WithDeleteNegatedArgs()},
			"root=/dev/sdb nogui console=ttyAMA0",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"-console=tty0", "nogui", "console=ttyAMA0"},
			[]AppendAllOption{WithDeleteNegatedArgs()},
			"console=ttyS0 console=ttyAMA0 root=/dev/sdb nogui",
		},
		{
			"console=tty0 console=ttyS0 root=/dev/sdb",
			[]string{"-console=tty0", "nogui", "console=ttyAMA0"},
			[]AppendAllOption{WithOverwriteArgs("console"), WithDeleteNegatedArgs()},
			"console=ttyAMA0 root=/dev/sdb nogui",
		},
	} {
		cmdline := NewCmdline(t.initial)
		err := cmdline.AppendAll(t.params, t.opts...)
		visited := map[string]bool{}

		for _, p := range cmdline.Parameters {
			if visited[p.key] {
				suite.FailNow(fmt.Sprintf("duplicate key %s", p.key))
			}

			visited[p.key] = true
		}

		suite.Assert().NoError(err)
		suite.Assert().Equal(t.expected, cmdline.String())
	}
}

func (suite *KernelSuite) TestCmdlineSetAll() {
	for _, t := range []struct { //nolint:govet
		initial  string
		params   []string
		expected string
	}{
		{
			"",
			[]string{"root=/dev/sda", "root=/dev/sdb"},
			"root=/dev/sda root=/dev/sdb",
		},
		{
			"root=/dev/sdb root=/dev/sdc aye=sir",
			[]string{"root=/dev/mmcblk0"},
			"root=/dev/mmcblk0 aye=sir",
		},
	} {
		cmdline := NewCmdline(t.initial)
		cmdline.SetAll(t.params)

		visited := map[string]bool{}
		for _, p := range cmdline.Parameters {
			if visited[p.key] {
				suite.FailNow(fmt.Sprintf("duplicate key %s", p.key))
			}

			visited[p.key] = true
		}

		suite.Assert().Equal(t.expected, cmdline.String())
	}
}

func (suite *KernelSuite) TestCmdlineString() {
	for _, t := range []struct {
		params   string
		expected string
	}{
		{
			"boot=xyz root=/dev/abc nogui console=tty0 console=ttyS0,9600",
			"boot=xyz root=/dev/abc nogui console=tty0 console=ttyS0,9600",
		},
	} {
		cmdline := NewCmdline(t.params)
		suite.Assert().Equal(t.expected, cmdline.String())
	}
}

func (suite *KernelSuite) TestCmdlineStrings() {
	for _, t := range []struct {
		params   string
		expected []string
	}{
		{
			"boot=xyz root=/dev/abc nogui console=tty0 console=ttyS0,9600",
			[]string{"boot=xyz", "root=/dev/abc", "nogui", "console=tty0", "console=ttyS0,9600"},
		},
	} {
		cmdline := NewCmdline(t.params)
		suite.Assert().Equal(t.expected, cmdline.Strings())
	}
}

func (suite *KernelSuite) TestParameterFirst() {
	for _, t := range []struct {
		value    *Parameter
		expected *string
	}{
		{&Parameter{values: []string{""}}, pointer("")},
		{&Parameter{values: []string{}}, nil},
		{&Parameter{values: []string{"a", "b", "c=d"}}, pointer("a")},
		{nil, nil},
	} {
		suite.Assert().Equal(t.expected, t.value.First())
	}
}

func (suite *KernelSuite) TestParameterGet() {
	for _, t := range []struct { //nolint:govet
		value    *Parameter
		idx      int
		expected *string
	}{
		{&Parameter{values: []string{"", "x", "/dev/sda"}}, 2, pointer("/dev/sda")},
		{&Parameter{values: []string{}}, 2, nil},
	} {
		suite.Assert().Equal(t.expected, t.value.Get(t.idx))
	}
}

func (suite *KernelSuite) TestParameterAppend() {
	for _, t := range []struct { //nolint:govet
		value    *Parameter
		app      string
		expected *Parameter
	}{
		{
			&Parameter{
				values: []string{"", "x", "/dev/sda"},
			},
			"ttyS0",
			&Parameter{
				values: []string{"", "x", "/dev/sda", "ttyS0"},
			},
		},
		{
			&Parameter{
				values: []string{},
			},
			"nogui",
			&Parameter{
				values: []string{"nogui"},
			},
		},
	} {
		suite.Assert().Equal(t.expected, t.value.Append(t.app))
	}
}

func (suite *KernelSuite) TestParameterContains() {
	for _, t := range []struct {
		value    *Parameter
		s        string
		expected bool
	}{
		{
			&Parameter{
				values: []string{"", "x", "/dev/sda"},
			},
			"x",
			true,
		},
		{
			&Parameter{
				values: []string{},
			},
			"x",
			false,
		},
	} {
		suite.Assert().Equal(t.expected, t.value.Contains(t.s))
	}
}

func pointer(s string) *string {
	return &s
}

func TestKernelSuite(t *testing.T) {
	suite.Run(t, new(KernelSuite))
}
