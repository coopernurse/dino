package dino

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func NewNameDotComProvider(username string, token string) *NameDotComProvider {
	return &NameDotComProvider{
		Username: username,
		Token:    token,
		Endpoint: "https://api.name.com/v4",
	}
}

type NameDotComProvider struct {
	Username string
	Token    string
	Endpoint string
}

func (p *NameDotComProvider) Put(rec Record) error {
	if rec.Id == "" {
		records, err := p.List(rec.Domain)
		if err != nil {
			return err
		}
		for _, r := range records {
			if r.Host == rec.Host && r.Type == rec.Type {
				rec.Id = r.Id
				break
			}
		}
	}

	if rec.Id == "" {
		return p.Create(rec)
	} else {
		return p.Update(rec)
	}
}

func (p *NameDotComProvider) Create(rec Record) error {
	url := fmt.Sprintf("%s/domains/%s/records", p.Endpoint, rec.Domain)
	return p.createUpdate("POST", url, rec)
}

func (p *NameDotComProvider) Update(rec Record) error {
	url := fmt.Sprintf("%s/domains/%s/records/%s", p.Endpoint, rec.Domain, rec.Id)
	return p.createUpdate("PUT", url, rec)
}

func (p *NameDotComProvider) Delete(domainName string, id string) error {
	url := fmt.Sprintf("%s/domains/%s/records/%s", p.Endpoint, domainName, id)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.Username, p.Token)
	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("namedotcom: invalid status code method=%s url=%s code=%d", "DELETE", url, resp.StatusCode)
	}
	return err
}

func (p *NameDotComProvider) createUpdate(method string, url string, rec Record) error {
	dotRec, err := fromRecord(rec)
	if err != nil {
		return err
	}
	var out NameDotComRecord
	_, err = p.reqJSON(url, method, dotRec, &out)
	return err
}

func (p *NameDotComProvider) List(domainName string) ([]Record, error) {
	url := fmt.Sprintf("%s/domains/%s/records", p.Endpoint, domainName)
	var out NameDotComListRecordsResponse
	_, err := p.getJSON(url, &out)
	if err != nil {
		return nil, err
	}
	return out.ToRecords(), nil
}

func (p *NameDotComProvider) getJSON(url string, output interface{}) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.Username, p.Token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("namedotcom: invalid status code method=%s url=%s code=%d", "GET", url, resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(output)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *NameDotComProvider) reqJSON(url string, method string, input interface{},
	output interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.Username, p.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respData, _ := ioutil.ReadAll(resp.Body)

		return nil, fmt.Errorf("namedotcom: invalid status code method=%s url=%s code=%d req=%s resp=%s", method, url,
			resp.StatusCode, string(jsonData), string(respData))
	}

	err = json.NewDecoder(resp.Body).Decode(output)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type NameDotComRecord struct {
	Id         int        `json:"id,omitempty"`
	DomainName string     `json:"domainName,omitempty"`
	Host       string     `json:"host"`
	Fqdn       string     `json:"fqdn,omitempty"`
	Type       RecordType `json:"type,omitempty"`
	Answer     string     `json:"answer,omitempty"`
	Ttl        uint32     `json:"ttl,omitempty"`
	Priority   uint32     `json:"priority,omitempty"`
}

func fromRecord(r Record) (NameDotComRecord, error) {
	rec := NameDotComRecord{
		DomainName: r.Domain,
		Host:       r.Host,
		Type:       r.Type,
		Answer:     r.Answer,
		Ttl:        r.Ttl,
		Priority:   r.Priority,
	}
	if r.Id != "" {
		id, err := strconv.Atoi(r.Id)
		if err != nil {
			return NameDotComRecord{}, err
		}
		rec.Id = id
	}
	return rec, nil
}

func (r NameDotComRecord) ToRecord() Record {
	rec := Record{
		Domain:   r.DomainName,
		Host:     r.Host,
		Type:     r.Type,
		Answer:   r.Answer,
		Ttl:      r.Ttl,
		Priority: r.Priority,
	}
	if r.Id != 0 {
		rec.Id = strconv.Itoa(r.Id)
	}
	return rec
}

type NameDotComListRecordsResponse struct {
	Records []NameDotComRecord
}

func (r NameDotComListRecordsResponse) ToRecords() []Record {
	out := make([]Record, len(r.Records))
	for x, rec := range r.Records {
		out[x] = rec.ToRecord()
	}
	return out
}
