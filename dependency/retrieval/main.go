// Copyright 2018-2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	id   = "appinternals-java"
	name = "Aternity AppInternals Java Agent"
	purl = "pkg:generic/aternity-appinternals-java-agent"

	bucket = "appint-pcf-instrumentation-rpm-master"
	host   = "https://pcf-instrumentation-download.steelcentral.net"
)

var (
	keyPattern   = regexp.MustCompile(`^riverbed-appinternals-agent-(\d+)\.(\d+)\.(\d+)_(.+)\.zip$`)
	buildPattern = regexp.MustCompile(`(\d+)$`)
)

type listBucketResult struct {
	IsTruncated           bool   `xml:"IsTruncated"`
	NextContinuationToken string `xml:"NextContinuationToken"`
	Contents              []struct {
		Key string `xml:"Key"`
	} `xml:"Contents"`
}

type version struct {
	version *semver.Version
	build   string
	key     string
}

func (v version) Version() *semver.Version {
	return v.version
}

func main() {
	retrieve.NewMetadata(id, getAllVersions, generateMetadata)
}

func getAllVersions() (versionology.VersionFetcherArray, error) {
	keys, err := listKeys(bucket)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects in %s\n%w", bucket, err)
	}

	latest := map[string]version{}
	for _, key := range keys {
		m := keyPattern.FindStringSubmatch(key)
		if m == nil {
			continue
		}

		versionString := fmt.Sprintf("%s.%s.%s", m[1], m[2], m[3])
		v, err := semver.NewVersion(versionString)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", versionString)
			continue
		}

		candidate := version{version: v, build: m[4], key: key}
		if existing, ok := latest[versionString]; !ok || buildLess(existing.build, candidate.build) {
			latest[versionString] = candidate
		}
	}

	var versions versionology.VersionFetcherArray
	for _, v := range latest {
		versions = append(versions, v)
	}

	return versions, nil
}

func generateMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	v, ok := versionFetcher.(version)
	if !ok {
		return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
	}

	versionString := v.version.String()

	uri := fmt.Sprintf("%s/%s", host, v.key)
	checksum, err := upstream.GetSHA256OfRemoteFile(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", uri, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:aternity:steelcentral_appinternals:%s:*:*:*:*:*:*:*", versionString),
		ID:       id,
		Licenses: []interface{}{
			map[string]string{
				"uri": "https://www.riverbed.com/document/legal/riverbed-end-user-license-agreement.pdf",
			},
		},
		Name:    name,
		PURL:    fmt.Sprintf("%s@%s?arch=amd64", purl, versionString),
		SHA256:  checksum,
		Stacks:  []string{"*"},
		URI:     uri,
		Version: versionString,
	}

	return versionology.NewDependencyArray(dependency, "")
}

// buildLess reports whether build a is older than build b. Builds are strings
// such as BL527; compare their numeric suffix when both have one, falling back
// to a plain string comparison.
func buildLess(a, b string) bool {
	an := buildPattern.FindString(a)
	bn := buildPattern.FindString(b)
	if an == "" || bn == "" {
		return a < b
	}

	ai, err := strconv.Atoi(an)
	if err != nil {
		return a < b
	}
	bi, err := strconv.Atoi(bn)
	if err != nil {
		return a < b
	}

	return ai < bi
}

func listKeys(bucket string) ([]string, error) {
	var keys []string
	continuationToken := ""

	for {
		q := url.Values{}
		q.Set("list-type", "2")
		if continuationToken != "" {
			q.Set("continuation-token", continuationToken)
		}

		resp, err := http.Get(fmt.Sprintf("https://%s.s3.amazonaws.com/?%s", bucket, q.Encode()))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("S3 returned status %d", resp.StatusCode)
		}

		var result listBucketResult
		if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
		_ = resp.Body.Close()

		for _, c := range result.Contents {
			keys = append(keys, c.Key)
		}

		if !result.IsTruncated || result.NextContinuationToken == "" {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	sort.Strings(keys)
	return keys, nil
}
