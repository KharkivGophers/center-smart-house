package dao

type DbWorker interface {
	FlushAll() (error)
	Publish(channel string, message interface{}) (int64, error)
	runDBConnection(host string, port uint) (*DbWorker, error)
	NewClient() (*DbWorker)
	Subscribe(cn chan []string, channel ...string) error

}