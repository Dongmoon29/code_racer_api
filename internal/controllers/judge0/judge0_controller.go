package judge0

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Dongmoon29/codeRacer_api/internal/services"
	"github.com/Dongmoon29/codeRacer_api/internal/util"
	"github.com/gin-gonic/gin"
)

type Judge0Controller struct {
	Judge0Service services.Judge0Service
}

var (
	instance *Judge0Controller
	once     sync.Once
)

func NewJudge0Controller(judge0Service services.Judge0Service) *Judge0Controller {
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

func (jc *Judge0Controller) CreateCodeSubmission(c *gin.Context) {
	c.JSON(200, "testing")
}

func (jc *Judge0Controller) GetCodeSubmission(c *gin.Context) {
	c.JSON(200, "testing")
}
