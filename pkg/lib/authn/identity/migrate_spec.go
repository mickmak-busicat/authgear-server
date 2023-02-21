package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type MigrateSpec struct {
	Type model.IdentityType `json:"type"`

	LoginID *LoginIDMigrateSpec `json:"login_id,omitempty"`
}

func (s *MigrateSpec) GetSpec() *Spec {
	return &Spec{
		Type: s.Type,
		LoginID: &LoginIDSpec{
			Type:  s.LoginID.Type,
			Key:   s.LoginID.Key,
			Value: s.LoginID.Value,
		},
	}
}

type LoginIDMigrateSpec struct {
	Key   string               `json:"key"`
	Type  model.LoginIDKeyType `json:"type"`
	Value string               `json:"value"`
}
