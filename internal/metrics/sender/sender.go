package sender

import (
	"time"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// AllMetricGeter interface for Metric repo
type AllMetricGeter interface {
	GetAllMetrics() ([][]string, [][]string, error)
}

// MetricSender interface for sender
type MetricSender interface {
	Send(typ string, name string, value string) error
}

// Start запускает процесс отправки метрик раз в reportInterval секунд
func Start(stor AllMetricGeter, sender MetricSender, reportInterval int64) {
	for {

		gauge, counter, err := stor.GetAllMetrics()
		if err != nil {
			//TODO error chanel
			panic("Stor unavailable!")
		}
		for _, metric := range gauge {

			err = sender.Send(format.Gauge, metric[0], metric[1])
			if err != nil {
				//TODO error chanel
				panic(err.Error())

			}
		}
		for _, metric := range counter {

			err = sender.Send(format.Counter, metric[0], metric[1])
			if err != nil {
				//TODO error chanel
				panic(err.Error())
			}
		}
		sleepSecond := time.Duration(reportInterval) * time.Second
		time.Sleep(sleepSecond)
	}
}
