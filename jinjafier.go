package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("v", false, "Print version")
	camelSplit := flag.Bool("camel-split", false, "Split camelCase into separate words (e.g. brokerURL -> BROKER_URL)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Jinjafier - Convert Java properties/YAML files to Jinja2 templates

Converts property names to environment variable format suitable for Spring Boot:
  - Replace '.' with '_'
  - Replace '-' with '_'
  - Convert to UPPERCASE
  - By default, camelCase is NOT split (e.g. firstName -> FIRSTNAME)

For .properties files, generates:
  .properties.j2     - Jinja2 template with original keys and env var placeholders
  .properties.env.j2 - Env var template (KEY={{ KEY }})
  .properties.yml    - YAML file mapping env var names to original values

For .yml files, generates:
  .yml.env            - Flat env var file (KEY=value)

Spring Boot binding:
  @Value requires exact env var match (no camelCase splitting) - use default mode.
  @ConfigurationProperties supports relaxed binding - both modes work.

Usage:
  jinjafier [flags] <file.properties|file.yml>

Examples:
  jinjafier app.properties                  # person.firstName -> PERSON_FIRSTNAME
  jinjafier -camel-split app.properties     # person.firstName -> PERSON_FIRST_NAME
  jinjafier config.yml                      # flatten YAML to env vars

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *versionFlag {
		fmt.Println("Jinjafier version:", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	filename := args[0]
	if strings.HasSuffix(filename, ".properties") {
		processPropertiesFile(filename, *camelSplit)
	} else if strings.HasSuffix(filename, ".yml") {
		processYamlFile(filename, *camelSplit)
	} else {
		fmt.Println("Unsupported file format. Please provide a .properties or .yml file.")
		os.Exit(1)
	}
}

func convertKey(key string, camelSplit bool) string {
	if camelSplit {
		re := regexp.MustCompile("([a-z0-9])([A-Z])")
		key = re.ReplaceAllString(key, "${1}_${2}")
	}
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ToUpper(key)
	return key
}

func processPropertiesFile(filename string, camelSplit bool) {
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

			key = convertKey(key, camelSplit)

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

func processYamlFile(filename string, camelSplit bool) {
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
	flattenYaml("", yamlData, &envTemplate, camelSplit)

	// Write env template to file
	err = ioutil.WriteFile(strings.ReplaceAll(filename, ".yml", ".yml.env"), []byte(envTemplate), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func flattenYaml(prefix string, data map[string]interface{}, envTemplate *string, camelSplit bool) {
	for key, value := range data {
		upperKey := convertKey(prefix+key, camelSplit)

		switch v := value.(type) {
		case map[string]interface{}:
			flattenYaml(upperKey+"_", v, envTemplate, camelSplit)
		default:
			*envTemplate += fmt.Sprintf("%s=%v\n", upperKey, v)
		}
	}
}
