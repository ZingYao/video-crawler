package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"video-crawler/internal/entities"
	"video-crawler/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type HistoryService interface {
	// 搜索历史
	AddSearchHistory(ctx *gin.Context, userEntity *entities.UserEntity, keyword, sourceId string) error
	GetSearchHistory(ctx *gin.Context, userName string) (entities.HistoryResponse, error)

	// 视频观看历史
	AddVideoHistory(ctx *gin.Context, userEntity *entities.UserEntity, videoId, videoTitle, videoUrl, sourceId, sourceName string, watchTime int64, progress float64) error
	GetVideoHistory(ctx *gin.Context, userName string) (entities.HistoryResponse, error)

	// 登录历史
	AddLoginHistory(ctx *gin.Context, userEntity *entities.UserEntity, password, token string)
	GetLoginHistory(ctx *gin.Context, userName string) (entities.HistoryResponse, error)
}

type historyService struct {
	searchHistoryDir string
	videoHistoryDir  string
	loginHistoryDir  string
	mutex            sync.RWMutex
}

var historyServiceInstance *historyService

func GetHistoryService() HistoryService {
	if historyServiceInstance == nil {
		historyServiceInstance = &historyService{
			searchHistoryDir: "./configs/search_history",
			videoHistoryDir:  "./configs/video_history",
			loginHistoryDir:  "./configs/login_history",
		}
		historyServiceInstance.initDirectories()
	}
	return historyServiceInstance
}

// 初始化目录
func (s *historyService) initDirectories() {
	// 确保数据目录存在
	if err := os.MkdirAll(s.searchHistoryDir, 0755); err != nil {
		logrus.WithError(err).Error("Failed to create search history directory")
	}
	if err := os.MkdirAll(s.videoHistoryDir, 0755); err != nil {
		logrus.WithError(err).Error("Failed to create video history directory")
	}
	if err := os.MkdirAll(s.loginHistoryDir, 0755); err != nil {
		logrus.WithError(err).Error("Failed to create login history directory")
	}
}

// 通用JSON文件加载函数
func (s *historyService) loadJSONFile(filePath string, v interface{}) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(v)
}

// 通用JSON文件保存函数
func (s *historyService) saveJSONFile(filePath string, v interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// 搜索历史相关方法
func (s *historyService) AddSearchHistory(ctx *gin.Context, userEntity *entities.UserEntity, keyword, sourceId string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.CtxLogger(ctx).WithFields(logrus.Fields{
		"user_id":   userEntity.Id,
		"keyword":   keyword,
		"source_id": sourceId,
	}).Info("Adding search history")

	// 确保目录存在
	if _, err := os.Stat(s.searchHistoryDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.searchHistoryDir, 0755); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("Failed to create search history directory")
			return fmt.Errorf("failed to create search history directory: %w", err)
		}
	}

	// 文件路径
	filePath := filepath.Join(s.searchHistoryDir, userEntity.Username+".json")

	// 读取文件内容（不存在则视为空）
	var searchHistoryList []entities.SearchHistory
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to read search history file")
		return fmt.Errorf("failed to read search history file: %w", err)
	}
	if len(content) > 0 {
		if err := json.Unmarshal(content, &searchHistoryList); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("Failed to unmarshal search history file")
			return fmt.Errorf("failed to unmarshal search history file: %w", err)
		}
	}

	// 最多保留 50 条记录（先进先出）
	if len(searchHistoryList) >= 50 {
		searchHistoryList = searchHistoryList[1:]
		logger.CtxLogger(ctx).Info("Removed oldest search history record due to limit")
	}

	// 添加新记录
	history := entities.SearchHistory{
		Id:        uuid.New().String(),
		UserId:    userEntity.Id,
		Keyword:   keyword,
		SourceId:  sourceId,
		CreatedAt: time.Now(),
	}
	searchHistoryList = append(searchHistoryList, history)

	// 写回文件（原子覆盖）
	content, err = json.Marshal(searchHistoryList)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to marshal search history file")
		return fmt.Errorf("failed to marshal search history file: %w", err)
	}
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to write search history file")
		return fmt.Errorf("failed to write search history file: %w", err)
	}

	logger.CtxLogger(ctx).WithFields(logrus.Fields{
		"history_id": history.Id,
		"total":      len(searchHistoryList),
	}).Info("Search history added successfully")

	return nil
}

