package presetsources

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"testing/fstest"
)

const (
	defaultGitHubAPI = "https://api.github.com"
	maxArchiveBytes  = 64 << 20
	maxCatalogFiles  = 10_000
	maxCatalogFile   = 4 << 20
	maxCatalogBytes  = 32 << 20
)

var (
	githubOwnerPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,38})$`)
	githubRepoPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,99}$`)
)

func ParseGitHubRepository(value string) (string, error) {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("parse GitHub URL: %w", err)
		}
		if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") ||
			parsed.Port() != "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", errors.New("GitHub URL must be credential-free HTTPS on github.com")
		}
		value = strings.Trim(parsed.EscapedPath(), "/")
		if decoded, err := url.PathUnescape(value); err == nil {
			value = decoded
		}
	}
	value = strings.TrimSuffix(value, ".git")
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return "", errors.New("GitHub repository must be owner/repository or its HTTPS URL")
	}
	owner := strings.ToLower(parts[0])
	repository := strings.ToLower(parts[1])
	if !githubOwnerPattern.MatchString(owner) || !githubRepoPattern.MatchString(repository) ||
		repository == "." || repository == ".." {
		return "", errors.New("GitHub repository has an invalid owner or name")
	}
	return owner + "/" + repository, nil
}

func validateGitHubRef(value string) error {
	if value == "" || len(value) > 256 || strings.TrimSpace(value) != value ||
		strings.Contains(value, "..") || strings.Contains(value, "@{") ||
		strings.ContainsAny(value, " ~^:?*[\\") || strings.HasPrefix(value, "/") ||
		strings.HasSuffix(value, "/") || strings.HasSuffix(value, ".") ||
		strings.Contains(value, "//") {
		return errors.New("GitHub ref is invalid")
	}
	for _, component := range strings.Split(value, "/") {
		if strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return errors.New("GitHub ref is invalid")
		}
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return errors.New("GitHub ref contains control characters")
		}
	}
	return nil
}

func (m Manager) fetchGitHub(ctx context.Context, repository, ref string) (fs.FS, string, string, error) {
	api := strings.TrimRight(strings.TrimSpace(m.GitHubAPI), "/")
	if api == "" {
		api = defaultGitHubAPI
	}
	client := m.HTTPClient
	if client == nil {
		client = defaultHTTPClient()
	}
	getenv := m.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	token := strings.TrimSpace(getenv("GITHUB_TOKEN"))
	if ref == "" {
		metadataURL := api + "/repos/" + repository
		metadataResponse, err := githubRequest(
			ctx,
			client,
			metadataURL,
			token,
			"application/vnd.github+json",
		)
		if err != nil {
			return nil, "", "", fmt.Errorf("resolve GitHub default branch: %w", err)
		}
		var metadata struct {
			DefaultBranch string `json:"default_branch"`
		}
		decodeErr := json.NewDecoder(io.LimitReader(metadataResponse.Body, 1<<20)).Decode(&metadata)
		closeErr := metadataResponse.Body.Close()
		if decodeErr != nil {
			return nil, "", "", fmt.Errorf("decode GitHub repository metadata: %w", decodeErr)
		}
		if closeErr != nil {
			return nil, "", "", fmt.Errorf("close GitHub repository metadata: %w", closeErr)
		}
		ref = strings.TrimSpace(metadata.DefaultBranch)
	}
	if err := validateGitHubRef(ref); err != nil {
		return nil, "", "", err
	}
	revisionURL := api + "/repos/" + repository + "/commits/" + url.PathEscape(ref)
	revisionResponse, err := githubRequest(ctx, client, revisionURL, token, "application/vnd.github+json")
	if err != nil {
		return nil, "", "", fmt.Errorf("resolve GitHub revision: %w", err)
	}
	defer revisionResponse.Body.Close()
	var value struct {
		SHA string `json:"sha"`
	}
	decoder := json.NewDecoder(io.LimitReader(revisionResponse.Body, 1<<20))
	if err := decoder.Decode(&value); err != nil {
		return nil, "", "", fmt.Errorf("decode GitHub revision: %w", err)
	}
	value.SHA = strings.ToLower(strings.TrimSpace(value.SHA))
	if !revisionPattern.MatchString(value.SHA) {
		return nil, "", "", errors.New("GitHub returned an invalid commit revision")
	}

	archiveURL := api + "/repos/" + repository + "/zipball/" + value.SHA
	archiveResponse, err := githubRequest(ctx, client, archiveURL, token, "application/vnd.github+json")
	if err != nil {
		return nil, "", "", fmt.Errorf("download GitHub snapshot: %w", err)
	}
	defer archiveResponse.Body.Close()
	if archiveResponse.ContentLength > maxArchiveBytes {
		return nil, "", "", fmt.Errorf("GitHub archive exceeds %d bytes", maxArchiveBytes)
	}
	data, err := io.ReadAll(io.LimitReader(archiveResponse.Body, maxArchiveBytes+1))
	if err != nil {
		return nil, "", "", fmt.Errorf("read GitHub archive: %w", err)
	}
	if len(data) > maxArchiveBytes {
		return nil, "", "", fmt.Errorf("GitHub archive exceeds %d bytes", maxArchiveBytes)
	}
	assets, err := archiveCatalog(data)
	if err != nil {
		return nil, "", "", err
	}
	return assets, ref, value.SHA, nil
}

