package judge0

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/services/judge0"
	"github.com/Dongmoon29/code_racer_api/internal/util"
	"github.com/gin-gonic/gin"
)

type Judge0Controller struct {
	Judge0Service judge0.Judge0Service
}

var (
	instance *Judge0Controller
	once     sync.Once
)

func NewJudge0Controller(judge0Service judge0.Judge0Service) *Judge0Controller {
	once.Do(func() {
		instance = &Judge0Controller{
			Judge0Service: judge0Service,
		}
	})
	return instance
}

func (jc *Judge0Controller) GetAbout(c *gin.Context) {
	client := util.Client
	res, err := client.GET("/about")
	if err != nil {

		c.JSON(500, gin.H{
			"error":   "Failed to fetch data from /about",
			"message": err.Error(),
		})
		return
	}

	var response map[string]interface{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to parse JSON response", "detail": err.Error()})
		return
	}
	c.JSON(200, gin.H{"response": response})
}

func (jc *Judge0Controller) HandleCreateCodeSubmission(c *gin.Context) {
	var req dtos.CodeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid dto"})
		return
	}
	result, err := jc.Judge0Service.CreateCodeSubmission(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (jc *Judge0Controller) HandleGetCodeSubmission(c *gin.Context) {
}
