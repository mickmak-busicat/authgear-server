package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type IntentSignupFlowVerifyTarget interface {
	GetVerifiableClaims(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error)
	GetPurpose(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) otp.Purpose
	GetMessageType(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) otp.MessageType
}

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowVerify{})
}

var IntentSignupFlowVerifySchema = validation.NewSimpleSchema(`{}`)

type IntentSignupFlowVerify struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowVerify{}

func (i *IntentSignupFlowVerify) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowVerify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ workflow.Intent = &IntentSignupFlowVerify{}

func (*IntentSignupFlowVerify) Kind() string {
	return "workflowconfig.IntentSignupFlowVerify"
}

func (*IntentSignupFlowVerify) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowVerifySchema
}

func (*IntentSignupFlowVerify) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Look up the claim to verify
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowVerify) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := i.current(deps)
	if err != nil {
		return nil, err
	}

	step := i.step(current)
	targetStepID := step.TargetStep

	// Find the target step from the root.
	targetStepWorkflow, err := FindTargetStep(workflows.Root, targetStepID)
	if err != nil {
		return nil, err
	}

	target, ok := targetStepWorkflow.Intent.(IntentSignupFlowVerifyTarget)
	if !ok {
		return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepID,
		})
	}

	claims, err := target.GetVerifiableClaims(ctx, deps, workflows.Replace(targetStepWorkflow))
	if err != nil {
		return nil, err
	}

	if len(claims) == 0 {
		// Nothing to verify. End this workflow.
		return workflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	var claimNames []model.ClaimName
	for claimName := range claims {
		claimNames = append(claimNames, claimName)
	}

	if len(claimNames) > 1 {
		// TODO(workflow): support verify more than 1 claim?
		return nil, InvalidTargetStep.NewWithInfo("target_step contains more than one claim to verify", apierrors.Details{
			"claims": claimNames,
		})
	}

	claimName := claimNames[0]
	switch claimName {
	case model.ClaimEmail:
		break
	case model.ClaimPhoneNumber:
		break
	default:
		return nil, InvalidTargetStep.NewWithInfo("target_step contains a claim that cannot be verified", apierrors.Details{
			"claim_name": claimName,
		})
	}

	purpose := target.GetPurpose(ctx, deps, workflows.Replace(targetStepWorkflow))
	messageType := target.GetMessageType(ctx, deps, workflows.Replace(targetStepWorkflow))
	claimValue := claims[claimName]
	return workflow.NewNodeSimple(&NodeVerifyClaimSelectChannel{
		UserID:      i.UserID,
		Purpose:     purpose,
		MessageType: messageType,
		ClaimName:   claimName,
		ClaimValue:  claimValue,
	}), nil
}

func (*IntentSignupFlowVerify) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentSignupFlowVerify) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentSignupFlowVerify) current(deps *workflow.Dependencies) (config.WorkflowObject, error) {
	root, err := findSignupFlow(deps.Config.Workflow, i.SignupFlow)
	if err != nil {
		return nil, err
	}

	entries, err := Traverse(root, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func (*IntentSignupFlowVerify) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}
