package commonsense

import (
	"fmt"
	"os"
	"io/ioutil"
	"net/http"
	"strings"
	"errors"
	"encoding/json"
	"net/url"
	"github.com/senseobservationsystems/marpessa/marpessa/v2adapter"
)

type CS_Credentials struct {
	Username 			string		`json:"username"`
	Password			string		`json:"password"`
}

type CS_Sensor struct {
	Id					string			`json:"id,omitempty"`
	Name				string			`json:"name"`
	Type				string			`json:"type,omitempty"`
	DeviceType			string			`json:"device_type"`
	DisplayName 		string			`json:"display_name"`
	UseDataStorage		bool			`json:"use_data_storage"`
	DataType			string			`json:"data_type"`
	DataStructure		string			`json:"data_structure,omitempty"`
}

type CS_SensorMetatags struct {
	Id					string			`json:"id,omitempty"`
	Name				string			`json:"name"`
	Type				string			`json:"type,omitempty"`
	DeviceType			string			`json:"device_type"`
	DisplayName 		string			`json:"display_name"`
	UseDataStorage		string			`json:"use_data_storage"`
	DataType			string			`json:"data_type"`
	DataStructure		string			`json:"data_structure,omitempty"`
	Metatags			interface{} 	`json:"metatags,omitempty"`
}

type CS_Data struct {
	Sensor_id 			int 					`json:"sensor_id,omitempty"`
	Value				string      			`json:"value"`
	Date				v2adapter.FloatDate     `json:"date,omitempty"`
}

type CS_SensorData struct {
	SensorId			string		`json:"sensor_id,omitempty"`
	Data				[]CS_Data	`json:"data"`
	Total 				int 		`json:"total,omitempty"`
}

type CS_Data_Wrapper struct {
	Data		[]CS_Data		`json:"data"`
	Total		int				`json:"total,omitempty"`
}

type CS_SensorData_Wrapper struct {
	Sensors		[]CS_SensorData `json:"sensors"`
}

type CS_Sensor_Wrapper struct {
	Sensor		CS_Sensor		`json:"sensor"`
}

type CS_Sensors_Metatag_Wrapper struct {
	Sensors		[]CS_SensorMetatags 	`json:"sensors"`
	Total		int						`json:"total"`
}

type CS_Sensors_Wrapper struct {
	Sensors		[]CS_Sensor		`json:"sensors"`
	Total		int				`json:"total"`
}

type CommonSenseClient struct {
	client			http.Client
	session_id 		string
	Debug			bool
}

func NewCommonSenseClient() (*CommonSenseClient) {
	return &CommonSenseClient{}
}

