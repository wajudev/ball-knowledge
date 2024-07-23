package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Username string    `json:"username" binding:"required"`
	Email    string    `json:"email" binding:"required"`
	Password string    `json:"password" binding:"required"`
}

func (user *User) BeforeCreate(*gorm.DB) (err error) {
	user.ID = uuid.New()
	return
}

type Match struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	HomeTeam string    `json:"home_team" binding:"required"`
	AwayTeam string    `json:"away_team" binding:"required"`
	Date     string    `json:"date" binding:"required"`
	Result   string    `json:"result"`
}

func (match *Match) BeforeCreate(*gorm.DB) (err error) {
	match.ID = uuid.New()
	return
}

type Prediction struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID             uuid.UUID `gorm:"type:uuid" json:"user_id"`
	MatchID            uuid.UUID `gorm:"type:uuid" json:"match_id"`
	PredictedScoreHome int       `json:"predicted_score_home" binding:"required"`
	PredictedScoreAway int       `json:"predicted_score_away" binding:"required"`
	Points             int       `json:"points"`
	User               User      `gorm:"foreignKey:UserID" json:"-"`
	Match              Match     `gorm:"foreignKey:MatchID" json:"-"`
}

func (prediction *Prediction) BeforeCreate(*gorm.DB) (err error) {
	prediction.ID = uuid.New()
	return
}
