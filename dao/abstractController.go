package dao

type DbWorker interface {
	FlushAll() (error)
	Publish(channel string, message interface{}) (int64, error)
	Connect()(error)
	Subscribe(cn chan []string, channel ...string) error
	Close() (error)
}
