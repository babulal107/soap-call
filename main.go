package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

//SOAP Request
type soapRequest struct {
	XMLName   xml.Name `xml:"soap:Envelope"`
	XMLNsSoap string   `xml:"xmlns:soap,attr"`
	XMLNsXSI  string   `xml:"xmlns:xsi,attr"`
	XMLNsXSD  string   `xml:"xmlns:xsd,attr"`
	Body      soapReqBody
}
type soapReqBody struct {
	XMLName xml.Name `xml:"soap:Body"`
	Payload interface{}
}

// Soap Response
type soapResponse struct {
	XMLName xml.Name
	Body    soapResBody
}
type soapResBody struct {
	XMLName            xml.Name
	GetIecDataResponse GetIecDataResponseData `xml:"GetIecDataResponse"`
}
type GetIecDataResponseData struct {
	XMLName          xml.Name         `xml:"GetIecDataResponse"`
	XMLNS            string           `xml:"xmlns,attr"`
	GetIecDataResult string `xml:"GetIecDataResult"`
}
//type GetIecDataResult struct {
//	XMLName    xml.Name   `xml:"GetIecDataResult"`
//	IECRequest IECRequest `xml:"IECRequest"`
//}

type IECRequest struct {
	XMLName     xml.Name    `xml:"IECRequest"`
	Transaction Transaction `xml:"Transaction"`
}

type Transaction struct {
	XMLName           xml.Name `xml:"Transaction"`
	RequestID         string   `xml:"Request_ID"`
	ChallanCode       string   `xml:"CHALLAN_CODE"`
	ChallnaNo         string   `xml:"CHALLAN_NO"`
	ClientAccountNo   string   `xml:"Client_AccountNo"`
	ClientName        string   `xml:"Client_Name"`
	Amount            string   `xml:"Amount"`
	RemitterName      string   `xml:"Remitter_Name"`
	RemitterAccountNo string   `xml:"Remitter_AccountNo"`
	RemitterIFSC      string   `xml:"Remitter_IFSC"`
	RemitterBank      string   `xml:"Remitter_Bank"`
	RemitterBranch    string   `xml:"Remitter_Branch"`
	RemitterUTR       string   `xml:"Remitter_UTR"`
	PayMethod         string   `xml:"Pay_Method"`
	CreditAccountNo   string   `xml:"Credit_AccountNo"`
	InwardRefNum      string   `xml:"Inward_Ref_Num"`
	CreditTime        string   `xml:"Credit_Time"`
	Reserve1          string   `xml:"Reserve1"`
	Reserve2          string   `xml:"Reserve2"`
	Reserve3          string   `xml:"Reserve3"`
	Reserve4          string   `xml:"Reserve4"`
	ResponseCode      string   `xml:"ResponseCode"`
	ResponseDesc      string   `xml:"ResponseDesc"`
}

//soapCall : call soap services
func soapCall(ws string, action string, payloadInterface interface{}) ([]byte, error) {
	v := soapRequest{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsXSD:  "http://www.w3.org/2001/XMLSchema",
		XMLNsXSI:  "http://www.w3.org/2001/XMLSchema-instance",
		Body: soapReqBody{
			Payload: payloadInterface,
		},
	}

	payload, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", ws, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/xml, multipart/related")
	req.Header.Set("SOAPAction", action)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%q\n", dump)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return bodyBytes, nil
}

func main() {

	result, err := GetIecData("ttt")
	if err != nil {
		log.Println(err)
	}
	fmt.Println("===============Result=============")
	fmt.Println(result)

	var resData IECRequest
	resxml := []byte(result.Body.GetIecDataResponse.GetIecDataResult)
	err = xml.Unmarshal(resxml, &resData)

	fmt.Println("============ IEC Data =============")
	fmt.Println(resData.Transaction.ResponseCode)
}

func GetIecData(customerTenderId string) (iecDataResponse soapResponse, err error) {

	var (
		response []byte
	)

	// prepare request payload
	type GetIecDataRequest struct {
		XMLName          xml.Name `xml:"GetIecData"`
		XMLNS            string   `xml:"xmlns,attr"`
		CustomerTenderID string   `xml:"CustomerTenderId"`
	}

	payload := GetIecDataRequest{
		XMLNS:            "http://tempuri.org/",
		CustomerTenderID: customerTenderId,
	}

	// Call SOAP service
	response, err = soapCall(
		"https://ibluatapig.indusind.com/app/uat/IBLeTender",
		"http://tempuri.org/IIBLeTender/GetIecData",
		payload,
	)
	if err != nil {
		log.Println("Failed to call SOAP : GetIecData : ", err)
	}

	//Convert SOAP response to go response
	err = xml.Unmarshal(response, &iecDataResponse)
	if err != nil {
		log.Println("Failed to do unmarshal : GetIecData : ", err)
	}

	return
}
