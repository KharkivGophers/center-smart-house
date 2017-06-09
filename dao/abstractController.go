package dao

type dbWorker interface {
	FlushAll() (error)
}