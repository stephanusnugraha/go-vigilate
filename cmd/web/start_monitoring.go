package main

type job struct {
	HostServiceID int
}

func (j job) Run() {

}

func startMonitoring() {
	if preferenceMap["monitoring_live"] == "1" {
		data := make(map[string]string)
		data["message"] = "starting"

		// TODO trigger a message to broadcast to all clients that app is starting to monitoring

		// get of the services that we want to monitor

		// range through the services

		// get the schedule unit and number

		// create a job

		// save the id of the job, so we can start/stop it

		// broadcast over websockets the fact that the service is scheduled

		// range
	}
}
