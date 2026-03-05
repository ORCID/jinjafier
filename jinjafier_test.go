package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Build the binary once before all tests
	cmd := exec.Command("go", "build", "-o", "jinjafier_test_bin")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build: " + err.Error())
	}
	code := m.Run()
	os.Remove("jinjafier_test_bin")
	os.Exit(code)
}

func runJinjafier(t *testing.T, dir string, args ...string) {
	t.Helper()
	bin, _ := filepath.Abs("jinjafier_test_bin")
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jinjafier %v failed: %v\n%s", args, err, out)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func setupTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func assertContains(t *testing.T, file, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("%s: expected to contain %q", file, substr)
	}
}

func assertNotContains(t *testing.T, file, content, substr string) {
	t.Helper()
	if strings.Contains(content, substr) {
		t.Errorf("%s: expected NOT to contain %q", file, substr)
	}
}

const testProperties = `# a comment
org.wibble.testCamelCaps=20
org.wibble.test.long.property=hello world

# testing = in value
org.wibble.brokerURL=tcp://localhost:61616?jms.useAsyncSend=true

# testing - in key
org.wibble.message-listener.retryCount=3

# testing weird cron
org.wibble.cronFormat=*/5 * * * * *
`

const testYaml = `org:
  wibble:
    testCamelCaps: 20
    test:
      long:
        property: hello world
    brokerURL: "tcp://localhost:61616?jms.useAsyncSend=true"
    message-listener:
      retryCount: 3
    cronFormat: "*/5 * * * * *"
`

func TestPropertiesDefaultMode(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.properties"), testProperties)

	runJinjafier(t, dir, "app.properties")

	// Check .j2 template
	j2 := readFile(t, filepath.Join(dir, "app.properties.j2"))
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_TESTCAMELCAPS }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_BROKERURL }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_MESSAGE_LISTENER_RETRYCOUNT }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_CRONFORMAT }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_TEST_LONG_PROPERTY }}")

	// camelCase should NOT be split
	assertNotContains(t, "j2", j2, "TEST_CAMEL_CAPS")
	assertNotContains(t, "j2", j2, "BROKER_URL")
	assertNotContains(t, "j2", j2, "CRON_FORMAT")
	assertNotContains(t, "j2", j2, "RETRY_COUNT")

	// Check .env.j2 template
	env := readFile(t, filepath.Join(dir, "app.properties.env.j2"))
	assertContains(t, "env.j2", env, "ORG_WIBBLE_TESTCAMELCAPS={{ ORG_WIBBLE_TESTCAMELCAPS }}")
	assertContains(t, "env.j2", env, "ORG_WIBBLE_BROKERURL={{ ORG_WIBBLE_BROKERURL }}")

	// Check .yml output
	yml := readFile(t, filepath.Join(dir, "app.properties.yml"))
	assertContains(t, "yml", yml, "ORG_WIBBLE_TESTCAMELCAPS:")
	assertContains(t, "yml", yml, "ORG_WIBBLE_BROKERURL:")
	assertNotContains(t, "yml", yml, "TEST_CAMEL_CAPS:")
}

func TestPropertiesCamelSplitMode(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.properties"), testProperties)

	runJinjafier(t, dir, "-camel-split", "app.properties")

	j2 := readFile(t, filepath.Join(dir, "app.properties.j2"))
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_TEST_CAMEL_CAPS }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_BROKER_URL }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_MESSAGE_LISTENER_RETRY_COUNT }}")
	assertContains(t, "j2", j2, "{{ ORG_WIBBLE_CRON_FORMAT }}")

	// Should NOT contain the unsplit versions
	assertNotContains(t, "j2", j2, "TESTCAMELCAPS")
	assertNotContains(t, "j2", j2, "BROKERURL")
	assertNotContains(t, "j2", j2, "CRONFORMAT")
}

