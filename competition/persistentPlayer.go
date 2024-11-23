package competition

import (
	"github.com/google/uuid"
	"sync"
)

type persistentPlayer struct {
	lock   sync.Mutex
	agents map[string]*persistentAgent
}

func (p *persistentPlayer) findAvailableAgent(game uuid.UUID) *persistentAgent {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, agent := range p.agents {
		if agent.tryNewGame(game) {
			return agent
		}
	}
	panic("todo")
}

type persistentAgent struct {
	agentURL string

	lock             sync.Mutex
	currentlyPlaying *persistentAgentGameInstance
}

type persistentAgentGameInstance struct {
	id uuid.UUID
}

func (p *persistentAgent) tryNewGame(gameID uuid.UUID) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.currentlyPlaying != nil {
		return false
	}

	p.currentlyPlaying = &persistentAgentGameInstance{id: gameID}
	return true
}
