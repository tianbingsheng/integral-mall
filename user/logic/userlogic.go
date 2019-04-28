package logic

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/yakaa/log4g"

	"integral-mall/common/baseerror"
	"integral-mall/common/rpcxclient/integralrpcmodel"
	"integral-mall/user/model"
)

type (
	UserLogic struct {
		userModel        *model.UserModel
		redisCache       *redis.Client
		integralRpcModel *integralrpcmodel.IntegralRpcModel
	}
	RegisterRequest struct {
		Mobile   string `json:"mobile" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	RegisterResponse struct {
	}

	LoginRequest struct {
		Mobile   string `json:"mobile" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	LoginResponse struct {
		Authorization string `json:"authorization"`
	}
)

var ErrRecordExist = baseerror.NewBaseError("此手机号已经存在")

func NewUserLogic(userModel *model.UserModel,
	redisCache *redis.Client,
	integralRpcModel *integralrpcmodel.IntegralRpcModel,
) *UserLogic {

	return &UserLogic{userModel: userModel, redisCache: redisCache, integralRpcModel: integralRpcModel}
}

func (l *UserLogic) Register(r *RegisterRequest) (*RegisterResponse, error) {
	response := new(RegisterResponse)
	b, err := l.userModel.ExistByMobile(r.Mobile)
	if err != nil {
		return nil, err
	}
	if b {
		return nil, ErrRecordExist
	}
	userId, err := l.userModel.Insert(&model.User{
		Mobile:   r.Mobile,
		Password: fmt.Sprintf("%x", md5.Sum([]byte(r.Password))),
	})
	if err != nil {
		return nil, err
	}
	if err := l.integralRpcModel.AddIntegral(int(userId), 1000); err != nil {
		log4g.ErrorFormat("AddIntegral err %+v", err)
	}
	return response, nil
}

func (l *UserLogic) Login(r *LoginRequest) (*LoginResponse, error) {
	response := new(LoginResponse)
	user, err := l.userModel.FindByMobile(r.Mobile)
	if err != nil {
		return nil, err
	}
	response.Authorization = fmt.Sprintf("%x", md5.Sum([]byte(user.Mobile+strconv.Itoa(int(user.Id)))))
	l.redisCache.Set(response.Authorization, user.Id, model.AuthorizationExpire)
	return response, nil
}
