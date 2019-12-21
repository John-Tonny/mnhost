package subscriber

import (
	"context"
	"github.com/micro/go-micro/util/log"

	monitor "github.com/John-Tonny/mnhost/monitor/proto/monitor"
)

type Monitor struct{}

func (e *Monitor) Handle(ctx context.Context, msg *monitor.Message) error {
	log.Log("Handler Received message: ", msg.Say)
	return nil
}

func Handler(ctx context.Context, msg *monitor.Message) error {
	log.Log("Function Received message: ", msg.Say)
	return nil
}
