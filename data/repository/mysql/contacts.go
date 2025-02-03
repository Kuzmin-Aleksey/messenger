package mysql

import (
	"database/sql"
	errorsutils "errors"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
)

type Contacts struct {
	db *sql.DB
}

func NewContacts(db *sql.DB) *Contacts {
	return &Contacts{db}
}

func (c *Contacts) Create(contact *models.Contact) *errors.Error {
	if _, err := c.db.Exec("INSERT INTO contacts (user_id, contact_user_id, contact_name) VALUES (?, ?, ?)",
		contact.UserId, contact.ContactId, contact.ContactName); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Contacts) GetContactsByUser(userId int) ([]models.Contact, *errors.Error) {
	rows, err := c.db.Query("SELECT id, user_id, contact_user_id, contact_name FROM contacts WHERE user_id=?", userId)
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

func (c *Contacts) GetContact(userId, contactId int) (*models.Contact, *errors.Error) {
	var contact models.Contact
	if err := c.db.QueryRow("SELECT id, user_id, contact_user_id, contact_name FROM contacts WHERE user_id=? AND contact_user_id=?",
		userId, contactId).Scan(&contact.Id, &contact.UserId, &contact.ContactId, &contact.ContactName); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &contact, nil
}

func (c *Contacts) Delete(userId, contactUserId int) (e *errors.Error) {
	if _, err := c.db.Exec("DELETE FROM contacts WHERE user_id=? AND contact_user_id=?", userId, contactUserId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Contacts) SetContactName(userid, contactId int, name string) *errors.Error {
	if _, err := c.db.Exec("UPDATE contacts SET contact_name=? WHERE user_id=? AND contact_user_id=?", name, userid, contactId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
