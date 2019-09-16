package idgen

import (
	"github.com/sony/sonyflake"
	"time"
)

var Gen = New()

type Generator interface {
	NextID() (uint64, error)
}

func New() Generator {
	return sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2019, 9, 15, 0, 0, 0, 0, time.UTC),
		MachineID: func() (u uint16, e error) {
			// TODO
			u = 0
			return
		},
		CheckMachineID: nil,
	})
}
