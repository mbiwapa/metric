package sender

import (
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
			panic("Stor unavailable!")
		}
		for _, metric := range gauge {

			err = sender.Send("gauge", metric[0], metric[1])
			if err != nil {
				//TODO ошибку сделать можно ли тут паниковать?
				panic(err.Error())

			}
		}
		for _, metric := range couner {

			err = sender.Send("counter", metric[0], metric[1])
			if err != nil {
				//TODO ошибку сделать
				panic(err.Error())
			}
		}
		sleepSecond := time.Duration(reportInterval) * time.Second
		time.Sleep(sleepSecond)
	}
}
