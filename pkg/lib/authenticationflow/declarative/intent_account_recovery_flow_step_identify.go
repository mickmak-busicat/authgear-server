package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepIdentify{})
}

type IntentAccountRecoveryFlowStepIdentifyData struct {
	Options []AccountRecoveryIdentificationOption `json:"options"`
}

var _ authflow.Data = IntentAccountRecoveryFlowStepIdentifyData{}

func (IntentAccountRecoveryFlowStepIdentifyData) Data() {}

type IntentAccountRecoveryFlowStepIdentify struct {
	JSONPointer jsonpointer.T                         `json:"json_pointer,omitempty"`
	StepName    string                                `json:"step_name,omitempty"`
	Options     []AccountRecoveryIdentificationOption `json:"options"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepIdentify{}

func (i *IntentAccountRecoveryFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentAccountRecoveryFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentAccountRecoveryFlowStepIdentify{}

func NewIntentAccountRecoveryFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentAccountRecoveryFlowStepIdentify) (*IntentAccountRecoveryFlowStepIdentify, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []AccountRecoveryIdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
			c := AccountRecoveryIdentificationOption{Identification: b.Identification}
			options = append(options, c)
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentAccountRecoveryFlowStepIdentify) Kind() string {
	return "IntentAccountRecoveryFlowStepIdentify"
}

func (i *IntentAccountRecoveryFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaStepAccountRecoveryIdentify{
			JSONPointer: i.JSONPointer,
			Options:     i.Options,
		}, nil
	}

	_, identityUsed := authflow.FindMilestone[MilestoneDoUseAccountRecoveryIdentificationMethod](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityUsed && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentAccountRecoveryFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountRecoveryIdentificationMethod inputTakeAccountRecoveryIdentificationMethod
		if authflow.AsInput(input, &inputTakeAccountRecoveryIdentificationMethod) {
			identification := inputTakeAccountRecoveryIdentificationMethod.GetAccountRecoveryIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}
			branch := step.OneOf[idx]

			switch identification {
			case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
				return authflow.NewNodeSimple(&NodeUseAccountRecoveryIdentity{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
					OnFailure:      branch.OnFailure,
				}), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, identityUsed := authflow.FindMilestone[MilestoneDoUseAccountRecoveryIdentificationMethod](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityUsed && !nestedStepsHandled:
		identification := i.identificationMethod(flows.Nearest)
		return authflow.NewSubFlow(&IntentAccountRecoveryFlowSteps{
			JSONPointer: i.jsonPointer(step, identification),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentAccountRecoveryFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return IntentAccountRecoveryFlowStepIdentifyData{
		Options: i.Options,
	}, nil
}

func (*IntentAccountRecoveryFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowAccountRecoveryFlowStep {
	step, ok := o.(*config.AuthenticationFlowAccountRecoveryFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentAccountRecoveryFlowStepIdentify) checkIdentificationMethod(
	deps *authflow.Dependencies,
	step *config.AuthenticationFlowAccountRecoveryFlowStep,
	im config.AuthenticationFlowAccountRecoveryIdentification,
) (idx int, err error) {
	idx = -1

	for index, branch := range step.OneOf {
		branch := branch
		if im == branch.Identification {
			idx = index
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (*IntentAccountRecoveryFlowStepIdentify) identificationMethod(w *authflow.Flow) config.AuthenticationFlowAccountRecoveryIdentification {
	m, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryIdentificationMethod](w)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneDoUseAccountRecoveryIdentificationMethod()

	return im
}

func (i *IntentAccountRecoveryFlowStepIdentify) jsonPointer(
	step *config.AuthenticationFlowAccountRecoveryFlowStep,
	im config.AuthenticationFlowAccountRecoveryIdentification,
) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}
