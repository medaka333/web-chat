// Masa作
package controllers

import (
	"net/http"
	"strconv"
	"web-chat/initializers"
	"web-chat/models"

	"github.com/gin-gonic/gin"
)

func Lists(c *gin.Context) {
	// 自分のIDに該当するfriendsテーブルを取得
	user, _ := c.Get("user")
	var friends []models.Friends
	userID := user.(models.Users).ID
	result := initializers.DB.Where("user_id1 = ? OR user_id2 = ?", userID, userID).Find(&friends)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	// friendsテーブルから、友達のIDのみを抽出
	var friendsIDs []uint
	for _, friend := range friends {
		if friend.UserID1 == userID {
			friendsIDs = append(friendsIDs, friend.UserID2)
		} else {
			friendsIDs = append(friendsIDs, friend.UserID1)
		}
	}
	// 友達IDをもとにusersテーブルから友達の名前を取得。 IN:スライスを扱える。Pluck():スライス形式で渡せる
	// var friends_name []string
	var users []models.Users
	// result = initializers.DB.Model(&models.Users{}).Where("id IN ?", friendsIDs).Pluck("user_name", &friends_name)
	result = initializers.DB.Where("id IN ?", friendsIDs).Find(&users)

	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	// 自分のIDに該当するgroupsテーブルを取得し、ルーム名を取得する
	groups := []models.Groups{}
	rooms := []models.Rooms{}
	var rooms_refer []uint
	result = initializers.DB.Where("users_refer = ?", userID).Find(&groups)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	for _, r := range groups {
		rooms_refer = append(rooms_refer, r.RoomsRefer)
	}
	result = initializers.DB.Where("id IN ?", rooms_refer).Find(&rooms)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	c.HTML(http.StatusOK, "home.html", gin.H{
		"title": "Home",
		// "friends_name": friends_name,
		"users": users,
		"rooms": rooms,
	})
}

func SearchFriend(c *gin.Context) {
	user, _ := c.Get("user")
	user_name := (c.PostForm("searchBar"))
	if user_name == user.(models.Users).UserName {
		Result(c, "自分は追加できません", models.Users{})
		return
	}
	if user_name == "" {
		Result(c, "ユーザー名を入力してください", models.Users{})
		return
	}
	user1 := models.Users{}
	result := initializers.DB.Where("user_name = ?", user_name).Find(&user1)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	if user1.UserName == "" {
		Result(c, "ユーザーが見つかりませんでした", models.Users{})
		return
	}
	// もう友達なら友達追加しない
	friend := models.Friends{}
	conditions := "(user_id1 = ? AND user_id2 = ?) OR (user_id1 = ? AND user_id2 = ?)"
	userID := user.(models.Users).ID
	result = initializers.DB.Where(conditions, userID, user1.ID, user1.ID, userID).First(&friend)
	if friend.UserID1 == user1.ID || friend.UserID2 == user1.ID {
		Result(c, "すでにフレンドです", models.Users{})
		return
	}
	Result(c, "ユーザーが見つかりました", user1)
}

func AddFriend(c *gin.Context) {
	user, _ := c.Get("user")
	user_id2, _ := strconv.Atoi(c.PostForm("userID"))
	user_name := (c.PostForm("userName"))
	room_name := user.(models.Users).UserName + user_name
	room := models.Rooms{RoomName: room_name}
	//　部屋作成
	result := initializers.DB.Create(&room)
	if result.Error != nil {
		Result(c, "ユーザーを追加できませんでした", models.Users{})
		return
	}
	friend := models.Friends{
		UserID1:    user.(models.Users).ID,
		UserID2:    uint(user_id2),
		RoomsRefer: room.ID,
	}
	// フレンド追加
	result = initializers.DB.Create(&friend)
	if result.Error != nil {
		// 一度作った部屋を消す
		initializers.DB.Where("room_name = ?", room_name).Delete(models.Rooms{})
		Result(c, "ユーザーを追加できませんでした", models.Users{})
		return
	}
	// Hubを作成
	// h := websocket.NewHub()
	// go h.Run()
	// models.RoomToHub[room.ID] = h

	c.Redirect(http.StatusSeeOther, "/home")
}

