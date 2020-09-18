package common

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func TestStepCreateCD_retrieveCDISOCreationCommand_msys2(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("test only applies to windows")
	}
	_, err := exec.LookPath("cygpath")
	if err != nil {
		t.Skip("test only applies to windows msys2")
	}
	c, err := retrieveCDISOCreationCommand(
		"test",
		"C:\\Windows\\Temp\\test-cd",
		"C:\\Windows\\Temp\\test-cd.iso")
	if !strings.HasSuffix(c.Path, "\\usr\\bin\\xorriso.exe") {
		t.Fatalf("expected a msys2 xorriso.exe command path but got %s", c.Path)
	}
	if len(c.Args) < 2 {
		t.Fatalf("expected a xorriso.exe command args length >= 2 but got %v", c.Args)
	}
	expectedSourcePath := "/c/Windows/Temp/test-cd"
	sourcePath := c.Args[len(c.Args)-1]
	if expectedSourcePath != sourcePath {
		t.Fatalf("expected the source path to be %s but got %s", expectedSourcePath, sourcePath)
	}
	expectedDestinationPath := "/c/Windows/Temp/test-cd.iso"
	destinationPath := c.Args[len(c.Args)-2]
	if expectedDestinationPath != destinationPath {
		t.Fatalf("expected the destination path to be %s but got %s", expectedDestinationPath, destinationPath)
	}
}

func TestStepCreateCD_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepCreateCD)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("StepCreateCD should be a step")
	}
}

func testStepCreateCDState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}

func TestStepCreateCD(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the disk management binaries.")
	}
	state := testStepCreateCDState(t)
	step := new(StepCreateCD)

	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	files := make([]string, 3)

	tempFileNames := []string{"test_cd_roms.tmp", "test cd files.tmp",
		"Test-Test-Test5.tmp"}
	for i, fname := range tempFileNames {
		files[i] = path.Join(dir, fname)

		_, err := os.Create(files[i])
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	step.Files = files
	action := step.Run(context.Background(), state)

	if err, ok := state.GetOk("error"); ok {
		t.Fatalf("state should be ok for %v: %s", step.Files, err)
	}

	if action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v for %v", action, step.Files)
	}

	CD_path := state.Get("cd_path").(string)

	if _, err := os.Stat(CD_path); err != nil {
		t.Fatalf("file not found: %s for %v", CD_path, step.Files)
	}

	if len(step.filesAdded) != 3 {
		t.Fatalf("expected 3 files, found %d for %v", len(step.filesAdded), step.Files)
	}

	step.Cleanup(state)

	if _, err := os.Stat(CD_path); err == nil {
		t.Fatalf("file found: %s for %v", CD_path, step.Files)
	}
}

func TestStepCreateCD_missing(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the disk management binaries.")
	}
	state := testStepCreateCDState(t)
	step := new(StepCreateCD)

	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	expected := 0

	step.Files = []string{"missing file.tmp"}
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v for %v", action, step.Files)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("state should not be ok for %v", step.Files)
	}

	CD_path := state.Get("cd_path")

	if CD_path != nil {
		t.Fatalf("CD_path is not nil for %v", step.Files)
	}

	if len(step.filesAdded) != expected {
		t.Fatalf("expected %d, found %d for %v", expected, len(step.filesAdded), step.Files)
	}
}
