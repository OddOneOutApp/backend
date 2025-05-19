package cleanup

import (
	"time"

	"github.com/OddOneOutApp/backend/internal/services"
	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/OddOneOutApp/backend/internal/websocket"
	"gorm.io/datatypes"
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
				if !game.AnswersFinished {
					game.SetAnswersFinished(db, true)
					err := db.Save(&game).Error
					if err != nil {
						utils.Logger.Errorf("Error saving game: %s", err)
						continue
					}
					time.Sleep(1 * time.Second)
					processGameEnd(db, &game)
					utils.Logger.Infof("Game %s answers finished", game.ID)
				}
			}
			db.Where("voting_end_time <= ?", now).Find(&games)
			for _, game := range games {
				if !game.VotingFinished {
					game.SetVotingFinished(db, true)
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
					votesMap := make(map[datatypes.UUID]datatypes.UUID)
					for _, vote := range votes {
						votesMap[vote.UserID] = vote.ID
					}

					websocket.SendVoteResultMessage(game.ID, votesMap)
					utils.Logger.Infof("Game %s voting finished", game.ID)
				}
			}
		}
	}()
}

func processGameEnd(db *gorm.DB, game *services.Game) {
	// send answers to all players
	answers, err := game.GetAnswers(db)
	if err != nil {
		utils.Logger.Errorf("Error fetching answers: %s", err)
		return
	}

	answersMap := make(map[datatypes.UUID]string)
	for _, ans := range answers {
		answersMap[ans.UserID] = ans.Answer
	}
	websocket.SendAnswersMessage(game.ID, answersMap, game.RegularQuestion)
	game.SetVotingEndTime(db, game.AnswersEndTime.Add(30*time.Second))
	game.SetVotingFinished(db, false)

}
