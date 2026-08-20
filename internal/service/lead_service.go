package service

import (
	"context"
	"log"
	"time"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/email"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/hubspot"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type LeadService struct {
	hubSpot hubspot.API
	email   email.EmailService
	logger  *log.Logger
}

func NewLeadService(hubSpot hubspot.API, emailService email.EmailService, logger *log.Logger) *LeadService {
	return &LeadService{
		hubSpot: hubSpot,
		email:   emailService,
		logger:  logger,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, lead models.LeadRequest) (models.LeadResponse, error) {
	domain, err := lead.CompanyDomain()
	if err != nil {
		return models.LeadResponse{}, apperrors.BadRequest(err.Error(), err)
	}

	company, err := s.hubSpot.FindCompanyByDomain(ctx, domain)
	if err != nil {
		s.logger.Printf("failed to search HubSpot company: domain=%s error=%v", domain, err)
		return models.LeadResponse{}, apperrors.ExternalService("failed to search company in HubSpot", err)
	}

	if company == nil {
		createdCompany, err := s.hubSpot.CreateCompany(ctx, lead.Company, domain)
		if err != nil {
			s.logger.Printf("failed to create HubSpot company: domain=%s error=%v", domain, err)
			return models.LeadResponse{}, apperrors.ExternalService("failed to create company in HubSpot", err)
		}
		company = &createdCompany
	}

	contact, err := s.hubSpot.FindContactByEmail(ctx, lead.WorkEmail)
	if err != nil {
		s.logger.Printf("failed to search HubSpot contact: email=%s error=%v", lead.WorkEmail, err)
		return models.LeadResponse{}, apperrors.ExternalService("failed to search contact in HubSpot", err)
	}

	if contact == nil {
		createdContact, err := s.hubSpot.CreateContact(ctx, lead)
		if err != nil {
			s.logger.Printf("failed to create HubSpot contact: email=%s error=%v", lead.WorkEmail, err)
			return models.LeadResponse{}, apperrors.ExternalService("failed to create contact in HubSpot", err)
		}
		contact = &createdContact
	} else {
		updatedContact, err := s.hubSpot.UpdateContact(ctx, contact.ID, lead)
		if err != nil {
			s.logger.Printf("failed to update HubSpot contact: contact_id=%s email=%s error=%v", contact.ID, lead.WorkEmail, err)
			return models.LeadResponse{}, apperrors.ExternalService("failed to update contact in HubSpot", err)
		}
		contact = &updatedContact
	}

	if err := s.hubSpot.AssociateContactToPrimaryCompany(ctx, contact.ID, company.ID); err != nil {
		s.logger.Printf("failed to associate HubSpot contact to company: contact_id=%s company_id=%s error=%v", contact.ID, company.ID, err)
		return models.LeadResponse{}, apperrors.ExternalService("failed to associate contact with company in HubSpot", err)
	}

	s.logger.Printf("Lead created in HubSpot: contact_id=%s email=%s", contact.ID, lead.WorkEmail)

	if err := s.email.SendLeadNotification(ctx, lead, time.Now()); err != nil {
		s.logger.Printf("failed to send internal notification email: contact_id=%s email=%s error=%v", contact.ID, lead.WorkEmail, err)
	}

	return models.LeadResponse{
		Message:          "lead submitted successfully",
		HubSpotContactID: contact.ID,
	}, nil
}
