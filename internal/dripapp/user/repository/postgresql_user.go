package repository

import (
	"context"
	"dripapp/configs"
	"dripapp/internal/dripapp/models"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const success = "Connection success (postgre) on: "

type PostgreUserRepo struct {
	Conn sqlx.DB
}

func NewPostgresUserRepository(config configs.PostgresConfig) (models.UserRepository, error) {
	ConnStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		configs.Postgres.User,
		configs.Postgres.Password,
		configs.Postgres.DBName)

	Conn, err := sqlx.Open("postgres", ConnStr)
	if err != nil {
		return nil, err
	}

	// query, err := ioutil.ReadFile("docker/postgres_scripts/dump.sql")
	// if err != nil {
	// 	return nil, err
	// }
	// strQuery := string(query)
	// if _, err := Conn.Exec(strQuery); err != nil {
	// 	return nil, err
	// }

	log.Printf("%s%s", success, ConnStr)
	return &PostgreUserRepo{*Conn}, nil
}

func (p PostgreUserRepo) GetUser(ctx context.Context, email string) (models.User, error) {
	var RespUser models.User
	err := p.Conn.QueryRow(GetUserQuery, email).
		Scan(&RespUser.ID, &RespUser.Name, &RespUser.Email, &RespUser.Password, &RespUser.Date, &RespUser.Description, pq.Array(&RespUser.Imgs))
	if err != nil {
		return models.User{}, err
	}

	// RespUser.Tags, err = p.getTagsByID(ctx, RespUser.ID)
	// if err != nil {
	// 	return models.User{}, err
	// }

	return RespUser, nil
}

func (p PostgreUserRepo) GetUserByID(ctx context.Context, userID uint64) (models.User, error) {
	var RespUser models.User
	err := p.Conn.GetContext(ctx, &RespUser, GetUserByIdAQuery, userID)
	if err != nil {
		return models.User{}, err
	}

	RespUser.Tags, err = p.getTagsByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	RespUser.Imgs, err = p.getImgsByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	return RespUser, nil
}

func (p PostgreUserRepo) CreateUser(ctx context.Context, logUserData models.LoginUser) (models.User, error) {
	var RespUser models.User
	err := p.Conn.GetContext(ctx, &RespUser, CreateUserQuery, logUserData.Email, logUserData.Password)
	return RespUser, err
}

func (p PostgreUserRepo) UpdateUser(ctx context.Context, newUserData models.User) (models.User, error) {
	var RespUser models.User
	err := p.Conn.GetContext(ctx, &RespUser, UpdateUserQuery, newUserData.Name, newUserData.Email, newUserData.Date,
		newUserData.Description, pq.Array(&newUserData.Imgs))

	if len(newUserData.Tags) != 0 {
		err = p.deleteTags(ctx, newUserData.ID)
		if err != nil {
			return models.User{}, err
		}
		err = p.insertTags(ctx, newUserData.ID, newUserData.Tags)
		if err != nil {
			return models.User{}, err
		}
	}

	if len(newUserData.Imgs) != 0 {
		RespUser.Imgs, err = p.getImgsByID(ctx, RespUser.ID)
		if err != nil {
			return models.User{}, err
		}
	}

	if len(newUserData.Tags) != 0 {
		RespUser.Tags, err = p.getTagsByID(ctx, RespUser.ID)
		if err != nil {
			return models.User{}, err
		}
	}

	return RespUser, err
}

func (p PostgreUserRepo) deleteTags(ctx context.Context, userId uint64) error {
	stmt, _ := p.Conn.Prepare(DeleteTagsQuery)
	_, err := stmt.Exec(userId)
	if err != nil {
		return err
	}

	return nil
}

func (p PostgreUserRepo) GetTags(ctx context.Context) (map[uint64]string, error) {
	var tags []models.Tag
	err := p.Conn.Select(&tags, GetTagsQuery)
	if err != nil {
		return nil, err
	}

	tagsMap := make(map[uint64]string)

	var i uint64
	for i = 0; i < uint64(len(tags)); i++ {
		tagsMap[i] = tags[i].Tag_Name
	}

	return tagsMap, nil
}

