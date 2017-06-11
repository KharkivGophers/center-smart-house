package dao

type DbWorker interface {
	FlushAll() (error)
}