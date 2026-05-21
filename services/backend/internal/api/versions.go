package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/catalog"
)

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
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", "coraza-spoa")),
	})
	if err != nil {
		log.Warn().Err(err).Msg("docker container list failed")
		return "unknown", "unknown"
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

	// Try exec /coraza-spoa --version in the running container
	if ver := execSPOAVersion(ctx, cli, c.ID); ver != "" {
		return ver, imageName
	}

	// Fallback: use image tag as version hint
	if len(containers[0].Names) > 0 {
		return c.Image, imageName
	}
	return "unknown", imageName
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
