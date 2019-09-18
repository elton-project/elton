package idgen

import (
	"fmt"
	"github.com/sony/sonyflake"
	"time"
)

var Gen = New()

type Generator interface {
	NextID() (uint64, error)
	NextStringID() (string, error)
}

func New() Generator {
	return NewSonyFlake()
}

func NewSonyFlake() Generator {
	return &sonyFlake{
		Sonyflake: sonyflake.NewSonyflake(sonyflake.Settings{
			StartTime: time.Date(2019, 9, 15, 0, 0, 0, 0, time.UTC),
			MachineID: func() (u uint16, e error) {
				// TODO
				u = 0
				return
			},
			CheckMachineID: nil,
		}),
	}
}

type sonyFlake struct {
	*sonyflake.Sonyflake
}

func (sf *sonyFlake) NextStringID() (string, error) {
	id, err := sf.NextID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", id), nil
}
