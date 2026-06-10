package hubspot

import (
	"context"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type API interface {
	FindCompanyByDomain(ctx context.Context, domain string) (*Company, error)
	CreateCompany(ctx context.Context, name string, domain string) (Company, error)
	FindContactByEmail(ctx context.Context, email string) (*Contact, error)
	CreateContact(ctx context.Context, lead models.LeadRequest) (Contact, error)
	UpdateContact(ctx context.Context, contactID string, lead models.LeadRequest) (Contact, error)
	AssociateContactToPrimaryCompany(ctx context.Context, contactID string, companyID string) error
}

type Company struct {
	ID string
}

type Contact struct {
	ID string
}
