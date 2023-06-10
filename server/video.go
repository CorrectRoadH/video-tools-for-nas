package server

import (
	"net/http"

	store "github.com/CorrectRoadH/video-tools-for-nas/store"
	"github.com/CorrectRoadH/video-tools-for-nas/types"
	"github.com/CorrectRoadH/video-tools-for-nas/utils"
	"github.com/gin-gonic/gin"
)

type DownloadVideoInput struct {
	Url     string `json:"url" binding:"required"`
	Storage string `json:"storage" binding:"required"`
}

type UpdateVideoStatusInput struct {
	Id      string `json:"id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Url     string `json:"url" binding:"required"`
	Status  string `json:"status" binding:"required"`
	Percent int    `json:"percent" binding:"required"`
	Size    int    `json:"size" binding:"required"` // 单位是字节
}

func composeResponse(data any) gin.H {
	return gin.H{"data": data}
}

func DownloadVideo(c *gin.Context) {
	var input DownloadVideoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := utils.HandlerDownloader(input.Url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	videoType := ""
	switch result {
	case "bilibili":
		videoType = "bilibili"
	case "youtube":
		videoType = "youtube"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown website"})
	}

	// 这里要不要用 goroutine 来做呢?
	store.GlobalVideoStatusMap.AddVideo(types.VideoStatus{
		Id:     utils.GenerateVideoIdFromURL(input.Url),
		Url:    input.Url,
		Status: "pending",
		Type:   videoType,
	})

	c.JSON(http.StatusOK, composeResponse("start downloading"))
}

// 这个是等 python 来调，来更新状态
func UpdateVideoStatus(c *gin.Context) {
	var input UpdateVideoStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store.GlobalVideoStatusMap.AddVideo(types.VideoStatus{
		Id:      input.Url, // 这个id通过url来取个hash值比较好
		Status:  input.Status,
		Percent: input.Percent,
	})
	if input.Percent == 100 {
		store.GlobalVideoStatusMap.DownloadComplete()
		store.GlobalVideoStatusMap.SchedulerDownload()
	}
	c.JSON(http.StatusOK, composeResponse("update video status"))
}

func GetAllVideoStatus(c *gin.Context) {
	c.JSON(http.StatusOK, composeResponse(store.GlobalVideoStatusMap.VideoStatusMap))
}
