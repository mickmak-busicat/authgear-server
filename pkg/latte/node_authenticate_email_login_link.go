package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticateEmailLoginLink{})
}

type NodeAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (n *NodeAuthenticateEmailLoginLink) Kind() string {
	return "latte.NodeAuthenticateEmailLoginLink"
}

func (n *NodeAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputCheckLoginLinkVerified{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputCheckLoginLinkVerified inputCheckLoginLinkVerified
	var inputResendOOBOTPCode inputResendOOBOTPCode
	switch {
	case workflow.AsInput(input, &inputResendOOBOTPCode):
		info := n.Authenticator
		_, err := (&SendOOBCode{
			WorkflowID:        workflow.GetWorkflowID(ctx),
			Deps:              deps,
			Stage:             authenticatorKindToStage(info.Kind),
			IsAuthenticating:  true,
			AuthenticatorInfo: info,
			OTPMode:           otp.OTPModeMagicLink,
		}).Do()
		if err != nil {
			return nil, err
		}
		return nil, workflow.ErrSameNode
	case workflow.AsInput(input, &inputCheckLoginLinkVerified):
		info := n.Authenticator
		_, err := deps.OTPCodes.VerifyMagicLinkCodeByTarget(info.OOBOTP.Email, true)
		if errors.Is(err, otp.ErrInvalidCode) {
			// Don't fire the AuthenticationFailedEvent
			// The event should be fired only when the user submits code through the login link
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
