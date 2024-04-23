package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pegov/fauth-backend-go/internal/http/render"
	"github.com/pegov/fauth-backend-go/internal/service"
)

type AdminHandler interface {
	Ban(w http.ResponseWriter, r *http.Request) error
}

type adminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) AdminHandler {
	return &adminHandler{
		adminService: adminService,
	}
}

func (h *adminHandler) Ban(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	if err := h.adminService.Ban(int32(id)); err != nil {
		return err
	}

	render.Status(w, http.StatusOK)
	return nil
}
