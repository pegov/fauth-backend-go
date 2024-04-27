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
	Unban(w http.ResponseWriter, r *http.Request) error
	Kick(w http.ResponseWriter, r *http.Request) error
	Unkick(w http.ResponseWriter, r *http.Request) error
}

type adminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) AdminHandler {
	return &adminHandler{
		adminService: adminService,
	}
}

func actionOnID(w http.ResponseWriter, r *http.Request, action func(int32) error) error {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	if err := action(int32(id)); err != nil {
		return err
	}

	render.Status(w, http.StatusOK)
	return nil
}

func (h *adminHandler) Ban(w http.ResponseWriter, r *http.Request) error {
	return actionOnID(w, r, h.adminService.Ban)
}

func (h *adminHandler) Unban(w http.ResponseWriter, r *http.Request) error {
	return actionOnID(w, r, h.adminService.Unban)
}

func (h *adminHandler) Kick(w http.ResponseWriter, r *http.Request) error {
	return actionOnID(w, r, h.adminService.Unban)
}

func (h *adminHandler) Unkick(w http.ResponseWriter, r *http.Request) error {
	return actionOnID(w, r, h.adminService.Unban)
}
