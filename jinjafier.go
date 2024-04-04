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

		if os.Args[1] == "-v" {
				fmt.Println("Jinjafier version:", version)
				os.Exit(0)
		}

    if len(os.Args) != 2 {
        fmt.Println("Usage: jinjafier <java_properties_file>")
        os.Exit(1)
    }

    filename := os.Args[1]
    file, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    jinjaTemplate := ""
    yamlMap := make(map[string]string)

    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "#") {
            jinjaTemplate += line + "\n"
        } else if strings.Contains(line, "=") {
            split := strings.SplitN(line, "=", 2) // Split at the first "=" only
            key := split[0]
            value := split[1]

            // Convert key to lowercase and replace dots and camel case with underscores
            re := regexp.MustCompile("([a-z0-9])([A-Z])")
            key = re.ReplaceAllString(key, "${1}_${2}")
            key = strings.ReplaceAll(key, ".", "_")
            key = strings.ReplaceAll(key, "-", "_")
            key = strings.ToUpper(key)

            // Add to jinja template
            jinjaTemplate += fmt.Sprintf("%s={{ %s }}\n", split[0], key)

            // Add to yaml map
            yamlMap[key] = value
        } else {
            jinjaTemplate += line + "\n"
        }
    }

    // Write jinja template to file
    err = ioutil.WriteFile(strings.ReplaceAll(filename, ".properties", ".properties.j2"), []byte(jinjaTemplate), 0644)
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

