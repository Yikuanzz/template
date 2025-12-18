package user

import (
	"context"

	userModel "backend/app/model/user"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type UserRepoParams struct {
	fx.In

	DB *gorm.DB
}

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(params UserRepoParams) *UserRepo {
	return &UserRepo{
		db: params.DB,
	}
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*userModel.User, error) {
	var user userModel.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, userID uint) (*userModel.User, error) {
	var user userModel.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user *userModel.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) UpdateUserInfo(ctx context.Context, userID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&userModel.User{}).Where("id = ?", userID).Updates(updates).Error
}
