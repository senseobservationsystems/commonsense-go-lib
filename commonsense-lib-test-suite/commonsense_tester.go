package main

import (
	"commonsense-go-lib/commonsense"
	"fmt"
	"os"
)

func main () {
	if len(os.Args) < 3 {
		fmt.Println("Usage: commonsense_tester <username> <password>")
		return
	}

	username := os.Args[1]
	password := os.Args[2]

	C := commonsense.CommonSenseClient{Debug: true}
	err := C.Login(username, password)
	
	if err != nil {
		fmt.Println("Login failed", err)
	}
	
	sensors, err := C.GetAllSensors()
	if err != nil {
		fmt.Println("GetSensors failed", err)
	}
	fmt.Println(sensors)
	
	s := commonsense.CS_Sensor{Name: "herp", DeviceType: "derp", DisplayName: "herpaderp", DataType: "float", UseDataStorage: true}
	id, err := C.PostSensor(s)
	if err != nil {
		fmt.Println("PostSensor failed", err)
	}
	fmt.Println("Sensor id", id)
	
	d := commonsense.CS_Data{Value: "100", Date: "1355321600"}
	data := commonsense.CS_SensorData{SensorId: id}
	data.Data = append(data.Data, d)
	err = C.PostSensorData([]commonsense.CS_SensorData{data})
	
	err = C.Logout()
	if err != nil {
		fmt.Println("Logout failed", err)
	}
}
