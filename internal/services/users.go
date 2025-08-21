package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"video-crawler/internal/entities"
	"video-crawler/internal/logger"
	"video-crawler/internal/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserServiceInterface interface {
	Login(ctx *gin.Context, username string, password string) (user *entities.UserEntity, token string, err error)
	Register(ctx *gin.Context, username string, password string, nickname string) (err error)
	UserDetail(ctx *gin.Context, userId string) (userInfo entities.UserDetailResponse, err error)
	UserDetailInner(userId string) (userInfo entities.UserEntity, exist bool)
	Save(ctx *gin.Context, userId string, userInfo *entities.UserEntity)
	UserList() (userList []entities.UserList)
	Delete(ctx *gin.Context, userId string)
	GenerateToken(user *entities.UserEntity) (string, error)
}

func NewUserService(jwtManager *utils.JWTManager) UserServiceInterface {
	userService := &userService{
		userMap:        &sync.Map{},
		userName2IdMap: &sync.Map{},
		jwtManager:     jwtManager,
		isWriting:      false,
	}
	const userConfigFilePath = "./configs/users.json"
	// 判断文件是否存在，不存在自动创建初始化内容为 "[]"
	_, err := os.Stat(userConfigFilePath)
	if os.IsNotExist(err) {
		file, err := os.Create(userConfigFilePath)
		if err != nil {
			panic(err)
		}
		_, err = file.Write([]byte("[]"))
		if err != nil {
			panic(err)
		}
		err = file.Sync()
		if err != nil {
			panic(err)
		}
		file.Close()
	}

	userService.reloadUsers()

	// 监听文件变化，当文件变化后重新刷新 map
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		defer watcher.Close()

		err = watcher.Add(userConfigFilePath)
		if err != nil {
			panic(err)
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// 如果正在写入文件，跳过重新加载，避免无限循环
					if userService.isWriting {
						logrus.Debug("skipping reload during file write")
						continue
					}
					// 文件被修改或创建时，重新加载用户信息
					userService.reloadUsers()
					logrus.Info("user config file changed, reload users")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.WithError(err).Error("failed to watch user config file")
				// 可以选择记录日志或处理错误
			}
		}
	}()
	return userService
}

type userService struct {
	userMap        *sync.Map
	userName2IdMap *sync.Map
	userList       []entities.UserEntity
	jwtManager     *utils.JWTManager
	isWriting      bool // 标记是否正在写入文件，避免监听器触发无限循环
}

// Delete implements UserServiceInterface.
func (s *userService) Delete(ctx *gin.Context, userId string) {
	s.userMap.Delete(userId)
	s.saveMapChange(ctx)
}

// UserList implements UserServiceInterface.
func (s *userService) UserList() (userList []entities.UserList) {
	s.userMap.Range(func(key, value interface{}) bool {
		user, _ := value.(entities.UserEntity)
		userList = append(userList, entities.UserList{
			Id:          user.Id,
			Nickname:    user.Nickname,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			IsSiteAdmin: user.IsSiteAdmin,
			AllowLogin:  user.AllowLogin,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			LastLoginAt: user.LastLoginAt,
			LoginCount:  user.LoginCount,
		})
		return true
	})
	return userList
}

func (s *userService) reloadUsers() {
	// 重新打开文件以确保文件指针在正确位置
	const userConfigFilePath = "./configs/users.json"
	file, err := os.Open(userConfigFilePath)
	if err != nil {
		logrus.WithError(err).Error("failed to open user config file for reload")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		logrus.WithError(err).Error("failed to read user config file")
		return
	}
	var userList []entities.UserEntity
	err = json.Unmarshal(content, &userList)
	if err != nil {
		logrus.WithError(err).Error("failed to unmarshal user config file")
		userList = []entities.UserEntity{}
	}
	s.userList = userList
	// 更新用户信息
	for _, user := range userList {
		s.userMap.Store(user.Id, user)
	}
	for _, user := range userList {
		s.userName2IdMap.Store(user.Username, user.Id)
	}
	// 删除不存在的用户
	s.userMap.Range(func(key, value interface{}) bool {
		exist := false
		for _, user := range s.userList {
			if user.Id == key {
				exist = true
				break
			}
		}
		if !exist {
			s.userMap.Delete(key)
		}
		return true
	})
	s.userName2IdMap.Range(func(key, value interface{}) bool {
		exist := false
		for _, user := range s.userList {
			if user.Username == key {
				exist = true
				break
			}
		}
		if !exist {
			s.userName2IdMap.Delete(key)
		}
		return true
	})
}

func (s *userService) saveMapChange(ctx *gin.Context) {
	userList := []entities.UserEntity{}
	s.userMap.Range(func(key, value interface{}) bool {
		userList = append(userList, value.(entities.UserEntity))
		return true
	})
	content, err := json.Marshal(userList)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to marshal user config file")
		return
	}

	// 设置写入标记，避免监听器触发无限循环
	s.isWriting = true
	defer func() {
		s.isWriting = false
	}()

	// 重新打开文件进行写入，避免文件指针问题
	const userConfigFilePath = "./configs/users.json"
	file, err := os.OpenFile(userConfigFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to open user config file for writing")
		return
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to write user config file")
		return
	}
	err = file.Sync()
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to sync user config file")
		return
	}
	logger.CtxLogger(ctx).Info("user config file saved")
}

