package result

import (
	"encoding/xml"
)

type Testsuites struct {
	XMLName   xml.Name    `xml:"testsuites"`
	Name      string      `xml:"name,attr"`
	Testsuite []Testsuite `xml:"testsuite"`
}

type Testsuite struct {
	XMLName  xml.Name   `xml:"testsuite"`
	Name     string     `xml:"name,attr"`
	Package  string     `xml:"package,attr"`
	Property Property   `xml:"property"`
	Testcase []Testcase `xml:"testcase"`
}

type Property struct {
	XMLName xml.Name `xml:"property"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type Testcase struct {
	XMLName   xml.Name `xml:"testcase"`
	Name      string   `xml:"name,attr"`
	Classname string   `xml:"classname,attr"`
	File      string   `xml:"file,attr"`
	Lineno    int      `xml:"lineno,attr"`
	Time      float64  `xml:"time,attr"`
	SimTimeNs float64  `xml:"sim_time_ns,attr"`
	RatioTime float64  `xml:"ratio_time,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
}

type Failure struct {
	XMLName xml.Name `xml:"failure"`
	Message string   `xml:"message,attr"`
}