func (s *historyService) GetSearchHistory(ctx *gin.Context, username string) (entities.HistoryResponse, error) {
	logger.CtxLogger(ctx).WithField("username", username).Info("Getting search history")

	// 读取文件数据
	filePath := filepath.Join(s.searchHistoryDir, username+".json")
	var searchHistoryList []entities.SearchHistory
	if err := s.loadJSONFile(filePath, &searchHistoryList); err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to load search history file")
		return entities.HistoryResponse{}, err
	}

	// 按创建时间倒序排序
	sort.Slice(searchHistoryList, func(i, j int) bool {
		return searchHistoryList[i].CreatedAt.After(searchHistoryList[j].CreatedAt)
	})

	logger.CtxLogger(ctx).WithField("total", len(searchHistoryList)).Info("Search history retrieved successfully")

	return entities.HistoryResponse{
		Total: len(searchHistoryList),
		Data:  searchHistoryList,
	}, nil
}

// 视频观看历史相关方法
func (s *historyService) AddVideoHistory(ctx *gin.Context, userEntity *entities.UserEntity, videoId, videoTitle, videoUrl, sourceId, sourceName string, watchTime int64, progress float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.CtxLogger(ctx).WithFields(logrus.Fields{
		"user_id":     userEntity.Id,
		"video_id":    videoId,
		"video_title": videoTitle,
		"source_id":   sourceId,
		"watch_time":  watchTime,
		"progress":    progress,
	}).Info("Adding video history")

	// 确保目录存在
	if _, err := os.Stat(s.videoHistoryDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.videoHistoryDir, 0755); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("Failed to create video history directory")
			return fmt.Errorf("failed to create video history directory: %w", err)
		}
	}

	// 文件路径
	filePath := filepath.Join(s.videoHistoryDir, userEntity.Username+".json")

	// 读取文件内容（不存在则视为空）
	var videoHistoryList []entities.VideoHistory
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to read video history file")
		return fmt.Errorf("failed to read video history file: %w", err)
	}
	if len(content) > 0 {
		if err := json.Unmarshal(content, &videoHistoryList); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("Failed to unmarshal video history file")
			return fmt.Errorf("failed to unmarshal video history file: %w", err)
		}
	}

	// 检查是否已存在相同视频的记录
	for i, history := range videoHistoryList {
		if history.VideoId == videoId {
			// 更新现有记录
			videoHistoryList[i].WatchTime = watchTime
			videoHistoryList[i].Progress = progress
			videoHistoryList[i].UpdatedAt = time.Now()

			logger.CtxLogger(ctx).WithFields(logrus.Fields{
				"video_id": videoId,
				"action":   "update",
			}).Info("Updated existing video history record")

			// 写回文件（原子覆盖）
			content, err = json.Marshal(videoHistoryList)
			if err != nil {
				logger.CtxLogger(ctx).WithError(err).Error("Failed to marshal video history file")
				return fmt.Errorf("failed to marshal video history file: %w", err)
			}
			if err := os.WriteFile(filePath, content, 0644); err != nil {
				logger.CtxLogger(ctx).WithError(err).Error("Failed to write video history file")
				return fmt.Errorf("failed to write video history file: %w", err)
			}
			return nil
		}
	}

	// 最多保留 100 条记录（先进先出）
	if len(videoHistoryList) >= 100 {
		videoHistoryList = videoHistoryList[1:]
		logger.CtxLogger(ctx).Info("Removed oldest video history record due to limit")
	}

	// 添加新记录
	history := entities.VideoHistory{
		Id:         uuid.New().String(),
		UserId:     userEntity.Id,
		VideoId:    videoId,
		VideoTitle: videoTitle,
		VideoUrl:   videoUrl,
		SourceId:   sourceId,
		SourceName: sourceName,
		WatchTime:  watchTime,
		Progress:   progress,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	videoHistoryList = append(videoHistoryList, history)

	// 写回文件（原子覆盖）
	content, err = json.Marshal(videoHistoryList)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to marshal video history file")
		return fmt.Errorf("failed to marshal video history file: %w", err)
	}
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to write video history file")
		return fmt.Errorf("failed to write video history file: %w", err)
	}

	logger.CtxLogger(ctx).WithFields(logrus.Fields{
		"history_id": history.Id,
		"total":      len(videoHistoryList),
		"action":     "add",
	}).Info("Video history added successfully")

	return nil
}

