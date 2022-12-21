package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"image-upload/internal/images"

	"github.com/gin-gonic/gin"
)

type (
	APIService struct {
		log        *log.Logger
		imgHandler images.Service
	}

	response struct {
		UUID string `json:"uuid"`
	}

	responseArray struct {
		Items []response `json:"items"`
	}
)

func New(log *log.Logger, f images.Service) *APIService {
	return &APIService{
		log:        log,
		imgHandler: f,
	}
}

func (s *APIService) BindAPI(r *gin.RouterGroup) {
	r.POST("image", s.create)
	r.GET("image/:id", s.getByID)
	r.GET("images", s.getAllIDs)
}

func (s *APIService) getAllIDs(c *gin.Context) {
	ids, err := s.imgHandler.GetAllIDs()
	if err != nil {
		s.log.Printf("cannot get ids: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error fetching ids",
		})
		return
	}

	data := []response{}
	for _, i := range ids {
		d := response{UUID: i}
		data = append(data, d)
	}

	responseData := responseArray{Items: data}
	c.JSON(http.StatusOK, responseData)
}

func (s *APIService) getByID(c *gin.Context) {
	id := c.Param("id")
	width := 0
	if query, ok := c.GetQuery("width"); ok {
		w, err := strconv.Atoi(query)
		if err != nil {
			s.log.Printf("image width not supported: %v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "image width not supported",
			})
			return

		}
		width = w
	}

	data, err := s.imgHandler.Download(id, width)
	if err != nil {
		s.log.Printf("image not found: %v", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "image not found",
		})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.jpg", id))
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (s *APIService) create(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		s.log.Printf("no file received: %v", err)
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"message": "no file received",
		})
		return
	}

	uuid, err := s.imgHandler.Upload(file)
	if err != nil {
		s.log.Printf("file upload failed: %v", err)
		if errors.Is(err, images.ErrInvalidFormat) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid image data",
			})
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "file upload failed",
			})
		}
		return

	}

	data := response{UUID: uuid}

	c.JSON(http.StatusOK, data)
}
