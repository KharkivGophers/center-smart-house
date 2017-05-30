package main

func main() {
	wg.Add(4)

	//db connection
	dbClient = runDBConnection()
	defer dbClient.Close()

	//http connection with browser
	go runDynamicServer()

	//websocket server
	go websocketServer()

	//-----TCP-Config
	go runConfigServer(configConnType, configHost, configPort)
	//-----TCP
	go runTCPServer()

	wg.Wait()
}
