// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cli

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignExistingV2(t *testing.T) {
	// first create archive, that we will be able to read
	updateTestDir, _ := ioutil.TempDir("", "update")
	defer os.RemoveAll(updateTestDir)

	priv, pub, err := generateKeys()
	assert.NoError(t, err)

	err = WriteArtifact(updateTestDir, 2, "")
	assert.NoError(t, err)

	err = MakeFakeUpdateDir(updateTestDir,
		[]TestDirEntry{
			{
				Path:    "private.key",
				Content: priv,
				IsDir:   false,
			},
			{
				Path:    "public.key",
				Content: pub,
				IsDir:   false,
			},
		})
	assert.NoError(t, err)

	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "artifact.rdfm")})
	assert.NoError(t, err)

	err = Run([]string{"rdfm-artifact", "validate",
		"-k", filepath.Join(updateTestDir, "public.key"),
		filepath.Join(updateTestDir, "artifact.rdfm.sig")})
	assert.NoError(t, err)

	// now check if signing already signed will fail
	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "artifact.rdfm.sig")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Artifact already signed")

	// and the same as above with force option
	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"), "-f",
		filepath.Join(updateTestDir, "artifact.rdfm.sig")})
	assert.NoError(t, err)
}

func TestSignExistingWithScripts(t *testing.T) {
	updateTestDir, _ := ioutil.TempDir("", "update")
	defer os.RemoveAll(updateTestDir)

	priv, pub, err := generateKeys()
	assert.NoError(t, err)

	err = MakeFakeUpdateDir(updateTestDir,
		[]TestDirEntry{
			{
				Path:    "private.key",
				Content: priv,
				IsDir:   false,
			},
			{
				Path:    "public.key",
				Content: pub,
				IsDir:   false,
			},
			{
				Path:    "update.ext4",
				Content: []byte("my update"),
				IsDir:   false,
			},
			{
				Path:    "ArtifactInstall_Enter_99",
				Content: []byte("this is first enter script"),
				IsDir:   false,
			},
			{
				Path:    "ArtifactInstall_Leave_01",
				Content: []byte("this is leave script"),
				IsDir:   false,
			},
		})
	assert.NoError(t, err)

	// write artifact
	err = Run([]string{"rdfm-artifact", "write", "rootfs-image", "-t", "my-device",
		"-n", "mender-1.1", "-f", filepath.Join(updateTestDir, "update.ext4"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm"),
		"-s", filepath.Join(updateTestDir, "ArtifactInstall_Enter_99"),
		"-s", filepath.Join(updateTestDir, "ArtifactInstall_Leave_01")})
	assert.NoError(t, err)

	// test sign exisiting
	err = Run([]string{"rdfm-artifact", "sign",
		"-k", "-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "artifact.rdfm")})
	assert.Error(t, err)

	// test sign exisiting
	err = Run([]string{"rdfm-artifact", "sign",
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "artifact.rdfm")})
	assert.Error(t, err)

	// test sign exisiting
	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "artifact.rdfm")})
	assert.NoError(t, err)

}

func TestSignExistingWithModules(t *testing.T) {
	updateTestDir, _ := ioutil.TempDir("", "update")
	defer os.RemoveAll(updateTestDir)

	priv, pub, err := generateKeys()
	assert.NoError(t, err)

	err = MakeFakeUpdateDir(updateTestDir,
		[]TestDirEntry{
			{
				Path:    "private.key",
				Content: priv,
				IsDir:   false,
			},
			{
				Path:    "public.key",
				Content: pub,
				IsDir:   false,
			},
			{
				Path:    "payload-file",
				Content: []byte("PayloadContent"),
				IsDir:   false,
			},
		})

	err = Run([]string{"rdfm-artifact", "write", "module-image", "-t", "my-device",
		"-n", "mender-1.1", "-T", "custom-update-type",
		"-f", filepath.Join(updateTestDir, "payload-file"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm")})
	assert.NoError(t, err)

	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "artifact.rdfm")})
	assert.NoError(t, err)

	err = Run([]string{"rdfm-artifact", "validate",
		"-k", filepath.Join(updateTestDir, "public.key"),
		filepath.Join(updateTestDir, "artifact.rdfm.sig")})
	assert.NoError(t, err)

	cmd := exec.Command("tar", "tvf", filepath.Join(updateTestDir, "artifact.rdfm"))
	origTar, err := cmd.Output()
	assert.NoError(t, err)
	cmd = exec.Command("tar", "tvf", filepath.Join(updateTestDir, "artifact.rdfm.sig"))
	newTar, err := cmd.Output()
	assert.NoError(t, err)

	// There should be no change in the tar output otherwise.
	origTarLines := strings.Split(string(origTar), "\n")
	newTarLines := strings.Split(string(newTar), "\n")
	manifestSigRemoved := make([]string, 0, len(newTarLines))
	sig := "manifest.sig"
	for _, newTarLine := range newTarLines {
		if len(newTarLine) >= len(sig) && newTarLine[len(newTarLine)-len(sig):] == sig {
			continue
		}
		manifestSigRemoved = append(manifestSigRemoved, newTarLine)
	}
	assert.Equal(t, origTarLines, manifestSigRemoved)
}

func TestSignExistingBrokenFiles(t *testing.T) {
	updateTestDir, _ := ioutil.TempDir("", "update")
	defer os.RemoveAll(updateTestDir)

	priv, pub, err := generateKeys()
	assert.NoError(t, err)

	err = MakeFakeUpdateDir(updateTestDir,
		[]TestDirEntry{
			{
				Path:    "private.key",
				Content: priv,
				IsDir:   false,
			},
			{
				Path:    "public.key",
				Content: pub,
				IsDir:   false,
			},
			{
				Path:    "empty-file",
				Content: []byte(""),
				IsDir:   false,
			},
			{
				Path:    "garbled-artifact",
				Content: []byte("DefinitelyNotTar"),
				IsDir:   false,
			},
		})

	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "empty-file")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Corrupt Artifact")

	err = Run([]string{"rdfm-artifact", "sign",
		"-k", filepath.Join(updateTestDir, "private.key"),
		"-o", filepath.Join(updateTestDir, "artifact.rdfm.sig"),
		filepath.Join(updateTestDir, "garbled-artifact")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Could not read tar header")
}
