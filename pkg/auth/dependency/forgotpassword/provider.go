package forgotpassword

import (
	"context"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	taskspec "github.com/authgear/authgear-server/pkg/auth/task/spec"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/log"
	"github.com/authgear/authgear-server/pkg/mail"
	"github.com/authgear/authgear-server/pkg/sms"
	"github.com/authgear/authgear-server/pkg/task"
	"github.com/authgear/authgear-server/pkg/template"
)

type AuthenticatorService interface {
	List(userID string, typ authn.AuthenticatorType) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
	WithSecret(ai *authenticator.Info, secret string) (bool, *authenticator.Info, error)
}

type LoginIDProvider interface {
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
	IsLoginIDKeyType(loginIDKey string, loginIDKeyType config.LoginIDKeyType) bool
}

type URLProvider interface {
	ResetPasswordURL(code string) *url.URL
}

type ProviderLogger struct{ *log.Logger }

func NewProviderLogger(lf *log.Factory) ProviderLogger {
	return ProviderLogger{lf.New("forgotpassword")}
}

type Provider struct {
	Context      context.Context
	ServerConfig *config.ServerConfig
	Localization *config.LocalizationConfig
	AppMetadata  config.AppMetadata
	Messaging    *config.MessagingConfig
	Config       *config.ForgotPasswordConfig

	Store          *Store
	Clock          clock.Clock
	URLs           URLProvider
	TemplateEngine *template.Engine
	TaskQueue      task.Queue

	Logger ProviderLogger

	LoginIDProvider LoginIDProvider
	Authenticators  AuthenticatorService
}

// SendCode checks if loginID is an existing login ID.
// For first matched login ID, a code is generated.
// Other matched login IDs are ignored.
// The code expires after a specific time.
// The code becomes invalid if it is consumed.
// Finally the code is sent to the login ID asynchronously.
func (p *Provider) SendCode(loginID string) (err error) {
	idens, err := p.LoginIDProvider.GetByLoginID(
		loginid.LoginID{
			Key:   "",
			Value: loginID,
		},
	)
	if err != nil {
		return
	}

	for _, iden := range idens {
		email := p.LoginIDProvider.IsLoginIDKeyType(iden.LoginIDKey, config.LoginIDKeyTypeEmail)
		phone := p.LoginIDProvider.IsLoginIDKeyType(iden.LoginIDKey, config.LoginIDKeyTypePhone)

		if !email && !phone {
			continue
		}

		code, codeStr := p.newCode(iden.UserID)

		err = p.Store.Create(code)
		if err != nil {
			return
		}

		if email {
			p.Logger.Debugf("sending email")
			err = p.sendEmail(iden.LoginID, codeStr)
			return
		}

		if phone {
			p.Logger.Debugf("sending sms")
			err = p.sendSMS(iden.LoginID, codeStr)
			return
		}
	}

	return
}

func (p *Provider) newCode(userID string) (code *Code, codeStr string) {
	createdAt := p.Clock.NowUTC()
	codeStr = GenerateCode()
	expireAt := createdAt.Add(p.Config.ResetCodeExpiry.Duration())
	code = &Code{
		CodeHash:  HashCode(codeStr),
		UserID:    userID,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		Consumed:  false,
	}
	return
}

func (p *Provider) sendEmail(email string, code string) (err error) {
	u := p.URLs.ResetPasswordURL(code)

	data := map[string]interface{}{
		"static_asset_url_prefix": p.ServerConfig.StaticAsset.URLPrefix,
		"email":                   email,
		"code":                    code,
		"link":                    u.String(),
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
	data["appname"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.Localization.FallbackLanguage), p.AppMetadata, "app_name")

	textBody, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordEmailTXT,
		data,
	)
	if err != nil {
		return
	}

	htmlBody, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordEmailHTML,
		data,
	)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: []mail.SendOptions{
				{
					MessageConfig: config.NewEmailMessageConfig(
						p.Messaging.DefaultEmailMessage,
						p.Config.EmailMessage,
					),
					Recipient: email,
					TextBody:  textBody,
					HTMLBody:  htmlBody,
				},
			},
		},
	})

	return
}

func (p *Provider) sendSMS(phone string, code string) (err error) {
	u := p.URLs.ResetPasswordURL(code)

	data := map[string]interface{}{
		"code": code,
		"link": u.String(),
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
	data["appname"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.Localization.FallbackLanguage), p.AppMetadata, "app_name")

	body, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordSMSTXT,
		data,
	)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			SMSMessages: []sms.SendOptions{
				{
					MessageConfig: config.NewSMSMessageConfig(
						p.Messaging.DefaultSMSMessage,
						p.Config.SMSMessage,
					),
					To:   phone,
					Body: body,
				},
			},
		},
	})

	return
}

// ResetPassword consumes code and reset password to newPassword.
// If the code is invalid, ErrInvalidCode is returned.
// If the code is found but expired, ErrExpiredCode is returned.
// if the code is found but used, ErrUsedCode is returned.
// Otherwise, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (p *Provider) ResetPassword(codeStr string, newPassword string) (userID string, newInfo *authenticator.Info, updateInfo *authenticator.Info, err error) {
	codeHash := HashCode(codeStr)
	code, err := p.Store.Get(codeHash)
	if err != nil {
		return
	}

	now := p.Clock.NowUTC()
	if now.After(code.ExpireAt) {
		err = ErrExpiredCode
		return
	}
	if code.Consumed {
		err = ErrUsedCode
		return
	}

	userID = code.UserID

	// First see if the user has password authenticator.
	ais, err := p.Authenticators.List(userID, authn.AuthenticatorTypePassword)
	if err != nil {
		return
	}

	// The user has no password. Create one for them.
	if len(ais) == 0 {
		p.Logger.Debugf("creating new password")
		newInfo, err = p.Authenticators.New(&authenticator.Spec{
			UserID: userID,
			Type:   authn.AuthenticatorTypePassword,
			Props:  map[string]interface{}{},
		}, newPassword)
		if err != nil {
			return
		}
	} else if len(ais) == 1 {
		p.Logger.Debugf("resetting password")
		// The user has 1 password. Reset it.
		var changed bool
		var ai *authenticator.Info
		changed, ai, err = p.Authenticators.WithSecret(ais[0], newPassword)
		if err != nil {
			return
		}
		if changed {
			updateInfo = ai
		}
	} else {
		// Otherwise the user has two passwords :(
		err = fmt.Errorf("forgotpassword: detected user %s having more than 1 password", userID)
		return
	}

	return
}

func (p *Provider) HashCode(code string) string {
	return HashCode(code)
}

func (p *Provider) AfterResetPassword(codeHash string) (err error) {
	code, err := p.Store.Get(codeHash)
	if err != nil {
		return
	}

	code.Consumed = true
	err = p.Store.Update(code)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.PwHousekeeperTaskName,
		Param: taskspec.PwHousekeeperTaskParam{
			AuthID: code.UserID,
		},
	})

	return
}
