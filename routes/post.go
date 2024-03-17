package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Bhar8at/bhar8at.github.io/database"
	"github.com/Bhar8at/bhar8at.github.io/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

var commentLimit = 10

func NewPost(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	// Checking if the user is logged in
	if id == nil {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}

	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "makepostT.html", nil)
	case "POST":
		var post models.Post

		if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to parse form.",
			})
			return
		}

		if err := c.ShouldBind(&post); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": err.Error(),
			})
			return
		}

		post.Id = uuid.NewString()
		post.CreatedAt = time.Now()

		file, header, _ := c.Request.FormFile("images[]")
		defer file.Close()

		// Read the file content into a byte slice
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to read image data.",
			})
			return
		}

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(header.Filename))
		path := filepath.Join("uploads", filename)

		err = os.MkdirAll("uploads", 0755)
		if err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to read image data.",
			})
			return
		}

		err = os.WriteFile(path, fileBytes, 0644)
		if err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to read image data.",
			})
			return
		}

		baseURL := "http://localhost:8080"
		imageURL := baseURL + "/" + path
		fmt.Println("HEre is the Image URL : ", imageURL)
		post.Images = imageURL

		if result := database.CreatePost(id.(string), &post); !result {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to create post, please try again later.",
			})
			return
		}
		c.Redirect(http.StatusFound, "/post/"+post.Id)
	}
}

func GetPost(c *gin.Context) {
	var self, voted bool
	session := sessions.Default(c)
	id := session.Get("userId")
	postId := c.Param("id")
	post := database.ReadPost(postId)
	if post == nil {
		c.HTML(http.StatusNotFound, "errorT.html", gin.H{
			"error":   "404 Not Found",
			"message": "Post not found or doesn't exist.",
		})
		return
	}
	commentLimit = 10
	comments := database.ReadComments(post.Id, 10, 0)
	for index := range comments {
		comments[index].Username = database.ReadUserById(comments[index].UserId).Username
		// Enable delete comment if its current user's comment
		if id != nil && id.(string) == comments[index].UserId {
			comments[index].Self = true
		}
	}
	if id != nil {
		// Check if current user has voted on post
		voted = database.Voted(id.(string), post.Id)
		// Enable delete post if its current user's post
		if id.(string) == post.UserId {
			self = true
		}
	}
	fmt.Println("\n\nHere is the image data : \n\n", post.Images)

	c.HTML(http.StatusOK, "getpostT.html", gin.H{
		"author":   database.ReadUserById(post.UserId),
		"post":     post,
		"self":     self,
		"voted":    voted,
		"voters":   database.ReadVotes(post.Id),
		"comments": comments,
		"imageURL": post.Images,
	})
}

// Return comments for loading through AJAX
func LoadMoreComments(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	postId := c.Param("id")
	comments := database.ReadComments(postId, 10, commentLimit)
	commentLimit += 10
	for index := range comments {
		comments[index].Username = database.ReadUserById(comments[index].UserId).Username
		// Enable delete comment if its current user's comment
		if id != nil && id.(string) == comments[index].UserId {
			comments[index].Self = true
		}
	}
	c.JSON(http.StatusOK, comments)
}

func DeletePost(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	postId := c.Param("id")
	post := database.ReadPost(postId)
	if id.(string) != post.UserId {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "Cannot perform this task.",
		})
		return
	}
	if result := database.DeletePost(post.Id); !result {
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to delete post, try again later.",
		})
		return
	}
	c.HTML(http.StatusOK, "responseT.html", gin.H{
		"message": "Post deleted successfully.",
	})
}

func ToggleVote(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	postId := c.Param("id")
	database.ToggleVote(id.(string), postId)
	c.Redirect(http.StatusFound, "/post/"+postId)
}

func Comment(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	var comment models.Comment
	if err := c.Request.ParseForm(); err != nil {
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to parse form.",
		})
		return
	}
	if err := c.ShouldBindWith(&comment, binding.Form); err != nil {
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": err.Error(),
		})
		return
	}
	postId := c.Param("id")
	comment.Id = uuid.NewString()
	comment.CreatedAt = time.Now()
	if result := database.CreateComment(id.(string), postId, &comment); !result {
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to add comment, try again later.",
		})
		return
	}
	c.Redirect(http.StatusFound, "/post/"+postId)
}

func DeleteComment(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	postId := c.Param("id")
	commentId := c.Query("commentId")
	comment := database.ReadComment(commentId)
	if comment == nil {
		c.HTML(http.StatusNotFound, "errorT.html", gin.H{
			"error":   "404 Not Found",
			"message": "Comment not found.",
		})
		return
	}
	if id.(string) != comment.UserId {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "Cannot perform this task.",
		})
		return
	}
	if result := database.DeleteComment(commentId); !result {
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to delete comment, try again later.",
		})
		return
	}
	c.Redirect(http.StatusFound, "/post/"+postId)
}
