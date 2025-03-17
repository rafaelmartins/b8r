package utils

import (
	"errors"
	"time"

	"rafaelmartins.com/p/octokeyz"
)

func IgnoreDisplayMissing(err error) error {
	if errors.Is(err, octokeyz.ErrDeviceDisplayNotSupported) {
		return nil
	}
	return err
}

func LedFlash3Times(dev *octokeyz.Device) error {
	for i := 0; i < 3; i++ {
		if err := dev.Led(octokeyz.LedFlash); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
