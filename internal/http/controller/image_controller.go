package controller

import (
	"booknest/internal/pkg/storage"
	"net/http"

	"github.com/gin-gonic/gin"

	"booknest/internal/middleware"
)

type imageController struct{}

func NewImageController() *imageController {
	return &imageController{}
}

func (c *imageController) RegisterRoutes(r gin.IRouter) {
	RegisterImageRoutes(r, getJWTConfig(), c)
}

func RegisterImageRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *imageController) {
	images := r.Group("/images")
	images.Use(middleware.JWTAuthMiddleware(jwtConfig), middleware.RequireAdmin())
	{
		images.POST("/upload", controller.UploadBookCover)
	}
}

// UploadBookCover godoc
// @Summary      Upload book cover
// @Description  Uploads a book cover image to S3 and returns the public URL
// @Tags         Images
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /images/upload [post]
func (controller *imageController) UploadBookCover(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("image")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "image required",
		})
		return
	}

	imageURL, err := storage.UploadImage(file, fileHeader)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"url": imageURL})
}
