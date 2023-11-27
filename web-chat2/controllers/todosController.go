package controllers

import (
	"log"
	"net/http"
	"strconv"
	"web-chat/initializers"
	"web-chat/models"

	"github.com/gin-gonic/gin"
)

func TodoCreate(c *gin.Context) {
	user, _ := c.Get("user")
	todo := models.Todo{
		UserID: user.(models.Users).ID,
	}
	content := (c.PostForm("content"))
	todo.Content = content

	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "何か書いてください",
		})
		return
	}

	result := initializers.DB.Debug().Create(&todo)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create todo",
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/todos/list")
}

func TodoUpdate(c *gin.Context) {
	user, _ := c.Get("user")
	todo := models.Todo{
		UserID: user.(models.Users).ID,
	}
	id, _ := strconv.Atoi(c.PostForm("id"))
	// .Where().Whereで繋ぐことで’かつ'AND演算子の役割をする
	initializers.DB.Debug().Where("id=?", id).Where("user_id=?", todo.UserID).Find(&todo)
	// 15
	// 2
	// 古い内容

	content := c.PostForm("content")
	todo.Content = content

	// todo
	// gorm.Model = todoリストのID　＝15（15番目に作られたtodoリスト）
	// UserID     user.ID 誰が作ったtodoかわかる=2（2番目に登録したユーザー）
	// Content    =新しい内容
	result := initializers.DB.Debug().Updates(&todo)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update todo",
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/todos/list")
}

func MoveToEdit(c *gin.Context) {
	user, _ := c.Get("user")
	todo := models.Todo{
		UserID: user.(models.Users).ID,
	}
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		log.Fatalln(err)
	}
	initializers.DB.Debug().Where("id=?", id).Where("user_id=?", todo.UserID).Find(&todo)

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"title": user.(models.Users).UserName + "'s Todos",
		"todo":  todo,
	})
}

func DeleteTodo(c *gin.Context) {
	user, _ := c.Get("user")
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		log.Fatalln(err)
	}
	// Unscoped:物理削除
	// Where("id = ? AND user_id = ?", id, user.(models.User).ID).Delete(&models.Todo{})
	result := initializers.DB.Debug().Unscoped().Where("id = ?", id).Where("user_id = ?", user.(models.Users).ID).Delete(&models.Todo{})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to delete todo",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/todos/list")
}

func ListTodo(c *gin.Context) {
	user, _ := c.Get("user")
	var todos []models.Todo
	initializers.DB.Debug().Find(&todos, "user_id = ?", user.(models.Users).ID)

	c.HTML(http.StatusOK, "list.html", gin.H{
		"title": user.(models.Users).UserName + "'s Todos",
		"todos": todos,
	})
}
