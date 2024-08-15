package proxy

import "time"

func Retry(fn func() error, attempts int, timeSleep time.Duration) (err error) {

	for attempts > 0 {

		if err = fn(); err != nil {

			attempts--
			time.Sleep(timeSleep)

			continue
		}

		return nil

	}

	return err

}
