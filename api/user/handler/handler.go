package handler

import (
	"context"
	"log"
	"net/http"

	userPb "github.com/John-Tonny/mnhost/interface/out/user"

	"github.com/gin-gonic/gin"
)

type UserAPIHandler struct {
	userClient userPb.UserService
}

func GetUserHandler(userClient userPb.UserService) *UserAPIHandler {
	return &UserAPIHandler{
		userClient: userClient,
	}
}

func (s *UserAPIHandler) Login(c *gin.Context) {
	log.Printf("start login")
	user := userPb.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.userClient.Auth(context.Background(), &user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": resp.Token,
	})
}

func (s *UserAPIHandler) Register(c *gin.Context) {
	user := userPb.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.userClient.Create(context.Background(), &user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp.User.Id})
}