func (p PostgreUserRepo) getTagsByID(ctx context.Context, id uint64) ([]string, error) {
	var tags []string
	err := p.Conn.Select(&tags, GetTagsByIdQuery, id)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (p PostgreUserRepo) getImgsByID(ctx context.Context, id uint64) ([]string, error) {
	var imgs []string
	if err := p.Conn.QueryRow(GetImgsByIDQuery, id).Scan(pq.Array(&imgs)); err != nil {
		return nil, err
	}

	return imgs, nil
}

func (p PostgreUserRepo) insertTags(ctx context.Context, id uint64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	vals := []interface{}{}
	vals = append(vals, id)
	for _, val := range tags {
		vals = append(vals, val)
	}

	var sb strings.Builder
	sb.WriteString(InsertTagsQueryFirstPart)
	var inserts []string
	for idx := range tags {
		str := fmt.Sprintf(InsertTagsQueryParts, idx+2)
		inserts = append(inserts, str)
	}
	sb.WriteString(strings.Join(inserts, ",\n"))
	sb.WriteString(";")
	insertTagsQuery := sb.String()

	stmt, _ := p.Conn.Prepare(insertTagsQuery)
	_, err := stmt.Exec(vals...)
	if err != nil {
		return err
	}

	return nil
}

func (p PostgreUserRepo) UpdateImgs(ctx context.Context, id uint64, imgs []string) error {
	var user_id uint64
	err := p.Conn.QueryRow(UpdateImgsQuery, id, pq.Array(&imgs)).Scan(&user_id)
	if err != nil {
		return err
	}

	return nil
}

func (p PostgreUserRepo) AddReaction(ctx context.Context, currentUserId uint64, swipedUserId uint64, reactionType uint64) error {
	stmt, _ := p.Conn.Prepare(AddReactionQuery)
	_, err := stmt.Exec(currentUserId, swipedUserId, reactionType)
	if err != nil {
		return err
	}

	return nil
}

func (p PostgreUserRepo) GetNextUserForSwipe(ctx context.Context, currentUserId uint64) ([]models.User, error) {
	var notSwipedUser []models.User
	err := p.Conn.Select(&notSwipedUser, GetNextUserForSwipeQuery, currentUserId)
	if err != nil {
		return nil, err
	}

	for idx := range notSwipedUser {
		notSwipedUser[idx].Age, err = models.GetAgeFromDate(notSwipedUser[idx].Date)
		if err != nil {
			return nil, err
		}

		notSwipedUser[idx].Imgs, err = p.getImgsByID(ctx, currentUserId)
		if err != nil {
			return nil, err
		}

		notSwipedUser[idx].Tags, err = p.getTagsByID(ctx, currentUserId)
		if err != nil {
			return nil, err
		}
	}

	return notSwipedUser, nil
}

func (p PostgreUserRepo) GetUsersMatches(ctx context.Context, currentUserId uint64) ([]models.User, error) {
	var matchesUsers []models.User
	err := p.Conn.Select(&matchesUsers, GetUsersForMatchesQuery, currentUserId)
	if err != nil {
		return nil, err
	}

	for idx := range matchesUsers {
		matchesUsers[idx].Age, err = models.GetAgeFromDate(matchesUsers[idx].Date)
		if err != nil {
			return nil, err
		}

		matchesUsers[idx].Imgs, err = p.getImgsByID(ctx, currentUserId)
		if err != nil {
			return nil, err
		}

		matchesUsers[idx].Tags, err = p.getTagsByID(ctx, currentUserId)
		if err != nil {
			return nil, err
		}
	}

	return matchesUsers, nil
}

func (p PostgreUserRepo) GetLikes(ctx context.Context, currentUserId uint64) ([]uint64, error) {
	// type = 1 is like (dislike - 2)

	var likes []uint64
	err := p.Conn.Select(&likes, GetLikesQuery, currentUserId)
	if err != nil {
		return nil, err
	}

	return likes, nil
}

func (p PostgreUserRepo) DeleteLike(ctx context.Context, firstUser uint64, secondUser uint64) error {
	stmt, _ := p.Conn.Prepare(DeleteLikeQuery)
	_, err := stmt.Exec(firstUser, secondUser)
	if err != nil {
		return err
	}

	return nil
}

func (p PostgreUserRepo) AddMatch(ctx context.Context, firstUser uint64, secondUser uint64) error {
	stmt, _ := p.Conn.Prepare(AddMatchQuery)
	_, err := stmt.Exec(firstUser, secondUser)
	if err != nil {
		return err
	}

	return nil
}

// func (p PostgreUserRepo) IsSwiped(ctx context.Context, userID, swipedUserID uint64) (bool, error) {
// 	query := `select exists(select id1, id2 from reactions where id1=$1 and id2=$2)`

// 	var resp bool
// 	err := p.Conn.GetContext(ctx, &resp, query, userID, swipedUserID)
// 	if err != nil {
// 		return false, err
// 	}
// 	return resp, nil
// }

// func (p PostgreUserRepo) CreateTag(ctx context.Context, tag_name string) error {
// 	sel := "insert into tag(tag_name) values($1);"

// 	if err := p.Conn.QueryRow(sel, tag_name).Scan(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p PostgreUserRepo) DropSwipes(ctx context.Context) error {
// 	query := `delete from reactions`

// 	if err := p.Conn.QueryRow(query).Scan(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p PostgreUserRepo) DropUsers(ctx context.Context) error {
// 	query := `
// 	delete from profile_tag;
// 	delete from matches;
// 	delete from reactions;
// 	delete from profile;`

// 	if err := p.Conn.QueryRow(query).Scan(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p PostgreUserRepo) CreateUserAndProfile(ctx context.Context, user models.User) (models.User, error) {
// 	query := `insert into profile(name, email, password, date, description, imgs)
// 		values($1,$2,$3,$4,$5,$6)
// 		RETURNING id, name, email, password, email, password, date, description;`

// 	var RespUser models.User
// 	err := p.Conn.GetContext(ctx, &RespUser, query, user.Name, user.Email, user.Password, user.Date,
// 		user.Description, pq.Array(&user.Imgs))
// 	if err != nil {
// 		return models.User{}, err
// 	}

// 	err = p.insertTags(ctx, RespUser.ID, user.Tags)
// 	if err != nil {
// 		return models.User{}, err
// 	}

// 	RespUser.Imgs, err = p.getImgsByID(ctx, RespUser.ID)
// 	if err != nil {
// 		return models.User{}, err
// 	}

// 	RespUser.Age, err = models.GetAgeFromDate(RespUser.Date)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	RespUser.Tags, err = p.getTagsByID(ctx, RespUser.ID)
// 	if err != nil {
// 		return models.User{}, err
// 	}

// 	return RespUser, err
// }

// func (p PostgreUserRepo) Init() error {
// 	query, err := ioutil.ReadFile("docker/postgres_scripts/dump.sql")
// 	if err != nil {
// 		return err
// 	}
// 	strQuery := string(query)

// 	if _, err := p.Conn.Exec(strQuery); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p PostgreUserRepo) DeleteUser(ctx context.Context, user models.User) error {
// 	query := `delete from profile where id=$1`

// 	if err := p.Conn.QueryRow(query, user.ID).Scan(); err != nil {
// 		return err
// 	}

// 	return nil
// }
