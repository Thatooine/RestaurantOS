package users

import (
	"net/http"
	"strconv"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// UserServiceRESTAdaptor exposes user read operations over a REST API.
type UserServiceRESTAdaptor struct {
	service UserService
}

// NewUserServiceRESTAdaptor returns a new UserServiceRESTAdaptor.
func NewUserServiceRESTAdaptor(service UserService) *UserServiceRESTAdaptor {
	return &UserServiceRESTAdaptor{service: service}
}

// GetUser

type GetUserRESTResponse struct {
	User User `json:"user"`
}

func (a *UserServiceRESTAdaptor) GetUser(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Param("email")

	resp, err := a.service.GetUser(ctx, GetUserRequest{Email: email})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get user")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetUserRESTResponse{User: resp.User})
}

// ListUsers

type ListUsersRESTResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}

func (a *UserServiceRESTAdaptor) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListUsers(ctx, ListUsersRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list users")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListUsersRESTResponse{
		Users: resp.Users,
		Total: resp.Total,
	})
}

// SearchUsers

type SearchUsersRESTResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}

func (a *UserServiceRESTAdaptor) SearchUsers(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	q := query.Get("q")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.SearchUsers(
		ctx,
		SearchUsersRequest{
			Query:  q,
			Offset: offset,
			Limit:  limit,
		})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to search users")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, SearchUsersRESTResponse{
		Users: resp.Users,
		Total: resp.Total,
	})
}
