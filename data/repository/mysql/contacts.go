package mysql

import (
	"context"
	"database/sql"
	errorsutils "errors"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
)

type Contacts struct {
	DB
}

func NewContacts(db DB) *Contacts {
	return &Contacts{db}
}

func (c *Contacts) Create(ctx context.Context, contact *models.Contact) *errors.Error {
	if _, err := c.DB.ExecContext(ctx, "INSERT INTO contacts (user_id, contact_user_id, contact_name) VALUES (?, ?, ?)",
		contact.UserId, contact.ContactId, contact.ContactName); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Contacts) GetContactsByUser(ctx context.Context, userId int) ([]models.Contact, *errors.Error) {
	rows, err := c.DB.QueryContext(ctx, "SELECT id, user_id, contact_user_id, contact_name FROM contacts WHERE user_id=?", userId)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return []models.Contact{}, nil
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		if err := rows.Scan(&contact.Id, &contact.UserId, &contact.ContactId, &contact.ContactName); err != nil {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

func (c *Contacts) GetContact(ctx context.Context, userId, contactId int) (*models.Contact, *errors.Error) {
	var contact models.Contact
	if err := c.DB.QueryRowContext(ctx, "SELECT id, user_id, contact_user_id, contact_name FROM contacts WHERE user_id=? AND contact_user_id=?",
		userId, contactId).Scan(&contact.Id, &contact.UserId, &contact.ContactId, &contact.ContactName); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &contact, nil
}

func (c *Contacts) Delete(ctx context.Context, userId, contactUserId int) (e *errors.Error) {
	if _, err := c.DB.ExecContext(ctx, "DELETE FROM contacts WHERE user_id=? AND contact_user_id=?", userId, contactUserId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Contacts) SetContactName(ctx context.Context, userid, contactId int, name string) *errors.Error {
	if _, err := c.DB.ExecContext(ctx, "UPDATE contacts SET contact_name=? WHERE user_id=? AND contact_user_id=?", name, userid, contactId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
