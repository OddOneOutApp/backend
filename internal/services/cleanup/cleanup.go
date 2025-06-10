package cleanup

import (
	"time"

	"github.com/OddOneOutApp/backend/internal/services"
	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/OddOneOutApp/backend/internal/websocket"
	"gorm.io/gorm"
)

func StartEndScheduler(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			var games []services.Game
			now := time.Now()
			db.Where("answers_end_time <= ?", now).Find(&games)
			for _, game := range games {
				if game.AnswersEndTime.IsZero() {
					//utils.Logger.Warnf("Game %s has zero answers end time, skipping", game.ID)
					continue
				}
				if game.State == services.GameStateAnswering {
					game.SetVotingEndTimeAndGameState(db, now.Add(30*time.Second))

					time.Sleep(1 * time.Second)
					answers, err := game.GetAnswers(db)
					if err != nil {
						utils.Logger.Errorf("Error fetching answers: %s", err)
						return
					}

					websocket.SendAnswersMessage(game.ID, answers, game.RegularQuestion, game.VotingEndTime)
					utils.Logger.Infof("Game %s answers finished", game.ID)
				}
			}
			db.Where("voting_end_time <= ?", now).Find(&games)
			for _, game := range games {
				if game.VotingEndTime.IsZero() {
					//utils.Logger.Warnf("Game %s has zero voting end time, skipping", game.ID)
					continue
				}
				if game.State == services.GameStateVoting {
					game.State = services.GameStateFinished
					err := db.Save(&game).Error
					if err != nil {
						utils.Logger.Errorf("Error saving game: %s", err)
						continue
					}
					time.Sleep(1 * time.Second)

					votes, err := game.GetVoteResults(db)
					if err != nil {
						utils.Logger.Errorf("Error fetching votes: %s", err)
						continue
					}

					websocket.SendVoteResultMessage(game.ID, votes)
					utils.Logger.Infof("Game %s voting finished", game.ID)
				}
			}
		}
	}()
}
