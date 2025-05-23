package handlers

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"votes/repositories"
	"votes/requests"
	"votes/response"
	"votes/services"
	"votes/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userRepo    *repositories.UserRepository
	userService *services.UserService
}

func NewUserHandler(userRepo *repositories.UserRepository, userService *services.UserService) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		userService: userService,
	}
}

func (u *UserHandler) GetUserAll(c *gin.Context) {
	page, err := utils.GetQueryParamPagination(c)
	if err != nil {
		response.BadRequest(c, "invalid query params")
		return
	}

	keyword := c.Query("name")
	keyword = strings.TrimSpace(keyword)
	users, err := u.userService.GetAll(page, keyword)
	if err != nil {
		log.Printf("[UserHandler.GetUserAll] failed to get user data: %v", err)
		response.InternalServerError(c)
		return
	}

	userMaps := make([]response.UserResponse, 0, len(users))
	for _, user := range users {
		userMaps = append(userMaps, response.UserResponse{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	response.OkPaginate(c, userMaps, page, "user data")
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req requests.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errFields := utils.ExtractValidationErrors(err)
		response.UnprocessableEntity(c, errFields)
		return
	}

	create, err := h.userService.Create(&req)
	if err != nil {
		log.Printf("[UserHandler.CreateUser] failed to create user: %v", err)
		response.InternalServerError(c)
		return
	}

	res := response.UserResponse{
		Id:   create.Id,
		Name: req.Name,
	}

	response.Created(c, res, "create user data")
}

func (h *UserHandler) GetUserById(c *gin.Context) {
	id := c.Param("id")
	convId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[UserHandler.GetUserById] invalid userId param: %v", err)
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.userService.GetById(convId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "user not found")
			return
		}

		log.Printf("[UserHandler.GetUserById] failed to get user by id: %v", err)
		response.InternalServerError(c)
		return
	}

	res := response.UserResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}

	response.Ok(c, res, "user")
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idUser := c.Param("id")
	convId, err := strconv.Atoi(idUser)
	if err != nil {
		log.Printf("[UserHandler.UpdateUser] invalid userId param: %v", err)
		response.BadRequest(c, "invalid user id")
		return
	}

	var req requests.UserRequestUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[UserHandler.UpdateUser] invalid request body: %v", err)
		response.UnprocessableEntityMalformedJSON(c)
		return
	}

	if req.IsEmpty() {
		log.Printf("[UserHandler.UpdateUser] empty request body: %v", err)
		response.UnprocessableEntityEmpty(c)
		return
	}

	if err = h.userService.Update(&req, convId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "user not found")
			return
		}

		log.Printf("[UserHandler,UpdateUser] failed to update user: %v", err)
		response.InternalServerError(c)
		return
	}

	res := response.UserResponseUpdate{
		Name:  req.Name,
		Email: req.Email,
	}

	response.Ok(c, res, "user update")
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idUser := c.Param("id")
	convId, err := strconv.Atoi(idUser)
	if err != nil {
		log.Printf("[UserHandler.UpdateUser] invalid userId param: %v", err)
		response.BadRequest(c, "invalid user id")
		return
	}

	if err = h.userService.Delete(convId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "user not found")
			return
		}

		log.Printf("[UserHandler.DeleteUser] failed to delete user: %v", err)
		response.InternalServerError(c)
		return
	}

	response.Deleted(c, "user has been deleted")
}
