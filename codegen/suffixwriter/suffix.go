package suffixwriter

import (
	"log/slog"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

var (
	listSuffixOnce = sync.OnceValue(listSuffix)
	preCachedList  = [...]string{
		"aix/ppc64",
		"android/386",
		"android/amd64",
		"android/arm",
		"android/arm64",
		"darwin/amd64",
		"darwin/arm64",
		"dragonfly/amd64",
		"freebsd/386",
		"freebsd/amd64",
		"freebsd/arm",
		"freebsd/arm64",
		"freebsd/riscv64",
		"illumos/amd64",
		"ios/amd64",
		"ios/arm64",
		"js/wasm",
		"linux/386",
		"linux/amd64",
		"linux/arm",
		"linux/arm64",
		"linux/loong64",
		"linux/mips",
		"linux/mips64",
		"linux/mips64le",
		"linux/mipsle",
		"linux/ppc64",
		"linux/ppc64le",
		"linux/riscv64",
		"linux/s390x",
		"netbsd/386",
		"netbsd/amd64",
		"netbsd/arm",
		"netbsd/arm64",
		"openbsd/386",
		"openbsd/amd64",
		"openbsd/arm",
		"openbsd/arm64",
		"openbsd/ppc64",
		"openbsd/riscv64",
		"plan9/386",
		"plan9/amd64",
		"plan9/arm",
		"solaris/amd64",
		"wasip1/wasm",
		"windows/386",
		"windows/amd64",
		"windows/arm",
		"windows/arm64",
	}
)

// as per https://pkg.go.dev/cmd/go#hdr-Build_constraints
//
// If a file's name, after stripping the extension and a possible _test suffix, matches any of the following patterns:
//
// *_GOOS
// *_GOARCH
// *_GOOS_GOARCH
func listSuffix() map[string]bool {
	lines, err := listDist()
	if err != nil {
		slog.Warn("go tool dist list failed. Falling back to pre-cached list", slog.Any("err", err))
		lines = preCachedList[:]
	}
	allSuffix := make(map[string]bool)
	for _, line := range lines {
		os, arch, _ := strings.Cut(line, "/")
		allSuffix["_"+os+"_"+arch] = true
		allSuffix["_"+os] = true
		allSuffix["_"+arch] = true
	}
	return allSuffix
}

func listDist() ([]string, error) {
	out, err := exec.Command("go", "tool", "dist", "list").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	return slices.Collect(
		xiter.Filter(
			func(s string) bool { return s != "" },
			xiter.Map(
				strings.TrimSpace,
				slices.Values(lines),
			),
		),
	), nil
}

func suffixFilename(f, suffix string) string {
	ext := filepath.Ext(filepath.Base(f))
	f, _ = strings.CutSuffix(f, ext)

	base, hadTestSuffix := strings.CutSuffix(f, "_test")
	sufTest := "_test"
	if !hadTestSuffix {
		sufTest = ""
	}

	var sufFormer, sufLatter string
	if idx := strings.LastIndex(base, "_"); idx >= 0 {
		sufLatter = base[idx:]
		base = base[:idx]
	}
	if idx := strings.LastIndex(base, "_"); idx >= 0 {
		sufFormer = base[idx:]
		base = base[:idx]
	}

	allSuffix := listSuffixOnce()
	switch {
	case sufFormer != "" && sufLatter != "":
		if allSuffix[sufFormer+sufLatter] {
			break
		}
		if allSuffix[sufLatter] {
			base += sufFormer
			sufFormer = ""
			break
		}
		base += sufFormer + sufLatter
		sufFormer = ""
		sufLatter = ""
	case sufLatter != "":
		if allSuffix[sufLatter] {
			break
		}
		base += sufLatter
		sufLatter = ""
	}

	return base + suffix + sufFormer + sufLatter + sufTest + ext
}
