package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type ApiCEP struct {
	Code       string `json:"code"`
	State      string `json:"state"`
	City       string `json:"city"`
	District   string `json:"district"`
	Address    string `json:"address"`
	Status     int    `json:"status"`
	Ok         bool   `json:"ok"`
	StatusText string `json:"statusText"`
}

func readCEPApiCEP(cep string, ch chan<- ApiCEP) {
	c := http.Client{}
	url := fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s.json", cep)
	res, err := c.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot request to %s: %v\n", url, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read response body: %v\n", err)
	}

	var respCEP ApiCEP
	err = json.Unmarshal(body, &respCEP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot unmarshal response to struct: %v\n", err)
		return
	}

	ch <- respCEP
}

func readCEPViaCEP(cep string, ch chan<- ViaCEP) {
	c := http.Client{}
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	res, err := c.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot request to %s: %v\n", url, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read response body: %v\n", err)
	}

	var respCEP ViaCEP
	err = json.Unmarshal(body, &respCEP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot unmarshal response to struct: %v\n", err)
		return
	}

	ch <- respCEP
}

func main() {
	cep := strings.Join(os.Args[1:], "")
	if len(cep) == 8 && !strings.Contains(cep, "-") {
		cep = fmt.Sprintf("%s-%s", cep[:5], cep[5:])
	}

	apiCEPChan := make(chan ApiCEP)
	viaCEPChan := make(chan ViaCEP)

	go readCEPApiCEP(cep, apiCEPChan)
	go readCEPViaCEP(cep, viaCEPChan)

	select {
	case msg := <-apiCEPChan:
		fmt.Fprintf(os.Stdout, "===== API CEP =====\nCEP:%s\nEndereço: %s, %s, %s\n", msg.Code, msg.Address, msg.City, msg.State)
	case msg := <-viaCEPChan:
		fmt.Fprintf(os.Stdout, "===== VIA CEP =====\nCEP:%s\nEndereço: %s, %s, %s\n", msg.Cep, msg.Logradouro, msg.Localidade, msg.Uf)
	case <-time.After(time.Second * 1):
		println("TIMEOUT!")
	}
}
