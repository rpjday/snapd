// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package bind_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/backendtest"
	"github.com/snapcore/snapd/interfaces/bind"
	"github.com/snapcore/snapd/osutil"
)

func Test(t *testing.T) {
	TestingT(t)
}

type backendSuite struct {
	backendtest.BackendSuite
}

var _ = Suite(&backendSuite{})

func (s *backendSuite) SetUpTest(c *C) {
	s.Backend = &bind.Backend{}
	s.BackendSuite.SetUpTest(c)

	err := os.MkdirAll(dirs.SnapBindMountPolicyDir, 0700)
	c.Assert(err, IsNil)

}

func (s *backendSuite) TearDownTest(c *C) {
	s.BackendSuite.TearDownTest(c)
}

func (s *backendSuite) TestName(c *C) {
	c.Check(s.Backend.Name(), Equals, "bind")
}

func (s *backendSuite) TestRemove(c *C) {
	canaryToGo := filepath.Join(dirs.SnapBindMountPolicyDir, "snap.hello-world.hello-world.bind")
	err := ioutil.WriteFile(canaryToGo, []byte("ni! ni! ni!"), 0644)
	c.Assert(err, IsNil)

	canaryToStay := filepath.Join(dirs.SnapBindMountPolicyDir, "snap.i-stay.really.bind")
	err = ioutil.WriteFile(canaryToStay, []byte("stay!"), 0644)
	c.Assert(err, IsNil)

	err = s.Backend.Remove("hello-world")
	c.Assert(err, IsNil)

	c.Assert(osutil.FileExists(canaryToGo), Equals, false)
	content, err := ioutil.ReadFile(canaryToStay)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "stay!")
}

var mockSnapYaml = `name: snap-name
version: 1
apps:
    app1:
    app2:
slots:
    iface:
`

func (s *backendSuite) TestSetupSetsup(c *C) {
	fsEntry := "/src-1 /dst-1 none bind,ro 0 0"
	s.Iface.PermanentSlotSnippetCallback = func(slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
		return []byte(fsEntry), nil
	}

	// devMode is irrelevant for this security backend
	s.InstallSnap(c, false, mockSnapYaml, 0)

	// FIXME: test combineSnipets implicitely somehow too
	fn1 := filepath.Join(dirs.SnapBindMountPolicyDir, "snap.snap-name.app1.bind")
	content, err := ioutil.ReadFile(fn1)
	c.Assert(err, IsNil)
	c.Check(string(content), Equals, fsEntry+"\n")

	fn2 := filepath.Join(dirs.SnapBindMountPolicyDir, "snap.snap-name.app2.bind")
	content, err = ioutil.ReadFile(fn2)
	c.Assert(err, IsNil)
	c.Check(string(content), Equals, fsEntry+"\n")

}
