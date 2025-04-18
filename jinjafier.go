package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-v" {
		fmt.Println("Jinjafier version:", version)
		os.Exit(0)
	}

	if len(os.Args) != 2 {
		fmt.Println("Usage: jinjafier <file.properties|file.yml>")
		os.Exit(1)
	}

	filename := os.Args[1]
	if strings.HasSuffix(filename, ".properties") {
		processPropertiesFile(filename)
	} else if strings.HasSuffix(filename, ".yml") {
		processYamlFile(filename)
	} else {
		fmt.Println("Unsupported file format. Please provide a .properties or .yml file.")
		os.Exit(1)
	}
}

func processPropertiesFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	jinjaTemplate := ""
	envTemplate := ""
	yamlMap := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			// Add comments to both templates
			jinjaTemplate += line + "\n"
			envTemplate += line + "\n"
		} else if line == "" {
			// Preserve blank lines in both templates
			jinjaTemplate += "\n"
			envTemplate += "\n"
		} else if strings.Contains(line, "=") {
			split := strings.SplitN(line, "=", 2) // Split at the first "=" only
			key := split[0]
			value := split[1]

			// Convert key to uppercase with underscores
			re := regexp.MustCompile("([a-z0-9])([A-Z])")
			key = re.ReplaceAllString(key, "${1}_${2}")
			key = strings.ReplaceAll(key, ".", "_")
			key = strings.ReplaceAll(key, "-", "_")
			key = strings.ToUpper(key)

			// Add to jinja template
			jinjaTemplate += fmt.Sprintf("%s={{ %s }}\n", split[0], key)

			// Add to env template
			envTemplate += fmt.Sprintf("%s={{ %s }}\n", key, key)

			// Add to yaml map
			yamlMap[key] = value
		} else {
			// Add non-comment, non-key-value lines to the Jinja template
			jinjaTemplate += line + "\n"
		}
	}

	// Write jinja template to file
	err = ioutil.WriteFile(strings.ReplaceAll(filename, ".properties", ".properties.j2"), []byte(jinjaTemplate), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Write env template to file
	err = ioutil.WriteFile(strings.ReplaceAll(filename, ".properties", ".properties.env.j2"), []byte(envTemplate), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Convert yaml map to yaml
	yamlData, err := yaml.Marshal(&yamlMap)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Write yaml to file
	err = ioutil.WriteFile(strings.ReplaceAll(filename, ".properties", ".properties.yml"), yamlData, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func processYamlFile(filename string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var yamlData map[string]interface{}
	err = yaml.Unmarshal(file, &yamlData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	envTemplate := ""
	flattenYaml("", yamlData, &envTemplate)

	// Write env template to file
	err = ioutil.WriteFile(strings.ReplaceAll(filename, ".yml", ".yml.env"), []byte(envTemplate), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func flattenYaml(prefix string, data map[string]interface{}, envTemplate *string) {
	for key, value := range data {
		// Convert key to uppercase with underscores
		re := regexp.MustCompile("([a-z0-9])([A-Z])")
		upperKey := re.ReplaceAllString(key, "${1}_${2}") // Add underscores before uppercase letters
		upperKey = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(prefix+upperKey, ".", "_"), "-", "_"))

		switch v := value.(type) {
		case map[string]interface{}:
			flattenYaml(upperKey+"_", v, envTemplate)
		default:
			*envTemplate += fmt.Sprintf("%s=%v\n", upperKey, v)
		}
	}
}
