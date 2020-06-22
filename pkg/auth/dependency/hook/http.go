package hook

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

const HeaderRequestBodySignature = "x-authgear-body-signature"

type SyncHTTPClient struct {
	*http.Client
}

func NewSyncHTTPClient(c *config.HookConfig) SyncHTTPClient {
	return SyncHTTPClient{
		httputil.NewExternalClient(c.SyncTimeout.Duration()),
	}
}

type AsyncHTTPClient struct {
	*http.Client
}

func NewAsyncHTTPClient() AsyncHTTPClient {
	return AsyncHTTPClient{
		httputil.NewExternalClient(60 * time.Second),
	}
}
