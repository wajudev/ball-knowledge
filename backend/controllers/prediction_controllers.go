package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"ball-knowledge/database"
	"ball-knowledge/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreatePredictionRequest struct {
	MatchID            uuid.UUID `json:"match_id" binding:"required"`
	PredictedScoreHome int       `json:"predicted_score_home" binding:"required,min=0"`
	PredictedScoreAway int       `json:"predicted_score_away" binding:"required,min=0"`
}

// CreatePrediction handles creating a new prediction
func CreatePrediction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreatePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if prediction already exists
	var existingPrediction models.Prediction
	if err := database.DB.Where("user_id = ? AND match_id = ?", userUUID, req.MatchID).First(&existingPrediction).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Prediction already exists for this match"})
		return
	}

	// Verify match exists
	var match models.Match
	if err := database.DB.Where("id = ?", req.MatchID).First(&match).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Create prediction
	prediction := models.Prediction{
		UserID:             userUUID,
		MatchID:            req.MatchID,
		PredictedScoreHome: req.PredictedScoreHome,
		PredictedScoreAway: req.PredictedScoreAway,
		Points:             CalculatePoints(req.PredictedScoreHome, req.PredictedScoreAway, match.Result),
	}

	if err := database.DB.Create(&prediction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prediction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Prediction created successfully",
		"prediction": prediction,
	})
}

// GetPrediction retrieves a user's prediction for a specific match
func GetPrediction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	matchID := c.Param("matchId")

	var prediction models.Prediction
	if err := database.DB.Where("user_id = ? AND match_id = ?", userID, matchID).First(&prediction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prediction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prediction": prediction})
}

// GetUserPredictions gets all predictions for the current user
func GetUserPredictions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var predictions []struct {
		models.Prediction
		HomeTeam string `json:"home_team"`
		AwayTeam string `json:"away_team"`
		Date     string `json:"date"`
		Result   string `json:"result"`
	}

	if err := database.DB.Table("predictions").
		Select("predictions.*, matches.home_team, matches.away_team, matches.date, matches.result").
		Joins("LEFT JOIN matches ON matches.id = predictions.match_id").
		Where("predictions.user_id = ?", userID).
		Order("matches.date DESC").
		Scan(&predictions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve predictions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"predictions": predictions,
		"count":       len(predictions),
	})
}

// UpdatePrediction allows updating a prediction (only before match starts)
func UpdatePrediction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	predictionID := c.Param("id")

	var req CreatePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find existing prediction
	var prediction models.Prediction
	if err := database.DB.Where("id = ? AND user_id = ?", predictionID, userID).First(&prediction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prediction not found"})
		return
	}

	// Get match to recalculate points
	var match models.Match
	if err := database.DB.Where("id = ?", prediction.MatchID).First(&match).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Match not found"})
		return
	}

	// Update prediction
	prediction.PredictedScoreHome = req.PredictedScoreHome
	prediction.PredictedScoreAway = req.PredictedScoreAway
	prediction.Points = CalculatePoints(req.PredictedScoreHome, req.PredictedScoreAway, match.Result)

	if err := database.DB.Save(&prediction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prediction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Prediction updated successfully",
		"prediction": prediction,
	})
}

// GetLeaderboard returns the leaderboard with user rankings
func GetLeaderboard(c *gin.Context) {
	var leaderboard []struct {
		UserID       string `json:"user_id"`
		Username     string `json:"username"`
		TotalPoints  int    `json:"total_points"`
		PredictionCount int `json:"prediction_count"`
		Rank         int    `json:"rank"`
	}

	// Get leaderboard data
	if err := database.DB.Table("predictions").
		Select("users.id as user_id, users.username, SUM(predictions.points) as total_points, COUNT(predictions.id) as prediction_count").
		Joins("LEFT JOIN users ON users.id = predictions.user_id").
		Group("users.id, users.username").
		Order("total_points DESC").
		Scan(&leaderboard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
		return
	}

	// Add ranking
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"count":       len(leaderboard),
	})
}

// CalculatePoints calculates points based on prediction accuracy
func CalculatePoints(predictedHome, predictedAway int, actualResult string) int {
	if actualResult == "" || actualResult == "0:0" {
		return 0 // No points if match hasn't been played
	}

	// Parse actual result
	parts := strings.Split(actualResult, ":")
	if len(parts) != 2 {
		return 0
	}

	actualHome, err1 := strconv.Atoi(parts[0])
	actualAway, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0
	}

	points := 0

	// Exact score: 10 points
	if predictedHome == actualHome && predictedAway == actualAway {
		points += 10
	}

	// Correct result (win/draw/loss): 5 points
	predictedResult := getMatchResult(predictedHome, predictedAway)
	actualMatchResult := getMatchResult(actualHome, actualAway)
	if predictedResult == actualMatchResult {
		points += 5
	}

	// Correct total goals: 3 points
	predictedTotal := predictedHome + predictedAway
	actualTotal := actualHome + actualAway
	if predictedTotal == actualTotal {
		points += 3
	}

	// Correct goal difference: 2 points
	predictedDiff := predictedHome - predictedAway
	actualDiff := actualHome - actualAway
	if predictedDiff == actualDiff {
		points += 2
	}

	return points
}

// getMatchResult determines the match result (home win, away win, or draw)
func getMatchResult(homeScore, awayScore int) string {
	if homeScore > awayScore {
		return "home_win"
	} else if awayScore > homeScore {
		return "away_win"
	}
	return "draw"
}