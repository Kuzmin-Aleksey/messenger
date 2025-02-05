package users

import (
	"context"
	errorsutils "errors"
	"github.com/nyaruka/phonenumbers"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/domain/service/auth"
	"messanger/pkg/db"
	"messanger/pkg/errors"
	"net/http"
	"regexp"
	"sync"
	"time"
)

type UsersService struct {
	phoneConf    PhoneConfirmator
	contactsRepo ports.ContactsRepo
	usersRepo    ports.UsersRepo
	chatsRepo    ports.ChatsRepo

	createUserMu     sync.Mutex
	updatePhoneMu    sync.Mutex
	updateUsernameMu sync.Mutex
}

type PhoneConfirmator interface {
	ToConfirming(ctx context.Context, userId int, phone string) *errors.Error
	ConfirmUser(ctx context.Context, code string) (int, *errors.Error)
}

func NewUsersService(
	usersRepo ports.UsersRepo,
	contactsRepo ports.ContactsRepo,
	chatsRepo ports.ChatsRepo,
	phoneConf PhoneConfirmator,
) *UsersService {
	return &UsersService{
		usersRepo:    usersRepo,
		contactsRepo: contactsRepo,
		chatsRepo:    chatsRepo,
		phoneConf:    phoneConf,
	}
}

var usernameRegexp = regexp.MustCompile("^[a-zA-Z0-9_]{4,32}$")

