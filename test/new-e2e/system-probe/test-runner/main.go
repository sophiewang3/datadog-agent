// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/sys/unix"
)

func init() {
	color.NoColor = false
}

type testConfig struct {
	retryCount      int
	includePackages []string
	excludePackages []string
	runTests        string
}

const (
	testDirRoot  = "/opt/system-probe-tests"
	ciVisibility = "/ci-visibility"
)

var baseEnv = []string{
	"GITLAB_CI=true", // force color output support to be detected
	"GOVERSION=" + runtime.Version(),
	"DD_SYSTEM_PROBE_BPF_DIR=" + filepath.Join(testDirRoot, "pkg/ebpf/bytecode/build"),
	"DD_SYSTEM_PROBE_JAVA_DIR=" + filepath.Join(testDirRoot, "pkg/network/protocols/tls/java"),
}

var timeouts = map[*regexp.Regexp]time.Duration{
	regexp.MustCompile("pkg/network/protocols/http$"): 15 * time.Minute,
	regexp.MustCompile("pkg/network/tracer$"):         55 * time.Minute,
	regexp.MustCompile("pkg/network/usm$"):            30 * time.Minute,
}

func getTimeout(pkg string) time.Duration {
	matchSize := 0
	to := 10 * time.Minute
	for re, rto := range timeouts {
		if re.MatchString(pkg) && len(re.String()) > matchSize {
			matchSize = len(re.String())
			to = rto
		}
	}
	return to
}

func pathEmbedded(fullPath, embedded string) bool {
	normalized := fmt.Sprintf("/%s/", strings.Trim(embedded, "/"))

	return strings.Contains(fullPath, normalized)
}

func glob(dir, filePattern string, filterFn func(path string) bool) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		present, err := regexp.Match(filePattern, []byte(d.Name()))
		if err != nil {
			return fmt.Errorf("file regexp match: %s", err)
		}

		if d.IsDir() || !present {
			return nil
		}
		if filterFn(path) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func pathToPackage(path string) string {
	dir, _ := filepath.Rel(testDirRoot, filepath.Dir(path))
	return dir
}

func buildCommandArgs(pkg string, xmlpath string, jsonpath string, file string, testConfig *testConfig) []string {
	args := []string{
		"--format", "dots",
		"--junitfile", xmlpath,
		"--jsonfile", jsonpath,
		fmt.Sprintf("--rerun-fails=%d", testConfig.retryCount),
		"--rerun-fails-max-failures=100",
		"--raw-command", "--",
		"/go/bin/test2json", "-t", "-p", pkg, file, "-test.v", "-test.count=1", "-test.timeout=" + getTimeout(pkg).String(),
	}

	if testConfig.runTests != "" {
		args = append(args, "-test.run", testConfig.runTests)
	}

	return args
}

// concatenateJsons combines all the test json output files into a single file.
func concatenateJsons(indir, outdir string) error {
	testJsonFile := filepath.Join(outdir, "out.json")
	matches, err := glob(indir, `.*\.json`, func(path string) bool { return true })
	if err != nil {
		return fmt.Errorf("json glob: %s", err)
	}

	f, err := os.OpenFile(testJsonFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("open %s: %s", testJsonFile, err)
	}
	defer f.Close()

	for _, jsonFile := range matches {
		jf, err := os.Open(jsonFile)
		if err != nil {
			return fmt.Errorf("open %s: %s", jsonFile, err)
		}

		_, err = io.Copy(f, jf)
		_ = jf.Close()
		if err != nil {
			return fmt.Errorf("%s copy: %s", jsonFile, err)
		}
	}
	return nil
}

func createDir(d string) error {
	if err := os.MkdirAll(d, 0o777); err != nil {
		return fmt.Errorf("failed to create directory %s", d)
	}
	return nil
}

