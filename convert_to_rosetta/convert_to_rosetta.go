package convert_to_rosetta

import (
	"fmt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

const (
	SITEURN               = "urn:site:imovirtualcom"
	CHARACTERISTICS_URN   = "urn:concept:characteristics"
	STATE_URN             = "urn:concept:state"
	CONSTRUCTION_YEAR_URN = "urn:concept:construction-year"
	BATHROOM_NUM_URN      = "urn:concept:number-of-bathrooms"
	GROSS_AREA_URN        = "urn:concept:gross-area-m2"
	CERTIFICATE_URN       = "urn:concept:energy_certificate"
	ROOMS_NUM_URN         = "urn:concept:number-of-rooms"

	STATE_OLD             = "condicao"
	CONSTRUCTION_YEAR_OLD = "ano_de_construcao"
	BATHROOM_NUM_OLD      = "casas_de_banho"
	GROSS_AREA_OLD        = "area_bruta_(m)"
	CERTIFICATE_OLD       = "certificado_energetico"
	CHARACTERISTICS_OLD   = "caracteristicas"
)

func ConvertJSONToRosetta(fullData map[string]interface{}, w http.ResponseWriter) (string, error) {

	for key, value := range fullData {
		fmt.Printf("Key: %s, Value: %v\n", key, value)
	}

	xmlData := ""

	// Create root <data> element
	xmlData += "<data>"

	// Create <header> element
	xmlData += "<header>"

	// Convert OwnerEmail
	ownerEmail := ConvertOwnerEmail(fullData)

	// Add <owner_email> to header element
	xmlData += "<owner_email>" + ownerEmail + "</owner_email>"

	// Add <site_urn> to header element
	xmlData += "<site_urn>" + SITEURN + "</site_urn>"

	// Close <header> element
	xmlData += "</header>"

	// Open <adverts> element
	xmlData += "<adverts>"

	logToFile("\n\n--------------------------------------------------------------------", "-")

	// Iterate through the JSON array and create an <advert> for each item
	adverts := fullData["Adverts"].([]interface{})
	for _, data := range adverts {
		advert := data.(map[string]interface{})

		// Create the <advert>
		xmlData += "<advert>"

		// Get title:
		title := advert["Title"].(string)

		// Create <title> element to XML with CDATA
		xmlData += "<title><![CDATA[" + title + "]]></title>"

		// Get description
		description := advert["Description"].(string)

		// Create <description> element to XML with CDATA
		xmlData += "<description><![CDATA[" + description + "]]></description>"

		// Convert category
		category := MapCategoryURN(advert["OfferType"].(string), advert["Category"].(string))

		// Create <category_urn> element to XML
		xmlData += "<category_urn><![CDATA[" + category + "]]></category_urn>"

		// Check if ConsultantEmail exists:
		consultant := MapConsulterContact(advert["ConsultantEmail"].(string), fullData)

		// Create <contact> element to XML case <consultant_email> exists with CDATA
		if len(consultant) > 0 {
			xmlData += "<consultant>"
			xmlData += "<email><![CDATA[" + consultant["email"] + "]]></email>"
			xmlData += "<name><![CDATA[" + consultant["name"] + "]]></name>"
			xmlData += "<phone><![CDATA[" + consultant["phone"] + "]]></phone>"
			xmlData += "<photo><![CDATA[" + consultant["photo"] + "]]></photo>"
			xmlData += "</consultant>"
		}

		// Convert price:
		price := MapPrice(advert["Price"].(string))

		// Create <price> element to XML
		xmlData += "<price>"
		xmlData += "<value>" + price["value"] + "</value>"
		xmlData += "<currency>" + price["currency"] + "</currency>"
		xmlData += "</price>"

		// Convert location:
		location := MapLocation(advert)

		// Create <location> contact element to XML
		xmlData += "<location>"
		xmlData += "<lat>" + location["lat"] + "</lat>"
		xmlData += "<lon>" + location["lon"] + "</lon>"
		xmlData += "<exact>" + location["exact"] + "</exact>"
		xmlData += "</location>"

		if imagesInterface, ok := advert["Images"].([]interface{}); ok {
			imageURLs := MapImages(imagesInterface)
			// Create <images> element
			if len(imageURLs) > 0 {
				xmlData += "<images>"
				for _, imageURL := range imageURLs {
					xmlData += "<image><url><![CDATA[" + imageURL + "]]></url></image>"
				}
				xmlData += "</images>"
			}
		}

		// Check if MovieURL exists:
		if movie, exists := advert["MovieURL"].(string); exists && movie != "" {
			// Create <movie> element to XML with CDATA
			xmlData += "<movie><![CDATA[" + movie + "]]></movie>"
		}

		// Check if NumOfUserLicence exists:
		if numOfLicence, exists := advert["NumOfUserLicence"].(string); exists && numOfLicence != "" && numOfLicence != "Isento" {
			// Create <number_of_user_license> element to XML with CDATA
			xmlData += "<number_of_user_license><![CDATA[" + numOfLicence + "]]></number_of_user_license>"
		}

		// Create <market> element to XML with CDATA
		if market, exists := advert["Market"].(string); exists && market != "" {
			xmlData += "<market><![CDATA[" + market + "]]></market>"
		} else {
			xmlData += "<market>secondary</market>"
		}

		// Create <custom_fields> element
		xmlData += "<custom_fields>"

		// Check if ReferenceID exists:
		/*if referenceId, exists := advert["ReferenceID"].(string); exists && referenceId != "" {
			// Create <reference_id> element to XML with CDATA
			xmlData += "<reference_id><![CDATA[" + referenceId + "]]></reference_id>"
		}

		// Check if ExternalID exists:
		if externalId, exists := advert["ExternalID"].(string); exists && externalId != "" {
			// Create <external_id> element to XML with CDATA
			xmlData += "<external_id><![CDATA[" + externalId + "]]></external_id>"
		}*/

		externalId := advert["ExternalID"].(string)
		xmlData += "<external_id><![CDATA[" + externalId + "]]></external_id>"
		referenceId := advert["ReferenceID"].(string)
		xmlData += "<reference_id><![CDATA[" + referenceId + "]]></reference_id>"

		xmlData += "</custom_fields>"

		// Convert attributes
		prepareAttributes := DefineAllAttributesToArray(advert, " | ExternalID: "+externalId+" | ReferenceId: "+referenceId+" | OwnerEmail: "+ownerEmail+" | ")

		// Create <attributes> element
		if len(prepareAttributes) > 0 {
			xmlData += "<attributes>"
			attributes := AddAllAttributesToXml(prepareAttributes)
			xmlData += attributes
			xmlData += "</attributes>"
		}

		xmlData += "</advert>"
	}

	xmlData += "</adverts>"

	xmlData += "</data>"

	return xmlData, nil
}

//-------------------------------------------------------------------- Add attributes to XML

func AddAllAttributesToXml(dataAttributes map[string]interface{}) string {
	xmlData := ""
	for urn, value := range dataAttributes {
		if values, isSlice := value.([]string); isSlice {
			for _, val := range values {
				xmlData += fmt.Sprintf("<attribute><urn>%s</urn><value>%s</value></attribute>", urn, val)
			}
		} else {
			xmlData += fmt.Sprintf("<attribute><urn>%s</urn><value>%v</value></attribute>", urn, value)
		}
	}
	return xmlData
}

//-------------------------------------------------------------------- Prepare attributes

func DefineAllAttributesToArray(adData map[string]interface{}, extra string) map[string]interface{} {

	dataAttributes := make(map[string]interface{})

	// Define area attribute
	if area, exists := adData["Area"]; exists {
		if areaStr, ok := area.(string); ok {
			if areaStr != "" {
				dataAttributes[GROSS_AREA_URN] = areaStr
			}
		}
	}

	// Define size attribute
	if size, exists := adData["Size"]; exists {
		if sizeStr, ok := size.(string); ok {
			if sizeStr != "" {
				dataAttributes[ROOMS_NUM_URN] = MapSize(sizeStr)
			}
		}
	}

	// Define year attribute
	if year, exists := adData["Year"]; exists {
		if yearStr, ok := year.(string); ok {
			if yearStr != "" {
				dataAttributes[CONSTRUCTION_YEAR_URN] = yearStr
			}
		}
	}

	// Rest of the attributes
	if attributesInterface, exists := adData["Attributes"]; exists {
		if attributesSlice, isSlice := attributesInterface.([]interface{}); isSlice {
			// Iterate through attributes
			for _, attr := range attributesSlice {
				if attribute, isMap := attr.(map[string]interface{}); isMap {
					attrName, nameExists := attribute["Name"].(string)
					attrValue, valueExists := attribute["Value"].(string)
					if nameExists && valueExists {
						characteristicTypes := getCharacteristicAttributesList()
						if mapping, exists := characteristicTypes[SanitizeString(attrName)]; exists {
							switch mapping {
							// Cases with direct input from client to value
							case CONSTRUCTION_YEAR_URN, GROSS_AREA_URN:
								if attrValue != "" {
									dataAttributes[mapping] = attrValue
								}
								break
							// Special case with specific conversion
							case CERTIFICATE_URN:
								conversion := ConvertCertificate(SanitizeString(attrValue))
								if conversion != "" {
									dataAttributes[mapping] = conversion
								}
								break
							// Special case with specific conversion as value and the url is named as 'urn:concept:characteristics'
							case CHARACTERISTICS_URN, STATE_URN, BATHROOM_NUM_URN:

								// Hammer to avoid duplicated value that actually exists on genesis
								if mapping == BATHROOM_NUM_URN && attrValue == "1" {
									attrValue = attrValue + "_bath"
								}

								conversion := Convert(SanitizeString(attrValue), true)
								if conversion != "" {
									if charValues, ok := dataAttributes[mapping].([]string); ok {
										// If there is a slice, append the new conversion
										dataAttributes[mapping] = append(charValues, conversion)
									} else {
										// If it doesn't exist, create a new slice with the conversion
										dataAttributes[mapping] = []string{conversion}
									}
								} else {
									logToFile("Attribute value: '"+attrValue+"' not possible to map for attr type '"+mapping+"'.", extra)
								}
								break
							// Default
							default:
								logToFile("Attr '"+attrName+"' and/or val '"+attrValue+"' not possible to map.", extra)
								if attrValue != "" {
									conversion := Convert(SanitizeString(attrValue), true)
									if conversion != "" {
										dataAttributes[mapping] = conversion
									}
								}
							}
						}
					} else {
						// Handle cases where "Name" or "Value" is missing
						fmt.Println("Attribute missing Name or Value")
					}
				} else {
					// Handle cases where the attribute is not a map
					fmt.Println("Attribute is not a map[string]interface{}")
				}
			}
		} else {
			// Handle cases where "Attributes" is not a slice
			fmt.Println("Attributes is not a []interface{}")
		}
	} else {
		// Here "Attributes" key are missing
		fmt.Println("Attributes key not found in advert")
	}
	return dataAttributes
}

//------------------------------------------------------------ Typology

func MapSize(size interface{}) string {
	// Check if size is an array, return an empty string
	if _, isArray := size.([]interface{}); isArray {
		return ""
	}

	// Map sizes to their respective values
	typology := GetTypologyList()

	// Convert size to string and make it lowercase
	strSize, ok := size.(string)
	if !ok {
		return ""
	}
	lowerSize := strings.ToLower(strSize)

	// If the lowercase size is not in the typology, return 'more'
	if _, exists := typology[lowerSize]; !exists {
		return "more"
	}

	// Return the value from the typology map
	return typology[lowerSize]
}

//------------------------------------------------------------ Images

func MapImages(imagesSlice []interface{}) []string {
	var imageList []string

	for _, image := range imagesSlice {
		if imageURL, ok := image.(string); ok {
			imageList = append(imageList, imageURL)
		}
	}

	return imageList
}

//------------------------------------------------------------- Price

func MapLocation(advert map[string]interface{}) map[string]string {
	locationData := make(map[string]string)

	// Assuming the latitude and longitude are retrieved from the advert map
	// For demonstration purposes, I'm setting them to "0" here
	locationData["lat"] = "0"
	locationData["lon"] = "0"

	exact, exactExists := advert["Exact"].(string)
	if exactExists {
		locationData["exact"] = exact
	} else {
		locationData["exact"] = "false"
	}

	return locationData
}

//------------------------------------------------------------- Price

func MapPrice(price string) map[string]string {
	priceData := make(map[string]string)

	priceData["value"] = price
	priceData["currency"] = "EUR"

	return priceData
}

//-------------------------------------------------------------- Consultant

func MapConsulterContact(consultantEmail string, fullData map[string]interface{}) map[string]string {
	consultantData := make(map[string]string)

	if consultantsData, ok := fullData["Consultants"].([]interface{}); ok {
		// Iterate consultant list
		for _, data := range consultantsData {
			consultant := data.(map[string]interface{})
			email := consultant["Email"].(string)

			if email == consultantEmail {
				// Found person
				consultantData["email"] = email
				consultantData["name"] = consultant["Name"].(string)
				consultantData["phone"] = consultant["Phone"].(string)
				consultantData["photo"] = consultant["Photo"].(string)
				break
			}
		}
	}

	return consultantData
}

//-------------------------------------------------------------- Owner

func ConvertOwnerEmail(fullData map[string]interface{}) string {

	// Access the "User" key inside fullData
	userMap, userExists := fullData["User"].(map[string]interface{})
	if !userExists {
		fmt.Println("Key 'User' not found or is not a map")
	}

	// Access the "Email" field inside the "User" map.
	userEmail, emailExists := userMap["Email"].(string)
	if !emailExists || len(userEmail) == 0 {
		fmt.Println("Key 'Email' not found or is not a string or is empty")
	} else {
		// Case exist exit
		return userEmail
	}

	// Check if the "Adverts" key exists and is an array of maps
	adverts, advertsExists := fullData["Adverts"].([]interface{})
	if !advertsExists {
		fmt.Println("Key 'Adverts' not found or is not an array")
	}

	// Check if there is at least one element in the "Adverts" array
	if len(adverts) == 0 {
		fmt.Println("Array 'Adverts' is empty")
	}

	// Access the first element of the "Adverts" array
	firstAdvert, isFirstAdvertMap := adverts[0].(map[string]interface{})
	if !isFirstAdvertMap {
		fmt.Println("First element of 'Adverts' is not a map")
	}

	// Check if the "Email" key exists in the first element of "Adverts"
	email, emailExists := firstAdvert["Email"].(string)
	if !emailExists {
		fmt.Println("Key 'Email' not found or is not a string")
	}

	return email
}

//------------------------------------------------------------------- Categories

func MapCategoryURN(offerType, category string) string {
	offerType = SanitizeString(offerType)
	category = SanitizeString(category)
	categoryMap := GetCategoryList()

	if category == "" {
		if offerType == "venda" {
			return "urn:concept:realestate-for-sale"
		}
		return "urn:concept:realestate-for-rent"
	}

	if val, ok := categoryMap[offerType][category]; ok {
		return val
	}
	return ""
}

//------------------------------------------------------------------- Helpers

// ConvertCertificate gets the certificate mapping
func ConvertCertificate(certificate string) string {
	// If certificate is not a valid string, return nil
	if certificate == "" {
		return ""
	}

	// Convert to lowercase for comparison
	lowerCertificate := strings.ToLower(certificate)

	// Map 'b-' to 'bminus' and 'a-' to 'aplus'
	if lowerCertificate == "b-" {
		return "bminus"
	} else if lowerCertificate == "a-" {
		return "aplus"
	}

	// If not 'b-' or 'a-', perform the conversion using the convert() function
	// Assuming convert() is a function that handles the other cases
	return Convert(certificate, true)
}

func getCharacteristicAttributesList() map[string]string {
	return map[string]string{
		CHARACTERISTICS_OLD:   CHARACTERISTICS_URN,
		STATE_OLD:             STATE_URN,
		CONSTRUCTION_YEAR_OLD: CONSTRUCTION_YEAR_URN,
		BATHROOM_NUM_OLD:      BATHROOM_NUM_URN,
		GROSS_AREA_OLD:        GROSS_AREA_URN,
		CERTIFICATE_OLD:       CERTIFICATE_URN,
	}
}

func Convert(param string, invert bool) string {
	// Define the URN values
	aux := urnValues

	if invert {
		// Invert the keys and values of the map
		invertedURNValues := make(map[string]string)
		for key, value := range aux {
			invertedURNValues[value] = key
		}
		aux = invertedURNValues
	}

	// Function to sanitize the string
	sanitizeString := func(s string) string {
		return strings.TrimSpace(strings.ToLower(s))
	}

	// Returns the value corresponding to the sanitized param string
	return aux[sanitizeString(param)]
}

// GetTypologyList gets the typology mapping
func GetTypologyList() map[string]string {
	return map[string]string{
		"zero": "zero",
		"mais": "more",
		"t0":   "0",
		"t1":   "1",
		"t2":   "2",
		"t3":   "3",
		"t4":   "4",
		"t5":   "5",
		"t6":   "6",
		"t7":   "7",
		"t8":   "8",
		"t9":   "9",
	}
}

// GetCategoryList Function that returns the correspondent offerType + category (URN)
func GetCategoryList() map[string]map[string]string {
	return map[string]map[string]string{
		"venda": {
			"apartamentos":              "urn:concept:apartments-for-sale",
			"moradias":                  "urn:concept:houses-for-sale",
			"terrenos":                  "urn:concept:lots-for-sale",
			"lojas":                     "urn:concept:stores-for-sale",
			"armazens":                  "urn:concept:warehouses-for-sale",
			"garagens_e_estacionamento": "urn:concept:garages-for-sale",
			"escritorios":               "urn:concept:offices-for-sale",
			"predios":                   "urn:concept:buildings-for-sale",
			"quintas_e_herdades":        "urn:concept:farms-and-estates-for-sale",
			"trespasse":                 "urn:concept:goodwill",
		},
		"arrendamento": {
			"apartamentos":              "urn:concept:apartments-for-rent",
			"moradias":                  "urn:concept:houses-for-rent",
			"quartos":                   "urn:concept:rooms-for-rent",
			"terrenos":                  "urn:concept:lots-for-rent",
			"lojas":                     "urn:concept:stores-for-rent",
			"armazens":                  "urn:concept:warehouses-for-rent",
			"garagens_e_estacionamento": "urn:concept:garages-for-rent",
			"escritorios":               "urn:concept:offices-for-rent",
			"predios":                   "urn:concept:buildings-for-rent",
			"quintas_e_herdades":        "urn:concept:farms-and-estates-for-rent",
			"apartamentos_para_ferias":  "urn:concept:apartments-for-vacation",
			"moradias_para_ferias":      "urn:concept:houses-for-vacation",
			"investimentos":             "urn:concept:investments",
		},
	}
}

// SanitizeString Function to sanitize a string
func SanitizeString(str string) string {
	// Convert to lowercase
	str = strings.ToLower(str)

	// Replace spaces with _
	str = strings.ReplaceAll(str, " ", "_")

	// Replace slashes with _
	str = strings.ReplaceAll(str, "/", "_")

	// Replace - with _
	str = strings.ReplaceAll(str, "-", "_")

	// Remove accents
	str = RemoveAccent(str)

	return str
}

// RemoveAccent Function to remove the accentuation from characters
func RemoveAccent(str string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, str)
	return result
}

func logToFile(message string, extra string) {
	file, err := os.OpenFile("notes_marta.log", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	content, err := os.ReadFile("notes_marta.log")
	if err != nil {
		log.Println(err)
		return
	}

	if strings.Contains(string(content), message) {
		return
	}

	_, err = file.Seek(0, 2)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetOutput(file)
	log.Println(extra + message)
}
