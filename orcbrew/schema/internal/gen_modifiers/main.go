// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates modifiers.go. It can be invoked by running
// go generate
package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"time"
)

var outputFilename = flag.String("output", "", "The file to use for output")

func main() {
	flag.Parse()

	if *outputFilename == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	f, err := os.Create(*outputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s", err)
		os.Exit(2)
	}
	defer f.Close()

	type Config struct {
		Key       string
		TypeName  string
		ValueType string
	}

	modifiers := []Config{
		Config{"armor-prof", "ModifierArmorProficiency", "Armor"},
		Config{"damage-immunity", "ModifierDamageImmunity", "Damage"},
		Config{"damage-resistance", "ModifierDamageResistance", "Damage"},
		Config{"flying-speed", "ModifierFlyingSpeed", "int"},
		Config{"flying-speed-equals-walking-speed", "ModifierFlyingSpeedEqualsWalkingSpeed", "int"},
		Config{"num-attacks", "ModifierExtraAttacks", "int"},
		Config{"saving-throw-advantage", "ModifierSavingThrowAdvantage", "Condition"},
		Config{"skill-prof", "ModifierSkillProficiency", "Skill"},
		Config{"spell", "ModifierSpell", "SpellWithAbility"},
		Config{"swimming-speed", "ModifierSwimmingSpeed", "int"},
		Config{"tool-prof", "ModifierToolProficiency", "string"},
		Config{"weapon-prof", "ModifierWeaponProficiency", "string"},
	}
	packageTemplate.Execute(f, struct {
		Timestamp time.Time
		Config    []Config
	}{
		Timestamp: time.Now(),
		Config:    modifiers,
	})
}

var packageTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}

package schema

import (
	"encoding/json"
	"fmt"
)

type levelModifierType string

type LevelModifier interface{
	Type() levelModifierType
}

type LevelModifierList []LevelModifier

func (list *LevelModifierList) UnmarshalJSON(b []byte) error {
	var rawList []*json.RawMessage
	err := json.Unmarshal(b, &rawList)
	if err != nil {
		return err
	}

	if len(rawList) == 0 {
		*list = make([]LevelModifier, 0)
	}

	var m map[string]interface{}
	for _, rawMessage := range rawList {
		err = json.Unmarshal(*rawMessage, &m)
		if err != nil {
			return err
		}

		entryType, ok := m["type"].(string)
		if !ok {
			return fmt.Errorf("Value type in map was not a string")
		}

		var entry LevelModifier
		switch entryType {
{{ range .Config }}
		case "{{ .Key }}":
			entry = &{{ .TypeName }}{}
			err = json.Unmarshal(*rawMessage, &entry)
{{ end }}
		default:
			return fmt.Errorf("Got unknown type: %s", entryType)
		}

		if err != nil {
			return err
		}

		*list = append(*list, entry)
	}

	return nil
}

{{ range .Config }}
type {{ .TypeName }} struct {
	Level int
	Value {{ .ValueType }}
}

func (m *{{ .TypeName }}) Type() levelModifierType {
	return "{{ .Key }}"
}

func (m *{{ .TypeName }}) MarshalJSON() (b []byte, e error) {
	var valueMap = map[string]interface{}{
		"type": m.Type(),
		"value": m.Value,
	}

	if m.Level != 0 {
		valueMap["level"] = m.Level
	}
	return json.Marshal(valueMap)
}
{{ end }}


{{ range .Config }}
var _ LevelModifier = &{{ .TypeName }}{}
{{ end }}
`))
