package user

import (
	"context"

	userModel "backend/app/model/user"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type UserRepo interface {
	GetUserByUsername(ctx context.Context, username string) (*userModel.User, error)
	GetUserByID(ctx context.Context, userID uint) (*userModel.User, error)
}

type UserRepoParams struct {
	fx.In

	DB *gorm.DB
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(params UserRepoParams) UserRepo {
	return &userRepo{
		db: params.DB,
	}
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*userModel.User, error) {
	var user userModel.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, userID uint) (*userModel.User, error) {
	var user userModel.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