func testPass(testConfig *testConfig, props map[string]string) error {
	testsuites, err := glob(testDirRoot, "testsuite", func(path string) bool {
		dir := pathToPackage(path)
		for _, p := range testConfig.excludePackages {
			if dir == p {
				return false
			}
		}
		if len(testConfig.includePackages) != 0 {
			for _, p := range testConfig.includePackages {
				if dir == p {
					return true
				}
			}
			return false
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("test glob: %s", err)
	}

	jsonDir := filepath.Join(ciVisibility, "pkgjson")
	jsonOutDir := filepath.Join(ciVisibility, "testjson")
	xmlDir := filepath.Join(ciVisibility, "junit")
	for _, d := range []string{jsonDir, jsonOutDir, xmlDir} {
		if err := createDir(d); err != nil {
			return err
		}
	}

	for _, testsuite := range testsuites {
		pkg := pathToPackage(testsuite)
		junitfilePrefix := strings.ReplaceAll(pkg, "/", "-")
		xmlpath := filepath.Join(xmlDir, fmt.Sprintf("%s.xml", junitfilePrefix))
		jsonpath := filepath.Join(jsonDir, fmt.Sprintf("%s.json", junitfilePrefix))
		args := buildCommandArgs(pkg, xmlpath, jsonpath, testsuite, testConfig)

		cmd := exec.Command("/go/bin/gotestsum", args...)
		cmd.Env = append(cmd.Environ(), baseEnv...)
		cmd.Dir = filepath.Dir(testsuite)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// log but do not return error
			fmt.Fprintf(os.Stderr, "cmd run %s: %s", testsuite, err)
		}

		if err := addProperties(xmlpath, props); err != nil {
			return fmt.Errorf("xml add props: %s", err)
		}
	}

	if err := concatenateJsons(jsonDir, jsonOutDir); err != nil {
		return fmt.Errorf("concat json: %s", err)
	}
	return nil
}

func fixAssetPermissions() error {
	matches, err := glob(testDirRoot, `.*\.o`, func(path string) bool {
		return pathEmbedded(path, "pkg/ebpf/bytecode/build")
	})
	if err != nil {
		return fmt.Errorf("glob assets: %s", err)
	}

	for _, file := range matches {
		if err := os.Chown(file, 0, 0); err != nil {
			return fmt.Errorf("chown %s: %s", file, err)
		}
	}
	return nil
}

func buildTestConfiguration() *testConfig {
	retryPtr := flag.Int("retry", 2, "number of times to retry testing pass")
	packagesPtr := flag.String("include-packages", "", "Comma separated list of packages to test")
	excludePackagesPtr := flag.String("exclude-packages", "", "Comma separated list of packages to exclude")
	runTestsPtr := flag.String("run-tests", "", "Regex for running specific tests")

	flag.Parse()

	var packagesLs []string
	var excludeLs []string

	if *packagesPtr != "" {
		packagesLs = strings.Split(*packagesPtr, ",")
	}
	if *excludePackagesPtr != "" {
		excludeLs = strings.Split(*excludePackagesPtr, ",")
	}

	return &testConfig{
		retryCount:      *retryPtr,
		includePackages: packagesLs,
		excludePackages: excludeLs,
		runTests:        *runTestsPtr,
	}
}

func readOSRelease() (map[string]string, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyvals := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// output contents for visibility
		fmt.Println(scanner.Text())
		k, v, found := bytes.Cut(scanner.Bytes(), []byte{'='})
		if found {
			keyvals[string(k)] = strings.Trim(string(v), "\"")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return keyvals, nil
}

func getProps() (map[string]string, error) {
	osrHash, err := readOSRelease()
	if err != nil {
		return nil, fmt.Errorf("os-release: %s", err)
	}
	osname := fmt.Sprintf("%s-%s", osrHash["ID"], osrHash["VERSION_ID"])
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return nil, fmt.Errorf("uname: %w", err)
	}
	arch, release := unix.ByteSliceToString(u.Machine[:]), unix.ByteSliceToString(u.Release[:])
	fmt.Printf("arch: %s\nrelease: %s\n", arch, release)
	return map[string]string{
		"dd_tags[os.platform]":     "linux",
		"dd_tags[os.name]":         osname,
		"dd_tags[os.architecture]": arch,
		"dd_tags[os.version]":      release,
	}, nil
}

func run() error {
	props, err := getProps()
	if err != nil {
		return fmt.Errorf("props: %s", err)
	}

	testConfig := buildTestConfiguration()
	if err := fixAssetPermissions(); err != nil {
		return fmt.Errorf("asset perms: %s", err)
	}

	if err := os.RemoveAll(ciVisibility); err != nil {
		return fmt.Errorf("failed to remove contents of %s: %w", ciVisibility, err)
	}
	if err := createDir(ciVisibility); err != nil {
		return err
	}

	return testPass(testConfig, props)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", color.RedString(err.Error()))
		os.Exit(1)
	}
}
