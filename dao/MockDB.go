package dao

import (
	. "github.com/KharkivGophers/center-smart-house/models"

)

type DBMock struct {
	Client   *DbMockClient
	//DbServer Server
}

// Implementation DbDriver-----------------------------------------------------------------
func (dbClien *DBMock) FlushAll() (error) {
	return nil
}
func (dbClien *DBMock) Publish(channel string, message interface{}) (int64, error) {
	return 0, nil
}
func (dbClien *DBMock) Connect() (error) {
	return nil
}
func (dbClien *DBMock) Subscribe(cn chan []string, channel ...string) error {
	return nil
}
func (dbClien *DBMock) Close() (error) {
	return nil
}
func (dbClien *DBMock) RunDBConnection() (error) {
	dbClien.Client = &DbMockClient{}
	return nil

}
func (dbClien *DBMock) GetAllDevices() ([]DevData) {
	return []DevData{}
}
func (dbClien *DBMock) GetDevice(devParamsKey string, devParamsKeysTokens []string) (DevData) {
	return DevData{}
}
func (dbClien *DBMock) GetClient() DbRedisDriver {
	return nil
}

// Implementation DbRedisDriver-----------------------------------------------------------------
type DbMockClient struct {
	Hash map[string][]interface{}
}

func (dbClien *DbMockClient) checkMap() {
	if dbClien.Hash == nil {
		dbClien.Hash = make(map[string][]interface{})
	}
}
func (dbClien *DbMockClient) SAdd(key string, member ...interface{}) (int64, error) {
	return 0, nil
}
func (dbClien *DbMockClient) ZAdd(key string, arguments ...interface{}) (int64, error) {
	return 0, nil
}

func (dbClien *DbMockClient) HMSet(key string, values ...interface{}) (string, error) {
	dbClien.checkMap()
	dbClien.Hash[key] = append(dbClien.Hash[key], values)
	return "", nil
}
func (dbClien *DbMockClient) HMGet(key string, fields ...string) ([]string, error) {

	return nil, nil
}
func (dbClien *DbMockClient) ZRangeByScore(key string, values ...interface{}) ([]string, error) {
	return nil, nil
}
func (dbClien *DbMockClient) SMembers(key string) ([]string, error) {
	return nil, nil
}
func (dbClien *DbMockClient) Close() error {
	return nil
}
func (dbClien *DbMockClient) Connect(host string, port uint) (err error) {
	return nil
}
func (dbClien *DbMockClient) Exists(key string) (bool, error) {
	return false, nil
}
