package subscriber

import (
	"context"

	"github.com/micro/go-micro/util/log"

	vps "github.com/John-Tonny/mnhost/interface/out/vps"
)

type Vps struct{}

func (e *Vps) Handle(ctx context.Context, msg *vps.Message) error {
	log.Log("Handler Received message: ", msg.Say)
	return nil
}

func Handler(ctx context.Context, msg *vps.Message) error {
	log.Log("Function Received message: ", msg.Say)
	return nil
}
