package torbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"scrapper/internal/config"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type Client struct {
	baseURL   *url.URL
	apiKey    string
	http      *http.Client
	cacheOnly bool
	logger    *log.Logger
}

type APIError struct {
	Code   string
	Detail string
}

func (e *APIError) Error() string {
	if e.Code == "" {
		return e.Detail
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

type Response[T any] struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Detail  string `json:"detail"`
	Data    T      `json:"data"`
}

type TorrentCreateResult struct {
	TorrentID int    `json:"torrent_id"`
	AuthID    string `json:"auth_id"`
	Hash      string `json:"hash"`
}

type TorrentFile struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Size      int64  `json:"size"`
	Mimetype  string `json:"mimetype"`
	S3Path    string `json:"s3_path"`
}

type Torrent struct {
	ID               int           `json:"id"`
	Hash             string        `json:"hash"`
	Name             string        `json:"name"`
	DownloadPresent  bool          `json:"download_present"`
	DownloadFinished bool          `json:"download_finished"`
	Files            []TorrentFile `json:"files"`
}

func New(torBox config.TorboxConfig, httpClient *http.Client, logger *log.Logger) (*Client, error) {
	apiKey := os.Getenv("TORBOX_API_KEY")
	if apiKey == "" {
		return nil, errors.New("missing Torbox API key")
	}

	baseUrl := torBox.BaseURL
	if baseUrl == "" {
		return nil, errors.New("missing Torbox base URL")
	}

	parsed, err := url.Parse(strings.TrimRight(baseUrl, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse torbox base url: %w", err)
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(torBox.Timeout) * time.Second,
		}
	}

	return &Client{
		baseURL:   parsed,
		apiKey:    apiKey,
		http:      httpClient,
		cacheOnly: torBox.CachedOnly,
		logger:    logger,
	}, nil
}

func (c *Client) CreateTorrent(ctx context.Context, magnet string, addOnlyIfCached bool) (TorrentCreateResult, error) {
	var result TorrentCreateResult

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("magnet", magnet)
	if addOnlyIfCached {
		_ = writer.WriteField("add_only_if_cached", "true")
	}
	_ = writer.Close()

	if err := c.doJSON(ctx, http.MethodPost, "/torrents/createtorrent", nil, body, writer.FormDataContentType(), &result); err != nil {
		return TorrentCreateResult{}, err
	}

	return result, nil
}

func (c *Client) GetTorrent(ctx context.Context, torrentID int) (Torrent, error) {
	var raw json.RawMessage
	q := url.Values{}
	q.Set("id", strconv.Itoa(torrentID))
	if err := c.doJSON(ctx, http.MethodGet, "/torrents/mylist", q, nil, "", &raw); err != nil {
		return Torrent{}, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return Torrent{}, errors.New("torrent not found")
	}

	var list []Torrent
	if err := json.Unmarshal(raw, &list); err == nil {
		if len(list) == 0 {
			return Torrent{}, errors.New("torrent not found")
		}
		return list[0], nil
	}

	var single Torrent
	if err := json.Unmarshal(raw, &single); err == nil {
		if single.ID == 0 {
			return Torrent{}, errors.New("torrent not found")
		}
		return single, nil
	}

	return Torrent{}, errors.New("torrent not found")
}

func (c *Client) FindTorrentByHash(ctx context.Context, hash string) (Torrent, error) {
	const pageSize = 100
	for offset := 0; offset < 500; offset += pageSize {
		var result []Torrent
		q := url.Values{}
		q.Set("offset", strconv.Itoa(offset))
		q.Set("limit", strconv.Itoa(pageSize))
		if err := c.doJSON(ctx, http.MethodGet, "/torrents/mylist", q, nil, "", &result); err != nil {
			return Torrent{}, err
		}
		if len(result) == 0 {
			break
		}
		for _, t := range result {
			if strings.EqualFold(t.Hash, hash) {
				return t, nil
			}
		}
	}
	return Torrent{}, errors.New("torrent not found")
}

func (c *Client) RequestDownload(ctx context.Context, torrentID int, fileID int) (string, error) {
	q := url.Values{}
	q.Set("token", c.apiKey)
	q.Set("torrent_id", strconv.Itoa(torrentID))
	q.Set("file_id", strconv.Itoa(fileID))
	q.Set("redirect", "false")
	var result string
	if err := c.doJSON(ctx, http.MethodGet, "/torrents/requestdl", q, nil, "", &result); err != nil {
		return "", err
	}
	return result, nil
}

func (c *Client) RequestDownloadWithTorrentId(ctx context.Context, torrentID int) (string, error) {
	q := url.Values{}
	q.Set("token", c.apiKey)
	q.Set("torrent_id", strconv.Itoa(torrentID))
	q.Set("redirect", "false")
	var result string
	if err := c.doJSON(ctx, http.MethodGet, "/torrents/requestdl", q, nil, "", &result); err != nil {
		return "", err
	}
	return result, nil
}

func (c *Client) CheckCached(ctx context.Context, hash string) (bool, error) {
	q := url.Values{}
	q.Set("hash", hash)
	q.Set("format", "list")
	q.Set("list_files", "false")
	var raw json.RawMessage
	if err := c.doJSON(ctx, http.MethodGet, "/torrents/checkcached", q, nil, "", &raw); err != nil {
		return false, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return false, nil
	}

	var list []struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(raw, &list); err == nil {
		return len(list) > 0, nil
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		return len(obj) > 0, nil
	}

	return false, nil
}

func (c *Client) doJSON(ctx context.Context, method string, endpoint string, query url.Values, body io.Reader, contentType string, out any) error {
	u := *c.baseURL
	endpoint = strings.TrimPrefix(endpoint, "/")
	u.Path = path.Join(u.Path, endpoint)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if method == http.MethodPost && contentType == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var wrapper Response[json.RawMessage]
	if err := dec.Decode(&wrapper); err != nil {
		return err
	}

	if !wrapper.Success {
		return &APIError{Code: wrapper.Error, Detail: wrapper.Detail}
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(wrapper.Data, out); err != nil {
		return err
	}
	return nil
}
