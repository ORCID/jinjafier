# jinjafier

## Introduction

Jinjafier is a small script to convert and existing java properties file to a jinja2 template with an included yml file using the current values of the property file.

## Rules for converting property files

The standards listed here are converted into jinja2 variables that are suitable to use as system environment variables (uppercase with _ as a delimiter.)

|Property	| Note |
|---------|------|
|my.main-project.person.first-name  | Kebab case, which is recommended for use in .properties and YAML files.|
|my.main-project.person.firstName   | Standard camel case syntax.|
|my.main-project.person.first\_name  | Underscore notation, which is an alternative format for use in .properties and YAML files.|
|MY\_MAINPROJECT\_PERSON\_FIRSTNAME     | Upper case format, which is recommended when using system environment variables.|

## Spring Boot binding behavior

Spring Boot resolves environment variables by replacing `.` with `_`, removing `-`, and converting to uppercase. It does **not** split camelCase words with underscores.

### `@Value` annotations

`@Value` does **not** support relaxed binding. The env var must match the exact property name after applying these rules:

- Replace `.` with `_`
- Remove `-`
- Convert to uppercase
- **No camelCase splitting**

Example: `@Value("${my.main-project.person.firstName}")` resolves to `MY_MAINPROJECT_PERSON_FIRSTNAME`

### `@ConfigurationProperties`

`@ConfigurationProperties` supports relaxed binding. Spring normalizes by stripping all delimiters and comparing case-insensitively, so both `FIRSTNAME` and `FIRST_NAME` will work.

### The `-camel-split` flag

By default, jinjafier follows Spring Boot's env var resolution rules (no camelCase splitting). If your application uses `@ConfigurationProperties` exclusively and you prefer underscores between camelCase words, use the `-camel-split` flag:

```
# Default (Spring Boot compatible)
./jinjafier example.properties
# person.firstName -> PERSON_FIRSTNAME

# With camelCase splitting
./jinjafier -camel-split example.properties
# person.firstName -> PERSON_FIRST_NAME
```

# Installing with asdf
```
# RUN asdf plugin add jinjafier https://github.com/ORCID/asdf-jinjafier.git
jinjafier latest
