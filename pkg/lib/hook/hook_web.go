package hook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

type WebHookImpl struct {
	Secret    *config.WebhookKeyMaterials
	SyncHTTP  SyncHTTPClient
	AsyncHTTP AsyncHTTPClient
}

var _ WebHook = &WebHookImpl{}

func (h *WebHookImpl) SupportURL(u *url.URL) bool {
	return u.Scheme == "http" || u.Scheme == "https"
}

func (h *WebHookImpl) CallSync(u *url.URL, body interface{}, timeout *time.Duration) (*http.Response, error) {
	request, err := h.prepareRequest(u, body)
	if err != nil {
		return nil, err
	}

	var client *http.Client = h.SyncHTTP.Client
	if timeout != nil {
		client = httputil.NewExternalClient(*timeout)
	}

	resp, err := h.performRequest(client, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *WebHookImpl) DeliverBlockingEvent(u *url.URL, e *event.Event) (*event.HookResponse, error) {
	request, err := h.prepareEventRequest(u, e)
	if err != nil {
		return nil, err
	}

	resp, err := h.performEventRequest(h.SyncHTTP.Client, request, true)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *WebHookImpl) DeliverNonBlockingEvent(u *url.URL, e *event.Event) error {
	request, err := h.prepareEventRequest(u, e)
	if err != nil {
		return err
	}

	_, err = h.performEventRequest(h.AsyncHTTP.Client, request, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *WebHookImpl) prepareRequest(u *url.URL, rawBody interface{}) (*http.Request, error) {
	body, err := json.Marshal(rawBody)
	if err != nil {
		return nil, err
	}

	key, err := jwkutil.ExtractOctetKey(h.Secret.Set, "")
	if err != nil {
		return nil, err
	}
	signature := crypto.HMACSHA256String(key, body)

	request, err := http.NewRequest("POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(HeaderRequestBodySignature, signature)

	return request, nil
}

func (h *WebHookImpl) prepareEventRequest(u *url.URL, event *event.Event) (*http.Request, error) {
	return h.prepareRequest(u, event)
}

func (h *WebHookImpl) performRequest(client *http.Client, request *http.Request) (resp *http.Response, err error) {
	resp, err = client.Do(request)
	if os.IsTimeout(err) {
		err = WebHookDeliveryTimeout.New("webhook delivery timeout")
		return
	} else if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = WebHookInvalidResponse.NewWithInfo("invalid status code", apierrors.Details{
			"status_code": resp.StatusCode,
		})
		return
	}

	return resp, nil
}

func (h *WebHookImpl) performEventRequest(client *http.Client, request *http.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	resp, err := h.performRequest(client, request)

	defer resp.Body.Close()

	if !withResponse {
		return
	}

	hookResp, err = event.ParseHookResponse(resp.Body)
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = WebHookInvalidResponse.NewWithInfo("invalid response body", apiError.Info)
		return
	}

	return
}
