# jinjafier

## Introduction

Jinjafier is a small script to convert and existing java properties file to a jinja2 template with an included yml file with the current values of the property file.

## Rules for converting property files

|Property	| Note |
|---------|------|
|my.main-project.person.first-name  | Kebab case, which is recommended for use in .properties and YAML files.|
|my.main-project.person.firstName   | Standard camel case syntax.|
|my.main-project.person.first\_name  | Underscore notation, which is an alternative format for use in .properties and YAML files.|
|MY\_MAINPROJECT\_PERSON\_FIRSTNAME     | Upper case format, which is recommended when using system environment variables.|


