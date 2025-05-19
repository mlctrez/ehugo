package hueapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/mlctrez/ehugo/ssdp"
	"github.com/mlctrez/servicego"
	"go.etcd.io/bbolt"
	"io"
	"net/http"
	"strings"
)

type HueApi struct {
	engine  *gin.Engine
	logger  servicego.Logger
	addr    string
	bridges []*ssdp.BridgeInfo
	boltDb  *bbolt.DB
}

func New(logger servicego.Logger, boltDb *bbolt.DB, addr string, bridges ...*ssdp.BridgeInfo) *HueApi {
	result := &HueApi{
		logger:  logger,
		boltDb:  boltDb,
		addr:    addr,
		bridges: bridges,
	}
	result.setupEngine()
	return result
}

func (h *HueApi) setupEngine() {
	//gin.SetMode(gin.ReleaseMode)
	h.engine = gin.New()
	engine := h.engine
	engine.Use(h.loggingHandler())
	engine.Use(gin.Recovery())
	for _, bridge := range h.bridges {
		bridge.Location = fmt.Sprintf("http://%s/bridge/%s/device.xml", h.addr, bridge.SerialNumber)
	}
	engine.GET("/bridge/:serial/device.xml", h.DeviceHandler)
	engine.POST("/api", h.Authenticate)
	engine.GET("/api/:user/lights", h.Lights)
	engine.PUT("/api/:user/lights", h.ApiPutLight)
	engine.GET("/api/:user/lights/:lightId", h.Light)
	engine.DELETE("/api/:user/lights/:lightId", h.Delete)
	engine.PUT("/api/:user/lights/:lightId/state", h.LightState)
}

func (h *HueApi) DeviceHandler(c *gin.Context) {
	serial := c.Param("serial")

	for _, bridge := range h.bridges {
		if bridge.SerialNumber == serial {
			urlBase := strings.Replace(bridge.Location, "device.xml", "", 1)
			friendlyName := fmt.Sprintf("eHueGo v0.0.1 (%s)", bridge.SerialNumber)
			description := NewDevice(urlBase, friendlyName, bridge.SerialNumber, bridge.UUID)
			c.XML(200, description)
			return
		}
	}
	c.AbortWithStatus(http.StatusNotFound)
}

func (h *HueApi) loggingHandler() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			h.logger.Infof("%s %s %s %s", params.ClientIP, params.Method, params.Path, params.Latency)
			if params.ErrorMessage != "" {
				h.logger.Errorf("error %s", params.ErrorMessage)
			}
			return ""
		},
		Output: io.Discard,
	})
}

func (h *HueApi) Handler() http.Handler {
	return h.engine
}

type AuthRequest struct {
	DeviceType string `json:"deviceType"`
}

type AuthResponse struct {
	Success struct {
		Username string `json:"username"`
	} `json:"success"`
}

func (h *HueApi) Authenticate(c *gin.Context) {
	authRequest := &AuthRequest{}
	if err := c.ShouldBindJSON(authRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := make([]AuthResponse, 1)
	response[0].Success.Username = "83b7780291a6ceffbe0bd049104df"
	c.JSON(http.StatusOK, response)
}

func (h *HueApi) Lights(c *gin.Context) {
	getLights, err := h.GetLights()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, getLights)
}

func (h *HueApi) Light(c *gin.Context) {
	light, err := h.GetLight(c.Param("lightId"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	h.LogJson("light", light)
	c.JSON(http.StatusOK, light)
}

func (h *HueApi) LightState(c *gin.Context) {
	id := c.Param("lightId")
	light, err := h.GetLight(id)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	stateChange := &StateChange{}
	if err = c.ShouldBindJSON(stateChange); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.LogJson("stateChange", stateChange)

	response := light.ApplyStateChange(id, stateChange)
	err = h.UpdateLight(id, light)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	h.LogJson("response", response)

	c.JSON(http.StatusOK, response)
}

func (h *HueApi) LogJson(name string, what any) {
	marshal, err := json.Marshal(what)
	if err != nil {
		h.logger.Errorf("Error marshalling json: %v", err)
	}
	h.logger.Infof("%s %+v", name, string(marshal))
}

func (h *HueApi) Delete(c *gin.Context) {
	lightId := c.Param("lightId")
	err := h.DeleteLight(lightId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *HueApi) ApiPutLight(c *gin.Context) {
	light := &LightInfo{}
	if err := c.ShouldBindJSON(light); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbLight, lightId, err := h.PutLight(light)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{lightId: dbLight}
	c.JSON(http.StatusOK, response)
}
