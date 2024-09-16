package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"game-knowledge/backend/database"
	"game-knowledge/backend/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"os"
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
	PredictedScoreHome *int      `json:"predicted_score_home" binding:"required"`
	PredictedScoreAway *int      `json:"predicted_score_away" binding:"required"`
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

	// Check if the prediction already exists
	var existingPrediction models.Prediction
	if err := database.DB.Where("user_id = ? AND match_id = ?", claims.UserID, input.MatchID).First(&existingPrediction).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Prediction already exists"})
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
		PredictedScoreHome: *input.PredictedScoreHome,
		PredictedScoreAway: *input.PredictedScoreAway,
	}

	prediction.Points = CalculatePoints(prediction, match.Result)

	// Add logging to verify data is being passed correctly
	fmt.Printf("Prediction: %+v\n", prediction)

	if err := database.DB.Create(&prediction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prediction})
}

func GetPrediction(c *gin.Context) {
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

	matchID := c.Param("matchId")
	var prediction models.Prediction
	if err := database.DB.Where("user_id = ? AND match_id = ?", claims.UserID, matchID).First(&prediction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prediction not found"})
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
	// Fetch and store matches from the API
	if err := fetchAndStoreMatches(); err != nil {
		fmt.Println("Error fetching matches:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch matches from the database
	var matches []models.Match
	if err := database.DB.Find(&matches).Error; err != nil {
		fmt.Println("Error retrieving matches from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the matches as a response
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

func GetMatchesForGameWeek(context *gin.Context) {
	gameWeek := context.Param("gameweek")
	var matches []models.Match
	if err := database.DB.Where("match_day = ?", gameWeek).Find(&matches).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": matches})

}

func fetchMatchesFromAPI() ([]models.Match, error) {
	apiKey := os.Getenv("API_FOOTBALL_KEY")
	leagueID := os.Getenv("LEAGUE_ID")
	season := os.Getenv("SEASON")

	if apiKey == "" {
		return nil, fmt.Errorf("error: Missing API key")
	}

	// Construct the API URL
	url := fmt.Sprintf("https://v3.football.api-sports.io/fixtures?league=%s&season=%s", leagueID, season)
	method := "GET"

	// Create a new HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("x-apisports-key", apiKey)

	// Send the request
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to API: %v", err)
	}
	defer res.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %v", err)
	}

	// Extract the fixtures from the response
	fixtures, ok := result["response"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("error: Unable to parse fixtures from response")
	}

	var matches []models.Match

	// Iterate over each fixture and extract match details
	for _, f := range fixtures {
		fixture := f.(map[string]interface{})["fixture"].(map[string]interface{})
		teams := f.(map[string]interface{})["teams"].(map[string]interface{})
		homeTeam := teams["home"].(map[string]interface{})["name"].(string)
		awayTeam := teams["away"].(map[string]interface{})["name"].(string)
		dateStr := fixture["date"].(string)

		// Parse the date from the API
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}

		// Extract round
		roundStr := f.(map[string]interface{})["league"].(map[string]interface{})["round"].(string)
		var matchDay int
		fmt.Sscanf(roundStr, "Regular Season - %d", &matchDay)

		// Extract full-time score
		score := f.(map[string]interface{})["score"].(map[string]interface{})
		fullTimeScore := score["fulltime"].(map[string]interface{})

		// Check if home/away scores are present before casting to float64
		var homeScore, awayScore int
		if fullTimeScore["home"] != nil {
			homeScore = int(fullTimeScore["home"].(float64))
		}
		if fullTimeScore["away"] != nil {
			awayScore = int(fullTimeScore["away"].(float64))
		}

		// Create the match model
		match := models.Match{
			HomeTeam: homeTeam,
			AwayTeam: awayTeam,
			Date:     date.String(),
			League:   "Premier League", // Can be made dynamic if needed
			Season:   os.Getenv("SEASON"),
			MatchDay: matchDay,
			Result:   fmt.Sprintf("%d:%d", homeScore, awayScore),
		}

		// Add the match to the list of matches
		matches = append(matches, match)
	}

	return matches, nil
}

func fetchAndStoreMatches() error {
	// Fetch matches from the API
	matches, err := fetchMatchesFromAPI()
	if err != nil {
		return fmt.Errorf("error fetching matches from API: %v", err)
	}

	// Loop through each match and store it in the database
	for _, match := range matches {
		// Check if the match already exists in the database to avoid duplicates
		var existingMatch models.Match
		if err := database.DB.Where("home_team = ? AND away_team = ? AND date = ?", match.HomeTeam, match.AwayTeam, match.Date).First(&existingMatch).Error; err == nil {
			fmt.Println("Match already exists, skipping:", match.HomeTeam, "vs", match.AwayTeam)
			continue // Skip if match already exists
		}

		// Store the match in the database
		if err := database.DB.Create(&match).Error; err != nil {
			return fmt.Errorf("error saving match to database: %v", err)
		}
	}

	return nil
}
