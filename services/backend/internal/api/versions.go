package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/catalog"
)

// allowedServices is a whitelist of service names for the generic logs endpoint.
var allowedServices = map[string]bool{
	"coraza-spoa": true,
	"backend":     true,
	"haproxy":     true,
	"fluentbit":   true,
	"postgres":    true,
}

// GetSystemVersions returns CRS versions for backend and coraza-spoa.
func (h *Handler) GetSystemVersions(w http.ResponseWriter, r *http.Request) {
	spoaVersion, spoaImage := getSPOAInfo()
	versions := map[string]string{
		"backend_crs_version": catalog.GetCRSVersion(),
		"spoa_crs_version":    spoaVersion,
		"spoa_image":          spoaImage,
	}
	jsonOK(w, versions)
}

// getSPOAInfo inspects the running coraza-spoa container and returns its CRS version + image name.
func getSPOAInfo() (crsVersion string, imageName string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Warn().Err(err).Msg("docker client init failed")
		return "unknown", "unknown"
	}
	defer cli.Close()

	ctx := context.Background()
	// Use label-based detection for robustness
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "com.docker.compose.service=coraza-spoa")
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true, // include stopped containers
		Filters: filterArgs,
	})
	if err != nil || len(containers) == 0 {
		// fallback: name-based search
		containers, err = cli.ContainerList(ctx, container.ListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("name", "coraza-spoa")),
		})
		if err != nil {
			log.Warn().Err(err).Msg("docker container list failed")
			return "unknown", "unknown"
		}
	}
	if len(containers) == 0 {
		log.Warn().Msg("coraza-spoa container not found")
		return "unknown", "unknown"
	}

	c := containers[0]
	imageName = c.Image
	if len(c.Names) > 0 {
		imageName = strings.TrimPrefix(c.Names[0], "/")
	}

	// Only exec --version if container is running
	if c.State == "running" {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if ver := execSPOAVersion(ctxTimeout, cli, c.ID); ver != "" {
			return ver, imageName
		}
	}

	// Fallback: use image tag as version hint
	return c.Image, imageName
}

// execSPOAVersion runs /coraza-spoa --version in the container and parses the CRS version.
func execSPOAVersion(ctx context.Context, cli *client.Client, containerID string) string {
	execCfg := container.ExecOptions{
		Cmd:          []string{"/coraza-spoa", "--version"},
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := cli.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		log.Warn().Err(err).Msg("exec create failed")
		return ""
	}

	resp, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("exec attach failed")
		return ""
	}
	defer resp.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Reader); err != nil && err != io.EOF {
		log.Warn().Err(err).Msg("exec read failed")
		return ""
	}

	output := buf.String()
	return parseCRSVersionFromOutput(output)
}

// parseCRSVersionFromOutput extracts a CRS version like "v4.25.0" from --version output or build info.
var crsVerRe = regexp.MustCompile(`(?:coraza-coreruleset(?:/v\d+)?)\s+(v\d+\.\d+\.\d+[\w.-]*)`)
var semverRe = regexp.MustCompile(`\b(v\d+\.\d+\.\d+[\w.-]*)\b`)

func parseCRSVersionFromOutput(output string) string {
	// Try to match dep line: "dep github.com/corazawaf/coraza-coreruleset/v4 v4.25.0 ..."
	if m := crsVerRe.FindStringSubmatch(output); m != nil {
		return m[1]
	}
	// Fallback: any semver in output (for --version output like "CRS v4.25.0")
	if strings.Contains(strings.ToLower(output), "crs") {
		if m := semverRe.FindStringSubmatch(output); m != nil {
			return m[1]
		}
	}
	return ""
}

// findContainerByService finds a container by compose service label, with name-based fallback.
// Pass all=true to include stopped containers.
func findContainerByService(ctx context.Context, cli *client.Client, service string, all bool) (string, bool, error) {
	// Primary: label-based
	labelArgs := filters.NewArgs()
	labelArgs.Add("label", "com.docker.compose.service="+service)
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     all,
		Filters: labelArgs,
	})
	if err != nil {
		return "", false, err
	}
	if len(containers) > 0 {
		return containers[0].ID, containers[0].State == "running", nil
	}
	// Fallback: name-based
	containers, err = cli.ContainerList(ctx, container.ListOptions{
		All:     all,
		Filters: filters.NewArgs(filters.Arg("name", service)),
	})
	if err != nil {
		return "", false, err
	}
	if len(containers) > 0 {
		return containers[0].ID, containers[0].State == "running", nil
	}
	return "", false, nil
}

// GetSPOALogs returns the last N log lines from the coraza-spoa container (backward compat).
func (h *Handler) GetSPOALogs(w http.ResponseWriter, r *http.Request) {
	// Delegate to generic handler with service=coraza-spoa
	q := r.URL.Query()
	if q.Get("service") == "" {
		// inject service param by wrapping
		q.Set("service", "coraza-spoa")
		r.URL.RawQuery = q.Encode()
	}
	h.GetServiceLogs(w, r)
}

// GetServiceLogs returns the last N log lines from any allowed service container.
// GET /api/system/logs?service=coraza-spoa&tail=100
func (h *Handler) GetServiceLogs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	if service == "" {
		service = "coraza-spoa"
	}
	if !allowedServices[service] {
		jsonError(w, "service not allowed", http.StatusBadRequest)
		return
	}

	tailStr := r.URL.Query().Get("tail")
	if tailStr == "" {
		tailStr = "100"
	}
	tailN, err := strconv.Atoi(tailStr)
	if err != nil || tailN < 1 || tailN > 500 {
		tailN = 100
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// Return 200 with error info — frontend should handle gracefully
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"lines": []string{},
			"count": 0,
			"error": "docker unavailable",
		})
		return
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	containerID, running, err := findContainerByService(ctx, cli, service, true)
	if err != nil {
		jsonError(w, "failed to list containers", http.StatusInternalServerError)
		return
	}
	if containerID == "" {
		// Return 200 with empty result — not 404 (frontend gets confused)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"lines":   []string{},
			"count":   0,
			"error":   "container not found",
			"running": false,
		})
		return
	}

	logOpts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(tailN),
		Timestamps: true,
	}

	// Docker allows fetching logs from stopped containers too
	reader, err := cli.ContainerLogs(ctx, containerID, logOpts)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"lines":   []string{},
			"count":   0,
			"error":   "failed to get logs: " + err.Error(),
			"running": running,
		})
		return
	}
	defer reader.Close()

	// Docker logs stream has 8-byte header per line (multiplexed stream)
	var buf bytes.Buffer
	stdcopy.StdCopy(&buf, &buf, reader)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"lines":   lines,
		"count":   len(lines),
		"running": running,
	})
}
