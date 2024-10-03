package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/EggsyOnCode/vm-manager/vm"
	"github.com/labstack/echo/v4"
)

const (
	ALPINE_LINUX = "alpine_linux"
)

type CreateVMRequest struct {
	Name    string `json:"name"`
	Mem     int    `json:"mem"`
	Storage int    `json:"storage"`
	OsType  string `json:"osType"`
}

func HandleVMCreateReq(ctx echo.Context) error {
	var req CreateVMRequest

	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": "bad request"})
	}

	if req.OsType != ALPINE_LINUX {
		return ctx.JSON(400, map[string]string{"error": "unsupported os type"})
	}

	if req.Storage > 200 {
		return ctx.JSON(400, map[string]string{"error": "storage too large"})
	}

	cfg := new(vm.Config)
	cfg.Name = req.Name
	cfg.Mem = req.Mem
	cfg.OsType = ALPINE_LINUX
	cfg.Storage = req.Storage

	xmlFile := cfg.CreateXMLConfig(req.Name, req.Mem, req.Storage, (req.OsType))

	cwd, _ := os.Getwd()
	basePath := cwd + "/templates"
	xmlFilePath := basePath + "/" + xmlFile
	cfg.XmlConfig = xmlFilePath

	v := new(vm.VM)
	if err := v.NewVM(cfg); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error(), "message": "vm creation failed"})
	}
	if err := v.Start(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error(), "message": "vm start failed"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok", "message": "vm created"})
}
