package users

import (
	"context"
	"messanger/domain/models"
	"messanger/domain/service/auth"
	"messanger/pkg/errors"
	"net/http"
)

func (s *UsersService) AddToContact(ctx context.Context, dto *CreateContactDTO) *errors.Error {
	phone, e := models.ParsePhone(dto.Phone)
	if e != nil {
		return errors.New(e, "invalid phone number", http.StatusBadRequest)
	}

	userId := auth.ExtractUser(ctx)
	contactUser, err := s.usersRepo.FindByPhone(ctx, phone)
	if err != nil {
		return err.Trace()
	}
	if len(dto.Name) == 0 {
		dto.Name = contactUser.Name
	}

	contact := &models.Contact{
		UserId:      userId,
		ContactId:   contactUser.Id,
		ContactName: dto.Name,
	}

	if err := s.contactsRepo.Create(ctx, contact); err != nil {
		return err.Trace()
	}

	return nil
}

func (s *UsersService) GetUserContacts(ctx context.Context) ([]models.Contact, *errors.Error) {
	userId := auth.ExtractUser(ctx)
	contacts, err := s.contactsRepo.GetContactsByUser(ctx, userId)
	if err != nil {
		return nil, err.Trace()
	}
	return contacts, nil
}

func (s *UsersService) RemoveUserFromContact(ctx context.Context, contactUserId int) *errors.Error {
	if contactUserId == 0 {
		return errors.New1Msg("missing contact user id", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	if err := s.contactsRepo.Delete(ctx, userId, contactUserId); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) RenameContact(ctx context.Context, contactUserId int, name string) *errors.Error {
	if len(name) == 0 {
		return errors.New1Msg("missing name", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	contact, err := s.contactsRepo.GetContact(ctx, userId, contactUserId)
	if err != nil {
		return err.Trace()
	}
	if contact.ContactName == name {
		return nil
	}
	if err := s.contactsRepo.SetContactName(ctx, userId, contactUserId, name); err != nil {
		return err.Trace()
	}
	return nil
}
