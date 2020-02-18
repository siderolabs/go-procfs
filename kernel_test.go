/*
Copyright 2020 Talos Systems, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kernel

import (
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
	for _, t := range []struct {
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
		params   string
		k        string
		v        *Parameter
		expected *Parameter
	}{
		{
			"root=/dev/sda root=/dev/sdb",
			"root",
			&Parameter{key: "root", values: []string{"/dev/sdc"}},
			&Parameter{key: "root", values: []string{"/dev/sdc"}},
		},
		{
			"boot=xyz root=/dev/abc nogui console=tty0 console=ttyS0,9600",
			"console",
			&Parameter{key: "console", values: nil},
			&Parameter{key: "console", values: nil},
		},
		{
			"initrd=initramfs.xz",
			"initrd",
			&Parameter{key: "initrd", values: []string{"/ROOT-A/initramfs.xz"}},
			&Parameter{key: "initrd", values: []string{"/ROOT-A/initramfs.xz"}},
		},
	} {
		cmdline := NewCmdline(t.params)
		cmdline.Set(t.k, t.v)
		suite.Assert().Equal(t.expected, cmdline.Get(t.k))
	}
}

func (suite *KernelSuite) TestCmdlineAppend() {
	for _, t := range []struct {
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
	for _, t := range []struct {
		initial  string
		params   []string
		expected string
	}{
		{
			"ip=dhcp console=x",
			[]string{"root=/dev/sda", "root=/dev/sdb"},
			"ip=dhcp console=x root=/dev/sda root=/dev/sdb",
		},
		{
			"root=/dev/sdb",
			[]string{"this=that=those"},
			"root=/dev/sdb this=that=those",
		},
	} {
		cmdline := NewCmdline(t.initial)
		err := cmdline.AppendAll(t.params)
		suite.Assert().NoError(err)
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
	for _, t := range []struct {
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
	for _, t := range []struct {
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
