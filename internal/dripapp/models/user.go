package models

import (
	"context"
	"dripapp/internal/pkg/hasher"
	"errors"
	"io"
	"strconv"
	"time"
)

type User struct {
	ID          uint64   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Email       string   `json:"email,omitempty"`
	Password    string   `json:"-"`
	Date        string   `json:"date,omitempty"`
	Age         string   `json:"age,omitempty"`
	Description string   `json:"description,omitempty"`
	Imgs        []string `json:"imgs,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type LoginUser struct {
	ID       uint64 `json:"id,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserReaction struct {
	Id       uint64 `json:"id"`
	Reaction uint64 `json:"reaction"`
}

type Match struct {
	Match bool `json:"match"`
}

type Tag struct {
	TagName string `json:"tagText"`
}

type Tags struct {
	AllTags map[uint64]Tag `json:"allTags"`
	Count   uint64         `json:"tagsCount"`
}

type Matches struct {
	AllUsers map[uint64]User `json:"allUsers"`
	Count    string          `json:"matchesCount"`
}

func MakeUser(id uint64, email string, password string) (User, error) {
	hashedPass, err := hasher.HashAndSalt(nil, password)
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, Password: hashedPass}, nil
}

func (user User) IsEmpty() bool {
	return len(user.Email) == 0
}

func GetAgeFromDate(date string) (string, error) {
	birthday, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", errors.New("failed on userYear")
	}

	age := uint(time.Now().Year() - birthday.Year())
	if time.Now().YearDay() < birthday.YearDay() {
		age -= 1
	}

	return strconv.Itoa(int(age)), nil
}

func (user *User) FillProfile(newUserData User) (err error) {
	user.Name = newUserData.Name
	user.Date = newUserData.Date
	user.Age, err = GetAgeFromDate(newUserData.Date)
	if err != nil {
		return errors.New("failed to save age")
	}
	user.Date = newUserData.Date
	user.Description = newUserData.Description
	user.Imgs = newUserData.Imgs
	user.Tags = newUserData.Tags

	return nil
}

// ArticleUsecase represent the article's usecases
type UserUsecase interface {
	CurrentUser(c context.Context) (User, HTTPError)
	EditProfile(c context.Context, newUserData User) (User, HTTPError)
	AddPhoto(c context.Context, photo io.Reader, fileName string) (Photo, HTTPError)
	DeletePhoto(c context.Context, photo Photo) HTTPError
	Login(c context.Context, logUserData LoginUser) (User, HTTPError)
	Signup(c context.Context, logUserData LoginUser) (User, HTTPError)
	NextUser(c context.Context) ([]User, HTTPError)
	GetAllTags(c context.Context) (Tags, HTTPError)
	UsersMatches(c context.Context) (Matches, HTTPError)
	Reaction(c context.Context, reactionData UserReaction) (Match, HTTPError)
}

// ArticleRepository represent the article's repository contract
type UserRepository interface {
	GetUser(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, userID uint64) (User, error)
	CreateUser(ctx context.Context, logUserData LoginUser) (User, error)
	UpdateUser(ctx context.Context, newUserData User) (User, error)
	GetTags(ctx context.Context) (map[uint64]string, error)
	UpdateImgs(ctx context.Context, id uint64, imgs []string) error
	AddReaction(ctx context.Context, currentUserId uint64, swipedUserId uint64, reactionType uint64) error
	GetNextUserForSwipe(ctx context.Context, currentUserId uint64) ([]User, error)
	GetUsersMatches(ctx context.Context, currentUserId uint64) ([]User, error)
	GetLikes(ctx context.Context, currentUserId uint64) ([]uint64, error)
	DeleteLike(ctx context.Context, firstUser uint64, secondUser uint64) error
	AddMatch(ctx context.Context, firstUser uint64, secondUser uint64) error

	// CreateUserAndProfile(ctx context.Context, user User) (User, error)
	// IsSwiped(ctx context.Context, userID, swipedUserID uint64) (bool, error)
	// CreateTag(ctx context.Context, tag_name string) error
	// Init() error
	// DropSwipes(ctx context.Context) error
	// DropUsers(ctx context.Context) error
	// DeleteTags(ctx context.Context, userId uint64) error
}
