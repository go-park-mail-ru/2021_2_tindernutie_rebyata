package models

import (
	"context"
	"io"
	"time"
)

type User struct {
	ID          uint64   `json:"id,omitempty"`
	Email       string   `json:"email,omitempty"`
	Password    string   `json:"-"`
	Name        string   `json:"name,omitempty"`
	Gender      string   `json:"gender,omitempty"`
	Prefer      string   `json:"prefer,omitempty"`
	FromAge     uint8    `json:"fromage,omitempty"`
	ToAge       uint8    `json:"toage,omitempty"`
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

type Likes struct {
	AllUsers map[uint64]User `json:"allUsers"`
	Count    string          `json:"likesCount"`
}

type Search struct {
	SearchingTmpl string `json:"searchTmpl"`
}

type Message struct {
	MessageID uint64    `json:"messageID" db:"message_id"`
	FromID    uint64    `json:"fromID" db:"from_id"`
	ToID      uint64    `json:"toID" db:"to_id"`
	Text      string    `json:"text"`
	Date      time.Time `json:"date"`
}

type Chat struct {
	FromUserID  uint64  `json:"fromUserID"`
	Name        string  `json:"name"`
	Img         string  `json:"img"`
	Messages    []Message `json:"messages"`
}

// ArticleUsecase represent the article's usecases
type UserUsecase interface {
	CurrentUser(c context.Context) (User, error)
	EditProfile(c context.Context, newUserData User) (User, error)
	AddPhoto(c context.Context, photo io.Reader, fileName string) (Photo, error)
	DeletePhoto(c context.Context, photo Photo) error
	Login(c context.Context, logUserData LoginUser) (User, error)
	Signup(c context.Context, logUserData LoginUser) (User, error)
	NextUser(c context.Context) ([]User, error)
	GetAllTags(c context.Context) (Tags, error)
	UsersMatches(c context.Context) (Matches, error)
	Reaction(c context.Context, reactionData UserReaction) (Match, error)
	UserLikes(c context.Context) (Likes, error)
	UsersMatchesWithSearching(c context.Context, searchData Search) (Matches, error)

	GetChats(c context.Context) ([]Chat, error)
	GetChat(c context.Context, fromId uint64, lastId uint64) ([]Message, error)
	SendMessage(currentUser User, message Message) (Message, error)
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
	GetNextUserForSwipe(ctx context.Context, currentUser User) ([]User, error)
	GetUsersMatches(ctx context.Context, currentUserId uint64) ([]User, error)
	GetLikes(ctx context.Context, currentUserId uint64) ([]uint64, error)
	DeleteLike(ctx context.Context, firstUser uint64, secondUser uint64) error
	AddMatch(ctx context.Context, firstUser uint64, secondUser uint64) error
	GetUsersLikes(ctx context.Context, currentUserId uint64) ([]User, error)
	GetUsersMatchesWithSearching(ctx context.Context, currentUserId uint64, searchTmpl string) ([]User, error)

	GetChats(ctx context.Context, currentUserId uint64) ([]Chat, error)
	GetChat(ctx context.Context, currentId uint64, fromId uint64, lastId uint64) ([]Message, error)
	SendMessage(ctx context.Context, currentId uint64, toId uint64, text string) (Message, error)
}