func CreateGroup(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(models.Users)
	type GroupCreationRequest struct {
		GroupName       string   `json:"groupName"`
		SelectedFriends []string `json:"selectedFriends"`
	}
	var request GroupCreationRequest
	c.ShouldBindJSON(&request)
	room_name := request.GroupName
	var u_ids []uint
	u_ids = append(u_ids, u.ID)
	for _, u_id := range request.SelectedFriends {
		id, _ := strconv.Atoi(u_id)
		u_ids = append(u_ids, uint(id))
	}
	room := models.Rooms{RoomName: room_name}
	// ルーム作成
	result := initializers.DB.Debug().Create(&room)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	// グループメンバー作成
	for _, u := range u_ids {
		group_member := models.Groups{UsersRefer: u, RoomsRefer: room.ID}
		result = initializers.DB.Debug().Create(&group_member)

		if result.Error != nil {
			// 一度作った部屋を消す
			initializers.DB.Where("id = ?", room.ID).Delete(&models.Rooms{})
			// 一度作ったgroup_memberを消す
			initializers.DB.Where("users_refer IN ? AND rooms_refer = ?", u_ids, room.ID).Delete(&models.Groups{})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
			return
		}
	}
}

func DeleteFriend(c *gin.Context) {
	user, _ := c.Get("user")
	u_id := user.(models.Users).ID
	f_id, _ := strconv.Atoi(c.PostForm("userid"))
	friend := models.Friends{}
	conditions := "(user_id1 = ? AND user_id2 = ?) OR (user_id1 = ? AND user_id2 = ?)"
	result := initializers.DB.Debug().
		Where(conditions, f_id, u_id, u_id, f_id).
		Find(&friend)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	result = initializers.DB.Debug().Where("id = ?", friend.RoomsRefer).Delete(&models.Rooms{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}

	result = initializers.DB.Debug().
		Where(conditions, f_id, u_id, u_id, f_id).
		Delete(&models.Friends{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func DeleteGroup(c *gin.Context) {
	room_id := c.PostForm("room_id")
	result := initializers.DB.Debug().
		Where("id = ?", room_id).Delete(&models.Rooms{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	result = initializers.DB.Debug().
		Where("rooms_refer = ?", room_id).Delete(&[]models.Groups{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func Result(c *gin.Context, title string, user models.Users) {
	c.HTML(http.StatusOK, "result.html", gin.H{
		"title": title,
		"user":  user,
	})
}

//非同期処理しようとした
// func loadFriendListAsync(user models.Users, ch chan []string) {
// 	// フレンドリストをデータベースから取得する処理
// 	var friends []models.Friends
// 	initializers.DB.Where("user_id1 = ? OR user_id2 = ?", user.ID, user.ID).Find(&friends)
// 	// 友達のユーザーIDのみを抽出
// 	var friendIDs []uint
// 	for _, friend := range friends {
// 		if friend.UserID1 == user.ID {
// 			friendIDs = append(friendIDs, friend.UserID2)
// 		} else {
// 			friendIDs = append(friendIDs, friend.UserID1)
// 		}
// 	}
// 	// 友達のユーザー名を取得
// 	var friendNames []string
// 	initializers.DB.Model(&models.Users{}).Where("id IN ?", friendIDs).Pluck("user_name", &friendNames)
// 	// フレンドリストをチャネルに送信
// 	ch <- friendNames
// }

//	func Lists(c *gin.Context) {
//		user, _ := c.Get("user")
//		// フレンドリストのチャネルとグループリストのチャネルを作成
//		friendListCh := make(chan []string)
//		groupListCh := make(chan []models.Rooms)
//		// 非同期でフレンドリストとグループリストを取得
//		go loadFriendListAsync(user.(models.Users), friendListCh)
//		go loadGroupListAsync(user.(models.Users), groupListCh)
//		// チャネルからデータを受け取る
//		friendNames := <-friendListCh
//		rooms := <-groupListCh
//		// テンプレートにデータを渡して表示
//		c.HTML(http.StatusOK, "home.html", gin.H{
//			"title":        "Home",
//			"friends_name": friendNames,
//			"rooms":        rooms,
//		})
//	}

// func loadGroupListAsync(user models.Users, ch chan []models.Rooms) {
// 	// グループリストをデータベースから取得する処理
// 	var groups []models.Groups
// 	initializers.DB.Where("users_refer = ?", user.ID).Find(&groups)
// 	// グループに関連するルーム情報を取得
// 	var roomIDs []uint
// 	for _, group := range groups {
// 		roomIDs = append(roomIDs, group.RoomsRefer)
// 	}
// 	var rooms []models.Rooms
// 	initializers.DB.Where("id IN ?", roomIDs).Find(&rooms)
// 	// グループリストをチャネルに送信
// 	ch <- rooms
// }

// todoの処理
// func TodoCreate(c *gin.Context) {
// 	user, _ := c.Get("user")
// 	todo := models.Todo{
// 		UserID: user.(models.User).ID,
// 	}
// 	content := (c.PostForm("content"))
// 	todo.Content = content
// 	if content == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "何か書いてください",
// 		})
// 		return
// 	}
// 	result := initializers.DB.Debug().Create(&todo)
// 	if result.Error != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Failed to create todo",
// 		})
// 		return
// 	}
// 	c.Redirect(http.StatusMovedPermanently, "/todos/list")
// }

// func TodoUpdate(c *gin.Context) {
// 	user, _ := c.Get("user")
// 	todo := models.Todo{
// 		UserID: user.(models.User).ID,
// 	}
// 	id, _ := strconv.Atoi(c.PostForm("id"))
// 	// .Where().Whereで繋ぐことで’かつ'AND演算子の役割をする
// 	initializers.DB.Debug().Where("id=?", id).Where("user_id=?", todo.UserID).Find(&todo)
// 	// 15
// 	// 2
// 	// 古い内容
// 	content := c.PostForm("content")
// 	todo.Content = content
// 	// todo
// 	// gorm.Model = todoリストのID　＝15（15番目に作られたtodoリスト）
// 	// UserID     user.ID 誰が作ったtodoかわかる=2（2番目に登録したユーザー）
// 	// Content    =新しい内容
// 	result := initializers.DB.Debug().Updates(&todo)
// 	if result.Error != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Failed to update todo",
// 		})
// 		return
// 	}
// 	c.Redirect(http.StatusMovedPermanently, "/todos/list")
// }

// func MoveToEdit(c *gin.Context) {
// 	user, _ := c.Get("user")
// 	todo := models.Todo{
// 		UserID: user.(models.User).ID,
// 	}
// 	id, err := strconv.Atoi(c.Query("id"))
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	initializers.DB.Debug().Where("id=?", id).Where("user_id=?", todo.UserID).Find(&todo)
// 	c.HTML(http.StatusOK, "edit.html", gin.H{
// 		"title": user.(models.User).UserName + "'s Todos",
// 		"todo":  todo,
// 	})
// }

//	func DeleteTodo(c *gin.Context) {
//		user, _ := c.Get("user")
//		id, err := strconv.Atoi(c.Query("id"))
//		if err != nil {
//			log.Fatalln(err)
//		}
//		// Unscoped:物理削除
//		// Where("id = ? AND user_id = ?", id, user.(models.User).ID).Delete(&models.Todo{})
//		result := initializers.DB.Debug().Unscoped().Where("id = ?", id).Where("user_id = ?", user.(models.User).ID).Delete(&models.Todo{})
//		if result.Error != nil {
//			c.JSON(http.StatusBadRequest, gin.H{
//				"error": "Failed to delete todo",
//			})
//			return
//		}
//		c.Redirect(http.StatusSeeOther, "/todos/list")
//	}
