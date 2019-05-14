package multipleldap

import (
	"github.com/grafana/grafana/pkg/models"
	LDAP "github.com/grafana/grafana/pkg/services/ldap"
)

// GetConfig gets LDAP config
var GetConfig = LDAP.GetConfig

// IsEnabled checks if LDAP is enabled
var IsEnabled = LDAP.IsEnabled

// ErrInvalidCredentials is returned if username and password do not match
var ErrInvalidCredentials = LDAP.ErrInvalidCredentials

// IMultipleLDAPs is interface for MultipleLDAPs
type IMultipleLDAPs interface {
	Login(query *models.LoginUserQuery) (
		*models.ExternalUserInfo, error,
	)

	Users(logins []string) (
		[]*models.ExternalUserInfo, error,
	)

	Add(dn string, values map[string][]string) error
}

// MultipleLDAPs is basic struct of LDAP authorization
type MultipleLDAPs struct {
	servers []*LDAP.ServerConfig
}

// New creates the new LDAP auth
func New(LDAPs []*LDAP.ServerConfig) IMultipleLDAPs {
	return &MultipleLDAPs{
		servers: LDAPs,
	}
}

// Add adds user to the *first* defined LDAP
func (multiples *MultipleLDAPs) Add(
	dn string,
	values map[string][]string,
) error {
	server := multiples.servers[0]
	ldap := LDAP.New(server)

	if err := ldap.Dial(); err != nil {
		return err
	}

	defer ldap.Close()

	err := ldap.Add(dn, values)
	if err != nil {
		return err
	}

	return nil
}

// Login tries to log in the user in multiples LDAP
func (multiples *MultipleLDAPs) Login(query *models.LoginUserQuery) (
	*models.ExternalUserInfo, error,
) {
	for _, server := range multiples.servers {
		ldap := LDAP.New(server)

		if err := ldap.Dial(); err != nil {
			return nil, err
		}

		defer ldap.Close()

		user, err := ldap.Login(query)

		if user != nil {
			return user, nil
		}

		// Continue if we couldn't find the user
		if err == LDAP.ErrInvalidCredentials {
			continue
		}

		if err != nil {
			return nil, err
		}

		return user, nil
	}

	// Return invalid credentials if we couldn't find the user anywhere
	return nil, ErrInvalidCredentials
}

// Users gets users from multiple LDAP servers
func (multiples *MultipleLDAPs) Users(logins []string) (
	[]*models.ExternalUserInfo,
	error,
) {

	var result []*models.ExternalUserInfo

	for _, server := range multiples.servers {
		ldap := LDAP.New(server)

		if err := ldap.Dial(); err != nil {
			return nil, err
		}

		defer ldap.Close()

		users, err := ldap.Users()
		if err != nil {
			return nil, err
		}
		result = append(result, users...)
	}

	return result, nil
}
