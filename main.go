package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Traits struct {
	Key         string           `json:"key"`
	Name        string           `json:"name"`
	Requirement string           `json:"requirement"`
	Options     []*FeatureOption `json:"options"`
}

type FeatureOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Image string `json:"image"`
}

func main() {
	a := "1"
	println("xxx", a)
}
func main1() {

	f, _ := os.Open("/Users/hxy/develops/角色.ini")
	scanner := bufio.NewScanner(f)

	var features = make([]*Traits, 0)

	var feature *Traits
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if line == "" {
			continue
		}

		//feature
		if strings.HasPrefix(line, "@") {
			featureTexts := strings.Split(line, "@")
			feature = &Traits{
				Key:         strings.Trim(featureTexts[2], " "),
				Name:        strings.Trim(featureTexts[1], " "),
				Requirement: "（单选项，如无特别要求，可不选）",
				Options:     make([]*FeatureOption, 0),
			}
			features = append(features, feature)
			continue
		}

		//feature enum
		log.Println(line)
		kv := strings.Split(line, "#")

		if len(kv) < 2 {
			continue
		}

		feature.Options = append(feature.Options, &FeatureOption{
			Key:   fmt.Sprintf("%s@%d", feature.Key, len(feature.Options)+1),
			Label: strings.Trim(kv[0], "\t"),
			Value: strings.Trim(kv[1], "\t"),
		})
	}

	log.Println(len(features))
	jsonBytes, _ := json.MarshalIndent(features, "", "\t")
	_ = ioutil.WriteFile("/Users/hxy/develops/角色.json", jsonBytes, fs.ModePerm)
}
