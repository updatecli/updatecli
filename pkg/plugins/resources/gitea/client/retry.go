package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"
)

// ErrAutoMergeAlreadyScheduled indicates that the pull request already has an auto-merge schedule.
var ErrAutoMergeAlreadyScheduled = errors.New("pull request is already scheduled to auto merge when checks succeed")

// The SDK discards merge response bodies, so inspect them before returning.
func retryMerge(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	if resp.Request.Method != http.MethodPost ||
		!strings.Contains(resp.Request.URL.Path, "/pulls/") ||
		!strings.HasSuffix(resp.Request.URL.Path, "/merge") {
		return false, nil
	}
	if resp.StatusCode != http.StatusMethodNotAllowed &&
		resp.StatusCode != http.StatusConflict {
		return false, nil
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return false, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	var rejection struct {
		Message string
	}
	if json.Unmarshal(body, &rejection) != nil {
		return false, nil
	}
	if resp.StatusCode == http.StatusConflict && strings.HasPrefix(rejection.Message, ErrAutoMergeAlreadyScheduled.Error()+" [pull_id: ") {
		return false, ErrAutoMergeAlreadyScheduled
	}
	if resp.StatusCode == http.StatusMethodNotAllowed && rejection.Message == "Please try again later" {
		log.Debugln("Gitea asked to retry the merge request")
		return true, nil
	}
	return false, nil
}
