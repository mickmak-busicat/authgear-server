package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	authSession "github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachGetHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/session/get", &GetHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type GetHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f GetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &GetHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h)
}

type GetRequestPayload struct {
	SessionID string `json:"session_id"`
}

// @JSONSchema
const GetRequestSchema = `
{
	"$id": "#SessionGetRequest",
	"type": "object",
	"properties": {
		"session_id": { "type": "string" }
	}
}
`

func (p GetRequestPayload) Validate() error {
	return nil
}

type GetResponse struct {
	Session model.Session `json:"session"`
}

// @JSONSchema
const GetResponseSchema = `
{
	"$id": "#SessionGetResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"session": { "$ref": "#Session" }
			}
		}
	}
}
`

/*
	@Operation POST /session/get - Get current user sessions
		Get the sessions with specified ID of current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the session ID.
			@JSONSchema {SessionGetRequest}

		@Response 200
			The requested session.
			@JSONSchema {SessionGetResponse}
*/
type GetHandler struct {
	AuthContext     coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	RequireAuthz    handler.RequireAuthz   `dependency:"RequireAuthz"`
	TxContext       db.TxContext           `dependency:"TxContext"`
	SessionProvider session.Provider       `dependency:"SessionProvider"`
}

func (h GetHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h GetHandler) WithTx() bool {
	return true
}

func (h GetHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := GetRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h GetHandler) Handle(req interface{}) (resp interface{}, err error) {
	authInfo, _ := h.AuthContext.AuthInfo()
	userID := authInfo.ID
	sessionID := req.(GetRequestPayload).SessionID

	s, err := h.SessionProvider.Get(sessionID)
	if err != nil {
		if err == session.ErrSessionNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "session not found")
		}
		return
	}
	if s.UserID != userID {
		err = skyerr.NewError(skyerr.ResourceNotFound, "session not found")
		return
	}

	resp = GetResponse{Session: authSession.Format(s)}
	return
}