func githubRequest(
	ctx context.Context,
	client *http.Client,
	requestURL,
	token,
	accept string,
) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", accept)
	request.Header.Set("User-Agent", "maisternia-preset-sources")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		response.Body.Close()
		return nil, fmt.Errorf("GitHub returned %s", response.Status)
	}
	return response, nil
}

func archiveCatalog(data []byte) (fstest.MapFS, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open GitHub archive: %w", err)
	}
	assets := fstest.MapFS{}
	var prefix string
	var total int64
	for _, file := range reader.File {
		name := file.Name
		if name == "" || strings.ContainsRune(name, '\\') || strings.HasPrefix(name, "/") ||
			path.Clean(name) != strings.TrimSuffix(name, "/") {
			return nil, fmt.Errorf("GitHub archive contains unsafe path %q", name)
		}
		parts := strings.Split(strings.TrimSuffix(name, "/"), "/")
		if len(parts) == 0 || parts[0] == "" || parts[0] == "." || parts[0] == ".." {
			return nil, fmt.Errorf("GitHub archive contains unsafe path %q", name)
		}
		if prefix == "" {
			prefix = parts[0]
		} else if parts[0] != prefix {
			return nil, errors.New("GitHub archive contains multiple repository roots")
		}
		if file.Mode()&fs.ModeSymlink != 0 || (!file.FileInfo().IsDir() && !file.Mode().IsRegular()) {
			return nil, fmt.Errorf("GitHub archive path %q is not a regular file", name)
		}
		if file.FileInfo().IsDir() || len(parts) < 3 || parts[1] != "config" {
			continue
		}
		relative := strings.Join(parts[1:], "/")
		if !fs.ValidPath(relative) || path.Clean(relative) != relative {
			return nil, fmt.Errorf("GitHub archive contains unsafe catalog path %q", name)
		}
		if _, exists := assets[relative]; exists {
			return nil, fmt.Errorf("GitHub archive repeats catalog path %q", relative)
		}
		if len(assets) >= maxCatalogFiles || file.UncompressedSize64 > maxCatalogFile {
			return nil, errors.New("GitHub catalog exceeds extraction limits")
		}
		handle, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open GitHub archive path %q: %w", name, err)
		}
		content, readErr := io.ReadAll(io.LimitReader(handle, maxCatalogFile+1))
		closeErr := handle.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read GitHub archive path %q: %w", name, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close GitHub archive path %q: %w", name, closeErr)
		}
		if len(content) > maxCatalogFile {
			return nil, errors.New("GitHub catalog exceeds extraction limits")
		}
		total += int64(len(content))
		if total > maxCatalogBytes {
			return nil, errors.New("GitHub catalog exceeds extraction limits")
		}
		assets[relative] = &fstest.MapFile{Data: content, Mode: 0o600}
	}
	if prefix == "" || len(assets) == 0 {
		return nil, errors.New("GitHub archive contains no catalog files")
	}
	return assets, nil
}
