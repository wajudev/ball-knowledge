package controllers

import (
	"errors"
	"game-knowledge/backend/database"
	"game-knowledge/backend/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var jwtKey = []byte("your_secret_key")

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type CreatePredictionInput struct {
	MatchID            uuid.UUID `json:"match_id" binding:"required"`
	PredictedScoreHome int       `json:"predicted_score_home" binding:"required"`
	PredictedScoreAway int       `json:"predicted_score_away" binding:"required"`
}

func RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Password = string(hashedPassword)

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func LoginUser(c *gin.Context) {
	var user models.User
	var input struct {
		UsernameOrEmail string `json:"usernameOrEmail" binding:"required"`
		Password        string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Where("username = ? OR email = ?", input.UsernameOrEmail, input.UsernameOrEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username/email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username/email or password"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID.String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func CreatePrediction(c *gin.Context) {
	tokenString := c.Request.Header.Get("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tokenString = tokenString[len("Bearer "):]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input CreatePredictionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch the match to get the actual result
	var match models.Match
	if err := database.DB.Where("id = ?", input.MatchID).First(&match).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Match not found"})
		return
	}

	// Create the prediction
	prediction := models.Prediction{
		UserID:             uuid.MustParse(claims.UserID),
		MatchID:            input.MatchID,
		PredictedScoreHome: input.PredictedScoreHome,
		PredictedScoreAway: input.PredictedScoreAway,
	}

	prediction.Points = CalculatePoints(prediction, match.Result)

	if err := database.DB.Create(&prediction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prediction})
}

func CreateMatch(c *gin.Context) {
	var matches []models.Match
	if err := c.ShouldBindJSON(&matches); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i := range matches {
		matches[i].ID = uuid.New()
		if err := database.DB.Create(&matches[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": matches})
}

func GetMatches(c *gin.Context) {
	var matches []models.Match
	if err := database.DB.Find(&matches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": matches})
}

func CalculatePoints(prediction models.Prediction, actualResult string) int {
	points := 0
	// parse actual results
	actualScores := strings.Split(actualResult, ":")
	if len(actualScores) != 2 {
		return points // Invalid result format
	}
	actualScoreHome, err1 := strconv.Atoi(actualScores[0])
	actualScoreAway, err2 := strconv.Atoi(actualScores[1])
	if err1 != nil || err2 != nil {
		return points // Invalid score
	}

	// Points for exact score
	if prediction.PredictedScoreHome == actualScoreHome && prediction.PredictedScoreAway == actualScoreAway {
		points += 10
	}

	// Points for predicting winning team
	predictedWinner := getWinner(prediction.PredictedScoreHome, prediction.PredictedScoreAway)
	actualWinner := getWinner(actualScoreHome, actualScoreAway)
	if predictedWinner == actualWinner {
		points += 5
	}

	// Points for correct total number of goals scored in the match
	if prediction.PredictedScoreHome+prediction.PredictedScoreAway == actualScoreHome+actualScoreAway {
		points += 3
	}

	return points
}

func GetLeaderboard(c *gin.Context) {
	tokenString := c.Request.Header.Get("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tokenString = tokenString[len("Bearer "):]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var leaderboard []struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Points   int    `json:"points"`
	}

	// Aggregate points by user
	if err := database.DB.Table("predictions").
		Select("users.id as user_id, users.username as username, SUM(predictions.points) as points").
		Joins("left join users on users.id = predictions.user_id").
		Group("users.id, users.username").
		Order("points DESC").
		Scan(&leaderboard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": leaderboard})
}

func GetUserProfile(c *gin.Context) {
	tokenString := c.Request.Header.Get("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tokenString = tokenString[len("Bearer "):]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func getWinner(score1, score2 int) string {
	if score1 > score2 {
		return "home-team"
	} else if score2 > score1 {
		return "away-team"
	} else {
		return "draw"
	}
}
