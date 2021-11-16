package usecase

import (
	"context"
	"dripapp/configs"
	"dripapp/internal/dripapp/models"
	"dripapp/internal/pkg/hasher"
	"errors"
	"io"
	"strconv"
	"time"
)

type userUsecase struct {
	UserRepo       models.UserRepository
	Session        models.SessionRepository
	File           models.FileRepository
	contextTimeout time.Duration
}

func NewUserUsecase(
	ur models.UserRepository,
	fileManager models.FileRepository,
	timeout time.Duration) models.UserUsecase {
	return &userUsecase{
		UserRepo:       ur,
		File:           fileManager,
		contextTimeout: timeout,
	}
}

func (h *userUsecase) CurrentUser(c context.Context) (models.User, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return models.User{}, errors.New(models.ErrContextNilError)
	}

	return currentUser, nil
}

func (h *userUsecase) EditProfile(c context.Context, newUserData models.User) (models.User, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return models.User{}, errors.New(models.ErrContextNilError)
	}

	err := currentUser.FillProfile(newUserData)
	if err != nil {
		return models.User{}, err
	}

	_, err = h.UserRepo.UpdateUser(c, currentUser)
	if err != nil {
		return models.User{}, err
	}

	return currentUser, nil
}

func (h *userUsecase) AddPhoto(c context.Context, photo io.Reader, fileName string) (models.Photo, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return models.Photo{}, errors.New(models.ErrContextNilError)
	}

	photoPath, err := h.File.SaveUserPhoto(currentUser, photo, fileName)
	if err != nil {
		return models.Photo{}, err
	}

	currentUser.AddNewPhoto(photoPath)

	err = h.UserRepo.UpdateImgs(c, currentUser.ID, currentUser.Imgs)
	if err != nil {
		return models.Photo{}, err
	}

	return models.Photo{Path: photoPath}, nil
}

func (h *userUsecase) DeletePhoto(c context.Context, photo models.Photo) error {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return errors.New(models.ErrContextNilError)
	}

	err := currentUser.DeletePhoto(photo)
	if err != nil {
		return err
	}

	err = h.UserRepo.UpdateImgs(c, currentUser.ID, currentUser.Imgs)
	if err != nil {
		return err
	}

	err = h.File.Delete(photo.Path)
	if err != nil {
		return err
	}

	return nil
}

// @Summary LogIn
// @Description log in
// @Tags login
// @Accept json
// @Produce json
// @Param input body LoginUser true "data for login"
// @Success 200 {object} JSON
// @Failure 400,404,500
// @Router /login [post]
func (h *userUsecase) Login(c context.Context, logUserData models.LoginUser) (models.User, error) {
	identifiableUser, err := h.UserRepo.GetUser(c, logUserData.Email)
	if err != nil {
		return models.User{}, err
	}

	if !hasher.CheckWithHash(identifiableUser.Password, logUserData.Password) {
		return models.User{}, err
	}

	return identifiableUser, nil
}

// @Summary SignUp
// @Description registration user
// @Tags registration
// @Accept json
// @Produce json
// @Param input body LoginUser true "data for registration"
// @Success 200 {object} JSON
// @Failure 400,404,500
// @Router /signup [post]
func (h *userUsecase) Signup(c context.Context, logUserData models.LoginUser) (models.User, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	identifiableUser, err := h.UserRepo.GetUser(ctx, logUserData.Email)
	if err != nil {
		return models.User{}, err
	}
	if !identifiableUser.IsEmpty() {
		return models.User{}, errors.New("")
	}

	logUserData.Password = hasher.HashAndSalt(nil, logUserData.Password)

	user, err := h.UserRepo.CreateUser(c, logUserData)
	if err != nil {
		return models.User{}, err
	}

	err = h.File.CreateFoldersForNewUser(user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (h *userUsecase) NextUser(c context.Context) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return nil, errors.New(models.ErrContextNilError)
	}

	nextUsers, err := h.UserRepo.GetNextUserForSwipe(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	return nextUsers, nil
}

func (h *userUsecase) GetAllTags(c context.Context) (models.Tags, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	allTags, err := h.UserRepo.GetTags(ctx)
	if err != nil {
		return models.Tags{}, err
	}
	var respTag models.Tag
	var currentAllTags = make(map[uint64]models.Tag)
	var respAllTags models.Tags
	counter := 0

	for _, value := range allTags {
		respTag.TagName = value
		currentAllTags[uint64(counter)] = respTag
		counter++
	}

	respAllTags.AllTags = currentAllTags
	respAllTags.Count = uint64(counter)

	return respAllTags, nil
}

func (h *userUsecase) UsersMatches(c context.Context) (models.Matches, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return models.Matches{}, errors.New(models.ErrContextNilError)
	}

	// find matches
	mathes, err := h.UserRepo.GetUsersMatches(ctx, currentUser.ID)
	if err != nil {
		return models.Matches{}, err
	}

	// count
	counter := 0
	var allMathesMap = make(map[uint64]models.User)
	for _, value := range mathes {
		allMathesMap[uint64(counter)] = value
		counter++
	}

	var allMatches models.Matches
	allMatches.AllUsers = allMathesMap
	allMatches.Count = strconv.Itoa(counter)

	return allMatches, nil
}

func (h *userUsecase) Reaction(c context.Context, reactionData models.UserReaction) (models.Match, error) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()

	currentUser, ok := ctx.Value(configs.ContextUser).(models.User)
	if !ok {
		return models.Match{}, errors.New(models.ErrContextNilError)
	}

	// added reaction in db
	err := h.UserRepo.AddReaction(ctx, currentUser.ID, reactionData.Id, reactionData.Reaction)
	if err != nil {
		return models.Match{}, err
	}

	// get users who liked current user
	var likes []uint64
	likes, err = h.UserRepo.GetLikes(ctx, currentUser.ID)
	if err != nil {
		return models.Match{}, err
	}

	var currMath models.Match
	currMath.Match = false
	for _, value := range likes {
		if value == reactionData.Id {
			currMath.Match = true
			err = h.UserRepo.DeleteLike(ctx, currentUser.ID, reactionData.Id)
			if err != nil {
				return models.Match{}, err
			}
			err = h.UserRepo.AddMatch(ctx, currentUser.ID, reactionData.Id)
			if err != nil {
				return models.Match{}, err
			}
		}
	}

	return currMath, nil
}
