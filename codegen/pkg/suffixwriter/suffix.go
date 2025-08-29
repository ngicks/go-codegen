package suffixwriter

import (
	"encoding/json"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	listSuffixOnce       sync.Once
	suffixOs, suffixArch map[string]bool
)

type goToolDistListJson struct {
	GOOS         string
	GOARCH       string
	CgoSupported bool
	FirstClass   bool
}

var preCachedList = []goToolDistListJson{
	{"aix", "ppc64", true, false},
	{"android", "386", true, false},
	{"android", "amd64", true, false},
	{"android", "arm", true, false},
	{"android", "arm64", true, false},
	{"darwin", "amd64", true, true},
	{"darwin", "arm64", true, true},
	{"dragonfly", "amd64", true, false},
	{"freebsd", "386", true, false},
	{"freebsd", "amd64", true, false},
	{"freebsd", "arm", true, false},
	{"freebsd", "arm64", true, false},
	{"freebsd", "riscv64", true, false},
	{"illumos", "amd64", true, false},
	{"ios", "amd64", true, false},
	{"ios", "arm64", true, false},
	{"js", "wasm", false, false},
	{"linux", "386", true, true},
	{"linux", "amd64", true, true},
	{"linux", "arm", true, true},
	{"linux", "arm64", true, true},
	{"linux", "loong64", true, false},
	{"linux", "mips", true, false},
	{"linux", "mips64", true, false},
	{"linux", "mips64le", true, false},
	{"linux", "mipsle", true, false},
	{"linux", "ppc64", false, false},
	{"linux", "ppc64le", true, false},
	{"linux", "riscv64", true, false},
	{"linux", "s390x", true, false},
	{"netbsd", "386", true, false},
	{"netbsd", "amd64", true, false},
	{"netbsd", "arm", true, false},
	{"netbsd", "arm64", true, false},
	{"openbsd", "386", true, false},
	{"openbsd", "amd64", true, false},
	{"openbsd", "arm", true, false},
	{"openbsd", "arm64", true, false},
	{"openbsd", "ppc64", false, false},
	{"openbsd", "riscv64", true, false},
	{"plan9", "386", false, false},
	{"plan9", "amd64", false, false},
	{"plan9", "arm", false, false},
	{"solaris", "amd64", true, false},
	{"wasip1", "wasm", false, false},
	{"windows", "386", true, true},
	{"windows", "amd64", true, true},
	{"windows", "arm", false, false},
	{"windows", "arm64", true, false},
}

// as per https://pkg.go.dev/cmd/go#hdr-Build_constraints
//
// If a file's name, after stripping the extension and a possible _test suffix, matches any of the following patterns:
//
// *_GOOS
// *_GOARCH
// *_GOOS_GOARCH
func listSuffix() (os, arch map[string]bool) {
	lines, err := listDist()
	if err != nil {
		slog.Warn("go tool dist list failed. Falling back to pre-cached list", slog.Any("err", err))
		lines = preCachedList[:]
	}

	os = make(map[string]bool, len(lines)/2)
	arch = make(map[string]bool, len(lines)/2)

	for _, line := range lines {
		os["_"+line.GOOS] = true
		arch["_"+line.GOARCH] = true
	}

	return
}

func listDist() ([]goToolDistListJson, error) {
	bin, err := exec.Command("go", "tool", "dist", "list", "-json").Output()
	if err != nil {
		return nil, err
	}

	var out []goToolDistListJson
	err = json.Unmarshal(bin, &out)

	return out, err
}

// SuffixFilename suffixes f by suffix.
// It moves implicit build constraints from original filename to suffix.
//
// For example, assuming passing ".suf" to suffix:
//   - foo_linux.go -> foo.suf_linux.go
//   - foo_amd64_test.go -> foo.suf_amd64_test.go
//   - foo_bar.go -> foo_bar.suf.go
//
// Basically SuffixFilename is intended to be used for ".go" files but actually can be used for any file extension,
// even no file extension is allowed.
func SuffixFilename(f, suffix string) string {
	base, sufFormer, sufLatter, sufTest, ext := stripBuildConstrains(f)
	return base + suffix + sufFormer + sufLatter + sufTest + ext
}

// IsSuffixed tests f is suffixed with suffix.
// The test ignores file extension and implicit build constraints suffix in f.
func IsSuffixed(f, suffix string) bool {
	base, _, _, _, _ := stripBuildConstrains(f)
	return strings.HasSuffix(base, suffix)
}

func stripBuildConstrains(f string) (base, sufFormer, sufLatter, sufTest, ext string) {
	ext = filepath.Ext(f)
	f, _ = strings.CutSuffix(f, ext)

	base, hadTestSuffix := strings.CutSuffix(f, "_test")
	sufTest = "_test"
	if !hadTestSuffix {
		sufTest = ""
	}

	listSuffixOnce.Do(func() {
		suffixOs, suffixArch = listSuffix()
	})

	var suffixedWithOs, suffixedWithArch bool
	if idx := strings.LastIndex(base, "_"); idx >= 0 {
		suf := base[idx:]
		switch {
		case suffixOs[suf]:
			suffixedWithOs = true
		case suffixArch[suf]:
			suffixedWithArch = true
		}
		if suffixedWithOs || suffixedWithArch {
			sufLatter = base[idx:]
			base = base[:idx]
		}
	}

	if suffixedWithArch {
		if idx := strings.LastIndex(base, "_"); idx >= 0 && suffixOs[base[idx:]] {
			sufFormer = base[idx:]
			base = base[:idx]
		}
	}

	return
}
