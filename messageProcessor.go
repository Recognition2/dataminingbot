package main

func messageProcessor(g global) {
	defer g.wg.Done()
	var totalLength int = 0

outer:
	for {
		select {
		case <-g.shutdown:
			break outer
		case msg := <-g.messages:
			totalLength += len(msg.Text)
		}
		logInfo.Printf("Current count: %d\n", totalLength)
	}

	logInfo.Println("Stopping message processor")
}