func (s *UsersService) CreateUser(ctx context.Context, dto *CreateUserDTO) (err *errors.Error) {
	if len(dto.Name) == 0 || len(dto.RealName) == 0 || len(dto.Password) == 0 || len(dto.Phone) == 0 {
		return errors.New1Msg("missing user info", http.StatusBadRequest)
	}
	if usernameRegexp.MatchString(dto.Name) {
		return errors.New1Msg("invalid username", http.StatusBadRequest)
	}

	var e error
	dto.Phone, e = parsePhone(dto.Phone)
	if e != nil {
		return errors.New(e, "invalid phone number", http.StatusBadRequest)
	}

	s.createUserMu.Lock()
	defer s.createUserMu.Unlock()

	if err := s.checkPhoneExist(ctx, dto.Phone); err != nil {
		return err.Trace()
	}
	if err := s.checkUsernameExist(ctx, dto.Name); err != nil {
		return err.Trace()
	}

	user := &models.User{
		Phone:      dto.Phone,
		Password:   dto.Password,
		Name:       dto.Name,
		RealName:   dto.RealName,
		ShowPhone:  true,
		LastOnline: time.Now(),
		Confirmed:  false,
	}
	ctx, err = db.WithTx(ctx, s.usersRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.usersRepo.New(ctx, user); err != nil {
		return err.Trace()
	}
	if err := s.phoneConf.ToConfirming(ctx, user.Id, user.Phone); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) ConfirmPhone(ctx context.Context, code string) *errors.Error {
	userId, err := s.phoneConf.ConfirmUser(ctx, code)
	if err != nil {
		return err.Trace()
	}
	if err := s.usersRepo.SetConfirm(ctx, userId, true); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) CheckUsername(ctx context.Context, username string) (bool, *errors.Error) {
	if len(username) == 0 {
		return false, errors.New1Msg("missing username", http.StatusBadRequest)
	}
	if usernameRegexp.MatchString(username) {
		return false, errors.New1Msg("invalid username", http.StatusBadRequest)
	}
	if _, err := s.usersRepo.FindByName(ctx, username); err != nil {
		if err.Code == http.StatusNotFound {
			return false, nil
		}
		return false, err.Trace()
	}
	return true, nil
}

func (s *UsersService) UpdateUser(ctx context.Context, dto *UpdateUserDTO) (err *errors.Error) {
	ctx, err = db.WithTx(ctx, s.usersRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	var isUpdating bool
	if len(dto.RealName) != 0 {
		if err := s.UpdateRealName(ctx, dto.RealName); err != nil {
			return err.Trace()
		}
		isUpdating = true
	}
	if len(dto.Name) != 0 {
		if err := s.UpdateUsername(ctx, dto.Name); err != nil {
			return err.Trace()
		}
		isUpdating = true
	}
	if len(dto.Password) != 0 {
		if err := s.UpdatePassword(ctx, dto.OldPassword, dto.Password); err != nil {
			return err.Trace()
		}
		isUpdating = true
	}
	if len(dto.Phone) != 0 {
		if err := s.UpdatePhone(ctx, dto.Phone); err != nil {
			return err.Trace()
		}
		isUpdating = true
	}
	if !isUpdating {
		return errors.New1Msg("all fields is null", http.StatusBadRequest)
	}
	return nil
}

func (s *UsersService) UpdateUsername(ctx context.Context, name string) *errors.Error {
	if len(name) == 0 {
		return errors.New1Msg("missing username", http.StatusBadRequest)
	}
	s.updateUsernameMu.Lock()
	defer s.updateUsernameMu.Unlock()

	if err := s.checkUsernameExist(ctx, name); err != nil {
		return err.Trace()
	}
	userId := auth.ExtractUser(ctx)
	if err := s.usersRepo.UpdateUsername(ctx, userId, name); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) UpdatePassword(ctx context.Context, oldPass, newPass string) *errors.Error {
	if len(newPass) == 0 {
		return errors.New1Msg("missing new password", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	_, err := s.usersRepo.GetByIdWithPass(ctx, userId, oldPass)
	if err != nil {
		return err.Trace()
	}

	if err := s.usersRepo.UpdatePassword(ctx, userId, newPass); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) UpdatePhone(ctx context.Context, phone string) *errors.Error {
	if len(phone) == 0 {
		return errors.New1Msg("missing phone", http.StatusBadRequest)
	}
	s.updatePhoneMu.Lock()

	if err := s.checkPhoneExist(ctx, phone); err != nil {
		s.updatePhoneMu.Unlock()
		return err.Trace()
	}

	userId := auth.ExtractUser(ctx)

	if err := s.usersRepo.UpdatePhone(ctx, userId, phone); err != nil {
		s.updatePhoneMu.Unlock()
		return err.Trace()
	}
	s.updatePhoneMu.Unlock()

	if err := s.phoneConf.ToConfirming(ctx, userId, phone); err != nil {
		return err.Trace()
	}
	if err := s.usersRepo.SetConfirm(ctx, userId, false); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) UpdateRealName(ctx context.Context, name string) *errors.Error {
	if len(name) == 0 {
		return errors.New1Msg("missing real name", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	if err := s.usersRepo.UpdateRealName(ctx, userId, name); err != nil {
		return err.Trace()
	}
	return nil
}

// return nil if not exist
func (s *UsersService) checkPhoneExist(ctx context.Context, phone string) *errors.Error {
	if _, err := s.usersRepo.FindByPhone(ctx, phone); err == nil {
		return errors.New1Msg("user with this phone already exists", http.StatusBadRequest)
	} else if err.Code != http.StatusNotFound {
		return err.Trace()
	}
	return nil
}

// return nil if not exist
func (s *UsersService) checkUsernameExist(ctx context.Context, name string) *errors.Error {
	if _, err := s.usersRepo.FindByName(ctx, name); err == nil {
		return errors.New1Msg("user with this name already exists", http.StatusBadRequest)
	} else if err.Code != http.StatusNotFound {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) CreateChatWithUser(ctx context.Context, userId int) (chat *models.Chat, err *errors.Error) {
	actionerId := auth.ExtractUser(ctx)
	chat = &models.Chat{
		Type: models.ChatTypeUser,
	}
	ctx, err = db.WithTx(ctx, s.chatsRepo)
	if err != nil {
		return nil, err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.chatsRepo.New(ctx, chat); err != nil {
		return nil, err.Trace()
	}
	if err := s.chatsRepo.AddUserToChat(ctx, chat.Id, actionerId); err != nil {
		return nil, err.Trace()
	}
	if err := s.chatsRepo.AddUserToChat(ctx, chat.Id, userId); err != nil {
		return nil, err.Trace()
	}

	return chat, nil
}

func (s *UsersService) SetShowPhone(ctx context.Context, v bool) *errors.Error {
	userId := auth.ExtractUser(ctx)
	if err := s.usersRepo.SetShowPhone(ctx, userId, v); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) FindUser(ctx context.Context, dto *FindUserDTO) (*UserResponseDTO, *errors.Error) {
	var user *models.User
	var err *errors.Error
	var foundByPhone bool

	if dto.UserId != 0 {
		user, err = s.GetById(ctx, dto.UserId)
	} else if len(dto.Username) != 0 {
		user, err = s.FindByName(ctx, dto.Username)
	} else if len(dto.Phone) != 0 {
		user, err = s.FindByPhone(ctx, dto.Phone)
		foundByPhone = true
	} else {
		return nil, errors.New1Msg("all fields is null", http.StatusBadRequest)
	}
	if err != nil {
		return nil, err.Trace()
	}

	resp := &UserResponseDTO{
		Id:       user.Id,
		Name:     user.Name,
		RealName: user.RealName,
	}

	actionerId := auth.ExtractUser(ctx)
	if actionerId == user.Id {
		resp.Phone = user.Phone
		resp.ShowPhone = user.ShowPhone
		return resp, nil
	}

	if foundByPhone || user.ShowPhone {
		resp.Phone = user.Phone
	}

	if contact, err := s.contactsRepo.GetContact(ctx, actionerId, user.Id); err != nil {
		if err.Code != http.StatusNotFound {
			return nil, err.Trace()
		}
	} else {
		resp.Phone = user.Phone
		resp.ContactName = contact.ContactName
	}

	return resp, nil
}

func (s *UsersService) GetById(ctx context.Context, id int) (*models.User, *errors.Error) {
	if id == 0 {
		return nil, errors.New1Msg("missing user id", http.StatusBadRequest)
	}
	user, err := s.usersRepo.GetById(ctx, id)
	if err != nil {
		return nil, err.Trace()
	}
	return user, nil
}

func (s *UsersService) GetLastOnline(ctx context.Context, userId int) (*time.Time, *errors.Error) {
	t, err := s.usersRepo.GetLastOnline(ctx, userId)
	if err != nil {
		if err.Code != http.StatusNotFound {
			return nil, err.Trace()
		}
		return nil, nil
	}
	return &t, nil
}

func (s *UsersService) FindByPhone(ctx context.Context, phoneNum string) (*models.User, *errors.Error) {
	phoneNum, e := parsePhone(phoneNum)
	if e != nil {
		return nil, errors.New(e, "invalid phone number", http.StatusBadRequest)
	}
	user, err := s.usersRepo.FindByPhone(ctx, phoneNum)
	if err != nil {
		return nil, err.Trace()
	}
	return user, nil
}

func (s *UsersService) FindByName(ctx context.Context, name string) (*models.User, *errors.Error) {
	user, err := s.usersRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err.Trace()
	}
	return user, nil
}

func (s *UsersService) DeleteUser(ctx context.Context) (err *errors.Error) {
	userId := auth.ExtractUser(ctx)

	ctx, err = db.WithTx(ctx, s.usersRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	chats, err := s.chatsRepo.GetByUserId(ctx, userId)
	if err != nil {
		return err.Trace()
	}

	for _, chat := range chats {
		count, err := s.chatsRepo.CountUsersInChat(ctx, chat.Id)
		if err != nil {
			return err.Trace()
		}
		if count == 1 {
			if err := s.chatsRepo.Delete(ctx, chat.Id); err != nil {
				return err.Trace()
			}
		}
	}

	if err := s.usersRepo.Delete(ctx, userId); err != nil {
		return err.Trace()
	}
	return nil
}

func parsePhone(phone string) (string, error) {
	num, err := phonenumbers.Parse(phone, "RU")
	if err != nil {
		return "", err
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", errorsutils.New("invalid phone number")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
