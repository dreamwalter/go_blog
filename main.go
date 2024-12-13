// main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string            `bson:"title" json:"title"`
	Content   string            `bson:"content" json:"content"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

var client *mongo.Client
var postsCollection *mongo.Collection

func init() {
	// 连接 MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	
	// 获取集合
	postsCollection = client.Database("blog").Collection("posts")
}

func main() {
	r := gin.Default()
	
	// 允许跨域
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
		}
		
		c.Next()
	})
	
	// 路由
	api := r.Group("/api")
	{
		api.GET("/posts", getPosts)
		api.GET("/posts/:id", getPost)
		api.POST("/posts", createPost)
		api.PUT("/posts/:id", updatePost)
		api.DELETE("/posts/:id", deletePost)
	}
	
	r.Run(":8080")
}

// 获取所有文章
func getPosts(c *gin.Context) {
	ctx := context.Background()
	
	cursor, err := postsCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)
	
	var posts []Post
	if err = cursor.All(ctx, &posts); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, posts)
}

// 获取单篇文章
func getPost(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}
	
	var post Post
	err = postsCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&post)
	if err != nil {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}
	
	c.JSON(200, post)
}

// 创建文章
func createPost(c *gin.Context) {
	var post Post
	if err := c.BindJSON(&post); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
	}
	
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	
	result, err := postsCollection.InsertOne(context.Background(), post)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	post.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(201, post)
}

// 更新文章
func updatePost(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}
	
	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	post.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": bson.M{
			"title":      post.Title,
			"content":    post.Content,
			"updated_at": post.UpdatedAt,
		},
	}
	
	result := postsCollection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	
	if result.Err() != nil {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}
	
	c.JSON(200, post)
}

// 删除文章
func deletePost(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}
	
	result, err := postsCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	if result.DeletedCount == 0 {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}
	
	c.JSON(200, gin.H{"message": "Post deleted"})
}