func (C *CommonSenseClient) apiCall (method, url, body string) (r_headers http.Header, r_body []byte, err error) {

	full_url := fmt.Sprintf("http://api.sense-os.nl%s", url)

	req, err := http.NewRequest(method, full_url, strings.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	if !strings.EqualFold("/login.json", url) {
		req.Header.Add("X-SESSION_ID", C.session_id)
	}
	req.Header.Add("Accept", "*")
	req.Header.Add("Content-Type", "application/json")

	if C.Debug {
		fmt.Println("\n===================")
		fmt.Printf("URL: %s\n", full_url)
		fmt.Printf("METHOD: %s\n", method)
		fmt.Printf("HEADER: %v\n", req.Header)
		fmt.Printf("BODY: %s\n", body)
		fmt.Println("===================")
	}

	resp, err := C.client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()
	r_body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if !(strings.Contains(resp.Status, "200") || strings.Contains(resp.Status, "201")) {
		err = errors.New(fmt.Sprintf("Call failed: %s, %s", resp.Status, r_body))
		if strings.Contains(resp.Status, "500") {
			f, err := os.OpenFile("recess_diagnostics.html", os.O_WRONLY | os.O_CREATE, 0)
			if err != nil {
				fmt.Println("Cant open 500 file!", err)
			} else {
				f.Write(r_body)
				f.Close()
			}
		}

		return nil, nil, err
	}

	if C.Debug {
		fmt.Println("===================")
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Headers: %v\n", resp.Header)
		fmt.Printf("Body: %s\n", r_body)
		fmt.Println("===================\n")
	}

	return resp.Header, r_body, nil
}

func (C *CommonSenseClient) Login (username, password string) (err error) {
	cred 		:= CS_Credentials{Username: username, Password: password}
	data, err 	:= json.Marshal(cred)

	if err != nil {
		return err
	}

	h, _, err := C.apiCall("POST", "/login.json", string(data))
	if err != nil {
		return err
	}

	C.session_id = h.Get("X-SESSION_ID")

	return nil
}

func (C *CommonSenseClient) Logout () (err error) {
	_, _, err = C.apiCall("POST", "/logout.json", "")

	if err != nil {
		return err
	}

	C.session_id = ""

	return nil
}

func (C *CommonSenseClient) GetSensors () (sensors []CS_Sensor, err error) {
	_, b, err := C.apiCall("GET", "/sensors.json?page=0&per_page=1000&shared=0&owned=1&physical=1&details=full", "")

	if err != nil {
		return nil, err
	}

	v := CS_Sensors_Wrapper{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}
	return v.Sensors, nil
}

func (C *CommonSenseClient) GetAllSensors () (sensors []CS_Sensor, err error) {
	i := 0

	for {
		_, b, err := C.apiCall("GET", fmt.Sprintf("/sensors.json?page=%d&per_page=100&shared=0&owned=1&physical=0&details=full", i), "")
		if err != nil {
			break
		}

		v := CS_Sensors_Wrapper{}
		err = json.Unmarshal(b, &v)
		if err != nil {
			break
		}

		sensors = append(sensors, v.Sensors...)
		if len(v.Sensors) < 100 {
			break
		}
		i++
	}

	return sensors, nil
}

func (C *CommonSenseClient) GetSensorsMetatags (namespace string) (sensors []CS_SensorMetatags, err error) {
	_, b, err := C.apiCall("GET", fmt.Sprintf("/sensors/metatags.json?namespace=%s&details=full", namespace), "")

	if err != nil {
		return nil, err
	}

	v := CS_Sensors_Metatag_Wrapper{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}

	return v.Sensors, nil
}

func (C *CommonSenseClient) PostSensor (s CS_Sensor) (id string, err error) {
	v := CS_Sensor_Wrapper{s}
	data, err := json.Marshal(v)
	if err != nil {
		return "0", err
	}

	h, _, err := C.apiCall("POST", "/sensors.json", string(data))
	if err != nil {
		return "0", err
	}

	loc := h.Get("Location")
	_, err = fmt.Sscanf(loc, "http://api.sense-os.nl/sensors/%s", &id)
	if err != nil {
		return "0", err
	}

	return id, nil
}

func (C *CommonSenseClient) PutSensor (sensor_id string, s CS_Sensor) (err error) {
	v := CS_Sensor_Wrapper{s}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, _, err = C.apiCall("PUT", fmt.Sprintf("/sensors/%s.json", sensor_id), string(data))
	if err != nil {
		return err
	}

	return nil
}

func (C *CommonSenseClient) DeleteSensor (sensor_id string) (err error) {
	
	_, _, err = C.apiCall("DELETE", fmt.Sprintf("/sensors/%s.json", sensor_id), "")
	if err != nil {
		return err
	}

	return nil
}

func (C *CommonSenseClient) PostSensorData(sensor_id string, d CS_SensorData) (err error) {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, _, err = C.apiCall("POST", fmt.Sprintf("/sensors/%s/data", sensor_id), string(data))
	if err != nil {
		return err
	}

	return nil
}

func (C *CommonSenseClient) PostSensorsData(d []CS_SensorData) (err error) {
	v := CS_SensorData_Wrapper{d}
	data, err := json.Marshal(&v)
	if err != nil {
		return err
	}

//	fmt.Printf("%s\n", data)

	_, _, err = C.apiCall("POST", "/sensors/data.json", string(data))
	if err != nil {
		return err
	}

	return nil
}

func (C *CommonSenseClient) GetSensorData(sensor_id string, parameters map[string]interface{}) ([]CS_Data, error) {

    par := urlEncode(parameters)

    _, b, err := C.apiCall("GET", fmt.Sprintf("/sensors/%s/data.json?%s", sensor_id, par.Encode()), "")

	if err != nil {
		return nil, err
	}

	v := CS_Data_Wrapper{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}

	return v.Data, nil
}

func (C *CommonSenseClient) GetSensorsData(sensor_ids []string, parameters map[string]interface{}) ([]CS_Data, error) {
	par := urlEncode(parameters)
	for _, s := range sensor_ids {
		par.Add("sensor_id[]", s)
	}

	_, b, err := C.apiCall("GET", fmt.Sprintf("/sensors/data.json?%s", par.Encode()), "")

	if err != nil {
		return nil, err
	}

	v := CS_Data_Wrapper{}
	if err = json.Unmarshal(b, &v); err != nil {
		return nil, err
	}

	return v.Data, nil
}

func urlEncode(m map[string]interface{}) (url.Values) {
	parameters := url.Values{}
	for k, v := range m {
		parameters.Add(k, fmt.Sprint(v))
	}
	return parameters
}
