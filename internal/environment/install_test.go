package environment

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestInstallRequiresConfirmationWithoutRunningCommands(t *testing.T) {
	t.Parallel()

	ran := false
	err := Install(validPack(), InstallOptions{
		Run: func([]string, io.Writer, io.Writer) error {
			ran = true
			return nil
		},
	})
	if !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("Install() error = %v, want confirmation", err)
	}
	if ran {
		t.Fatal("Install() ran a command without confirmation")
	}
}

func TestInstallRunsMissingRequirementsInDependencyOrderAndVerifies(t *testing.T) {
	t.Parallel()

	pack := validPack()
	installed := map[string]bool{"herdr": true}
	plugins := map[string]bool{}
	var commands [][]string
	var stdout, stderr bytes.Buffer
	err := Install(pack, InstallOptions{
		Confirmed: true,
		GOOS:      "darwin",
		LookPath: func(command string) (string, error) {
			if installed[command] || command == "brew" || command == "go" {
				return "/bin/" + command, nil
			}
			return "", exec.ErrNotFound
		},
		InspectPlugin: func(host, pluginID string) (bool, error) {
			if host != "herdr" {
				return false, fmt.Errorf("unexpected host %q", host)
			}
			return plugins[pluginID], nil
		},
		Run: func(command []string, output, errorOutput io.Writer) error {
			commands = append(commands, slices.Clone(command))
			if slices.Equal(command, []string{"brew", "install", "zellij"}) {
				installed["zellij"] = true
			}
			if slices.Equal(command, []string{"brew", "install", "tatami"}) {
				installed["tatami"] = true
			}
			if len(command) >= 4 && slices.Equal(command[:4], []string{"herdr", "plugin", "install", "owner/repo"}) {
				plugins["example-plugin"] = true
			}
			fmt.Fprintln(output, strings.Join(command, " "))
			fmt.Fprintln(errorOutput, "unsafe\x1b]52;c;clipboard")
			return nil
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("Install() error = %v, stderr = %s", err, stderr.String())
	}
	if strings.Contains(stderr.String(), "\x1b") {
		t.Fatalf("installer stderr contains terminal escape: %q", stderr.String())
	}
	want := [][]string{
		{"brew", "install", "zellij"},
		{"brew", "tap", "OleksandrBesan/tap"},
		{"brew", "install", "tatami"},
		{"herdr", "plugin", "install", "owner/repo", "--ref", "v1.2.3", "--yes"},
	}
	if !slices.EqualFunc(commands, want, slices.Equal[[]string]) {
		t.Fatalf("commands = %#v, want %#v", commands, want)
	}
	for _, expected := range []string{
		"installed zellij",
		"installed tatami",
		"installed herdr-plugin",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("output missing %q: %s", expected, stdout.String())
		}
	}
}

func TestInstallSkipsSatisfiedRequirementsAndRejectsUnverifiableInstall(t *testing.T) {
	t.Parallel()

	t.Run("satisfied", func(t *testing.T) {
		pack := validPack()
		ran := false
		err := Install(pack, InstallOptions{
			Confirmed: true,
			GOOS:      "darwin",
			LookPath: func(command string) (string, error) {
				return "/bin/" + command, nil
			},
			InspectPlugin: func(string, string) (bool, error) {
				return true, nil
			},
			Run: func([]string, io.Writer, io.Writer) error {
				ran = true
				return nil
			},
		})
		if err != nil {
			t.Fatalf("Install() error = %v", err)
		}
		if ran {
			t.Fatal("Install() ran commands for satisfied requirements")
		}
	})

	t.Run("verification failure", func(t *testing.T) {
		pack := validPack()
		err := Install(pack, InstallOptions{
			Confirmed: true,
			GOOS:      "darwin",
			LookPath: func(command string) (string, error) {
				if command == "brew" || command == "go" || command == "herdr" {
					return "/bin/" + command, nil
				}
				return "", exec.ErrNotFound
			},
			InspectPlugin: func(string, string) (bool, error) {
				return false, nil
			},
			Run: func([]string, io.Writer, io.Writer) error {
				return nil
			},
		})
		if err == nil || !strings.Contains(err.Error(), "verification") {
			t.Fatalf("Install() error = %v, want verification failure", err)
		}
	})
}

func TestNPMInstallerIsPinnedAndNonInteractiveHostPluginIsExact(t *testing.T) {
	t.Parallel()

	npm := planInstaller(Installer{
		ID: "npm", Kind: InstallerNPMGlobal, Platforms: []string{"darwin"},
		Package: "mdmaid", Version: "0.1.14",
	})
	if got := npm.Commands; !slices.EqualFunc(got, [][]string{{"npm", "install", "--global", "mdmaid@0.1.14"}}, slices.Equal[[]string]) {
		t.Fatalf("npm commands = %#v", got)
	}
	plugin := planInstaller(Installer{
		ID: "herdr", Kind: InstallerHostPlugin, Platforms: []string{"darwin"},
		Host: "herdr", Repository: "owner/repo", Ref: "0123456789abcdef0123456789abcdef01234567",
	})
	if got := plugin.Commands; !slices.EqualFunc(got, [][]string{{
		"herdr", "plugin", "install", "owner/repo", "--ref",
		"0123456789abcdef0123456789abcdef01234567", "--yes",
	}}, slices.Equal[[]string]) {
		t.Fatalf("plugin commands = %#v", got)
	}
}

func TestDecodePluginInspection(t *testing.T) {
	t.Parallel()

	data := []byte(`{"result":{"plugins":[{"plugin_id":"hail"}]}}`)
	present, err := decodePluginInspection(data, "hail")
	if err != nil || !present {
		t.Fatalf("decodePluginInspection() = %t, %v", present, err)
	}
	present, err = decodePluginInspection(data, "herdr-bar")
	if err != nil || present {
		t.Fatalf("decodePluginInspection(absent) = %t, %v", present, err)
	}
	if _, err := decodePluginInspection([]byte(`{"result":`), "hail"); err == nil {
		t.Fatal("decodePluginInspection() accepted malformed JSON")
	}
}

func TestLimitedBufferRejectsOversizedPluginInspection(t *testing.T) {
	t.Parallel()

	buffer := limitedBuffer{remaining: 2}
	if _, err := buffer.Write([]byte("abc")); err == nil {
		t.Fatal("limitedBuffer.Write() accepted oversized data")
	}
	buffer = limitedBuffer{remaining: 3}
	if count, err := buffer.Write([]byte("ok")); err != nil || count != 2 || buffer.String() != "ok" {
		t.Fatalf("limitedBuffer.Write() = %d, %v, %q", count, err, buffer.String())
	}
}

func TestInstallReportsExecutionInspectionAndInstallerErrors(t *testing.T) {
	t.Parallel()

	t.Run("execution", func(t *testing.T) {
		pack := validPack()
		err := Install(pack, InstallOptions{
			Confirmed: true,
			GOOS:      "darwin",
			LookPath: func(command string) (string, error) {
				if command == "brew" {
					return "/bin/brew", nil
				}
				return "", exec.ErrNotFound
			},
			Run: func([]string, io.Writer, io.Writer) error {
				return errors.New("installer failed")
			},
		})
		if err == nil || !strings.Contains(err.Error(), "installer failed") {
			t.Fatalf("Install() error = %v", err)
		}
	})

	t.Run("plugin inspection", func(t *testing.T) {
		pack := validPack()
		err := Install(pack, InstallOptions{
			Confirmed: true,
			GOOS:      "darwin",
			LookPath: func(command string) (string, error) {
				return "/bin/" + command, nil
			},
			InspectPlugin: func(string, string) (bool, error) {
				return false, errors.New("registry unavailable")
			},
		})
		if err == nil || !strings.Contains(err.Error(), "registry unavailable") {
			t.Fatalf("Install() error = %v", err)
		}
	})

	t.Run("manual only", func(t *testing.T) {
		requirement := validPack().Requirements[0]
		requirement.Installers = requirement.Installers[1:]
		pack := validPack()
		pack.Requirements = []Requirement{requirement}
		err := Install(pack, InstallOptions{
			Confirmed: true,
			GOOS:      "darwin",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		})
		if err == nil || !strings.Contains(err.Error(), "follow https://") {
			t.Fatalf("Install() error = %v", err)
		}
	})
}

func TestSuggestedInstallerPrefersAvailableThenManual(t *testing.T) {
	t.Parallel()

	requirement := PlannedRequirement{Installers: []PlannedInstaller{
		{ID: "missing", Commands: [][]string{{"missing"}}},
		{ID: "manual", Instructions: "Follow guide"},
		{ID: "available", Available: true, Commands: [][]string{{"available"}}},
	}}
	if got, ok := requirement.SuggestedInstaller(); !ok || got.ID != "available" {
		t.Fatalf("SuggestedInstaller() = %#v, %t", got, ok)
	}
	requirement.Installers[2].Available = false
	if got, ok := requirement.SuggestedInstaller(); !ok || got.ID != "manual" {
		t.Fatalf("SuggestedInstaller() fallback = %#v, %t", got, ok)
	}
	requirement.Installers = nil
	if _, ok := requirement.SuggestedInstaller(); ok {
		t.Fatal("SuggestedInstaller() found an installer in an empty plan")
	}
}

func TestDefaultCommandAndPluginInspectionFailuresAreBounded(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if err := runCommand([]string{"go", "version"}, &stdout, &stderr); err != nil {
		t.Fatalf("runCommand(go version) error = %v, stderr = %s", err, stderr.String())
	}
	if _, err := inspectHerdrPlugin("maisternia-command-that-does-not-exist", "hail"); err == nil {
		t.Fatal("inspectHerdrPlugin() accepted a missing host")
	}
}