func TestYamlDefaultMode(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.yml"), testYaml)

	runJinjafier(t, dir, "app.yml")

	env := readFile(t, filepath.Join(dir, "app.yml.env"))
	assertContains(t, "yml.env", env, "ORG_WIBBLE_TESTCAMELCAPS=20")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_BROKERURL=")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_MESSAGE_LISTENER_RETRYCOUNT=3")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_CRONFORMAT=")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_TEST_LONG_PROPERTY=hello world")

	assertNotContains(t, "yml.env", env, "TEST_CAMEL_CAPS")
	assertNotContains(t, "yml.env", env, "BROKER_URL")
}

func TestYamlCamelSplitMode(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.yml"), testYaml)

	runJinjafier(t, dir, "-camel-split", "app.yml")

	env := readFile(t, filepath.Join(dir, "app.yml.env"))
	assertContains(t, "yml.env", env, "ORG_WIBBLE_TEST_CAMEL_CAPS=20")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_BROKER_URL=")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_MESSAGE_LISTENER_RETRY_COUNT=3")
	assertContains(t, "yml.env", env, "ORG_WIBBLE_CRON_FORMAT=")

	assertNotContains(t, "yml.env", env, "TESTCAMELCAPS")
	assertNotContains(t, "yml.env", env, "BROKERURL")
}

func TestPropertiesPreservesComments(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.properties"), testProperties)

	runJinjafier(t, dir, "app.properties")

	j2 := readFile(t, filepath.Join(dir, "app.properties.j2"))
	assertContains(t, "j2", j2, "# a comment")
	assertContains(t, "j2", j2, "# testing = in value")
	assertContains(t, "j2", j2, "# testing - in key")
	assertContains(t, "j2", j2, "# testing weird cron")

	env := readFile(t, filepath.Join(dir, "app.properties.env.j2"))
	assertContains(t, "env.j2", env, "# a comment")
}

func TestPropertiesPreservesBlankLines(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.properties"), testProperties)

	runJinjafier(t, dir, "app.properties")

	j2 := readFile(t, filepath.Join(dir, "app.properties.j2"))
	assertContains(t, "j2", j2, "\n\n")
}

func TestPropertiesOriginalKeysPreservedInJ2(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.properties"), testProperties)

	runJinjafier(t, dir, "app.properties")

	j2 := readFile(t, filepath.Join(dir, "app.properties.j2"))
	assertContains(t, "j2", j2, "org.wibble.testCamelCaps=")
	assertContains(t, "j2", j2, "org.wibble.brokerURL=")
	assertContains(t, "j2", j2, "org.wibble.message-listener.retryCount=")
}

func TestPropertiesEqualsInValue(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, filepath.Join(dir, "app.properties"), testProperties)

	runJinjafier(t, dir, "app.properties")

	yml := readFile(t, filepath.Join(dir, "app.properties.yml"))
	assertContains(t, "yml", yml, "tcp://localhost:61616?jms.useAsyncSend=true")
}

func TestVersionFlag(t *testing.T) {
	bin, _ := filepath.Abs("jinjafier_test_bin")
	out, err := exec.Command(bin, "-v").CombinedOutput()
	if err != nil {
		t.Fatalf("-v failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Jinjafier version:") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestUnsupportedFileFormat(t *testing.T) {
	bin, _ := filepath.Abs("jinjafier_test_bin")
	cmd := exec.Command(bin, "file.txt")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for unsupported file format")
	}
	if !strings.Contains(string(out), "Unsupported file format") {
		t.Errorf("expected unsupported format message, got: %s", out)
	}
}

func TestNoArgs(t *testing.T) {
	bin, _ := filepath.Abs("jinjafier_test_bin")
	cmd := exec.Command(bin)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error with no args")
	}
	if !strings.Contains(string(out), "Jinjafier - Convert Java properties/YAML") {
		t.Errorf("expected usage message, got: %s", out)
	}
}

func TestHyphenInKeyConvertedToUnderscore(t *testing.T) {
	dir := setupTempDir(t)
	props := "my.app.some-key=value\n"
	writeFile(t, filepath.Join(dir, "app.properties"), props)

	runJinjafier(t, dir, "app.properties")

	j2 := readFile(t, filepath.Join(dir, "app.properties.j2"))
	assertContains(t, "j2", j2, "{{ MY_APP_SOME_KEY }}")
}
