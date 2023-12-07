package sender

import (
	"fmt"
	"time"
)

// MetricGeter interface for Metric repo
type MetricGeter interface {
	GetAllMetrics() ([][]string, [][]string, error)
}

// MetricSender interface for sender
type MetricSender interface {
	Send(typ string, name string, value string) error
}

// Start запускает процесс отправки метрик раз в reportInterval секунд
func Start(stor MetricGeter, sender MetricSender, reportInterval int64) {
	for {

		gauge, couner, err := stor.GetAllMetrics()
		if err != nil {
			// TODO можно ли тут паниковать?
			fmt.Errorf("Stor unavailable!")
			return
		}
		for _, metric := range gauge {

			err = sender.Send("gauge", metric[0], metric[1])
			if err != nil {
				//TODO ошибку сделать
				fmt.Errorf("Не отправилось почему-то")
				// continue
			}
		}
		for _, metric := range couner {

			err = sender.Send("counter", metric[0], metric[1])
			if err != nil {
				//TODO ошибку сделать
				fmt.Errorf("Не отправилось почему-то")
				// continue
			}
		}
		sleepSeconds := time.Duration(reportInterval) * time.Second
		time.Sleep(sleepSeconds)
	}
}
