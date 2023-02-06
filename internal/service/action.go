package service

import (
	"fmt"
	"time"

	"github.com/KKopilka/discord-bot/internal/misc"
)

type Action func(goBot *Service) error

type ActionTask struct {
	act    Action
	ticker *time.Ticker
	done   chan bool
	// dc     chan bool
}

func NewActionTask(act Action, d time.Duration) *ActionTask {
	return &ActionTask{
		act:    act,
		ticker: time.NewTicker(d), // make an action ticker
		done:   make(chan bool),   // make a done channel
	}
}

func (at *ActionTask) ActFuncName() string {
	return misc.GetFunctionName(at.act)
}

func (at *ActionTask) Run(goBot *Service) {
	// run go-routine
	go func() {
		for {
			select {
			case <-at.done:
				// stop routine
				return
			case t := <-at.ticker.C:
				// run action
				if err := at.act(goBot); err != nil {
					fmt.Println("tick:", t, "err:", err.Error())
					return
				}
			}
		}
	}()
}

func (at *ActionTask) Stop() {
	at.ticker.Stop()
	at.done <- true
}
