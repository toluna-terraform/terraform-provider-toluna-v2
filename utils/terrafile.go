package toluna

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jessevdk/go-flags"
	ver_sort "golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

type module struct {
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

var opts struct {
	ModulePath    string
	TerrafilePath string
}

// To be set by goreleaser on build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var running_modules []string

func init() {

}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func scanModules(paths []string) ([]string, error) {
	var out []string
	for _, path := range paths {
		module, err := tfconfig.LoadModule(path)
		if err != nil {
			return nil, fmt.Errorf("read terraform module %q: %w", path, err)
		}
		for _, call := range module.ModuleCalls {
			if call == nil {
				continue
			}
			out = append(out, call.Name)
		}
	}
	return out, nil
}

func gitClone(ctx context.Context, repository string, version string, moduleName string) {
	tflog.Info(ctx, "[*] Checking out "+version+" of "+repository+"\n")
	cmd := exec.Command("git", "clone", "--single-branch", "--depth=1", "-b", version, repository, moduleName)
	cmd.Dir = opts.ModulePath
	err := cmd.Run()
	if err != nil {
		tflog.Error(ctx, err.Error())
	}
}

type getModuleVersions struct {
	ModuleList []struct {
		VersionList []struct {
			Version string `json:"version"`
		} `json:"versions"`
	} `json:"modules"`
}

func tfRegistryGetVersionList(ctx context.Context, module string, version string, operator string) string {
	baseURL := "https://registry.terraform.io/v1/modules/"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+module+"/versions", nil)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		tflog.Error(ctx, "Error reading terraform registry")
	}
	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	var module_versions getModuleVersions
	if err := json.Unmarshal(resp_body, &module_versions); err != nil { // Parse []byte to the go struct pointer
		tflog.Info(ctx, "Can not get remote version")
	}
	parsedVerionList := []string{}
	for _, pVersion := range module_versions.ModuleList[0].VersionList {
		if operator == "~>" {
			s_version := strings.Split(version, ".")
			major_minor := fmt.Sprintf("%s%s%s%s", s_version[0], ".", s_version[1], ".")
			if strings.HasPrefix(pVersion.Version, major_minor) {
				parsedVerionList = append(parsedVerionList, pVersion.Version)
			}
		}
		lv, _ := semver.NewVersion(version)
		rv, _ := semver.NewVersion(pVersion.Version)
		if operator == ">" {
			if lv.LessThan(rv) {
				parsedVerionList = append(parsedVerionList, pVersion.Version)
			}
		}
		if operator == ">=" {
			if lv.LessThan(rv) || lv.Equal(rv) {
				parsedVerionList = append(parsedVerionList, pVersion.Version)
			}
		}
		if operator == "<" {
			if lv.GreaterThan(rv) {
				parsedVerionList = append(parsedVerionList, pVersion.Version)
			}
		}
		if operator == "<=" {
			if lv.GreaterThan(rv) || lv.Equal(rv) {
				parsedVerionList = append(parsedVerionList, pVersion.Version)
			}
		}
	}
	ver_sort.Sort(ver_sort.ByVersion(parsedVerionList))
	return parsedVerionList[len(parsedVerionList)-1]
}

func parseVersion(ctx context.Context, module string, version string) string {
	parsed_version := strings.ReplaceAll(version, " ", "")
	if strings.HasPrefix(version, "=") {
		parsed_version = strings.TrimPrefix(parsed_version, "=")
	} else if strings.HasPrefix(version, ">=") {
		parsed_version = strings.TrimPrefix(parsed_version, ">=")
		parsed_version = tfRegistryGetVersionList(ctx, module, parsed_version, ">=")
	} else if strings.HasPrefix(version, ">") {
		parsed_version = strings.TrimPrefix(parsed_version, ">")
		parsed_version = tfRegistryGetVersionList(ctx, module, parsed_version, ">")
	} else if strings.HasPrefix(version, "<=") {
		parsed_version = strings.TrimPrefix(parsed_version, "<=")
		parsed_version = tfRegistryGetVersionList(ctx, module, parsed_version, "<=")
	} else if strings.HasPrefix(version, "<") {
		parsed_version = strings.TrimPrefix(parsed_version, "<")
		parsed_version = tfRegistryGetVersionList(ctx, module, parsed_version, "<")
	} else if strings.HasPrefix(version, "~>") {
		parsed_version = strings.TrimPrefix(parsed_version, "~>")
		parsed_version = tfRegistryGetVersionList(ctx, module, parsed_version, "~>")
	}

	return parsed_version
	//return fmt.Sprint(v)
}

func FetchModules(ctx context.Context, terrafile_path string, module_path string) {
	opts.TerrafilePath = terrafile_path
	opts.ModulePath = module_path
	fmt.Printf("Terrafile: version %v, commit %v, built at %v \n", version, commit, date)
	_, err := flags.Parse(&opts)

	// Invalid choice
	if err != nil {
		os.Exit(1)
	}

	// Read File
	yamlFile, err := ioutil.ReadFile(opts.TerrafilePath)
	if err != nil {
		tflog.Error(ctx, err.Error())
	}

	// Parse File
	var config map[string]module
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		tflog.Error(ctx, err.Error())
	}
	//get layer directory
	dir, err := os.Getwd()
	if err != nil {
		tflog.Error(ctx, err.Error())
	}
	tflog.Info(ctx, dir)
	running_modules = append(running_modules, dir)
	module_list, err := scanModules(running_modules)
	// Clone modules
	var wg sync.WaitGroup
	var parsed_version string
	_ = os.RemoveAll(opts.ModulePath)
	_ = os.MkdirAll(opts.ModulePath, os.ModePerm)
	for key, mod := range config {
		wg.Add(1)
		go func(m module, key string) {
			defer wg.Done()
			if contains(module_list, key) {
				if !strings.HasPrefix(m.Source, "git") && !strings.HasPrefix(m.Source, "https") {
					parsed_version = parseVersion(ctx, m.Source, m.Version)
					path_name := strings.Split(m.Source, "/")
					org_name := path_name[0]
					provider_name := path_name[2]
					module_name := path_name[1]
					str := []string{"terraform", provider_name, module_name}
					parsed_module_name := strings.Join(str, "-")
					m.Source = "git@github.com:" + org_name + "/" + parsed_module_name
				} else {
					parsed_version = m.Version
				}
				gitClone(ctx, m.Source, parsed_version, key)
				_ = os.RemoveAll(filepath.Join(opts.ModulePath, key, ".git"))
			}
		}(mod, key)
	}
	wg.Wait()
}
