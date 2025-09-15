package controllers

import (
"encoding/json"
"fmt"
"io/ioutil"
"net/http"
"os"
"strconv"
"time"

"ball-knowledge/database"
"ball-knowledge/models"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

// GetMatches fetches matches from API and database
func GetMatches(c *gin.Context) {
	// First, try to fetch and store latest matches from API
	if err := fetchAndStoreMatches(); err != nil {
		fmt.Printf("Warning: Failed to fetch latest matches from API: %v\n", err)
		// Continue with database matches even if API fails
	}

	// Get matches from database
	var matches []models.Match
	if err := database.DB.Order("date ASC").Find(&matches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  matches,
		"count": len(matches),
	})
}

// GetMatchesForGameWeek retrieves matches for a specific gameweek
func GetMatchesForGameWeek(c *gin.Context) {
	gameWeekStr := c.Param("gameweek")
	gameWeek, err := strconv.Atoi(gameWeekStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gameweek number"})
		return
	}

	var matches []models.Match
	if err := database.DB.Where("match_day = ?", gameWeek).Order("date ASC").Find(&matches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     matches,
		"gameweek": gameWeek,
		"count":    len(matches),
	})
}

// CreateMatch manually creates matches (admin function)
func CreateMatch(c *gin.Context) {
	var matches []models.Match
	if err := c.ShouldBindJSON(&matches); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i := range matches {
		matches[i].ID = uuid.New()
		if err := database.DB.Create(&matches[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to create match: %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Matches created successfully",
		"data":    matches,
		"count":   len(matches),
	})
}

// GetMatchDetails gets detailed information about a specific match
func GetMatchDetails(c *gin.Context) {
	matchID := c.Param("id")

	var match models.Match
	if err := database.DB.Where("id = ?", matchID).First(&match).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Get prediction count for this match
	var predictionCount int64
	database.DB.Model(&models.Prediction{}).Where("match_id = ?", matchID).Count(&predictionCount)

	c.JSON(http.StatusOK, gin.H{
		"match":           match,
		"prediction_count": predictionCount,
	})
}

// fetchAndStoreMatches fetches matches from the Football API and stores them
func fetchAndStoreMatches() error {
	matches, err := fetchMatchesFromAPI()
	if err != nil {
		return fmt.Errorf("error fetching matches from API: %v", err)
	}

	successCount := 0
	skipCount := 0

	for _, match := range matches {
		// Check if match already exists
		var existingMatch models.Match
		if err := database.DB.Where("home_team = ? AND away_team = ? AND date = ?",
			match.HomeTeam, match.AwayTeam, match.Date).First(&existingMatch).Error; err == nil {
			skipCount++
			continue
		}

		// Store new match
		if err := database.DB.Create(&match).Error; err != nil {
			fmt.Printf("Error saving match %s vs %s: %v\n", match.HomeTeam, match.AwayTeam, err)
			continue
		}
		successCount++
	}

	fmt.Printf("Matches processed: %d new, %d skipped\n", successCount, skipCount)
	return nil
}

// fetchMatchesFromAPI fetches matches from the Football API
func fetchMatchesFromAPI() ([]models.Match, error) {
	apiKey := os.Getenv("API_FOOTBALL_KEY")
	leagueID := os.Getenv("LEAGUE_ID")
	season := os.Getenv("SEASON")

	if apiKey == "" {
		return nil, fmt.Errorf("API_FOOTBALL_KEY environment variable not set")
	}

	if leagueID == "" {
		leagueID = "39" // Default to Premier League
	}

	if season == "" {
		season = "2024" // Default to current season
	}

	// Build API URL
	url := fmt.Sprintf("https://v3.football.api-sports.io/fixtures?league=%s&season=%s", leagueID, season)

	// Create HTTP client and request
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("x-apisports-key", apiKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse JSON response
	var apiResponse struct {
		Response []struct {
			Fixture struct {
				ID   int    `json:"id"`
				Date string `json:"date"`
			} `json:"fixture"`
			League struct {
				Round string `json:"round"`
			} `json:"league"`
			Teams struct {
				Home struct {
					Name string `json:"name"`
				} `json:"home"`
				Away struct {
					Name string `json:"name"`
				} `json:"away"`
			} `json:"teams"`
			Score struct {
				Fulltime struct {
					Home *int `json:"home"`
					Away *int `json:"away"`
				} `json:"fulltime"`
			} `json:"score"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	var matches []models.Match

	for _, fixture := range apiResponse.Response {
		// Parse date
		date, err := time.Parse(time.RFC3339, fixture.Fixture.Date)
		if err != nil {
			fmt.Printf("Error parsing date %s: %v\n", fixture.Fixture.Date, err)
			continue
		}

		// Extract match day from round string
		var matchDay int
		fmt.Sscanf(fixture.League.Round, "Regular Season - %d", &matchDay)
		if matchDay == 0 {
			// Try other formats
			fmt.Sscanf(fixture.League.Round, "Matchday %d", &matchDay)
		}

		// Build result string
		result := "0:0" // Default
		if fixture.Score.Fulltime.Home != nil && fixture.Score.Fulltime.Away != nil {
			result = fmt.Sprintf("%d:%d", *fixture.Score.Fulltime.Home, *fixture.Score.Fulltime.Away)
		}

		match := models.Match{
			HomeTeam: fixture.Teams.Home.Name,
			AwayTeam: fixture.Teams.Away.Name,
			Date:     date.Format(time.RFC3339),
			League:   "Premier League", // Can be made dynamic
			Season:   season,
			MatchDay: matchDay,
			Result:   result,
		}

		matches = append(matches, match)
	}

	return matches, nil
}