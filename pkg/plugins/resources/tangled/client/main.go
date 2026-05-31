package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/sirupsen/logrus"
)

// Client wraps an indigo atproto APIClient authenticated with an app password.
type Client struct {
	spec Spec

	mu  sync.Mutex
	api *atclient.APIClient
	did syntax.DID
}

// Clients are keyed by (PDS, identifier) and shared process-wide. Pipelines
// that share credentials reuse one atproto session, otherwise the per-IP
// rate limit on com.atproto.server.createSession kicks in after a few
// concurrent SCMs.
var (
	clientCacheMu sync.Mutex
	clientCache   = map[string]*Client{}
)

// cacheKey identifies a Client by its full visible configuration. Appview is
// part of the key because two specs targeting the same account but
// different Tangled appviews must not share a Client — otherwise the second
// caller would generate pull request links against the appview the first
// caller registered.
func cacheKey(s Spec) string {
	id := s.DID
	if id == "" {
		id = s.Handle
	}
	return s.PDS + "|" + id + "|" + s.Appview
}

// New returns a Client. Clients constructed with the same (PDS, identifier,
// appview) triple share an authenticated atproto session.
func New(s Spec) (*Client, error) {
	key := cacheKey(s)

	clientCacheMu.Lock()
	defer clientCacheMu.Unlock()
	if existing, ok := clientCache[key]; ok {
		return existing, nil
	}

	c := &Client{spec: s}
	clientCache[key] = c
	return c, nil
}

// DID returns the DID of the authenticated account.
func (c *Client) DID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.did != "" {
		return c.did.String()
	}
	return c.spec.DID
}

// Appview returns the configured Tangled appview URL.
func (c *Client) Appview() string {
	return c.spec.Appview
}

// PDS returns the configured PDS URL.
func (c *Client) PDS() string {
	return c.spec.PDS
}

// API returns the underlying APIClient, creating an authenticated session if needed.
func (c *Client) API(ctx context.Context) (*atclient.APIClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.api != nil {
		return c.api, nil
	}

	identifier := c.spec.DID
	if identifier == "" {
		identifier = strings.TrimPrefix(c.spec.Handle, "@")
	}

	var (
		api *atclient.APIClient
		err error
	)
	if c.spec.PDS != "" {
		api, err = atclient.LoginWithPasswordHost(ctx, c.spec.PDS, identifier, c.spec.AppPassword, "", nil)
	} else {
		atIdent, perr := syntax.ParseAtIdentifier(identifier)
		if perr != nil {
			return nil, fmt.Errorf("parse atproto identifier %q: %w", identifier, perr)
		}
		api, err = atclient.LoginWithPassword(ctx, identity.DefaultDirectory(), atIdent, c.spec.AppPassword, "", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("login with app password: %w", err)
	}

	c.api = api
	if api.AccountDID != nil {
		c.did = *api.AccountDID
	}
	logrus.Debugf("tangled: authenticated session for %s", c.did)
	return api, nil
}

// PutRecord wraps com.atproto.repo.putRecord.
func (c *Client) PutRecord(ctx context.Context, collection, rkey string, record lexutil.CBOR) error {
	api, err := c.API(ctx)
	if err != nil {
		return err
	}

	_, err = atproto.RepoPutRecord(ctx, api, &atproto.RepoPutRecord_Input{
		Repo:       c.DID(),
		Collection: collection,
		Rkey:       rkey,
		Record:     &lexutil.LexiconTypeDecoder{Val: record},
	})
	if err != nil {
		return fmt.Errorf("put record: %w", err)
	}
	return nil
}

// UploadBlob wraps com.atproto.repo.uploadBlob.
func (c *Client) UploadBlob(ctx context.Context, data []byte, mimeType string) (*lexutil.LexBlob, error) {
	api, err := c.API(ctx)
	if err != nil {
		return nil, err
	}

	// indigo's RepoUploadBlob drops the Content-Type header; force it back on
	// the returned ref so the record advertises the mime type we actually
	// uploaded.
	out, err := atproto.RepoUploadBlob(ctx, api, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("upload blob: %w", err)
	}
	if out.Blob != nil && mimeType != "" {
		out.Blob.MimeType = mimeType
	}
	return out.Blob, nil
}

// ListRecords iterates all records of the given collection on the given repo DID.
func (c *Client) ListRecords(ctx context.Context, repo, collection string) ([]*atproto.RepoListRecords_Record, error) {
	api, err := c.API(ctx)
	if err != nil {
		return nil, err
	}

	var all []*atproto.RepoListRecords_Record
	cursor := ""
	for {
		out, err := atproto.RepoListRecords(ctx, api, collection, cursor, 100, repo, false)
		if err != nil {
			return nil, fmt.Errorf("list records: %w", err)
		}
		all = append(all, out.Records...)
		if out.Cursor == nil || *out.Cursor == "" || len(out.Records) == 0 {
			break
		}
		cursor = *out.Cursor
	}
	return all, nil
}

// ResolveHandle resolves an atproto handle to a DID via the default identity directory.
func (c *Client) ResolveHandle(ctx context.Context, handle string) (string, error) {
	h, err := syntax.ParseHandle(strings.TrimPrefix(handle, "@"))
	if err != nil {
		return "", fmt.Errorf("parse handle: %w", err)
	}
	ident, err := identity.DefaultDirectory().LookupHandle(ctx, h)
	if err != nil {
		return "", fmt.Errorf("resolve handle: %w", err)
	}
	return ident.DID.String(), nil
}

// ResolvedDID returns the DID for the configured spec, resolving the handle if no DID is set.
func (c *Client) ResolvedDID(ctx context.Context) (string, error) {
	if c.spec.DID != "" {
		return c.spec.DID, nil
	}
	if c.spec.Handle == "" {
		return "", fmt.Errorf("neither did nor handle configured")
	}
	return c.ResolveHandle(ctx, c.spec.Handle)
}
