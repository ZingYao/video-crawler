package services

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"video-crawler/internal/entities"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type VideoSourceService interface {
	List() ([]entities.VideoSourceListResponse, error)
	Detail(videoSourceId string) (entities.VideoSourceEntity, error)
	Save(videoSource entities.VideoSourceEntity) error
	Delete(videoSourceId string) error
}

type videoSourceService struct {
	videoSourceMap  *sync.Map
	videoSourceList []entities.VideoSourceEntity
	isWriting       bool
}

func NewVideoSourceService() VideoSourceService {
	videoSourceService := &videoSourceService{
		videoSourceMap:  &sync.Map{},
		videoSourceList: []entities.VideoSourceEntity{},
		isWriting:       false,
	}
	const videoSourceConfigFilePath = "./configs/video-source.json"
	_, err := os.Stat(videoSourceConfigFilePath)
	if os.IsNotExist(err) {
		file, err := os.Create(videoSourceConfigFilePath)
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
	}
	videoSourceService.reloadVideoSource()

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		defer watcher.Close()
		err = watcher.Add(videoSourceConfigFilePath)
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
					if videoSourceService.isWriting {
						logrus.Debug("skipping reload during file write")
						continue
					}
					videoSourceService.reloadVideoSource()
					logrus.Info("video source config file changed, reload video source")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.WithError(err).Error("failed to watch video source config file")
			}
		}
	}()
	return videoSourceService
}

func (s *videoSourceService) List() ([]entities.VideoSourceListResponse, error) {
	videoSourceList := []entities.VideoSourceListResponse{}
	s.videoSourceMap.Range(func(key, value interface{}) bool {
		videoSource := value.(entities.VideoSourceEntity)
		videoSourceList = append(videoSourceList, entities.VideoSourceListResponse{
			Id:         videoSource.Id,
			Name:       videoSource.Name,
			Domain:     videoSource.Domain,
			Status:     videoSource.Status,
			SourceType: videoSource.SourceType,
			LuaScript:  videoSource.LuaScript,
		})
		return true
	})
	return videoSourceList, nil
}

func (s *videoSourceService) Detail(videoSourceId string) (entities.VideoSourceEntity, error) {
	videoSource, ok := s.videoSourceMap.Load(videoSourceId)
	if !ok {
		return entities.VideoSourceEntity{}, errors.New("video source not found")
	}
	return videoSource.(entities.VideoSourceEntity), nil
}

func (s *videoSourceService) reloadVideoSource() {
	const videoSourceConfigFilePath = "./configs/video-source.json"
	file, err := os.Open(videoSourceConfigFilePath)
	if err != nil {
		logrus.WithError(err).Error("failed to open video source config file")
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		logrus.WithError(err).Error("failed to read video source config file")
		return
	}
	var videoSourceList []entities.VideoSourceEntity
	err = json.Unmarshal(content, &videoSourceList)
	if err != nil {
		logrus.WithError(err).Error("failed to unmarshal video source config file")
		return
	}
	s.videoSourceList = videoSourceList
	for _, videoSource := range videoSourceList {
		s.videoSourceMap.Store(videoSource.Id, videoSource)
	}
	// 删除不存在的站点
	s.videoSourceMap.Range(func(key, value interface{}) bool {
		exist := false
		for _, videoSource := range videoSourceList {
			if videoSource.Id == key {
				exist = true
				break
			}
		}
		if !exist {
			s.videoSourceMap.Delete(key)
		}
		return true
	})

}

func (s *videoSourceService) Save(videoSource entities.VideoSourceEntity) error {
	// 如果是新站点（ID为空），生成新的UUID
	if videoSource.Id == "" {
		videoSource.Id = uuid.New().String()
	}

	// 检查站点是否已存在
	_, exists := s.videoSourceMap.Load(videoSource.Id)

	// 更新内存中的数据
	s.videoSourceMap.Store(videoSource.Id, videoSource)

	// 更新文件中的数据
	err := s.saveToFile()
	if err != nil {
		// 如果保存失败，回滚内存中的数据
		if !exists {
			s.videoSourceMap.Delete(videoSource.Id)
		}
		return err
	}

	return nil
}

func (s *videoSourceService) Delete(videoSourceId string) error {
	// 检查站点是否存在
	videoSource, exists := s.videoSourceMap.Load(videoSourceId)
	if !exists {
		return errors.New("video source not found")
	}

	// 从内存中删除
	s.videoSourceMap.Delete(videoSourceId)

	// 更新文件中的数据
	err := s.saveToFile()
	if err != nil {
		// 如果保存失败，恢复内存中的数据
		s.videoSourceMap.Store(videoSourceId, videoSource)
		return err
	}

	return nil
}

func (s *videoSourceService) saveToFile() error {
	const videoSourceConfigFilePath = "./configs/video-source.json"

	// 设置写入标志，防止文件监听器触发重载
	s.isWriting = true
	defer func() {
		s.isWriting = false
	}()

	// 从内存中获取所有站点数据
	var videoSourceList []entities.VideoSourceEntity
	s.videoSourceMap.Range(func(key, value interface{}) bool {
		videoSource := value.(entities.VideoSourceEntity)
		videoSourceList = append(videoSourceList, videoSource)
		return true
	})

	// 序列化为JSON
	content, err := json.MarshalIndent(videoSourceList, "", "    ")
	if err != nil {
		return err
	}

	// 写入文件
	err = os.WriteFile(videoSourceConfigFilePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
}
