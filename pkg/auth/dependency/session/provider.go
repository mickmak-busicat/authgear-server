package session

type Provider interface {
	// Make makes a session from authn session
	MakeSession(attrs *Attrs) (session *Session, token string)
	// Create creates a session
	Create(session *Session) error
	// GetByToken gets the session identified by the token
	GetByToken(token string) (*Session, error)
	// Get gets the session identified by the ID
	Get(id string) (*Session, error)
	// Update updates the session attributes.
	Update(session *Session) error
	// Access updates the session info when it is being accessed by user
	Access(*Session) error
	// Invalidate invalidates session with the ID
	Invalidate(*Session) error
	// InvalidateBatch invalidates sessions
	InvalidateBatch([]*Session) error
	// InvalidateAll invalidates all sessions of the user, except specified session
	InvalidateAll(userID string, sessionID string) error
	// List lists the sessions belonging to the user, in ascending creation time order
	List(userID string) ([]*Session, error)
}
