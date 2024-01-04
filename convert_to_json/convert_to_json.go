package convert_to_json

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

type User struct {
	Email       string `xml:"email"`
	FirstName   string `xml:"first_name"`
	LastName    string `xml:"last_name"`
	Address     string `xml:"address"`
	PostalCode  string `xml:"postal_code"`
	Phone       string `xml:"phone"`
	Ami         string `xml:"ami"`
	TaxID       string `xml:"taxid"`
	CompanyName string `xml:"company_name"`
}

type Consultant struct {
	Email string `xml:"email"`
	Name  string `xml:"name"`
	Phone string `xml:"phone"`
	Photo string `xml:"photo"`
}

type Advert struct {
	ExternalID       string      `xml:"external_id"`
	Email            string      `xml:"email"`
	PostalCode       string      `xml:"postal_code"`
	Category         string      `xml:"category"`
	OfferType        string      `xml:"offer_type"`
	Title            string      `xml:"title"`
	Price            string      `xml:"price"`
	Area             string      `xml:"area"`
	AreaGround       string      `xml:"area_ground"`
	Size             string      `xml:"size"`
	Images           []string    `xml:"images>image"`
	MovieURL         string      `xml:"movie_url"`
	ReferenceID      string      `xml:"reference_id"`
	Description      string      `xml:"description"`
	ConsultantEmail  string      `xml:"consultant_email"`
	Year             string      `xml:"year"`
	NumOfUserLicence string      `xml:"number_of_user_license"`
	Market           string      `xml:"market"`
	Attributes       []Attribute `xml:"attributes>attribute"`
}

type Attribute struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type Data struct {
	XMLName     xml.Name     `xml:"data"`
	User        User         `xml:"user"`
	Consultants []Consultant `xml:"consultant"` // Slice
	Adverts     []Advert     `xml:"advert"`
}

func ConvertXMLToJSON(xmlContent []byte) ([]byte, error) {
	var data Data
	err := xml.Unmarshal(xmlContent, &data)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling XML: %v", err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling JSON: %v", err)
	}

	return jsonData, nil
}