func (s *historyService) GetVideoHistory(ctx *gin.Context, username string) (entities.HistoryResponse, error) {
	logger.CtxLogger(ctx).WithField("username", username).Info("Getting video history")

	// 读取文件数据
	filePath := filepath.Join(s.videoHistoryDir, username+".json")
	var videoHistoryList []entities.VideoHistory
	if err := s.loadJSONFile(filePath, &videoHistoryList); err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to load video history file")
		return entities.HistoryResponse{}, err
	}

	// 按更新时间倒序排序
	sort.Slice(videoHistoryList, func(i, j int) bool {
		return videoHistoryList[i].UpdatedAt.After(videoHistoryList[j].UpdatedAt)
	})

	logger.CtxLogger(ctx).WithField("total", len(videoHistoryList)).Info("Video history retrieved successfully")

	return entities.HistoryResponse{
		Total: len(videoHistoryList),
		Data:  videoHistoryList,
	}, nil
}

// 登录历史相关方法
func (s *historyService) AddLoginHistory(ctx *gin.Context, userEntity *entities.UserEntity, password, token string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保目录存在
	loginHistoryDir := "./configs/login_history"
	if _, err := os.Stat(loginHistoryDir); os.IsNotExist(err) {
		if err := os.MkdirAll(loginHistoryDir, 0755); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("failed to create login history directory")
		}
	}

	// 文件路径
	loginHistoryFilePath := fmt.Sprintf("%s/%s.json", loginHistoryDir, userEntity.Username)

	// 读取文件内容（不存在则视为空）
	var loginHistoryList []entities.LoginHistory
	content, err := os.ReadFile(loginHistoryFilePath)
	if err != nil && !os.IsNotExist(err) {
		logger.CtxLogger(ctx).WithError(err).Error("failed to read login history file")
	}
	if len(content) > 0 {
		if err := json.Unmarshal(content, &loginHistoryList); err != nil {
			logger.CtxLogger(ctx).WithError(err).Error("failed to unmarshal login history file")
			loginHistoryList = []entities.LoginHistory{}
		}
	}

	// 最多保留 10 条记录（先进先出）
	if len(loginHistoryList) >= 10 {
		loginHistoryList = loginHistoryList[1:]
	}
	success := token != ""

	// 追加新的登录记录
	loginHistoryList = append(loginHistoryList, entities.LoginHistory{
		LoginAt:   time.Now(),
		Ip:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
		Password:  password,
		Success:   success,
		Token:     token,
	})

	// 写回文件（原子覆盖）
	content, err = json.Marshal(loginHistoryList)
	if err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to marshal login history file")
		return
	}
	if err := os.WriteFile(loginHistoryFilePath, content, 0644); err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("failed to write login history file")
		return
	}
}

func (s *historyService) GetLoginHistory(ctx *gin.Context, username string) (entities.HistoryResponse, error) {
	logger.CtxLogger(ctx).WithField("username", username).Info("Getting login history")

	// 读取文件数据
	filePath := filepath.Join(s.loginHistoryDir, username+".json")
	var loginHistoryList []entities.LoginHistory
	if err := s.loadJSONFile(filePath, &loginHistoryList); err != nil {
		logger.CtxLogger(ctx).WithError(err).Error("Failed to load login history file")
		return entities.HistoryResponse{}, err
	}

	// 按创建时间倒序排序
	sort.Slice(loginHistoryList, func(i, j int) bool {
		return loginHistoryList[i].LoginAt.After(loginHistoryList[j].LoginAt)
	})

	logger.CtxLogger(ctx).WithField("total", len(loginHistoryList)).Info("Login history retrieved successfully")

	return entities.HistoryResponse{
		Total: len(loginHistoryList),
		Data:  loginHistoryList,
	}, nil
}