func (s *userService) Login(ctx *gin.Context, username string, password string) (retUser *entities.UserEntity, token string, err error) {
	// 通过用户名获取用户 ID
	userId, ok := s.userName2IdMap.Load(username)
	if !ok {
		logger.CtxLogger(ctx).WithFields(logrus.Fields{
			"username": username,
		}).Error("user not found")
		return nil, "", errors.New("username or password is incorrect")
	}
	user, ok := s.userMap.Load(userId)
	if !ok {
		logger.CtxLogger(ctx).WithFields(logrus.Fields{
			"username": username,
		}).Error("user not found")
		return nil, "", errors.New("username or password is incorrect")
	}
	userEntity, _ := user.(entities.UserEntity)
	retUser = &userEntity
	saltedPass := utils.SaltedMd5Password(password, userEntity.Salt)
	if userEntity.Password != saltedPass {
		logger.CtxLogger(ctx).WithFields(logrus.Fields{
			"username": username,
		}).Error("password is incorrect")
		return retUser, "", errors.New("username or password is incorrect")
	}

	if !userEntity.AllowLogin {
		logger.CtxLogger(ctx).WithFields(logrus.Fields{
			"username": username,
		}).Error("user is not allowed to login")
		return retUser, "", errors.New("user is not allowed to login")
	}

	// 所有校验已经完成，更新用户登录时间，保存回文件
	userEntity.LastLoginAt = time.Now()
	userEntity.LoginCount++
	s.userMap.Store(userId, userEntity)
	// 将用户信息保存回文件
	s.saveMapChange(ctx)
	// 生成登录记录
	// 生成 token
	token, err = s.jwtManager.GenerateToken(userEntity.Id, userEntity.Username, userEntity.IsAdmin, userEntity.IsSiteAdmin)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to generate token")
		return retUser, "", errors.New("failed to generate token")
	}
	return retUser, token, nil
}

// Register implements UserServiceInterface.
func (s *userService) Register(ctx *gin.Context, username string, password string, nickname string) (err error) {
	// 判断用户名是否存在
	_, ok := s.userName2IdMap.Load(username)
	if ok {
		logger.CtxLogger(ctx).WithFields(logrus.Fields{
			"username": username,
		}).Error("user already exists")
		return nil // 用户名已存在，返回成功，迷惑非法用户
	}
	userId := uuid.New().String()
	salt := uuid.New().String()
	password = utils.SaltedMd5Password(password, salt)
	if nickname == "" {
		nickname = username
	}
	userEntity := entities.UserEntity{
		Id:          userId,
		Username:    username,
		Password:    password,
		IsAdmin:     false,
		IsSiteAdmin: false,
		Salt:        salt,
		Nickname:    nickname,
		AllowLogin:  false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastLoginAt: time.Now(),
		LoginCount:  0,
	}
	s.userMap.Store(userId, userEntity)
	s.userName2IdMap.Store(username, userId)
	s.saveMapChange(ctx)
	return nil
}

// Save implements UserServiceInterface.
func (s *userService) Save(ctx *gin.Context, userId string, userInfo *entities.UserEntity) {
	s.userMap.Store(userId, *userInfo)
	s.saveMapChange(ctx)
}

// UserDetail implements UserServiceInterface.
func (s *userService) UserDetail(ctx *gin.Context, userId string) (userInfo entities.UserDetailResponse, err error) {
	user, ok := s.userMap.Load(userId)
	if !ok {
		logger.CtxLogger(ctx).WithFields(logrus.Fields{
			"user_id": userId,
		}).Error("user not found")
		return userInfo, errors.New("user not found")
	}
	userEntity, _ := user.(entities.UserEntity)
	userInfo = entities.UserDetailResponse{
		Id:           userEntity.Id,
		Nickname:     userEntity.Nickname,
		Username:     userEntity.Username,
		IsAdmin:      userEntity.IsAdmin,
		IsSiteAdmin:  userEntity.IsSiteAdmin,
		AllowLogin:   userEntity.AllowLogin,
		CreatedAt:    userEntity.CreatedAt,
		UpdatedAt:    userEntity.UpdatedAt,
		LastLoginAt:  userEntity.LastLoginAt,
		LoginCount:   userEntity.LoginCount,
		LoginHistory: []entities.LoginHistory{},
	}
	// 补充登录历史
	loginHistoryFilePath := fmt.Sprintf("./configs/login_history/%s.json", userEntity.Username)
	content, err := os.ReadFile(loginHistoryFilePath)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to read login history file")
	}
	if len(content) > 0 {
		if err := json.Unmarshal(content, &userInfo.LoginHistory); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("failed to unmarshal login history file")
		}
	}
	return userInfo, nil
}

func (s *userService) UserDetailInner(userId string) (userInfo entities.UserEntity, exist bool) {
	user, ok := s.userMap.Load(userId)
	if ok {
		userInfo, _ = user.(entities.UserEntity)
	}
	return userInfo, ok
}

func (s *userService) GenerateToken(user *entities.UserEntity) (string, error) {
	return s.jwtManager.GenerateToken(user.Id, user.Username, user.IsAdmin, user.IsSiteAdmin)
}
