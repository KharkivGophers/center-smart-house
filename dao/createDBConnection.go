package dao



func GetDBDriver(db DbClient) (DbClient) {
	db.RunDBConnection()
	return db
}
//func GetDBDriver(db Server) (DbClient) {
//	client := &DBMock{DbServer: db}
//	client.RunDBConnection()
//	return client
//}
