package handler

import (
	"context"
	"log"
	"net/http"

	vpsPb "github.com/John-Tonny/mnhost/interface/out/vps"

	"github.com/gin-gonic/gin"
)

type VpsAPIHandler struct {
	vpsClient vpsPb.VpsService
}

func GetVpsHandler(vpsClient vpsPb.VpsService) *VpsAPIHandler {
	return &VpsAPIHandler{
		vpsClient: vpsClient,
	}
}

func (s *VpsAPIHandler) NewNode(c *gin.Context) {
	log.Printf("start new node")
	vps := vpsPb.Request{}
	if err := c.ShouldBindJSON(&vps); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.vpsClient.CreateNode(context.Background(), &vps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Order":  resp,
		"Errno":  resp.Errno,
		"Errmsg": resp.Errmsg,
	})
}

func (s *VpsAPIHandler) DelNode(c *gin.Context) {
	log.Printf("start del node")
	vps := vpsPb.Request{}
	if err := c.ShouldBindJSON(&vps); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.vpsClient.RemoveNode(context.Background(), &vps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Node":   resp,
		"Errno":  resp.Errno,
		"Errmsg": resp.Errmsg,
	})
}

func (s *VpsAPIHandler) ExpandVolume(c *gin.Context) {
	log.Printf("start new node")
	vps := vpsPb.VolumeRequest{}
	if err := c.ShouldBindJSON(&vps); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.vpsClient.ExpandVolume(context.Background(), &vps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Vps":    resp,
		"Errno":  resp.Errno,
		"Errmsg": resp.Errmsg,
	})
}

func (s *VpsAPIHandler) GetAllVps(c *gin.Context) {
	log.Printf("start get all vps")
	vps := vpsPb.Request{}
	if err := c.ShouldBindJSON(&vps); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.vpsClient.GetAllVps(context.Background(), &vps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Vps":    resp.Vpss,
		"Errno":  resp.Errno,
		"Errmsg": resp.Errmsg,
	})
}

func (s *VpsAPIHandler) GetAllNodeFromUser(c *gin.Context) {
	log.Printf("get all node from vps")
	vps := vpsPb.Request{}
	if err := c.ShouldBindJSON(&vps); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := s.vpsClient.GetAllNodeFromUser(context.Background(), &vps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "err", "errmsg": err})
		//c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Nodes":  resp.Nodes,
		"Errno":  resp.Errno,
		"Errmsg": resp.Errmsg,
	})
}
